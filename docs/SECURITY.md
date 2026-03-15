# Security Policy

webtool supports network-level request interception via a YAML policy file. When a policy is configured, matching requests are aborted before reaching the server. This enables safe agent-driven browsing by preventing destructive operations (deleting emails, sending messages, etc.) while allowing read-only access.

## Usage

Start the daemon with a policy file:

```bash
webtool start -p policy.yml
```

The policy is validated at startup — invalid files are rejected immediately.

## Policy File Format

```yaml
version: "1"

deny:
  - method: "DELETE"
    url: "*api.example.com*"
  - url: "*api.example.com/sync*"
    body: "delete_action"

allow:
  - url: "*api.example.com/sync*"
    body: "read_action"
```

### Top-Level Fields

| Field | Required | Description |
|-------|----------|-------------|
| `version` | No | Policy format version. Currently `"1"`. |
| `deny` | Yes | List of rules. At least one deny rule is required. |
| `allow` | No | List of exception rules that override deny matches. |

### Rule Fields

Each rule in `deny` or `allow` has three optional fields. All specified fields must match for the rule to trigger (AND logic). Multiple rules in a list are checked in order (OR logic — first match wins).

| Field | Type | Description |
|-------|------|-------------|
| `method` | string | HTTP method, case-insensitive. Exact match. Example: `"DELETE"`, `"POST"` |
| `url` | string | URL pattern using CDP wildcards: `*` matches any characters, `?` matches a single character. Example: `"*api.example.com/sync*"` |
| `body` | string | Regular expression matched against the request body. Uses Go `regexp` syntax. Example: `"delete\|archive"` |

If a field is omitted, it matches anything. A rule with only `method: "DELETE"` blocks all DELETE requests to any URL.

### How Deny and Allow Work Together

`deny` rules block requests. `allow` rules are exceptions to the deny list — they let specific requests through that would otherwise be blocked. You must have at least one `deny` rule; `allow` is only useful when a deny rule matches and you want to carve out an exception.

### URL Pattern Syntax

URL patterns use CDP wildcard syntax (not regex):

| Pattern | Meaning |
|---------|---------|
| `*` | Matches any sequence of characters |
| `?` | Matches exactly one character |
| No wildcards | Exact match |

Examples:

| Pattern | Matches | Does Not Match |
|---------|---------|----------------|
| `*api.example.com/sync*` | `https://api.example.com/sync/data` | `https://other.example.com/path` |
| `https://api.example.com/*` | `https://api.example.com/users/1` | `https://other.example.com/users/1` |
| `https://example.com/v?/api` | `https://example.com/v2/api` | `https://example.com/v10/api` |

### Body Regex

Body patterns use Go regular expression syntax. Common patterns:

| YAML | Matches | Notes |
|------|---------|-------|
| `"delete"` | Request body containing "delete" | Simple substring |
| `"delete\|archive"` | Body containing "delete" or "archive" | Alternation |
| `"\\baction\\b"` | Body containing the word "action" | Word boundary match. Double backslash in YAML for single backslash in regex. |

Note: YAML requires double backslash (`\\`) to produce a single backslash in the regex. `"\\baction\\b"` in YAML becomes the regex `\baction\b`.

## Examples

### Block All DELETE Requests

```yaml
version: "1"

deny:
  - method: "DELETE"
```

### Block API Writes with Login Exception

Block all POST requests to an API, but allow the login endpoint:

```yaml
version: "1"

deny:
  - method: "POST"
    url: "*api.example.com*"

allow:
  - method: "POST"
    url: "*api.example.com/login*"
```

### Block Destructive Actions by Request Body

Block sync requests that contain delete or archive operations:

```yaml
version: "1"

deny:
  - url: "*api.example.com/sync*"
    body: "delete|archive"

allow:
  - url: "*api.example.com/sync*"
    body: "read"
```

### Block All Requests to a Domain

```yaml
version: "1"

deny:
  - url: "*evil.example.com*"
```

## Error Messages

When a request is blocked, the CLI returns:

```
request blocked by policy: POST https://api.example.com/sync/data (rule: url=*api.example.com/sync* body=delete|archive)
```

The error includes the HTTP method, full URL, and the matching rule for debugging.

## Limitations

- Policies are set at daemon startup and cannot be changed without restarting (`webtool stop && webtool start -p new-policy.yml`).
- Body inspection reads the full request body into memory. Very large request bodies may impact performance.
- URL patterns use wildcard syntax, not regular expressions. Use `body` for regex matching.
