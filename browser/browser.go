package browser

import (
	"os"
	"path/filepath"

	"github.com/go-rod/rod"
)

// Browser holds a connection to a running Chrome instance.
type Browser struct {
	rod           *rod.Browser
	// DataDir is the directory where webtool persists state (e.g. state.json).
	DataDir string `json:"-"`
	// WSUrl is the Chrome DevTools WebSocket URL used for CDP communication.
	WSUrl string `json:"ws_url"`
	// ChromeDataDir is the Chrome user data directory used for DevToolsActivePort discovery.
	ChromeDataDir string `json:"chrome_data_dir"`
}

// New creates a new Browser with default settings.
func New() *Browser {
	home, _ := os.UserHomeDir()
	return &Browser{
		DataDir: filepath.Join(home, ".webtool"),
	}
}

// WithDataDir sets the directory where webtool persists state (e.g. state.json).
func (b *Browser) WithDataDir(dir string) *Browser {
	b.DataDir = dir
	return b
}

// WithChromeDataDir sets the Chrome user data directory for DevToolsActivePort discovery.
func (b *Browser) WithChromeDataDir(dir string) *Browser {
	b.ChromeDataDir = dir
	return b
}

// WithURL sets an explicit debugging WebSocket URL, skipping DevToolsActivePort discovery.
func (b *Browser) WithURL(wsURL string) *Browser {
	b.WSUrl = wsURL
	return b
}
