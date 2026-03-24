package policy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/go-rod/rod/lib/proto"
	"github.com/spf13/viper"
)

// Policy is the top-level structure loaded from a YAML policy file.
type Policy struct {
	Version string        `mapstructure:"version"`
	Network NetworkPolicy `mapstructure:"network"`
	Actions ActionsPolicy `mapstructure:"actions"`
}

// ActionsPolicy defines which CLI actions are allowed or denied.
// Only one of DenyList or AllowList may be specified.
type ActionsPolicy struct {
	DenyList  []string `mapstructure:"deny"`
	AllowList []string `mapstructure:"allow"`
}

// knownActions is the set of valid action names for validation.
// Excludes "health" and "stop" which always bypass the action policy.
var knownActions = map[string]bool{
	"open": true, "snapshot": true, "click": true, "type": true,
	"key": true, "back": true, "forward": true, "reload": true,
	"eval": true, "select": true, "extract": true, "tab": true,
	"wait": true, "upload": true, "hover": true, "tabs": true,
}

// IsActionAllowed checks if the given action is allowed by the policy.
// Returns true if no policy or no action rules are configured.
func (p *Policy) IsActionAllowed(action string) bool {
	// No policy loaded — everything is allowed.
	if p == nil {
		return true
	}
	a := p.Actions
	// No action rules configured — everything is allowed.
	if len(a.DenyList) == 0 && len(a.AllowList) == 0 {
		return true
	}
	// Deny list: block listed actions, allow everything else.
	if len(a.DenyList) > 0 {
		return !slices.Contains(a.DenyList, action)
	}
	// Allow list: permit only listed actions, block everything else.
	return slices.Contains(a.AllowList, action)
}

// NetworkPolicy defines network request interception rules.
type NetworkPolicy struct {
	DenyList  []Rule `mapstructure:"deny"`
	AllowList []Rule `mapstructure:"allow"`
}

// IsEnabled returns true if any network rules are configured.
func (p NetworkPolicy) IsEnabled() bool {
	return len(p.DenyList) > 0 || len(p.AllowList) > 0
}

// IsEnabled returns true if any action rules are configured.
func (a ActionsPolicy) IsEnabled() bool {
	return len(a.DenyList) > 0 || len(a.AllowList) > 0
}

// Rule matches network requests by method, URL component patterns, and body regex.
type Rule struct {
	Method string `mapstructure:"method"` // regex pattern, case-insensitive
	Host   string `mapstructure:"host"`   // CDP wildcard pattern matched against parsed URL host
	Path   string `mapstructure:"path"`   // regex pattern matched against parsed URL path
	Query  string `mapstructure:"query"`  // regex pattern matched against parsed URL query
	Header string `mapstructure:"header"` // regex pattern matched against serialized request headers
	Body   string `mapstructure:"body"`   // regex pattern

	// Compiled patterns, set by Load.
	methodRegex *regexp.Regexp
	hostRegex   *regexp.Regexp
	pathRegex   *regexp.Regexp
	queryRegex  *regexp.Regexp
	headerRegex *regexp.Regexp
	bodyRegex   *regexp.Regexp
}

// String returns a human-readable description of the rule.
func (r Rule) String() string {
	var parts []string
	if r.Method != "" {
		parts = append(parts, "method="+r.Method)
	}
	if r.Host != "" {
		parts = append(parts, "host="+r.Host)
	}
	if r.Path != "" {
		parts = append(parts, "path="+r.Path)
	}
	if r.Query != "" {
		parts = append(parts, "query="+r.Query)
	}
	if r.Header != "" {
		parts = append(parts, "header="+r.Header)
	}
	if r.Body != "" {
		parts = append(parts, "body="+r.Body)
	}
	if len(parts) == 0 {
		return "deny all"
	}
	return strings.Join(parts, " ")
}

