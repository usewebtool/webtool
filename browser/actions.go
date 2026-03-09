package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// keyMap maps lowercase key names to Rod's input.Key constants. The CLI
// argument is lowercased before lookup, so "Enter", "enter", and "ENTER"
// all work. Key names follow the Playwright/W3C naming convention.
var keyMap = map[string]input.Key{
	"enter":      input.Enter,
	"escape":     input.Escape,
	"tab":        input.Tab,
	"backspace":  input.Backspace,
	"delete":     input.Delete,
	"space":      input.Space,
	"arrowup":    input.ArrowUp,
	"arrowdown":  input.ArrowDown,
	"arrowleft":  input.ArrowLeft,
	"arrowright": input.ArrowRight,
	"home":       input.Home,
	"end":        input.End,
	"pageup":     input.PageUp,
	"pagedown":   input.PageDown,
	"return":     input.Enter,
}

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

// Key sends a single key press to the active page. The key name is
// case-insensitive and follows Playwright/W3C naming (e.g. "Enter",
// "ArrowDown", "Escape"). This dispatches real keyDown/keyUp CDP events,
// unlike Type which uses insertText.
func (b *Browser) Key(ctx context.Context, name string) error {
	if err := b.Connect(); err != nil {
		return err
	}

	key, ok := keyMap[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("unknown key %q", name)
	}

	page, err := b.activePage()
	if err != nil {
		return err
	}

	if err := page.Context(ctx).Keyboard.Press(key); err != nil {
		return fmt.Errorf("pressing key %q: %w", name, err)
	}

	return nil
}
