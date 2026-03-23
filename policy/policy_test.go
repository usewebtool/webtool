package policy

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsAllowed_AllowExceptionOverridesDeny(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Method: "POST", Host: "*api.example.com"},
		},
		AllowList: []Rule{
			{Method: "POST", Host: "*api.example.com", Path: "/login"},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}
	if err := compileRules(np.AllowList); err != nil {
		t.Fatal(err)
	}

	// Login endpoint matches allow exception — allowed.
	req, _ := http.NewRequest("POST", "https://api.example.com/login", nil)
	allowed, rule, err := np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed (allow exception), got denied")
	}
	if rule == nil {
		t.Fatal("expected matched allow rule, got nil")
	}

	// Other endpoint matches deny but no allow exception — denied.
	req, _ = http.NewRequest("POST", "https://api.example.com/users/delete", nil)
	allowed, rule, err = np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied (no allow exception)")
	}
	if rule == nil {
		t.Fatal("expected matched deny rule, got nil")
	}
}

func TestIsAllowed_DenyWithNoException(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Method: "DELETE", Host: "*api.example.com"},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("DELETE", "https://api.example.com/users/1", nil)
	allowed, rule, err := np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied, got allowed")
	}
	if rule == nil {
		t.Fatal("expected matched deny rule, got nil")
	}
	if rule.Method != "DELETE" {
		t.Errorf("expected deny rule method DELETE, got %s", rule.Method)
	}
}

func TestIsAllowed_BodyRegex(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Host: "*api.example.com", Path: "/sync", Body: `danger`},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}

	// Request with matching body — should be denied.
	req, _ := http.NewRequest("POST", "https://api.example.com/sync/data", strings.NewReader("do something danger here"))
	allowed, _, err := np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied for body matching regex")
	}

	// Request without matching body — should be allowed.
	req, _ = http.NewRequest("POST", "https://api.example.com/sync/data", strings.NewReader("safe content"))
	allowed, _, err = np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed for body without match")
	}
}

func TestIsAllowed_MethodOnly(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Method: "delete"},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}

	// DELETE should be denied (case-insensitive).
	req, _ := http.NewRequest("DELETE", "https://anything.com/whatever", nil)
	allowed, _, err := np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected DELETE denied")
	}

	// GET should be allowed.
	req, _ = http.NewRequest("GET", "https://anything.com/whatever", nil)
	allowed, _, err = np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected GET allowed")
	}
}

func TestIsAllowed_MethodRegex(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Method: "POST|PUT|DELETE"},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}

	// POST, PUT, DELETE should all be denied.
	for _, method := range []string{"POST", "PUT", "DELETE", "post", "Put"} {
		req, _ := http.NewRequest(method, "https://example.com/api", nil)
		allowed, _, err := np.IsAllowed(req)
		if err != nil {
			t.Fatal(err)
		}
		if allowed {
			t.Fatalf("expected %s denied", method)
		}
	}

	// GET and HEAD should be allowed.
	for _, method := range []string{"GET", "HEAD"} {
		req, _ := http.NewRequest(method, "https://example.com/api", nil)
		allowed, _, err := np.IsAllowed(req)
		if err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Fatalf("expected %s allowed", method)
		}
	}
}

func TestIsAllowed_AllFieldsMustMatch(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Method: "POST", Host: "*api.example.com", Body: "dangerous"},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}

	// All fields match — denied.
	req, _ := http.NewRequest("POST", "https://api.example.com/action", strings.NewReader("do something dangerous"))
	allowed, _, err := np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied when all fields match")
	}

	// Method and host match but body doesn't — allowed.
	req, _ = http.NewRequest("POST", "https://api.example.com/action", strings.NewReader("safe content"))
	allowed, _, err = np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed when body doesn't match")
	}

	// Wrong method — allowed.
	req, _ = http.NewRequest("GET", "https://api.example.com/action", nil)
	allowed, _, err = np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed when method doesn't match")
	}
}

