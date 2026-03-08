package browser

import (
	"context"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Tab represents an open browser tab.
type Tab struct {
	// ID is the CDP target ID.
	ID string `json:"id"`
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

	if err := page.Context(ctx).Navigate(url); err != nil {
		return fmt.Errorf("navigating to %s: %w", url, err)
	}

	if err := page.Context(ctx).WaitLoad(); err != nil {
		return fmt.Errorf("waiting for page load: %w", err)
	}

	b.TargetID = string(page.TargetID)
	return nil
}

// Tabs returns all open browser tabs.
func (b *Browser) Tabs(ctx context.Context) ([]Tab, error) {
	if err := b.Connect(); err != nil {
		return nil, err
	}

	pages, err := b.rod.Pages()
	if err != nil {
		return nil, fmt.Errorf("listing pages: %w", err)
	}

	tabs := make([]Tab, 0, len(pages))
	for _, p := range pages {
		info, err := p.Context(ctx).Info()
		if err != nil {
			return nil, fmt.Errorf("getting page info: %w", err)
		}
		tabs = append(tabs, Tab{
			ID:    string(p.TargetID),
			Title: info.Title,
			URL:   info.URL,
		})
	}

	return tabs, nil
}

// activePage returns the page for the saved TargetID, or the first available page.
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
