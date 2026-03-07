package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// discoverWSURL reads Chrome's DevToolsActivePort file to find the debugging WebSocket URL.
func discoverWSURL() (string, error) {
	dataDir, err := chromeUserDataDir()
	if err != nil {
		return "", err
	}
	return discoverWSURLFromDir(dataDir)
}

// discoverWSURLFromDir reads DevToolsActivePort from a given directory.
func discoverWSURLFromDir(dataDir string) (string, error) {
	portFile := filepath.Join(dataDir, "DevToolsActivePort")
	data, err := os.ReadFile(portFile)
	if err != nil {
		return "", fmt.Errorf("could not read DevToolsActivePort: %w\nEnable remote debugging at chrome://inspect#remote-debugging", err)
	}

	lines := strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)
	if len(lines) < 2 {
		return "", fmt.Errorf("unexpected DevToolsActivePort format")
	}

	port := strings.TrimSpace(lines[0])
	path := strings.TrimSpace(lines[1])
	return fmt.Sprintf("ws://127.0.0.1:%s%s", port, path), nil
}

// chromeUserDataDir returns the default Chrome user data directory for the current OS.
// Override with WEBTOOL_CHROME_USER_DATA_DIR env var.
func chromeUserDataDir() (string, error) {
	if dir := os.Getenv("WEBTOOL_CHROME_USER_DATA_DIR"); dir != "" {
		return dir, nil
	}

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
