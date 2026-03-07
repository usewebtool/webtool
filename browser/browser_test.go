package browser

import (
	"testing"
)

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
