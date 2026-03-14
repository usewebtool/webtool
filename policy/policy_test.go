package policy

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsAllowed_AllowExceptionOverridesDeny(t *testing.T) {
	p := &Policy{
		DenyList: []Rule{
			{Method: "POST", URL: "*api.example.com*"},
		},
		AllowList: []Rule{
			{Method: "POST", URL: "*api.example.com/login*"},
		},
	}
	if err := compileRules(p.DenyList); err != nil {
		t.Fatal(err)
	}
	if err := compileRules(p.AllowList); err != nil {
		t.Fatal(err)
	}

	// Login endpoint matches allow exception — allowed.
	req, _ := http.NewRequest("POST", "https://api.example.com/login", nil)
	allowed, rule, err := p.IsAllowed(req)
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
	allowed, rule, err = p.IsAllowed(req)
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
	p := &Policy{
		DenyList: []Rule{
			{Method: "DELETE", URL: "*api.example.com*"},
		},
	}
	if err := compileRules(p.DenyList); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("DELETE", "https://api.example.com/users/1", nil)
	allowed, rule, err := p.IsAllowed(req)
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
	p := &Policy{
		DenyList: []Rule{
			{URL: "*api.example.com/sync*", Body: `danger`},
		},
	}
	if err := compileRules(p.DenyList); err != nil {
		t.Fatal(err)
	}

	// Request with matching body — should be denied.
	req, _ := http.NewRequest("POST", "https://api.example.com/sync/data", strings.NewReader("do something danger here"))
	allowed, _, err := p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied for body matching regex")
	}

	// Request without matching body — should be allowed.
	req, _ = http.NewRequest("POST", "https://api.example.com/sync/data", strings.NewReader("safe content"))
	allowed, _, err = p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed for body without ^k")
	}
}

func TestIsAllowed_MethodOnly(t *testing.T) {
	p := &Policy{
		DenyList: []Rule{
			{Method: "delete"},
		},
	}
	if err := compileRules(p.DenyList); err != nil {
		t.Fatal(err)
	}

	// DELETE should be denied (case-insensitive).
	req, _ := http.NewRequest("DELETE", "https://anything.com/whatever", nil)
	allowed, _, err := p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected DELETE denied")
	}

	// GET should be allowed.
	req, _ = http.NewRequest("GET", "https://anything.com/whatever", nil)
	allowed, _, err = p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected GET allowed")
	}
}

func TestIsAllowed_AllFieldsMustMatch(t *testing.T) {
	p := &Policy{
		DenyList: []Rule{
			{Method: "POST", URL: "*api.example.com*", Body: "dangerous"},
		},
	}
	if err := compileRules(p.DenyList); err != nil {
		t.Fatal(err)
	}

	// All three match — denied.
	req, _ := http.NewRequest("POST", "https://api.example.com/action", strings.NewReader("do something dangerous"))
	allowed, _, err := p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied when all fields match")
	}

	// Method and URL match but body doesn't — allowed.
	req, _ = http.NewRequest("POST", "https://api.example.com/action", strings.NewReader("safe content"))
	allowed, _, err = p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed when body doesn't match")
	}

	// Wrong method — allowed.
	req, _ = http.NewRequest("GET", "https://api.example.com/action", nil)
	allowed, _, err = p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed when method doesn't match")
	}
}

func TestIsAllowed_MultipleRulesOR(t *testing.T) {
	p := &Policy{
		DenyList: []Rule{
			{Method: "DELETE"},
			{URL: "*evil.com*"},
		},
	}
	if err := compileRules(p.DenyList); err != nil {
		t.Fatal(err)
	}

	// Matches first rule only — denied.
	req, _ := http.NewRequest("DELETE", "https://safe.com/resource", nil)
	allowed, rule, err := p.IsAllowed(req)
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
	allowed, rule, err = p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected denied by second rule")
	}
	if rule.URL != "*evil.com*" {
		t.Errorf("expected second rule to match, got: %s", rule)
	}

	// Matches neither — allowed.
	req, _ = http.NewRequest("GET", "https://safe.com/resource", nil)
	allowed, _, err = p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed when no rule matches")
	}
}

func TestDenyPatterns_Deduplicates(t *testing.T) {
	p := &Policy{
		DenyList: []Rule{
			{URL: "*api.example.com/sync*", Body: "action1"},
			{URL: "*api.example.com/sync*", Body: "action2"},
			{URL: "*api.example.com/delete*"},
		},
	}
	patterns := p.DenyPatterns()
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
}

