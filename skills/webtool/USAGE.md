# webtool — Command Reference

A fast CLI tool that drives Chrome via Chrome DevTools Protocol. Designed for agent-driven workflows.

**Requirements:** Chrome 144+ with dynamic remote debugging enabled.

## Quick Start

`webtool` does **not** auto-start its daemon on first use. Start it explicitly before running normal commands.

```bash
webtool start                      # start daemon first
webtool open https://example.com   # navigate to a URL
webtool snapshot                   # see interactive elements
webtool click 43821                # click an element by its ID
webtool type 43822 "hello"         # type into an input
```

Chrome may show a permission dialog when the daemon connects. Ask the user to click Allow.

## Agent Workflow

The core loop is: **snapshot → reason → action → snapshot**

1. Take a `snapshot` to see interactive elements with their IDs
2. Decide which element to interact with
3. Perform an action (`click`, `type`, `select`, etc.)
4. Take another `snapshot` to see the result

Every action command automatically waits for the DOM to stabilize before returning, so the next snapshot reflects the settled page state.

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--timeout` | `30s` | Timeout for the command (e.g. `5s`, `1m`) |

## Selectors

Most action commands accept a `<selector>` argument. Three formats are supported:

| Format | Example | Resolution |
|--------|---------|------------|
| Integer | `43821` | backendNodeId from snapshot (immediate, no retry) |
| `/` or `//` prefix | `//button[@id='submit']` | XPath (retries until found or timeout) |
| Anything else | `#submit` | CSS selector (retries until found or timeout) |

Use backendNodeId from snapshots for the most reliable targeting. CSS/XPath selectors retry until the element appears or the timeout expires.

## Commands

### Navigation

#### `open <url>`

Navigate the browser to a URL. Waits for the page to load.

```bash
webtool open https://example.com
webtool open "https://google.com/search?q=hello+world"
webtool open --new https://example.com   # open in a new tab
```

| Flag | Description |
|------|-------------|
| `--new` | Open the URL in a new tab instead of navigating the current tab. The new tab becomes active. |

#### `back`

Navigate back in browser history.

```bash
webtool back
```

#### `forward`

Navigate forward in browser history.

```bash
webtool forward
```

#### `reload`

Reload the current page.

```bash
webtool reload
```

### Page Inspection

#### `snapshot`

Print a text snapshot of the current page. Returns a compact, token-efficient list of elements with their backendNodeId, role, and label. The core tool in the snapshot → reason → action loop.

```bash
webtool snapshot                   # default mode
webtool snapshot -i                # interactive only
webtool snapshot -a                # all content
```

| Flag | Description |
|------|-------------|
| `-i`, `--interactive` | Show only interactive elements and structural grouping. Strips headings, images, and labels. Use when you only need to find buttons, links, and form controls — e.g. filling out a form or navigating a menu. Lowest token count. |
| `-a`, `--all` | Show everything in default mode plus text content (paragraphs, blockquotes, code blocks, static text). Use when you need to read or compare page content — e.g. extracting article text, comparing search results, or verifying displayed information. Highest token count. |
| (none) | **Default mode.** Interactive elements + structural grouping + headings + content-container summaries + status/alert messages. The workhorse mode for most tasks. Content containers like list items and articles show a summary of their non-interactive text (sender, date, price, etc.) so you can identify items without extracting each one. |

The flags are mutually exclusive — use at most one.

**When to use each mode:**

- **Default** — start here. Gives you enough context to identify elements and understand page structure. Content-container summaries let you scan repeated items (inbox rows, search results, product cards) without reading every detail.
- **Interactive (`-i`)** — use when the page is complex and you already know what you're looking for. Cuts noise from headings, images, and labels to focus purely on actionable elements.
- **All (`-a`)** — use when you need to read page content, not just interact with it. Shows paragraphs, static text, blockquotes, and code blocks. If you still need more detail, use `webtool extract <id>` on a specific element.

**Output format:**

