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

func TestResolveWSUrl_AlreadySet(t *testing.T) {
	b := New().WithURL("ws://127.0.0.1:9222/devtools/browser/abc")

	err := b.resolveWSUrl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.WSUrl != "ws://127.0.0.1:9222/devtools/browser/abc" {
		t.Errorf("WSUrl changed unexpectedly to %q", b.WSUrl)
	}
}

func TestResolveWSUrl_WithURLSkipsDiscovery(t *testing.T) {
	b := New().WithURL("ws://127.0.0.1:9222/devtools/browser/abc").WithChromeDataDir("/nonexistent")

	err := b.resolveWSUrl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ChromeDataDir != "/nonexistent" {
		t.Errorf("got ChromeDataDir %q, want %q", b.ChromeDataDir, "/nonexistent")
	}
}

func TestResolveWSUrl_DiscoversFromDataDir(t *testing.T) {
	dir := t.TempDir()
	portFile := filepath.Join(dir, "DevToolsActivePort")
	os.WriteFile(portFile, []byte("9222\n/devtools/browser/abc123\n"), 0644)

	b := New().WithChromeDataDir(dir)

	err := b.resolveWSUrl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.WSUrl != "ws://127.0.0.1:9222/devtools/browser/abc123" {
		t.Errorf("got WSUrl %q, want %q", b.WSUrl, "ws://127.0.0.1:9222/devtools/browser/abc123")
	}
}

func TestResolveWSUrl_MissingPortFileErrors(t *testing.T) {
	dir := t.TempDir()
	b := New().WithChromeDataDir(dir)

	err := b.resolveWSUrl()
	if err == nil {
		t.Fatal("expected error for missing DevToolsActivePort, got nil")
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"bare port", "9222", "ws://127.0.0.1:9222"},
		{"colon port", ":9222", "ws://127.0.0.1:9222"},
		{"host and port", "localhost:9222", "ws://localhost:9222"},
		{"ws scheme", "ws://127.0.0.1:9222/devtools/browser/abc", "ws://127.0.0.1:9222/devtools/browser/abc"},
		{"http to ws", "http://127.0.0.1:9222/devtools/browser/abc", "ws://127.0.0.1:9222/devtools/browser/abc"},
		{"https to wss", "https://host:9222/path", "wss://host:9222/path"},
		{"empty defaults to 9222", "", "ws://127.0.0.1:9222"},
		{"whitespace trimmed", "  9222  ", "ws://127.0.0.1:9222"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeURL(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("normalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
