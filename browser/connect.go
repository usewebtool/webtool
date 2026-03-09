package browser

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-rod/rod"
)

// Connect establishes a connection to Chrome.
// If no URL was set, it discovers Chrome via DevToolsActivePort in the user data directory.
// Safe to call multiple times — returns immediately if already connected.
func (b *Browser) Connect() error {
	if b.rod != nil {
		return nil
	}

	var err error

	if err = b.resolveWSUrl(); err != nil {
		return err
	}

	if b.WSUrl, err = normalizeURL(b.WSUrl); err != nil {
		return err
	}

	rb := rod.New().ControlURL(b.WSUrl).NoDefaultDevice()
	if err = rb.Connect(); err != nil {
		return friendlyError(err)
	}
	b.rod = rb

	return nil
}

// resolveWSUrl populates b.WSUrl if not already set.
func (b *Browser) resolveWSUrl() error {
	if b.WSUrl != "" {
		return nil
	}

	if b.ChromeDataDir == "" {
		dir, err := DefaultChromeUserDataDir()
		if err != nil {
			return err
		}
		b.ChromeDataDir = dir
	}

	wsURL, err := discoverWSURLFromDir(b.ChromeDataDir)
	if err != nil {
		return err
	}
	b.WSUrl = wsURL
	return nil
}

// Close disconnects from Chrome without closing it.
// We do not call rod.Close() because it sends Browser.close which kills Chrome.
// The WebSocket connection drops naturally when the process exits.
func (b *Browser) Close() error {
	b.rod = nil
	return nil
}

// IsConnected returns true if the browser has an active connection.
func (b *Browser) IsConnected() bool {
	return b.rod != nil
}

// RodBrowser returns the underlying rod browser instance.
func (b *Browser) RodBrowser() *rod.Browser {
	return b.rod
}

// friendlyError rewrites known connection errors into user-friendly messages.
// Returns nil for nil input, the original error if unrecognized.
func friendlyError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "403"):
		return fmt.Errorf("Chrome rejected the connection. Run webtool start and click Allow in Chrome")
	case strings.Contains(msg, "connection refused"):
		return fmt.Errorf("Chrome is not running or remote debugging is not enabled.\nStart Chrome and enable remote debugging at chrome://inspect#remote-debugging")
	default:
		return err
	}
}

// discoverWSURLFromDir reads DevToolsActivePort from a given directory.
func discoverWSURLFromDir(dataDir string) (string, error) {
	portFile := filepath.Join(dataDir, "DevToolsActivePort")
	data, err := os.ReadFile(portFile)
	if err != nil {
		return "", fmt.Errorf("could not find Chrome's DevToolsActivePort file.\nStart Chrome and enable remote debugging at chrome://inspect#remote-debugging")
	}

	lines := strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)
	if len(lines) < 2 {
		return "", fmt.Errorf("unexpected DevToolsActivePort format")
	}

	port := strings.TrimSpace(lines[0])
	path := strings.TrimSpace(lines[1])
	return fmt.Sprintf("ws://127.0.0.1:%s%s", port, path), nil
}

// DefaultChromeUserDataDir returns the default Chrome user data directory for the current OS.
func DefaultChromeUserDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Google", "Chrome"), nil
	case "linux":
		return filepath.Join(home, ".config", "google-chrome"), nil
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "Google", "Chrome", "User Data"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// Copied from go-rod/rod/lib/launcher (ResolveURL normalization, without the HTTP request).
var (
	regPort     = regexp.MustCompile(`^\:?(\d+)$`)
	regProtocol = regexp.MustCompile(`^\w+://`)
)

// normalizeURL normalizes a CDP URL into a ws:// URL that rod can connect to.
// Accepts formats: "9222", ":9222", "host:9222", "ws://host:9222/path",
// "http://host:9222/path".
func normalizeURL(u string) (string, error) {
	if u == "" {
		u = "9222"
	}

	u = strings.TrimSpace(u)
	u = regPort.ReplaceAllString(u, "127.0.0.1:$1")

	if !regProtocol.MatchString(u) {
		u = "ws://" + u
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	// Ensure ws scheme
	switch parsed.Scheme {
	case "http":
		parsed.Scheme = "ws"
	case "https":
		parsed.Scheme = "wss"
	}

	return parsed.String(), nil
}
