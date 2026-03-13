# webtool

A fast, single-binary CLI that drives your Chrome browser. No Playwright, no Node.js, no browser downloads — just a Go binary and Chrome.

Built for AI agent workflows. Webtool turns web pages into token-efficient text snapshots that LLMs can read, and provides simple commands to click, type, and navigate.

## Install

```bash
go install github.com/machinae/webtool@latest
```

Or download a binary from the [releases page](https://github.com/machinae/webtool/releases).

## Setup

**Chrome 144+** is required.

1. Open Chrome and navigate to `chrome://inspect/#remote-debugging`
2. Enable remote debugging for your current browser session
3. Run any webtool command — a daemon starts automatically and connects to Chrome
4. Chrome will show a permission dialog on the first connection — click **Accept**

That's it. The daemon keeps the connection open, so you only approve once per session.

To stop the daemon and close the connection run `webtool stop`.

## Usage

```bash
webtool open https://example.com    # navigate to a URL
webtool snapshot                    # text snapshot of interactive elements
webtool click 43821                 # click an element by its ID from the snapshot
webtool type 43822 "hello world"    # type into an input field
webtool key Enter                   # press a key
webtool extract --main              # extract the main content as markdown
webtool tabs                        # list open tabs
webtool tab 2                       # switch to tab 2
webtool stop                        # close the connection 
```

The workflow is **snapshot → action → snapshot**: take a snapshot to see what's on the page, act on an element by its ID, then snapshot again to see the result.

See [docs/usage.md](docs/usage.md) for the full command reference.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBTOOL_HOME` | `~/.webtool` | Base directory for runtime files (socket, PID, logs) |
| `WEBTOOL_CHROME_DATA_DIR` | OS default | Chrome user data directory for DevToolsActivePort discovery |

## FAQ

<details>
<summary>Does webtool support headless mode?</summary>

Webtool doesn't launch Chrome — it connects to your already-running instance. If you launch Chrome in headless mode yourself, webtool will connect to it just fine.

What webtool intentionally avoids is *managing* a headless Chrome instance for you. Why? Because connecting to your real Chrome session is fundamentally better for agent workflows:

- **No bot detection.** Automated Chrome instances (launched with `--remote-debugging-port` or via Playwright/Puppeteer) set `navigator.webdriver=true`, which triggers CAPTCHAs and blocks on most websites. Your normal Chrome session doesn't have this flag — webtool inherits that advantage by connecting to it rather than launching its own instance.
- **Real logins.** Your Chrome already has your cookies, sessions, and saved passwords. No need to automate login flows or manage auth tokens.
- **You can see what's happening.** When an agent controls your browser, you watch it work in real time. This builds trust and makes debugging trivial.

If you need headless for CI or server environments, you can launch Chrome yourself with `chrome --headless --remote-debugging-port=0` and webtool will connect to it — but you'll be on your own for bot detection.

</details>

<details>
<summary>How do I use webtool with a different Chrome profile?</summary>

Switch profiles in Chrome's profile picker. Webtool connects to your Chrome process via a single shared connection, so it works with whichever profile is active — no configuration needed.

</details>
