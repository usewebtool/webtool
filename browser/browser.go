package browser

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// Browser holds a connection to a running Chrome instance.
type Browser struct {
	rod         *rod.Browser
	WSUrl       string `json:"ws_url"`
	UserDataDir string `json:"user_data_dir"`
}

// New connects to the user's running Chrome instance via CDP.
// It discovers Chrome's debugging URL from the DevToolsActivePort file.
func New() (*Browser, error) {
	dataDir, err := chromeUserDataDir()
	if err != nil {
		return nil, err
	}

	wsURL, err := discoverWSURLFromDir(dataDir)
	if err != nil {
		return nil, err
	}

	b, err := NewFromURL(wsURL)
	if err != nil {
		return nil, err
	}
	b.UserDataDir = dataDir
	return b, nil
}

// NewFromURL connects to Chrome using an explicit debugging WebSocket URL.
func NewFromURL(wsURL string) (*Browser, error) {
	u, err := launcher.ResolveURL(wsURL)
	if err != nil {
		return nil, err
	}

	b := rod.New().ControlURL(u).NoDefaultDevice()
	if err := b.Connect(); err != nil {
		return nil, err
	}

	return &Browser{
		rod:   b,
		WSUrl: wsURL,
	}, nil
}

// Close disconnects from Chrome without closing it.
func (b *Browser) Close() error {
	return b.rod.Close()
}
