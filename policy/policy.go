package policy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/go-rod/rod/lib/proto"
	"github.com/spf13/viper"
)

// Policy defines network request rules loaded from a YAML file.
type Policy struct {
	Version   string `mapstructure:"version"`
	DenyList  []Rule `mapstructure:"deny"`
	AllowList []Rule `mapstructure:"allow"`
}

// Rule matches network requests by method, URL pattern, and body regex.
type Rule struct {
	Method string `mapstructure:"method"` // regex pattern, case-insensitive
	URL    string `mapstructure:"url"`    // CDP wildcard pattern (* and ?)
	Body   string `mapstructure:"body"`   // regex pattern

	// Compiled patterns, set by Load.
	methodRegex *regexp.Regexp
	urlRegex    *regexp.Regexp
	bodyRegex   *regexp.Regexp
}

// String returns a human-readable description of the rule.
func (r Rule) String() string {
	var parts []string
	if r.Method != "" {
		parts = append(parts, "method="+r.Method)
	}
	if r.URL != "" {
		parts = append(parts, "url="+r.URL)
	}
	if r.Body != "" {
		parts = append(parts, "body="+r.Body)
	}
	return strings.Join(parts, " ")
}

// Load reads and validates a policy YAML file.
// Returns an error if the file cannot be read, parsed, or contains invalid patterns.
func Load(path string) (*Policy, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading policy file: %w", err)
	}

	var p Policy
	if err := v.Unmarshal(&p); err != nil {
		return nil, fmt.Errorf("parsing policy file: %w", err)
	}

	if len(p.DenyList) == 0 {
		return nil, fmt.Errorf("policy must have at least one deny rule")
	}

	if err := compileRules(p.DenyList); err != nil {
		return nil, fmt.Errorf("deny rule: %w", err)
	}
	if err := compileRules(p.AllowList); err != nil {
		return nil, fmt.Errorf("allow rule: %w", err)
	}

	return &p, nil
}

// compileRules compiles URL and body patterns for a slice of rules.
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
		if r.URL != "" {
			if err := validateURLPattern(r.URL); err != nil {
				return err
			}
			pattern := proto.PatternToReg(r.URL)
			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("invalid url pattern %q: %w", r.URL, err)
			}
			r.urlRegex = re
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

// validateURLPattern checks for regex metacharacters that indicate the user
// is trying to use regex instead of CDP wildcard syntax.
func validateURLPattern(url string) error {
	if strings.ContainsAny(url, `|\^$`) {
		return fmt.Errorf("invalid url in policy: %q. Regular expressions are not supported in url policy rules. Only wildcard characters * and ? are supported", url)
	}
	return nil
}

// IsAllowed checks if the request is allowed by the policy.
// Deny rules are checked first. If a deny matches, allow rules are checked
// as exceptions. If no allow exception is found, the request is denied.
// Returns (false, matched deny rule, nil) if denied.
// Returns (true, nil, nil) if allowed (no deny match).
// Returns (true, matched allow rule, nil) if allowed by exception.
func (p *Policy) IsAllowed(r *http.Request) (bool, *Rule, error) {
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
func (p *Policy) needsBody() bool {
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
// body is the pre-read request body (empty if no rules need body inspection).
func (p *Policy) matchRules(rules []Rule, r *http.Request, body string) (bool, *Rule) {
	for i := range rules {
		rule := &rules[i]

		if rule.methodRegex != nil && !rule.methodRegex.MatchString(r.Method) {
			continue
		}
		if rule.urlRegex != nil && !rule.urlRegex.MatchString(r.URL.String()) {
			continue
		}
		if rule.bodyRegex != nil && !rule.bodyRegex.MatchString(body) {
			continue
		}

		return true, rule
	}

	return false, nil
}

// DenyPatterns returns deduplicated URL patterns from deny rules for CDP registration.
// If any deny rule has no URL pattern, returns ["*"] (catch-all).
func (p *Policy) DenyPatterns() []string {
	seen := make(map[string]bool)
	for _, r := range p.DenyList {
		if r.URL == "" {
			return []string{"*"}
		}
		seen[r.URL] = true
	}
	patterns := make([]string, 0, len(seen))
	for u := range seen {
		patterns = append(patterns, u)
	}
	sort.Strings(patterns)
	return patterns
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
