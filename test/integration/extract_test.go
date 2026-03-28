//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
)

func TestExtract_FullPageMarkdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/extract"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	md, err := b.Extract(ctx, "", false)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	if !strings.Contains(md, "Main Heading") {
		t.Errorf("expected heading in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "main content paragraph") {
		t.Errorf("expected paragraph text in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "Example Link") {
		t.Errorf("expected link text in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "Footer content") {
		t.Errorf("expected footer in full page extract, got:\n%s", md)
	}
}

func TestExtract_ScopedToMain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/extract"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	md, err := b.Extract(ctx, "main, [role='main']", false)
	if err != nil {
		t.Fatalf("Extract --main: %v", err)
	}

	if !strings.Contains(md, "Main Heading") {
		t.Errorf("expected heading in main extract, got:\n%s", md)
	}
	if !strings.Contains(md, "main content paragraph") {
		t.Errorf("expected paragraph in main extract, got:\n%s", md)
	}
	if strings.Contains(md, "Footer content") {
		t.Errorf("main extract should not contain footer, got:\n%s", md)
	}
}

func TestExtract_RawHTML(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/extract"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	html, err := b.Extract(ctx, "", true)
	if err != nil {
		t.Fatalf("Extract --html: %v", err)
	}

	if !strings.Contains(html, "<h1>") {
		t.Errorf("expected HTML tags in raw extract, got:\n%s", html)
	}
	if !strings.Contains(html, "Main Heading") {
		t.Errorf("expected heading text in raw extract, got:\n%s", html)
	}
}

func TestExtract_ScopedToElement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := b.Open(ctx, pageURL("/extract"), false); err != nil {
		t.Fatalf("Open: %v", err)
	}

	md, err := b.Extract(ctx, "ul", false)
	if err != nil {
		t.Fatalf("Extract scoped to ul: %v", err)
	}

	if !strings.Contains(md, "Item one") || !strings.Contains(md, "Item two") {
		t.Errorf("expected list items in scoped extract, got:\n%s", md)
	}
	if strings.Contains(md, "Main Heading") {
		t.Errorf("scoped extract should not contain heading, got:\n%s", md)
	}
}
