package browser

import (
	"github.com/go-rod/rod"
	"github.com/machinae/webtool/policy"
)

// Browser holds a connection to a running Chrome instance.
type Browser struct {
	rod *rod.Browser
	// WSUrl is the Chrome DevTools WebSocket URL used for CDP communication.
	WSUrl string
	// ChromeDataDir is the Chrome user data directory used for DevToolsActivePort discovery.
	ChromeDataDir string
	// active is the current agent-controlled tab.
	active *tab
	// tabs tracks all agent-touched tabs by target ID.
	tabs map[string]*tab
	// policy is the security policy for request interception. Nil means no policy.
	policy *policy.Policy
}

// New creates a new Browser with default settings.
func New() *Browser {
	return &Browser{
		tabs: make(map[string]*tab),
	}
}

// WithChromeDataDir sets the Chrome user data directory for DevToolsActivePort discovery.
func (b *Browser) WithChromeDataDir(dir string) *Browser {
	b.ChromeDataDir = dir
	return b
}

// WithPolicy sets the security policy for request interception.
func (b *Browser) WithPolicy(p *policy.Policy) *Browser {
	b.policy = p
	return b
}

// WithURL sets an explicit debugging WebSocket URL, skipping DevToolsActivePort discovery.
func (b *Browser) WithURL(wsURL string) *Browser {
	b.WSUrl = wsURL
	return b
}
