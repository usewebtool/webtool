//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/usewebtool/webtool/browser"
)

func TestWait_DelayedElementAppears(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/dynamic"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := b.Wait(ctx, "#delayed-notification"); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	snap, err := b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	text := snap.String()
	if !strings.Contains(text, "New deployment ready") {
		t.Errorf("expected delayed notification in snapshot, got:\n%s", text)
	}
}
