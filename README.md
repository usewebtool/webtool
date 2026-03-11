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

## Usage

```bash
webtool open https://example.com    # navigate to a URL
webtool snapshot                    # text snapshot of interactive elements
webtool click 43821                 # click an element by its ID from the snapshot
webtool type 43822 "hello world"    # type into an input field
webtool key Enter                   # press a key
webtool extract --main              # extract the main content as markdown
webtool tabs                        # list open tabs
webtool back                        # go back in history
```

The workflow is **snapshot → action → snapshot**: take a snapshot to see what's on the page, act on an element by its ID, then snapshot again to see the result.

See [docs/usage.md](docs/usage.md) for the full command reference.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBTOOL_HOME` | `~/.webtool` | Base directory for runtime files (socket, PID, logs) |
| `WEBTOOL_CHROME_DATA_DIR` | OS default | Chrome user data directory for DevToolsActivePort discovery |
