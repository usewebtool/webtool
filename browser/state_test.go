package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	b := New().WithDataDir(dir).WithURL("ws://127.0.0.1:9222/devtools/browser/abc")
	b.ChromeDataDir = "/chrome/data"
	b.TargetID = "TARGET123"

	if err := b.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := New().WithDataDir(dir)
	if err := loaded.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.WSUrl != b.WSUrl {
		t.Errorf("WSUrl: got %q, want %q", loaded.WSUrl, b.WSUrl)
	}
	if loaded.ChromeDataDir != b.ChromeDataDir {
		t.Errorf("ChromeDataDir: got %q, want %q", loaded.ChromeDataDir, b.ChromeDataDir)
	}
	if loaded.TargetID != b.TargetID {
		t.Errorf("TargetID: got %q, want %q", loaded.TargetID, b.TargetID)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()

	b := New().WithDataDir(dir)
	if err := b.Load(); err != nil {
		t.Fatalf("Load from missing file should return nil, got: %v", err)
	}

	if b.WSUrl != "" {
		t.Errorf("WSUrl should be empty, got %q", b.WSUrl)
	}
}

func TestLoadMissingDir(t *testing.T) {
	b := New().WithDataDir("/nonexistent/path/that/does/not/exist")
	if err := b.Load(); err != nil {
		t.Fatalf("Load from missing dir should return nil, got: %v", err)
	}
}

func TestSaveCreatesDataDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")

	b := New().WithDataDir(dir)
	if err := b.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, stateFileName)); err != nil {
		t.Fatalf("state file should exist: %v", err)
	}
}

func TestDataDirNotPersisted(t *testing.T) {
	dir := t.TempDir()

	b := New().WithDataDir(dir).WithURL("ws://localhost:1234")
	b.Save()

	loaded := New().WithDataDir(dir)
	loaded.Load()

	if loaded.DataDir != dir {
		t.Errorf("DataDir should remain %q from WithDataDir, got %q", dir, loaded.DataDir)
	}
}
