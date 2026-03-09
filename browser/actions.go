package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

// stableQuietPeriod is the duration an element's position and size must remain
// unchanged before it is considered stable. This ensures animations, layout
// shifts, and transitions have settled before acting on the element.
const stableQuietPeriod = 500 * time.Millisecond

// Click finds an element by selector and clicks it. The element is resolved
// via resolveElement (backendNodeId, XPath, or CSS). Rod's built-in
// actionability checks (scroll into view, hover, wait interactable, wait
// enabled) run before the click. WaitStable is called first to handle
// animations.
func (b *Browser) Click(ctx context.Context, selector string) error {
	if err := b.Connect(); err != nil {
		return err
	}

	page, err := b.activePage()
	if err != nil {
		return err
	}

	el, err := resolveElement(ctx, page, selector)
	if err != nil {
		return fmt.Errorf("resolving element %q: %w", selector, err)
	}

	el = el.Context(ctx)

	if err := el.WaitStable(stableQuietPeriod); err != nil {
		return fmt.Errorf("waiting for element stability: %w", err)
	}

	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking element: %w", err)
	}

	return nil
}

// Type finds an element by selector and types text into it. Uses Rod's
// Element.Input which calls CDP Input.insertText — a single CDP call that
// inserts the entire string at once (like a paste), rather than dispatching
// individual keystrokes. This avoids the bot-detection fingerprint of
// uniform-timed synthetic keystrokes, while still firing isTrusted input
// events compatible with React/Vue controlled inputs.
//
// Existing text is selected first so the new text replaces it, matching
// human behavior (select all → type overwrites).
func (b *Browser) Type(ctx context.Context, selector string, text string) error {
	if err := b.Connect(); err != nil {
		return err
	}

	page, err := b.activePage()
	if err != nil {
		return err
	}

	el, err := resolveElement(ctx, page, selector)
	if err != nil {
		return fmt.Errorf("resolving element %q: %w", selector, err)
	}

	el = el.Context(ctx)

	if err := el.WaitStable(stableQuietPeriod); err != nil {
		return fmt.Errorf("waiting for element stability: %w", err)
	}

	if err := el.SelectAllText(); err != nil {
		return fmt.Errorf("selecting existing text: %w", err)
	}

	if err := el.Input(text); err != nil {
		return fmt.Errorf("typing text: %w", err)
	}

	return nil
}
