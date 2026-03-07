package browser

import (
	"testing"
)

func TestWithUserDataDir(t *testing.T) {
	b := New().WithUserDataDir("/custom/chrome/data")

	if b.UserDataDir != "/custom/chrome/data" {
		t.Errorf("got UserDataDir %q, want %q", b.UserDataDir, "/custom/chrome/data")
	}
}

func TestWithURL(t *testing.T) {
	b := New().WithURL("ws://127.0.0.1:9222/devtools/browser/abc")

	if b.WSUrl != "ws://127.0.0.1:9222/devtools/browser/abc" {
		t.Errorf("got WSUrl %q, want %q", b.WSUrl, "ws://127.0.0.1:9222/devtools/browser/abc")
	}
}

func TestChaining(t *testing.T) {
	b := New().WithUserDataDir("/data").WithURL("ws://localhost:1234")

	if b.UserDataDir != "/data" {
		t.Errorf("got UserDataDir %q, want %q", b.UserDataDir, "/data")
	}
	if b.WSUrl != "ws://localhost:1234" {
		t.Errorf("got WSUrl %q, want %q", b.WSUrl, "ws://localhost:1234")
	}
}