```
[url] https://example.com
[title] Example Domain

[1] main
  [10] form "Login"
    [11] textbox "Email" value="user@example.com"
    [12] textbox "Password"
    [13] button "Sign in"
  [20] navigation "Primary"
    [21] link "Home" url="https://example.com/"
    [22] link "About" url="https://example.com/about"
[30] heading[1] "Welcome"
[31] checkbox "Remember me" checked
```

Each element line: `[backendNodeId] role "name"` followed by optional attributes:
- `value="..."` — current input value
- `url="..."` — link href (query params stripped)
- State flags: `focused`, `checked`, `disabled`, `readonly`, `required`, `selected`, `expanded`, `collapsed`, `invalid`

Structural containers (landmarks, forms, lists, articles, sections) are shown with 2-space indentation. Headings show their level as `heading[1]`, `heading[2]`, etc.

**Content-container summaries** (default and all modes): List items and articles without an explicit accessible name show a synthetic summary built from their non-interactive text, joined with ` | `. For example, an inbox row might show:

```
[201] listitem "John Doe | Mar 10"
  [200] checkbox "Select"
  [202] link "Meeting Tomorrow - Hi, can we meet..."
```

All text in snapshots is truncated to 160 characters. Use `webtool extract <id>` to read the full content of any element.

#### `extract [selector]`

Extract page content as markdown. Default timeout is **1 second** (not the global 30s).

```bash
webtool extract                    # extract entire page as markdown
webtool extract 43821              # extract a specific element
webtool extract "#article"         # extract by CSS selector
webtool extract --main             # extract only the main content area
webtool extract --html             # extract as raw HTML
webtool extract --html 43821       # extract a specific element as HTML
```

| Flag | Default | Description |
|------|---------|-------------|
| `--main` | `false` | Extract only the main content area (`<main>` or `[role='main']`). Mutually exclusive with a selector. |
| `--html` | `false` | Return raw HTML instead of markdown |

#### `html [selector]`

Alias for `extract --html`. Extracts page content as HTML.

```bash
webtool html                       # full page HTML
webtool html 43821                 # specific element HTML
```

### Actions

#### `click <selector>`

Click an element.

```bash
webtool click 43821                # click by backendNodeId
webtool click "#submit"            # click by CSS selector
webtool click "//button[1]"        # click by XPath
```

Before clicking, the reliability pipeline verifies the element is visible, enabled, not obscured, and stable. After clicking, it waits for the DOM to settle and detects navigation.

#### `type <selector> <text>`

Type text into an element. Uses paste-like insertion (CDP `Input.insertText`) — enters the full string in a single operation.

```bash
webtool type 43823 "user@example.com"
webtool type "#search" "search query"
```

Replaces existing content — the field is selected-all before insertion, so the new text overwrites whatever was there.

#### `select <selector> <text>`

Select a dropdown option by its visible text.

```bash
webtool select 43826 "United States"
webtool select "#country" "Canada"
```

Returns an error if no option matches the given text. Use `extract` on the select element to see available options.

#### `key <name>`

Send a key press. Key names are case-insensitive and follow Playwright/W3C naming.

```bash
webtool key Enter
webtool key escape
webtool key Tab
webtool key ArrowDown
```

Supported keys:

| Key | Aliases |
|-----|---------|
| `Enter` | `Return` |
| `Escape` | |
| `Tab` | |
| `Backspace` | |
| `Delete` | |
| `Space` | |
| `ArrowUp` | |
| `ArrowDown` | |
| `ArrowLeft` | |
| `ArrowRight` | |
| `Home` | |
| `End` | |
| `PageUp` | |
| `PageDown` | |

#### `hover <selector>`

Move the mouse over an element without clicking. Triggers CSS `:hover` states and JS `mouseenter`/`mouseover` events. Useful for revealing dropdown menus, tooltips, and hidden action buttons that only appear on hover.

```bash
webtool hover 43821
webtool hover "#dropdown-trigger"
```

