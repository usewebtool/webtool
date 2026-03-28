//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usewebtool/webtool/browser"
)

func TestUpload_FileInputShowsFilename(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	// Create a temp file to upload.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	if err := b.Open(ctx, pageURL("/controlled"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	if _, err := b.Upload(ctx, "#attachment", []string{tmpFile}); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	snap, err := b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	text := snap.String()
	if !strings.Contains(text, "testfile.txt") {
		t.Errorf("expected filename in snapshot after upload, got:\n%s", text)
	}
}