func TestIsAllowed_MultipleRulesOR(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Method: "DELETE"},
			{Host: "*evil.com"},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}

	// Matches first rule only — denied.
	req, _ := http.NewRequest("DELETE", "https://safe.com/resource", nil)
	allowed, rule, err := np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied by first rule")
	}
	if rule.Method != "DELETE" {
		t.Errorf("expected first rule to match, got: %s", rule)
	}

	// Matches second rule only — denied.
	req, _ = http.NewRequest("GET", "https://evil.com/steal", nil)
	allowed, rule, err = np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied by second rule")
	}
	if rule.Host != "*evil.com" {
		t.Errorf("expected second rule to match, got: %s", rule)
	}

	// Matches neither — allowed.
	req, _ = http.NewRequest("GET", "https://safe.com/resource", nil)
	allowed, _, err = np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed when no rule matches")
	}
}

func TestDenyPatterns_Deduplicates(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Host: "*api.example.com", Path: "/sync", Body: "action1"},
			{Host: "*api.example.com", Path: "/sync", Body: "action2"},
			{Host: "*api.example.com", Path: "/delete"},
		},
	}
	patterns := np.DenyPatterns()
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
}

func TestDenyPatterns_CatchAllWhenNoHost(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Host: "*api.example.com"},
			{Method: "DELETE"}, // no host
		},
	}
	patterns := np.DenyPatterns()
	if len(patterns) != 1 || patterns[0] != "*" {
		t.Fatalf("expected [*], got %v", patterns)
	}
}