func TestDenyPatterns_CatchAllWhenNoURL(t *testing.T) {
	p := &Policy{
		DenyList: []Rule{
			{URL: "*api.example.com*"},
			{Method: "DELETE"}, // no URL
		},
	}
	patterns := p.DenyPatterns()
	if len(patterns) != 1 || patterns[0] != "*" {
		t.Fatalf("expected [*], got %v", patterns)
	}
}

func TestLoad_ValidPolicy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yml")
	content := `version: "1"
deny:
  - method: "DELETE"
    url: "*api.example.com*"
  - url: "*api.example.com/sync*"
    body: "delete_action"
allow:
  - url: "*api.example.com/sync*"
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
	if len(p.DenyList) != 2 {
		t.Fatalf("deny rules: got %d, want 2", len(p.DenyList))
	}
	if len(p.AllowList) != 1 {
		t.Fatalf("allow rules: got %d, want 1", len(p.AllowList))
	}
	if p.DenyList[0].Method != "DELETE" {
		t.Errorf("deny[0].Method: got %q, want %q", p.DenyList[0].Method, "DELETE")
	}
	if p.DenyList[1].urlRegex == nil {
		t.Error("deny[1].urlRegex not compiled")
	}
	if p.DenyList[1].bodyRegex == nil {
		t.Error("deny[1].bodyRegex not compiled")
	}
	if p.AllowList[0].bodyRegex == nil {
		t.Error("allow[0].bodyRegex not compiled")
	}
}

func TestLoad_InvalidRegex(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yml")
	content := "deny:\n  - body: \"[\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
	if !strings.Contains(err.Error(), "invalid body regex") {
		t.Errorf("expected 'invalid body regex' in error, got: %s", err)
	}
}

func TestCompileRules_URLPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		match   string
		noMatch string
	}{
		{
			name:    "wildcard both ends",
			pattern: "*api.example.com/sync*",
			match:   "https://api.example.com/sync/u/0/i/s",
			noMatch: "https://other.example.com/path",
		},
		{
			name:    "wildcard suffix only",
			pattern: "https://api.example.com/*",
			match:   "https://api.example.com/users/1",
			noMatch: "https://other.example.com/users/1",
		},
		{
			name:    "question mark single char",
			pattern: "https://example.com/v?/api",
			match:   "https://example.com/v2/api",
			noMatch: "https://example.com/v10/api",
		},
		{
			name:    "no wildcards exact match",
			pattern: "https://example.com/exact",
			match:   "https://example.com/exact",
			noMatch: "https://example.com/exact/more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := []Rule{{URL: tt.pattern}}
			if err := compileRules(rules); err != nil {
				t.Fatal(err)
			}
			if rules[0].urlRegex == nil {
				t.Fatal("expected compiled url regex")
			}
			if !rules[0].urlRegex.MatchString(tt.match) {
				t.Errorf("expected pattern %q to match %q", tt.pattern, tt.match)
			}
			if rules[0].urlRegex.MatchString(tt.noMatch) {
				t.Errorf("expected pattern %q not to match %q", tt.pattern, tt.noMatch)
			}
		})
	}
}

func TestIsAllowed_AllowExceptionByBody(t *testing.T) {
	p := &Policy{
		DenyList: []Rule{
			{Method: "POST", URL: "*api.example.com/sync*", Body: "delete|archive"},
		},
		AllowList: []Rule{
			{Method: "POST", URL: "*api.example.com/sync*", Body: "read"},
		},
	}
	if err := compileRules(p.DenyList); err != nil {
		t.Fatal(err)
	}
	if err := compileRules(p.AllowList); err != nil {
		t.Fatal(err)
	}

	// Body matches deny but also matches allow exception — allowed.
	req, _ := http.NewRequest("POST", "https://api.example.com/sync/data", strings.NewReader("read delete"))
	allowed, rule, err := p.IsAllowed(req)
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
	allowed, rule, err = p.IsAllowed(req)
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
	p := &Policy{
		DenyList: []Rule{
			{URL: "*api.example.com*", Body: "dangerous"},
		},
	}
	if err := compileRules(p.DenyList); err != nil {
		t.Fatal(err)
	}

	// Nil body — body regex won't match, so request is allowed.
	req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)
	allowed, _, err := p.IsAllowed(req)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected allowed when body is nil and body regex can't match")
	}
}

func TestLoad_NoDenyRules(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yml")
	content := "version: \"1\"\nallow:\n  - url: \"*example.com*\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for no deny rules, got nil")
	}
	if !strings.Contains(err.Error(), "at least one deny rule") {
		t.Errorf("expected 'at least one deny rule' in error, got: %s", err)
	}
}