// Load reads and validates a policy from a file path or URL.
// If source starts with http:// or https://, it is fetched via HTTP.
// Returns an error if the source cannot be read, parsed, or contains invalid patterns.
func Load(ctx context.Context, source string) (*Policy, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		body, err := fetchURL(ctx, source)
		if err != nil {
			return nil, err
		}
		defer body.Close()
		if err := v.ReadConfig(body); err != nil {
			return nil, fmt.Errorf("reading policy: %w", err)
		}
	} else {
		v.SetConfigFile(source)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("reading policy file: %w", err)
		}
	}

	var p Policy
	if err := v.Unmarshal(&p); err != nil {
		return nil, fmt.Errorf("parsing policy: %w", err)
	}

	// If allow rules exist but no deny rules, insert an implicit deny-all.
	// An empty Rule{} matches everything (all nil regexes = no checks fail).
	if len(p.Network.DenyList) == 0 && len(p.Network.AllowList) > 0 {
		p.Network.DenyList = []Rule{{}}
	}

	if err := compileRules(p.Network.DenyList); err != nil {
		return nil, fmt.Errorf("deny rule: %w", err)
	}
	if err := compileRules(p.Network.AllowList); err != nil {
		return nil, fmt.Errorf("allow rule: %w", err)
	}

	// Lowercase action names before validation.
	for i, name := range p.Actions.DenyList {
		p.Actions.DenyList[i] = strings.ToLower(name)
	}
	for i, name := range p.Actions.AllowList {
		p.Actions.AllowList[i] = strings.ToLower(name)
	}

	if err := validateActions(p.Actions); err != nil {
		return nil, err
	}

	return &p, nil
}

// validateActions checks that the actions policy is valid.
// Returns an error if both deny and allow lists are specified,
// or if any action name is not recognized.
func validateActions(a ActionsPolicy) error {
	if len(a.DenyList) > 0 && len(a.AllowList) > 0 {
		return fmt.Errorf("actions policy: specify either deny or allow, not both")
	}
	for _, name := range a.DenyList {
		if !knownActions[name] {
			return fmt.Errorf("unknown action %q in actions deny list", name)
		}
	}
	for _, name := range a.AllowList {
		if !knownActions[name] {
			return fmt.Errorf("unknown action %q in actions allow list", name)
		}
	}
	return nil
}

// compileRules compiles host, path, query, and body patterns for a slice of rules.
func compileRules(rules []Rule) error {
	for i := range rules {
		r := &rules[i]
		if r.Method != "" {
			re, err := regexp.Compile("(?i)" + r.Method)
			if err != nil {
				return fmt.Errorf("invalid method regex %q: %w", r.Method, err)
			}
			r.methodRegex = re
		}
		if r.Host != "" {
			if err := validateHostPattern(r.Host); err != nil {
				return err
			}
			re, err := regexp.Compile(proto.PatternToReg(r.Host))
			if err != nil {
				return fmt.Errorf("invalid host pattern %q: %w", r.Host, err)
			}
			r.hostRegex = re
		}
		if r.Path != "" {
			re, err := regexp.Compile(r.Path)
			if err != nil {
				return fmt.Errorf("invalid path regex %q: %w", r.Path, err)
			}
			r.pathRegex = re
		}
		if r.Query != "" {
			re, err := regexp.Compile(r.Query)
			if err != nil {
				return fmt.Errorf("invalid query regex %q: %w", r.Query, err)
			}
			r.queryRegex = re
		}
		if r.Header != "" {
			re, err := regexp.Compile(r.Header)
			if err != nil {
				return fmt.Errorf("invalid header regex %q: %w", r.Header, err)
			}
			r.headerRegex = re
		}
		if r.Body != "" {
			re, err := regexp.Compile(r.Body)
			if err != nil {
				return fmt.Errorf("invalid body regex %q: %w", r.Body, err)
			}
			r.bodyRegex = re
		}
	}
	return nil
}

// validateHostPattern checks for regex metacharacters in host patterns.
// Only applied to host — path and query can legitimately contain these characters.
func validateHostPattern(host string) error {
	if strings.ContainsAny(host, `|\^$`) {
		return fmt.Errorf("invalid host pattern in policy: %q. Regular expressions are not supported. Only wildcard characters * and ? are supported", host)
	}
	return nil
}

// checkRequest validates that the request has the fields needed for policy matching.
func checkRequest(r *http.Request) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	if r.URL == nil {
		return fmt.Errorf("request has nil URL")
	}
	return nil
}

