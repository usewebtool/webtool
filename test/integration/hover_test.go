//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/usewebtool/webtool/browser"
)

func TestHover_RevealsHiddenActions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/dynamic"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	// Before hover, the action buttons should not be visible.
	snap, err := b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot before hover: %v", err)
	}
	if strings.Contains(snap.String(), "Archive") {
		t.Fatalf("expected Archive button to be hidden before hover, got:\n%s", snap.String())
	}

	// Hover over the card to reveal actions.
	if _, err := b.Hover(ctx, "#project-card"); err != nil {
		t.Fatalf("Hover: %v", err)
	}

	snap, err = b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot after hover: %v", err)
	}

	text := snap.String()
	if !strings.Contains(text, "Archive") {
		t.Errorf("expected Archive button visible after hover, got:\n%s", text)
	}
	if !strings.Contains(text, "Delete") {
		t.Errorf("expected Delete button visible after hover, got:\n%s", text)
	}
}
