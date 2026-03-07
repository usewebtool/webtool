package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDefaultDataDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	b := New()
	want := filepath.Join(home, ".webtool")
	if b.DataDir != want {
		t.Errorf("got DataDir %q, want %q", b.DataDir, want)
	}
}

func TestWithDataDir(t *testing.T) {
	b := New().WithDataDir("/custom/state")
	if b.DataDir != "/custom/state" {
		t.Errorf("got DataDir %q, want %q", b.DataDir, "/custom/state")
	}
}

func TestWithChromeDataDir(t *testing.T) {
	b := New().WithChromeDataDir("/custom/chrome/data")

	if b.ChromeDataDir != "/custom/chrome/data" {
		t.Errorf("got ChromeDataDir %q, want %q", b.ChromeDataDir, "/custom/chrome/data")
	}
}

func TestWithURL(t *testing.T) {
	b := New().WithURL("ws://127.0.0.1:9222/devtools/browser/abc")

	if b.WSUrl != "ws://127.0.0.1:9222/devtools/browser/abc" {
		t.Errorf("got WSUrl %q, want %q", b.WSUrl, "ws://127.0.0.1:9222/devtools/browser/abc")
	}
}

func TestChaining(t *testing.T) {
	b := New().WithChromeDataDir("/data").WithURL("ws://localhost:1234")

	if b.ChromeDataDir != "/data" {
		t.Errorf("got ChromeDataDir %q, want %q", b.ChromeDataDir, "/data")
	}
	if b.WSUrl != "ws://localhost:1234" {
		t.Errorf("got WSUrl %q, want %q", b.WSUrl, "ws://localhost:1234")
	}
}
