//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/usewebtool/webtool/browser"
)

func TestShadowDOM_SnapshotAndClick(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/dynamic"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	// Question 1: Does the snapshot include shadow DOM content?
	snap, err := b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	text := snap.String()
	t.Logf("Snapshot:\n%s", text)

	if !strings.Contains(text, "Shadow Heading") {
		t.Errorf("snapshot does not contain shadow DOM heading")
	}
	if !strings.Contains(text, "Shadow Action") {
		t.Errorf("snapshot does not contain shadow DOM button")
	}

	// Question 2: Can we click a shadow DOM element by backendNodeId?
	btnID := findElement(t, text, "Shadow Action")

	if _, err := b.Click(ctx, btnID); err != nil {
		t.Fatalf("Click shadow button by backendNodeId: %v", err)
	}

	// Question 3: Did the click actually work inside the shadow root?
	snap, err = b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot after click: %v", err)
	}

	if !strings.Contains(snap.String(), "Shadow clicked") {
		t.Errorf("expected 'Shadow clicked' after clicking shadow button, got:\n%s", snap.String())
	}
}
