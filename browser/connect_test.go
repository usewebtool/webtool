package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverWSURL_ValidPortFile(t *testing.T) {
	dir := t.TempDir()
	portFile := filepath.Join(dir, "DevToolsActivePort")
	os.WriteFile(portFile, []byte("9222\n/devtools/browser/abc123\n"), 0644)

	// Override the discovery to use our temp dir
	wsURL, err := discoverWSURLFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "ws://127.0.0.1:9222/devtools/browser/abc123"
	if wsURL != expected {
		t.Errorf("got %q, want %q", wsURL, expected)
	}
}

func TestDiscoverWSURL_MissingPortFile(t *testing.T) {
	dir := t.TempDir()

	_, err := discoverWSURLFromDir(dir)
	if err == nil {
		t.Fatal("expected error for missing DevToolsActivePort, got nil")
	}
}

func TestDiscoverWSURL_MalformedPortFile(t *testing.T) {
	dir := t.TempDir()
	portFile := filepath.Join(dir, "DevToolsActivePort")
	os.WriteFile(portFile, []byte("9222\n"), 0644)

	_, err := discoverWSURLFromDir(dir)
	if err == nil {
		t.Fatal("expected error for malformed DevToolsActivePort, got nil")
	}
}
