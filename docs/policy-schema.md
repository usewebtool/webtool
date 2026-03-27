# Policy Schema Reference

Complete field reference for the webtool security policy YAML. Use this to generate or validate policy files.

## Top-Level Structure

```yaml
version: "1"        # optional
network:             # optional, network request interception rules
  deny: [...]        # optional, deny rules (implicit deny-all if omitted with allow present)
  allow: [...]       # optional, exception rules
actions:             # optional, CLI action restrictions
  deny: [...]        # optional, block specific actions
  allow: [...]       # optional, allow only specific actions
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | No | Policy schema version. Currently `"1"`. |
| `network` | object | No | Network request interception rules. |
| `actions` | object | No | CLI action restrictions. |

## Network

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `deny` | list of rules | No | Rules that block matching requests. If omitted but `allow` is present, all requests are denied by default. |
| `allow` | list of rules | No | Exception rules that override deny matches. At least one of `deny` or `allow` is required. |

## Actions

Controls which CLI actions the agent can perform.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `deny` | list of strings | No | Actions to block. Everything else is allowed. |
| `allow` | list of strings | No | Actions to allow. Everything else is blocked. |

**Validation rules:**
- Only one of `deny` or `allow` may be specified, not both.
- Action names are case-insensitive (lowercased on load).
- Unknown action names are rejected at load time.

**Valid action names:** `open`, `snapshot`, `click`, `type`, `key`, `back`, `forward`, `reload`, `eval`, `select`, `extract`, `switch`, `wait`, `upload`, `hover`, `tabs`.

`health` and `stop` always bypass the action policy.

## Rule

Each rule in `deny` or `allow`. All specified fields must match (AND logic). Multiple rules in a list use OR logic — first match wins. If a field is omitted, it matches anything.

| Field | Syntax | Matched Against | Description |
|-------|--------|-----------------|-------------|
| `method` | Regex (case-insensitive) | HTTP method (`r.Method`) | Example: `"POST\|PUT\|DELETE\|PATCH"` |
| `host` | CDP wildcard | Parsed URL host (`r.URL.Host`) | Includes port if present. Example: `"*api.example.com"` |
| `path` | Regex | Parsed URL path (`r.URL.Path`) | Example: `"/sync"`, `"^/api/"`, `"/(users\|accounts)"` |
| `query` | Regex | Parsed URL query string (`r.URL.RawQuery`) | Example: `"action=delete"`, `"action=(delete\|archive)"` |
| `header` | Regex | Serialized request headers | Headers in wire format: `"Name: Value\r\n"` per header. Header names use Go canonical form (e.g. `Content-Type`, `Authorization`). Example: `"Authorization:.*Bearer"` |
| `body` | Regex | Request body | Example: `"delete\|archive"` |

## Syntax Details

### CDP Wildcard (host only)

The `host` field uses CDP wildcard syntax, not regex:

| Pattern | Meaning |
|---------|---------|
| `*` | Matches any sequence of characters (zero or more) |
| `?` | Matches exactly one character |
| No wildcards | Exact match |

### Regex (method, path, query, header, body)

All fields except `host` use Go `regexp` syntax. `method` is automatically case-insensitive (`(?i)` prefix). All other regex fields are case-sensitive — add `(?i)` to your pattern if needed.

Regex fields use **substring matching** by default. Use `^` and `$` anchors for exact matching:

| Pattern | Behavior |
|---------|----------|
| `/sync` | Matches any path containing `/sync` (e.g. `/sync`, `/sync/data`, `/old/sync`) |
| `^/sync` | Matches paths starting with `/sync` |
| `^/sync$` | Matches exactly `/sync` |

## Component-Level Matching

Each field is matched against the **parsed** URL component, not the raw URL string. The request URL is parsed with Go's `url.Parse()`, then:

- `host` matches against `URL.Host` (includes port if present)
- `path` matches against `URL.Path`
- `query` matches against `URL.RawQuery`


## Deny/Allow Evaluation

1. If no deny rules are specified but allow rules exist, requests are denied unless they match an allow rule.
2. Check deny rules. No match -> **allowed** (default pass-through).
3. Deny matched -> check allow rules. Match -> **allowed** (exception overrides deny).
4. Deny matched, no allow exception -> **denied**.

## Examples

### Block all writes

```yaml
network:
  deny:
    - method: "POST|PUT|DELETE|PATCH"
```

### Block a domain

```yaml
network:
  deny:
    - host: "*evil.example.com"
```

### Block writes to an API with a login exception

```yaml
network:
  deny:
    - method: "POST"
      host: "*api.example.com"

  allow:
    - method: "POST"
      host: "*api.example.com"
      path: "/login"
```

### Block destructive actions by request body

```yaml
network:
  deny:
    - host: "*api.example.com"
      path: "/sync"
      body: "delete|archive"

  allow:
    - host: "*api.example.com"
      path: "/sync"
      body: "read"
```

### Block requests with auth headers

```yaml
network:
  deny:
    - header: "Authorization:.*Bearer"
```

### Block DELETE to specific paths

```yaml
network:
  deny:
    - method: "DELETE"
      host: "*api.example.com"
      path: "^/api/(users|accounts)"
```

### Allow only specific domains

```yaml
network:
  allow:
    - host: "*.example.com"
    - host: "docs.google.com"
```

### Block specific query parameters

```yaml
network:
  deny:
    - host: "*api.example.com"
      query: "action=(delete|archive)"
```

### Block JavaScript execution

```yaml
actions:
  deny:
    - eval
```

### Read-only agent (observe but don't interact)

```yaml
actions:
  allow:
    - snapshot
    - extract
    - tabs
```

### Combined network and action restrictions

```yaml
network:
  allow:
    - host: "*.example.com"
actions:
  deny:
    - eval
```

## Error Messages

When a network request is blocked:

```
request blocked by policy: POST https://api.example.com/sync (rule: host=*api.example.com path=/sync body=delete|archive)
```

When an action is blocked:

```
action "eval" is blocked by policy
```
