package browser

import (
	"context"
	"encoding/json"
	"fmt"
)

// CDP sends a raw Chrome DevTools Protocol command to the active page and
// returns the JSON result. This is a low-level escape hatch for CDP methods
// not covered by dedicated commands (e.g. Input.insertText for canvas-based
// apps like Google Docs).
func (b *Browser) CDP(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	tab, err := b.activeTab()
	if err != nil {
		return nil, err
	}
	page := tab.page

	// Unmarshal params into a generic map for rod's Call interface.
	// An empty/nil params is valid (some CDP methods take no arguments).
	var p map[string]any
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params JSON: %w", err)
		}
	}

	res, err := page.Context(ctx).Call(ctx, string(page.SessionID), method, p)
	if err != nil {
		return nil, fmt.Errorf("cdp %s: %w", method, err)
	}

	return res, nil
}
