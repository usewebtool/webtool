package browser

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Tab represents an open browser tab.
type Tab struct {
	// Index is the 1-based position in the tab list.
	Index int `json:"index"`
	// Title is the page title.
	Title string `json:"title"`
	// URL is the page URL.
	URL string `json:"url"`
}

// Open navigates the active page to the given URL.
// If a TargetID is saved from a previous command, it reuses that page.
// Otherwise it picks the first available page.
func (b *Browser) Open(ctx context.Context, url string) error {
	if err := b.Connect(); err != nil {
		return err
	}

	page, err := b.activePage()
	if err != nil {
		return err
	}

	if err := waitPageLoad(ctx, page, func() error {
		if err := page.Context(ctx).Navigate(url); err != nil {
			return fmt.Errorf("navigating to %s: %w", url, err)
		}
		return nil
	}); err != nil {
		return err
	}

	b.TargetID = string(page.TargetID)
	return nil
}

// Tabs returns all open browser tabs, filtering out DevTools and other non-page targets.
func (b *Browser) Tabs(ctx context.Context) ([]Tab, error) {
	if err := b.Connect(); err != nil {
		return nil, err
	}

	all, err := b.rod.Pages()
	if err != nil {
		return nil, fmt.Errorf("listing pages: %w", err)
	}

	tabs := make([]Tab, 0, len(all))
	for _, p := range all {
		info, err := p.Context(ctx).Info()
		if err != nil {
			return nil, fmt.Errorf("getting page info: %w", err)
		}
		if !isUserTab(info) {
			continue
		}
		tabs = append(tabs, Tab{
			Index: len(tabs) + 1,
			Title: info.Title,
			URL:   info.URL,
		})
	}

	return tabs, nil
}

// Switch activates the tab at the given 1-based index and sets it as the active page.
func (b *Browser) Switch(ctx context.Context, index int) error {
	if err := b.Connect(); err != nil {
		return err
	}

	pages, err := b.pageTargets(ctx)
	if err != nil {
		return err
	}

	if index < 1 || index > len(pages) {
		return fmt.Errorf("tab %d out of range (have %d tabs)", index, len(pages))
	}

	page := pages[index-1]
	if _, err := page.Context(ctx).Activate(); err != nil {
		return fmt.Errorf("activating tab: %w", err)
	}

	b.TargetID = string(page.TargetID)
	return nil
}

// activePage returns the page for the saved TargetID, or the first available page.
// This method intentionally does NOT accept a context.Context parameter.
// Rod's .Context(ctx) returns a shallow copy, but Page objects returned by
// the clone inherit the request-scoped context. When the request ends, those
// pages carry a cancelled context, breaking subsequent operations. Callers
// should apply context to the returned page directly: page.Context(ctx).
// See commit 408af4f.
func (b *Browser) activePage() (*rod.Page, error) {
	if b.TargetID != "" {
		page, err := b.rod.PageFromTarget(proto.TargetTargetID(b.TargetID))
		if err == nil {
			return page, nil
		}
		// Stale target — fall through to first page.
	}

	pages, err := b.rod.Pages()
	if err != nil {
		return nil, fmt.Errorf("listing pages: %w", err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no open pages found")
	}

	return pages[0], nil
}

// pageTargets returns only "page" type targets, filtering out DevTools, extensions, etc.
func (b *Browser) pageTargets(ctx context.Context) ([]*rod.Page, error) {
	all, err := b.rod.Pages()
	if err != nil {
		return nil, fmt.Errorf("listing pages: %w", err)
	}

	pages := make([]*rod.Page, 0, len(all))
	for _, p := range all {
		info, err := p.Context(ctx).Info()
		if err != nil {
			return nil, fmt.Errorf("getting page info: %w", err)
		}
		if isUserTab(info) {
			pages = append(pages, p)
		}
	}

	return pages, nil
}

// isUserTab returns true if the target is a user-visible tab (web pages and extension pages).
// Excludes DevTools, about:, and chrome:// internal pages.
func isUserTab(info *proto.TargetTargetInfo) bool {
	if info.Type != proto.TargetTargetInfoTypePage {
		return false
	}
	if strings.HasPrefix(info.URL, "devtools://") ||
		strings.HasPrefix(info.URL, "about:") ||
		strings.HasPrefix(info.URL, "chrome://") {
		return false
	}
	return true
}