// IsAllowed checks if the request is allowed by the network policy.
// Deny rules are checked first. If a deny matches, allow rules are checked
// as exceptions. If no allow exception is found, the request is denied.
// Returns (false, matched deny rule, nil) if denied.
// Returns (true, nil, nil) if allowed (no deny match).
// Returns (true, matched allow rule, nil) if allowed by exception.
func (p *NetworkPolicy) IsAllowed(r *http.Request) (bool, *Rule, error) {
	if err := checkRequest(r); err != nil {
		return false, nil, err
	}

	// Read body once upfront if any rule needs it.
	var body string
	if p.needsBody() {
		var err error
		body, err = readBody(r)
		if err != nil {
			return false, nil, err
		}
	}

	denied, denyRule := p.matchRules(p.DenyList, r, body)
	if !denied {
		return true, nil, nil
	}

	excepted, allowRule := p.matchRules(p.AllowList, r, body)
	if excepted {
		return true, allowRule, nil
	}

	return false, denyRule, nil
}

// needsBody returns true if any rule has a body pattern.
func (p *NetworkPolicy) needsBody() bool {
	for i := range p.DenyList {
		if p.DenyList[i].bodyRegex != nil {
			return true
		}
	}
	for i := range p.AllowList {
		if p.AllowList[i].bodyRegex != nil {
			return true
		}
	}
	return false
}

// matchRules checks if any rule in the list matches the request.
// URL components are extracted from the parsed request URL.
// Header serialization is deferred until a rule actually needs it.
func (p *NetworkPolicy) matchRules(rules []Rule, r *http.Request, body string) (bool, *Rule) {
	// Serialize headers lazily — only if a rule has a header regex.
	var header string
	var headerReady bool

	for i := range rules {
		rule := &rules[i]

		if rule.methodRegex != nil && !rule.methodRegex.MatchString(r.Method) {
			continue
		}
		if rule.hostRegex != nil && !rule.hostRegex.MatchString(r.URL.Host) {
			continue
		}
		if rule.pathRegex != nil && !rule.pathRegex.MatchString(r.URL.Path) {
			continue
		}
		if rule.queryRegex != nil && !rule.queryRegex.MatchString(r.URL.RawQuery) {
			continue
		}
		if rule.headerRegex != nil {
			if !headerReady {
				header = serializeHeader(r.Header)
				headerReady = true
			}
			if !rule.headerRegex.MatchString(header) {
				continue
			}
		}
		if rule.bodyRegex != nil && !rule.bodyRegex.MatchString(body) {
			continue
		}

		return true, rule
	}

	return false, nil
}

// DenyPatterns returns deduplicated CDP URL patterns from deny rules for registration.
// Constructs coarse patterns from host/path fields. If any rule has no host, returns ["*"].
func (p *NetworkPolicy) DenyPatterns() []string {
	seen := make(map[string]bool)
	for _, r := range p.DenyList {
		pattern := cdpPattern(&r)
		if pattern == "*" {
			return []string{"*"}
		}
		seen[pattern] = true
	}
	patterns := make([]string, 0, len(seen))
	for u := range seen {
		patterns = append(patterns, u)
	}
	sort.Strings(patterns)
	return patterns
}

// cdpPattern constructs a coarse CDP Fetch URL pattern from a rule's host and path.
func cdpPattern(r *Rule) string {
	if r.Host == "" {
		return "*"
	}
	if r.Path != "" {
		return "*://" + r.Host + r.Path + "*"
	}
	return "*://" + r.Host + "*"
}

// serializeHeader formats request headers in wire format ("Name: Value\r\n").
// Uses http.Header.Write which outputs canonical header names.
func serializeHeader(h http.Header) string {
	var b strings.Builder
	h.Write(&b)
	return b.String()
}

// fetchURL fetches a policy from a URL and returns the response body.
func fetchURL(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching policy: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching policy: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fetching policy: %s", resp.Status)
	}
	return resp.Body, nil
}

// readBody reads the request body and returns it as a string.
// The body is reset so the request can still be forwarded.
// Returns empty string if the body is nil.
func readBody(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", nil
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("reading request body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(data))
	return string(data), nil
}
