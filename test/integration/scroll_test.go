//go:build integration

package integration

import (
	"context"
	"testing"
)

func TestScroll_DownAndUpChangesScrollPosition(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/dynamic"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	// Get initial scroll position.
	before, err := b.Eval(ctx, "window.scrollY")
	if err != nil {
		t.Fatalf("Eval scrollY before: %v", err)
	}

	// Scroll down by 300 pixels.
	if err := b.Scroll(ctx, 300); err != nil {
		t.Fatalf("Scroll down: %v", err)
	}

	after, err := b.Eval(ctx, "window.scrollY")
	if err != nil {
		t.Fatalf("Eval scrollY after scroll down: %v", err)
	}

	if after == before {
		t.Errorf("expected scroll position to change after scrolling down, got %s both times", before)
	}

	// Scroll back up.
	if err := b.Scroll(ctx, -300); err != nil {
		t.Fatalf("Scroll up: %v", err)
	}

	restored, err := b.Eval(ctx, "window.scrollY")
	if err != nil {
		t.Fatalf("Eval scrollY after scroll up: %v", err)
	}

	if restored != before {
		t.Errorf("expected scroll position to restore to %s after scrolling up, got %s", before, restored)
	}
}
