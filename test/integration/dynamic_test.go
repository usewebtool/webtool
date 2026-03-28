//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/usewebtool/webtool/browser"
)

func TestControlledForm_ClickTypeKeyAddsItem(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := b.Open(ctx, pageURL("/controlled"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	snap, err := b.Snapshot(ctx, browser.ModeDefault)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	text := snap.String()
	if !strings.Contains(text, "Task name") {
		t.Fatalf("expected Task name textbox in snapshot, got:\n%s", text)
	}

	inputID := findElement(t, text, `textbox "Task name"`)
	if _, err := b.Click(ctx, inputID); err != nil {
		t.Fatalf("Click input: %v", err)
	}

	if _, err := b.Type(ctx, inputID, "Ship integration coverage"); err != nil {
		t.Fatalf("Type: %v", err)
	}

	if err := b.Key(ctx, "Enter"); err != nil {
		t.Fatalf("Key Enter: %v", err)
	}

	snap, err = b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot after enter: %v", err)
	}

	text = snap.String()
	if !strings.Contains(text, "Added Ship integration coverage") {
		t.Errorf("expected status text after submit, got:\n%s", text)
	}
	if !strings.Contains(text, "Ship integration coverage") {
		t.Errorf("expected saved list item after submit, got:\n%s", text)
	}
}

func TestSPA_ClickBackForwardRestoresRoute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := b.Open(ctx, pageURL("/spa"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	snap, err := b.Snapshot(ctx, browser.ModeDefault)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	text := snap.String()
	if !strings.Contains(text, "Home") || !strings.Contains(text, "Route: home") {
		t.Fatalf("expected home route in initial snapshot, got:\n%s", text)
	}

	settingsID := findElement(t, text, `button "Settings"`)
	if _, err := b.Click(ctx, settingsID); err != nil {
		t.Fatalf("Click settings nav: %v", err)
	}

	snap, err = b.Snapshot(ctx, browser.ModeDefault)
	if err != nil {
		t.Fatalf("Snapshot after click: %v", err)
	}

	text = snap.String()
	if !strings.Contains(text, "Settings") || !strings.Contains(text, "Route: settings") {
		t.Fatalf("expected settings route after click, got:\n%s", text)
	}

	if err := b.Back(ctx); err != nil {
		t.Fatalf("Back: %v", err)
	}

	snap, err = b.Snapshot(ctx, browser.ModeDefault)
	if err != nil {
		t.Fatalf("Snapshot after back: %v", err)
	}

	text = snap.String()
	if !strings.Contains(text, `heading[1] "Home"`) || !strings.Contains(text, "Route: home") {
		t.Fatalf("expected home route after back, got:\n%s", text)
	}

	if err := b.Forward(ctx); err != nil {
		t.Fatalf("Forward: %v", err)
	}

	snap, err = b.Snapshot(ctx, browser.ModeDefault)
	if err != nil {
		t.Fatalf("Snapshot after forward: %v", err)
	}

	text = snap.String()
	if !strings.Contains(text, `heading[1] "Settings"`) || !strings.Contains(text, "Route: settings") {
		t.Errorf("expected settings route after forward, got:\n%s", text)
	}
}
