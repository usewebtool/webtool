# FAQ

## Does webtool support headless mode?

Webtool doesn't launch Chrome — it connects to your already-running instance. If you launch Chrome in headless mode yourself, webtool will connect to it just fine.

What webtool intentionally avoids is *managing* a headless Chrome instance for you. Why? Because connecting to your real Chrome session is fundamentally better for agent workflows:

- **No bot detection.** Automated Chrome instances (launched with `--remote-debugging-port` or via Playwright/Puppeteer) set `navigator.webdriver=true`, which triggers CAPTCHAs and blocks on most websites. Your normal Chrome session doesn't have this flag — webtool inherits that advantage by connecting to it rather than launching its own instance.
- **Real logins.** Your Chrome already has your cookies, sessions, and saved passwords. No need to automate login flows or manage auth tokens.
- **You can see what's happening.** When an agent controls your browser, you watch it work in real time. This builds trust and makes debugging trivial.

If you need headless for CI or server environments, you can launch Chrome yourself with `chrome --headless --remote-debugging-port=0` and webtool will connect to it — but you'll be on your own for bot detection.

## How do I use webtool with a different Chrome profile?

Switch profiles in Chrome's profile picker. Webtool connects to your Chrome process via a single shared connection, so it works with whichever profile is active — no configuration needed.

## What version of Chrome do I need?

Chrome 144 or newer. That's when Chrome added dynamic remote debugging — the feature that lets webtool connect to your browser without restarting it with special flags. If you're on an older version, just update Chrome. You can check your version at `chrome://version`.

## How is this different from Playwright/Puppeteer?

Playwright and Puppeteer launch a separate browser instance. That instance has no cookies, no logins, no extensions, and sets `navigator.webdriver=true` — which gets you blocked on most websites.

webtool connects to the Chrome you're already using. Your sessions, your cookies, your extensions — all there. No `navigator.webdriver` flag, no bot detection. It's also a single Go binary with no Node.js runtime, no browser downloads, and no dependency tree.

The tradeoff: webtool is designed for agent-driven automation and scraping, not browser testing. If you need cross-browser test suites, Playwright is the right tool.

## Can I use webtool without an AI agent?

Yes. It's a regular CLI tool. You can use it interactively, in shell scripts, or pipe commands together. The agent skill is just a layer on top that teaches LLMs how to use the CLI — underneath it's all `webtool open`, `webtool click`, `webtool extract`.
