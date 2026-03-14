package browser

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// TabInfo is a display-only snapshot of a browser tab, returned by Tabs().
type TabInfo struct {
	// Index is the 1-based position in the tab list.
	Index int `json:"index"`
	// Title is the page title.
	Title string `json:"title"`
	// URL is the page URL.
	URL string `json:"url"`
	// TargetID is the CDP target ID for reference.
	TargetID string `json:"targetID"`
}

// tab is a live browser tab the agent is controlling.
// Tracked in Browser.tabs. Holds the rod page and per-tab state.
type tab struct {
	targetID     string
	page         *rod.Page
	hijackRouter *rod.HijackRouter
}

// Open navigates the active page to the given URL.
// If a TargetID is saved from a previous command, it reuses that page.
// Otherwise it picks the first available page.
func (b *Browser) Open(ctx context.Context, url string) error {
	if err := b.Connect(); err != nil {
		return err
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
		tabs = append(tabs, TabInfo{
			Index:    len(tabs) + 1,
			Title:    info.Title,
			URL:      info.URL,
			TargetID: string(info.TargetID),
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

	b.getOrCreateTab(page)
	return nil
}

// activeTab returns the current active tab, or finds a suitable one.
// This method intentionally does NOT accept a context.Context parameter.
// Rod's .Context(ctx) returns a shallow copy, but Page objects returned by
// the clone inherit the request-scoped context. When the request ends, those
// pages carry a cancelled context, breaking subsequent operations. Callers
// should apply context to the returned page directly: tab.page.Context(ctx).
// See commit 408af4f.
func (b *Browser) activeTab() (*tab, error) {
	if b.active != nil {
		// Verify the session is still alive — the page may be a zombie
		// whose CDP session was destroyed (e.g. tab closed).
		if _, err := b.active.page.Info(); err == nil {
			b.ensureHijacked(b.active)
			return b.active, nil
		}
		// Stale tab — remove from map and fall through.
		delete(b.tabs, b.active.targetID)
		b.active = nil
	}

	pages, err := b.rod.Pages()
	if err != nil {
		return nil, fmt.Errorf("listing pages: %w", err)
	}

	// Pick the last user-visible page, skipping internal Chrome targets
	// (omnibox popup, devtools, etc.). The last page is typically the
	// most recently active tab.
	for i := len(pages) - 1; i >= 0; i-- {
		info, err := pages[i].Info()
		if err != nil {
			continue
		}
		if info.Type == proto.TargetTargetInfoTypePage && !strings.HasPrefix(info.URL, "chrome://") {
			return b.getOrCreateTab(pages[i]), nil
		}
	}

	// No user page found — fall back to the last page.
	if len(pages) == 0 {
		return nil, fmt.Errorf("no open pages found")
	}
	return b.getOrCreateTab(pages[len(pages)-1]), nil
}

// getOrCreateTab looks up or creates a tab for the given rod page and sets it as active.
func (b *Browser) getOrCreateTab(page *rod.Page) *tab {
	targetID := string(page.TargetID)
	t, ok := b.tabs[targetID]
	if !ok {
		t = &tab{targetID: targetID, page: page}
		b.tabs[targetID] = t
	}
	b.active = t
	b.ensureHijacked(t)
	return t
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
