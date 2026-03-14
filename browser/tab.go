package browser

import (
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
	// Active is true if this is the tab webtool will act on.
	Active bool `json:"active"`
}

// tab is a live browser tab the agent is controlling.
// Tracked in Browser.tabs. Holds the rod page and per-tab state.
type tab struct {
	targetID     string
	page         *rod.Page
	hijackRouter *rod.HijackRouter

	// asyncErr receives errors from async operations (e.g. blocked requests).
	// Buffer size 1 — only the first error is kept.
	asyncErr chan error
}

// Err returns and drains the first async error, or nil if none.
// The error is removed from the channel — subsequent calls return nil
// until a new error is sent.
func (t *tab) Err() error {
	select {
	case err := <-t.asyncErr:
		return err
	default:
		return nil
	}
}

// sendErr sends an error to the async error channel without blocking.
// If an error is already buffered, the new error is dropped.
func (t *tab) sendErr(err error) {
	select {
	case t.asyncErr <- err:
	default:
	}
}

// activeTab connects to Chrome (if needed) and returns the current active tab.
// This method intentionally does NOT accept a context.Context parameter.
// Rod's .Context(ctx) returns a shallow copy, but Page objects returned by
// the clone inherit the request-scoped context. When the request ends, those
// pages carry a cancelled context, breaking subsequent operations. Callers
// should apply context to the returned page directly: tab.page.Context(ctx).
// See commit 408af4f.
func (b *Browser) activeTab() (*tab, error) {
	if err := b.Connect(); err != nil {
		return nil, err
	}
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
		t = &tab{targetID: targetID, page: page, asyncErr: make(chan error, 1)}
		b.tabs[targetID] = t
	}
	b.active = t
	b.ensureHijacked(t)
	return t
}
