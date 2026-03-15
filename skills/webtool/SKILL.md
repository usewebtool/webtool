---
name: webtool
description: Drive Chrome browser via CLI for web automation, form filling, scraping, and testing. Use when the user asks to interact with websites, fill forms, extract content, click buttons, or automate any browser task. Provides text snapshots of pages with element IDs for precise interaction.
compatibility: Requires webtool binary (Go) and Chrome with remote debugging enabled.
allowed-tools: Bash(webtool:*)
---

# webtool — Browser Automation via CLI

webtool drives your local Chrome browser through the Chrome DevTools Protocol. It connects to your running Chrome instance — no separate browser, no Playwright, no Node.js. You interact with the user's actual authenticated browser session.

## Setup

Before first use, the user must enable remote debugging in Chrome at `chrome://inspect/#remote-debugging`. This only needs to be done once — Chrome remembers the setting.

To connect, run `webtool start`. Chrome will show a permission dialog — **ask the user to click Allow**. This happens once per daemon session. After that, all commands work.

```bash
webtool start    # start daemon — user must click Allow in Chrome
```

The user can run `webtool stop` to shut down the daemon and close the Chrome connection.

**If a command returns `"daemon not running"`**, run `webtool start` and ask the user to approve the Chrome dialog.

## Core Workflow

The agent loop is: **snapshot → reason → action → snapshot**

```bash
webtool open https://example.com     # navigate
webtool snapshot                      # see the page
webtool click 43821                   # act on an element
webtool snapshot                      # see the result
```

Every action waits for the DOM to stabilize before returning. The next snapshot reflects the settled page.

## Commands

### Navigation

```bash
webtool open <url>              # navigate to URL
webtool open --new <url>        # open in a new tab
webtool tabs                    # list tabs — [active] marks the current one
webtool tab <index>             # switch to tab by 1-based index
```

### Snapshots

```bash
webtool snapshot                # interactive elements + structure + summaries
webtool snapshot -i             # interactive only — lowest tokens
webtool snapshot -a             # all content — includes paragraphs, static text
```

Use default `snapshot` for most tasks. Use `-i` for complex pages when you only need actionable elements. Use `-a` when you need to read page content.

**Output format:** each line is `[backendNodeId] role "name"` with optional attributes.

```
[url] https://example.com
[title] Example Page

[10] form "Login"
  [11] textbox "Email" value="user@example.com"
  [12] textbox "Password"
  [13] button "Sign in"
[20] link "Forgot password?" url="/reset"
[30] heading[1] "Welcome"
```

- `[11]` is the backendNodeId — use it in action commands
- Roles: `button`, `link`, `textbox`, `checkbox`, `radio`, `combobox`, `heading[N]`, etc.
- Attributes: `value="..."`, `url="..."`, `focused`, `checked`, `disabled`, `expanded`
- Indentation shows containment (form contains its inputs)

### Interacting with the Page

```bash
webtool click <selector>        # click an element
webtool type <selector> "text"  # type into an input (replaces existing text)
webtool select <selector> "Option Text"  # select dropdown option by visible text
webtool key Enter               # press a key: Enter, Escape, Tab, ArrowDown, etc.
webtool hover <selector>        # hover to reveal hidden menus/buttons
webtool upload <selector> file  # set files on a file input
webtool wait 2s                 # sleep for a duration
webtool wait "#results"         # wait until element exists (CSS/XPath)
```

### Scraping Data

Extract page content as **markdown** (default) or **raw HTML** (with `--html`).

```bash
webtool extract                 # full page as markdown
webtool extract <selector>      # specific element as markdown
webtool extract --main          # main content area only as markdown
webtool extract --html          # full page as raw HTML
webtool extract --html <selector>  # specific element as raw HTML
```

Note: `extract` defaults to a 1-second timeout (not 30s) so typos in selectors fail fast. Override with `--timeout` if the page is slow to render.

## Selectors

Commands accept three selector formats:

| Format | Example | Notes |
|--------|---------|-------|
| Integer | `43821` | backendNodeId from snapshot — most reliable |
| CSS | `#submit`, `.btn` | Retries until found or timeout |
| XPath | `//button[@type='submit']` | Retries until found or timeout |

Always prefer backendNodeId from the most recent snapshot.

## Key Patterns

**Hidden elements appear on hover.** Some buttons (delete, edit, menu) only render when the parent is hovered. If you expect an action button but don't see it, hover over the containing element and re-snapshot.

```bash
webtool hover 329              # hover over the item
webtool snapshot               # now the delete button appears
webtool click 330              # click the revealed button
```

**File inputs appear as buttons.** Chrome's accessibility tree shows `<input type="file">` as `button "Choose File"`. Target them by backendNodeId like any other element.

**Form filling.** `type` replaces existing content (select-all then insert). No need to clear first.

**After navigation, always re-snapshot.** backendNodeIds become stale after page changes.

## Troubleshooting

Most issues come from the page still loading JavaScript after the snapshot was taken. When something fails or an expected element is missing, **re-snapshot first** before trying anything else. The goal is to fail fast and retry — don't debug, just take a fresh snapshot.

When re-snapshotting doesn't help, these commands let you bypass the accessibility tree and work with the page directly:

```bash
webtool eval "<js>"              # run JavaScript on the page
webtool html                     # get full page HTML
webtool html <selector>          # get HTML of a specific element
webtool cdp <method> [params]    # send a raw Chrome DevTools Protocol command
```

**Use `eval`** when an element won't respond to `click` or `type` — e.g. dismiss a `beforeunload` dialog (`webtool eval "window.onbeforeunload = null"`), scroll to exact coordinates, or trigger a JS handler directly.

**Use `html`** as a last resort when multiple re-snapshots still miss elements you expect. The accessibility tree can omit elements without accessible roles — raw HTML shows everything.

**Use `cdp`** as a last resort for low-level browser control — e.g. `webtool cdp Input.insertText '{"text":"hello"}'` for canvas-based apps like Google Docs where normal `type` doesn't work.

## Common Errors

| Error | Recovery |
|-------|----------|
| `stale node` | Re-snapshot — the page re-rendered |
| `element not found` | Check selector, re-snapshot |
| `element not visible` | Scroll or wait for it to appear |
| `element obscured` | Dismiss the covering element (modal, banner) |
| `element not clickable` | `pointer-events: none` in CSS — find an alternative element |
| `element not stable` | Position/size still changing — wait and retry |
| `element disabled` | Wait for it to become enabled |
| `option not found` | Use `extract` on the select to see available options |
| `request blocked by policy` | A security policy is blocking this network request. This is intentional and cannot be bypassed. Do not retry. |

## Global Flags

```bash
webtool --timeout 60s <command>   # override default 30s timeout
```

## Full Reference

See [USAGE.md](USAGE.md) for complete command documentation with all flags and detailed examples.
