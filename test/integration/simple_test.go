//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/usewebtool/webtool/browser"
)

func TestSimple_SnapshotShowsElements(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/simple"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	snap, err := b.Snapshot(ctx, browser.ModeDefault)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	text := snap.String()
	if !strings.Contains(text, "Click me") {
		t.Errorf("expected 'Click me' button in snapshot, got:\n%s", text)
	}
	if !strings.Contains(text, "Hello") {
		t.Errorf("expected 'Hello' heading in snapshot, got:\n%s", text)
	}
}

func TestSimple_ClickUpdatesDOM(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/simple"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	snap, err := b.Snapshot(ctx, browser.ModeDefault)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	btnID := findElement(t, snap.String(), "Click me")

	if _, err := b.Click(ctx, btnID); err != nil {
		t.Fatalf("Click: %v", err)
	}

	// After clicking, the JS should have written "clicked" into the output div.
	snap, err = b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot after click: %v", err)
	}

	if !strings.Contains(snap.String(), "clicked") {
		t.Errorf("expected 'clicked' in snapshot after button click, got:\n%s", snap.String())
	}
}