After hovering, take a `snapshot` to see newly revealed elements.

#### `upload <selector> <file> [file...]`

Set one or more files on a `<input type="file">` element. File paths are resolved to absolute paths.

```bash
webtool upload 43825 document.pdf
webtool upload "#file-input" photo1.jpg photo2.jpg photo3.jpg
```

Note: Chrome's accessibility tree exposes file inputs as `button "Choose File"` — they appear in snapshots as buttons, not as a distinct file input role.

### Waiting

#### `wait <duration|selector>`

Wait for a duration or until an element exists in the DOM.

```bash
webtool wait 2s                    # sleep for 2 seconds
webtool wait 500ms                 # sleep for 500 milliseconds
webtool wait "#results"            # wait until element exists (CSS selector)
webtool wait "//div[@class='loaded']"   # wait until element exists (XPath)
```

If the argument parses as a Go duration (e.g. `2s`, `500ms`, `1m`), it sleeps for that long. Otherwise it treats the argument as a selector and polls until the element appears or `--timeout` expires.

Governed by the global `--timeout` flag. `webtool wait 60s` will time out at the default 30s unless you pass `--timeout 60s`.

### JavaScript

#### `eval <js>`

Execute a JavaScript expression and print the result.

```bash
webtool eval "document.title"
webtool eval "window.location.href"
webtool eval "document.querySelectorAll('a').length"
```

Only expressions are supported, not statements (`const`, `let`, `var`). For multi-statement code, wrap in an IIFE:

```bash
webtool eval "(function(){ const a = 1; return a; })()"
```

### Tab Management

#### `tabs`

List open browser tabs. Output is one tab per line: `<index> <title> <url>`. The tab webtool will operate on is marked `[active]`. DevTools, `about:`, and `chrome://` tabs are filtered out.

```bash
webtool tabs
```

Output:

```
1 Example Domain https://example.com [active]
2 Google https://google.com
```

`[active]` indicates which tab the next command will operate on. This is the tab last navigated or switched to by webtool, which may differ from the tab visually focused in Chrome's tab bar.

#### `tab <index>`

Switch to a tab by its 1-based index (as shown by `webtool tabs`).

```bash
webtool tab 2
```

### Daemon Management

#### `start`

Start the daemon in the background. The daemon holds the Chrome WebSocket connection so you only approve the Chrome permission dialog once.

Required before running normal `webtool` commands. The CLI does not auto-start the daemon on first command.

```bash
webtool start
webtool start --policy policy.yml   # start with a security policy
```

| Flag | Description |
|------|-------------|
| `-p`, `--policy` | Path to a security policy YAML file for request interception |

#### `stop`

Stop the daemon. Idempotent — exits cleanly if no daemon is running.

```bash
webtool stop
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Any error (element not found, timeout, connection failed, etc.) |

## Errors

Actionability errors are returned when an element cannot be interacted with. Each error includes a clear message with a recommended recovery action.

| Error | Meaning | Recovery |
|-------|---------|----------|
| `stale node` | backendNodeId no longer in DOM (React/Vue re-render) | Run `snapshot` again |
| `element not found` | No element matches the selector | Check selector, run `snapshot` |
| `element not visible` | Element has no visible shape or is off-screen | Scroll or wait for it to appear |
| `element obscured` | Covered by overlay/modal/banner | Dismiss the covering element first |
| `element not clickable` | `pointer-events: none` in CSS | Find an alternative element |
| `element not stable` | Position/size still changing (animation) | Wait and retry |
| `option not found` | No matching option in select dropdown | Use `extract` to see available options |
| `element disabled` | Element is disabled | Wait for it to become enabled |
| `request blocked by policy` | Network request blocked by security policy | Check the policy rules |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBTOOL_HOME` | `~/.webtool` | Directory for daemon socket, PID file, and logs |
| `WEBTOOL_CHROME_DATA_DIR` | OS default | Chrome user data directory for DevToolsActivePort discovery |
