//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/usewebtool/webtool/browser"
)

func TestEval_ReturnsDocumentTitle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/simple"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	result, err := b.Eval(ctx, "document.title")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}

	if result != "Simple Test" {
		t.Errorf("expected 'Simple Test', got %q", result)
	}
}

func TestEval_Arithmetic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/simple"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	result, err := b.Eval(ctx, "1 + 1")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}

	if result != "2" {
		t.Errorf("expected '2', got %q", result)
	}
}

func TestEval_MutatesDOMVisibleInSnapshot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/simple"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	_, err := b.Eval(ctx, `document.querySelector('h1').textContent = 'Modified'`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}

	snap, err := b.Snapshot(ctx, browser.ModeAll)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	text := snap.String()
	if !strings.Contains(text, "Modified") {
		t.Errorf("expected 'Modified' in snapshot after eval, got:\n%s", text)
	}
}
