# webtool

**Your agent hates Playwright.**

webtool is a fast, zero-dependency CLI that drives your Chrome browser. 

Your logins, your extensions, your cookies. You don't need a cloud browser, Playwright, stealth browsers or captcha workarounds. Just let your agent control the browser you are already using. 
webtool doesn't trigger bot detection because it is driving your real browser. **It just works**.

Webtool turns web pages into highly optimized token-efficient text snapshots that LLMs can read, and provides simple commands to click, type, and navigate.

**What about security?** It can be unsettling to give an AI agent full access to your browser. webtool has a powerful security policy engine that filters requests at the network level. Lock down the agent to specific pages or limit access with fine-grained request filtering. See [docs/SECURITY.md](docs/SECURITY.md) for details.

## Install

### 1. Install with npm (recommended)

```bash
npm i -g @usewebtool/webtool
```
webtool does not require Node.js, but npm is the easiest path to a cross-platform install.

### 2. Add the skill to your agent

```bash
npx skills add machinae/webtool
```

### Alternate install methods
Install from source with Go:

```bash
go install github.com/machinae/webtool@latest
```

Or download a binary from the [releases page](https://github.com/machinae/webtool/releases) and put it somewhere in your PATH.

To manually install the skill, clone the repo and copy `skills/webtool` to your agent's skills directory. For example, to install the skill into OpenClaw:

```bash
git clone https://github.com/machinae/webtool.git
cp -r webtool/skills/webtool ~/.openclaw/skills/
```

## Setup

Make sure you're on the latest version of Chrome. You'll need to enable remote debugging so webtool can connect to your browser.

1. Open Chrome, navigate to `chrome://inspect/#remote-debugging` and enable remote debugging.
2. Run `webtool start` to start the webtool daemon.
3. Chrome will show a permission dialog. Click **Accept**.

That's it. The daemon keeps the connection open, so you only approve once per session.

To stop the daemon and close the connection run `webtool stop`.

## Agent Usage

Once you have installed the agent skill, just ask your agent to do things online.
```
"Open my Gmail and archive invitation emails."
```

The LLM is the brain. webtool gives your agent the "hands" it needs to interact with the web. Works in Codex, Claude Code, OpenClaw, or any other agent that supports skills.

## CLI Usage

You can use webtool from the command line or in shell scripts to programmatically control Chrome.

```bash
webtool start                       # connect to Chrome
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

See [docs/usage.md](docs/usage.md) for the full command reference.

## Security Policy

You can create a simple YAML security policy file to filter your agent's Chrome traffic at the network level.
Create a policy file:

```yaml
# Read-only mode: block non-idempotent methods
deny:
  - method: "POST|PUT|DELETE|PATCH"
```

Or block all requests to specific sites. Wrap URLs in `*` wildcards to match all pages on the domain.

```yaml
deny:
  - url: "*mail.google.com*"
  - url: "*bank.example.com*"
  - url: "*admin.example.com*"
```

Then start the daemon with it:

```bash
webtool start -p policy.yml
```

See [docs/SECURITY.md](docs/SECURITY.md) for the full policy format.

## Snapshots

Instead of feeding raw HTML or screenshots to your agent, webtool generates compact semantic snapshots. A page that would be 50k+ tokens as HTML becomes a few hundred tokens. The snapshots go beyond simple accessibility trees and create a structured map that is naturally easy for an LLM to understand.

Generate a snapshot of the current tab:
```bash
webtool open https://mail.google.com
webtool snapshot
```

```
[url] https://mail.google.com/mail/u/0/#inbox
[title] Inbox - Gmail

[1] navigation "Main"
  [2] link "Inbox" url="/mail/u/0/#inbox"
  [3] link "Starred" url="/mail/u/0/#starred"
  [4] link "Sent" url="/mail/u/0/#sent"
[10] list "Messages"
  [11] listitem "Alice Chen | Meeting tomorrow"
    [12] checkbox "Select"
    [13] link "Meeting tomorrow - Hey, are we still on for..."
  [14] listitem "GitHub | New issue assigned"
    [15] checkbox "Select"
    [16] link "New issue assigned - You've been assigned #421..."
```

## FAQ

See [docs/FAQ.md](docs/FAQ.md).
