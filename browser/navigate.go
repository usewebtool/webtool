package browser

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Open navigates the active page to the given URL.
// If newTab is true, a new tab is created instead of navigating the current one.
func (b *Browser) Open(ctx context.Context, url string, newTab bool) error {
	if newTab {
		return b.openNewTab(ctx, url)
	}

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

// openNewTab creates a new tab, navigates it to the URL, and sets it as active.
func (b *Browser) openNewTab(ctx context.Context, url string) error {
	if err := b.Connect(); err != nil {
		return err
	}

	page, err := b.rod.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return fmt.Errorf("creating new tab: %w", err)
	}

	page = page.Context(ctx)

	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("waiting for page load: %w", err)
	}

	if _, err := page.Activate(); err != nil {
		return fmt.Errorf("activating new tab: %w", err)
	}

	b.getOrCreateTab(page)
	return nil
}

// Tabs returns all open browser tabs, filtering out DevTools and other non-page targets.
// Uses pageTargets() for enumeration so tab indices are consistent with Switch().
func (b *Browser) Tabs(ctx context.Context) ([]TabInfo, error) {
	if err := b.Connect(); err != nil {
		return nil, err
	}

	pages, err := b.pageTargets(ctx)
	if err != nil {
		return nil, err
	}

	tabs := make([]TabInfo, 0, len(pages))
	for _, p := range pages {
		info, err := p.Context(ctx).Info()
		if err != nil {
			return nil, fmt.Errorf("getting page info: %w", err)
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

// pageTargets returns only "page" type targets, filtered and sorted by TargetID.
// Sorting by TargetID ensures deterministic tab indices across calls — without this,
// CDP can return targets in different orders, causing the index from `tabs` to point
// at a different tab in a subsequent `tab N` call.
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

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].TargetID < pages[j].TargetID
	})

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
