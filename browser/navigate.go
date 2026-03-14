package browser

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Open navigates the active page to the given URL.
// If a TargetID is saved from a previous command, it reuses that page.
// Otherwise it picks the first available page.
func (b *Browser) Open(ctx context.Context, url string) error {
	tab, err := b.activeTab()
	if err != nil {
		return err
	}
	page := tab.page

	if err := waitPageLoad(ctx, page, func() error {
		if err := page.Context(ctx).Navigate(url); err != nil {
			return fmt.Errorf("navigating to %s: %w", url, err)
		}
		return nil
	}); err != nil {
		return err
	}

	// Bring the tab to the foreground so the user can see what the agent is doing.
	if _, err := page.Context(ctx).Activate(); err != nil {
		return fmt.Errorf("activating page: %w", err)
	}

	return nil
}

// Tabs returns all open browser tabs, filtering out DevTools and other non-page targets.
func (b *Browser) Tabs(ctx context.Context) ([]TabInfo, error) {
	if err := b.Connect(); err != nil {
		return nil, err
	}

	all, err := b.rod.Pages()
	if err != nil {
		return nil, fmt.Errorf("listing pages: %w", err)
	}

	tabs := make([]TabInfo, 0, len(all))
	for _, p := range all {
		info, err := p.Context(ctx).Info()
		if err != nil {
			return nil, fmt.Errorf("getting page info: %w", err)
		}
		if !isUserTab(info) {
			continue
		}
		t := TabInfo{
			Index:    len(tabs) + 1,
			Title:    info.Title,
			URL:      info.URL,
			TargetID: string(info.TargetID),
		}
		if b.active != nil && b.active.targetID == t.TargetID {
			t.Active = true
		}
		tabs = append(tabs, t)
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

	b.getOrCreateTab(page)
	return nil
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
