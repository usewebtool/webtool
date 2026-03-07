package browser

import "github.com/go-rod/rod"

// Browser holds a connection to a running Chrome instance.
type Browser struct {
	rod         *rod.Browser
	WSUrl       string `json:"ws_url"`
	UserDataDir string `json:"user_data_dir"`
}

// New creates a new Browser with default settings.
func New() *Browser {
	return &Browser{}
}

// WithUserDataDir sets the Chrome user data directory for DevToolsActivePort discovery.
func (b *Browser) WithUserDataDir(dir string) *Browser {
	b.UserDataDir = dir
	return b
}

// WithURL sets an explicit debugging WebSocket URL, skipping DevToolsActivePort discovery.
func (b *Browser) WithURL(wsURL string) *Browser {
	b.WSUrl = wsURL
	return b
}