func TestLoad_ValidPolicy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yml")
	content := `version: "1"
network:
  deny:
    - method: "DELETE"
      host: "*api.example.com"
    - host: "*api.example.com"
      path: "/sync*"
      body: "delete_action"
  allow:
    - host: "*api.example.com"
      path: "/sync*"
      body: "read_action"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Version != "1" {
		t.Errorf("version: got %q, want %q", p.Version, "1")
	}
	if len(p.Network.DenyList) != 2 {
		t.Fatalf("deny rules: got %d, want 2", len(p.Network.DenyList))
	}
	if len(p.Network.AllowList) != 1 {
		t.Fatalf("allow rules: got %d, want 1", len(p.Network.AllowList))
	}
	if p.Network.DenyList[0].Method != "DELETE" {
		t.Errorf("deny[0].Method: got %q, want %q", p.Network.DenyList[0].Method, "DELETE")
	}
	if p.Network.DenyList[1].hostRegex == nil {
		t.Error("deny[1].hostRegex not compiled")
	}
	if p.Network.DenyList[1].pathRegex == nil {
		t.Error("deny[1].pathRegex not compiled")
	}
	if p.Network.DenyList[1].bodyRegex == nil {
		t.Error("deny[1].bodyRegex not compiled")
	}
	if p.Network.AllowList[0].bodyRegex == nil {
		t.Error("allow[0].bodyRegex not compiled")
	}
}

func TestCompileRules_InvalidRegex(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		errText string
	}{
		{"invalid path regex", Rule{Path: "["}, "invalid path regex"},
		{"invalid query regex", Rule{Query: "["}, "invalid query regex"},
		{"invalid body regex", Rule{Body: "["}, "invalid body regex"},
		{"invalid method regex", Rule{Method: "["}, "invalid method regex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compileRules([]Rule{tt.rule})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errText) {
				t.Errorf("expected %q in error, got: %s", tt.errText, err)
			}
		})
	}
}

func TestCompileRules_HostPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		match   string
		noMatch string
	}{
		{
			name:    "exact host",
			pattern: "api.example.com",
			match:   "api.example.com",
			noMatch: "evil.com",
		},
		{
			name:    "wildcard subdomain",
			pattern: "*.example.com",
			match:   "api.example.com",
			noMatch: "example.org",
		},
		{
			name:    "wildcard prefix",
			pattern: "*example.com",
			match:   "api.example.com",
			noMatch: "example.org",
		},
		{
			name:    "host with port",
			pattern: "api.example.com:8080",
			match:   "api.example.com:8080",
			noMatch: "api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := []Rule{{Host: tt.pattern}}
			if err := compileRules(rules); err != nil {
				t.Fatal(err)
			}
			if rules[0].hostRegex == nil {
				t.Fatal("expected compiled host regex")
			}
			if !rules[0].hostRegex.MatchString(tt.match) {
				t.Errorf("expected pattern %q to match %q", tt.pattern, tt.match)
			}
			if rules[0].hostRegex.MatchString(tt.noMatch) {
				t.Errorf("expected pattern %q not to match %q", tt.pattern, tt.noMatch)
			}
		})
	}
}

func TestCompileRules_PathPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		match   string
		noMatch string
	}{
		{
			name:    "exact path",
			pattern: "^/api/users$",
			match:   "/api/users",
			noMatch: "/api/users/1",
		},
		{
			name:    "prefix match",
			pattern: "^/api/",
			match:   "/api/users/1",
			noMatch: "/other/path",
		},
		{
			name:    "suffix match",
			pattern: "/delete$",
			match:   "/api/users/delete",
			noMatch: "/api/users/1",
		},
		{
			name:    "alternation",
			pattern: "/(users|accounts)",
			match:   "/api/users",
			noMatch: "/api/settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := []Rule{{Path: tt.pattern}}
			if err := compileRules(rules); err != nil {
				t.Fatal(err)
			}
			if rules[0].pathRegex == nil {
				t.Fatal("expected compiled path regex")
			}
			if !rules[0].pathRegex.MatchString(tt.match) {
				t.Errorf("expected pattern %q to match %q", tt.pattern, tt.match)
			}
			if rules[0].pathRegex.MatchString(tt.noMatch) {
				t.Errorf("expected pattern %q not to match %q", tt.pattern, tt.noMatch)
			}
		})
	}
}

func TestCompileRules_QueryPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		match   string
		noMatch string
	}{
		{
			name:    "exact query param",
			pattern: "action=delete",
			match:   "action=delete",
			noMatch: "action=read",
		},
		{
			name:    "alternation",
			pattern: "action=(delete|archive)",
			match:   "action=delete&id=1",
			noMatch: "action=read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := []Rule{{Query: tt.pattern}}
			if err := compileRules(rules); err != nil {
				t.Fatal(err)
			}
			if rules[0].queryRegex == nil {
				t.Fatal("expected compiled query regex")
			}
			if !rules[0].queryRegex.MatchString(tt.match) {
				t.Errorf("expected pattern %q to match %q", tt.pattern, tt.match)
			}
			if rules[0].queryRegex.MatchString(tt.noMatch) {
				t.Errorf("expected pattern %q not to match %q", tt.pattern, tt.noMatch)
			}
		})
	}
}

func TestIsAllowed_AllowExceptionByBody(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Method: "POST", Host: "*api.example.com", Path: "/sync", Body: "delete|archive"},
		},
		AllowList: []Rule{
			{Method: "POST", Host: "*api.example.com", Path: "/sync", Body: "read"},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}
	if err := compileRules(np.AllowList); err != nil {
		t.Fatal(err)
	}

	// Body matches deny but also matches allow exception — allowed.
	req, _ := http.NewRequest("POST", "https://api.example.com/sync/data", strings.NewReader("read delete"))
	allowed, rule, err := np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed (allow exception by body)")
	}
	if rule == nil || rule.Body != "read" {
		t.Fatalf("expected allow rule with body=read, got %v", rule)
	}

	// Body matches deny only — denied.
	req, _ = http.NewRequest("POST", "https://api.example.com/sync/data", strings.NewReader("archive stuff"))
	allowed, rule, err = np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied (no allow exception)")
	}
	if rule == nil || rule.Body != "delete|archive" {
		t.Fatalf("expected deny rule with body=delete|archive, got %v", rule)
	}
}

func TestIsAllowed_NilBodyWithBodyRules(t *testing.T) {
	np := &NetworkPolicy{
		DenyList: []Rule{
			{Host: "*api.example.com", Body: "dangerous"},
		},
	}
	if err := compileRules(np.DenyList); err != nil {
		t.Fatal(err)
	}

	// Nil body — body regex won't match, so request is allowed.
	req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)
	allowed, _, err := np.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed when body is nil and body regex can't match")
	}
}

func TestCompileRules_RejectsRegexInHost(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{"pipe alternation", "(mail|calendar).google.com"},
		{"backslash escape", `mail\.google\.com`},
		{"caret anchor", "^api.example.com"},
		{"dollar anchor", "api.example.com$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := []Rule{{Host: tt.host}}
			err := compileRules(rules)
			if err == nil {
				t.Fatalf("expected error for host %q, got nil", tt.host)
			}
			if !strings.Contains(err.Error(), "Regular expressions are not supported") {
				t.Errorf("expected regex warning in error, got: %s", err)
			}
		})
	}
}

func TestLoad_EmptyNetworkPolicy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yml")
	content := "version: \"1\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Network.DenyList) != 0 {
		t.Errorf("expected empty deny list, got %d rules", len(p.Network.DenyList))
	}
}
