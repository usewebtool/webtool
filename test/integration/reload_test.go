//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/usewebtool/webtool/browser"
)

func TestReload_RestoresOriginalContent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/dynamic"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	// Increment the counter to change the page state.
	if _, err := b.Click(ctx, "#increment-btn"); err != nil {
		t.Fatalf("Click increment: %v", err)
	}

	// Verify increment took effect using Eval to read the exact counter value.
	result, err := b.Eval(ctx, `document.getElementById("counter-value").textContent`)
	if err != nil {
		t.Fatalf("Eval counter after increment: %v", err)
	}
	if result != "1" {
		t.Fatalf("expected counter to be '1' after increment, got %q", result)
	}

	// Reload should reset counter back to 0.
	if err := b.Reload(ctx); err != nil {
		t.Fatalf("Reload: %v", err)
	}

	snap, err := b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot after reload: %v", err)
	}

	text := snap.String()
	if !strings.Contains(text, "Dashboard") {
		t.Errorf("expected page content after reload, got:\n%s", text)
	}

	// Verify the counter reset by checking the counter value element.
	result, err = b.Eval(ctx, `document.getElementById("counter-value").textContent`)
	if err != nil {
		t.Fatalf("Eval counter after reload: %v", err)
	}
	if result != "0" {
		t.Errorf("expected counter to reset to 0 after reload, got %q", result)
	}
}
