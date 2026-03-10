package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
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

// Select finds a <select> element by selector and selects the option matching
// the given visible text. Uses rod's built-in Element.Select which handles
// scrolling into view, waiting for visibility, and dispatching change events.
func (b *Browser) Select(ctx context.Context, selector string, value string) error {
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

	if err := el.Select([]string{value}, true, rod.SelectorTypeText); err != nil {
		return fmt.Errorf("selecting option %q: %w", value, err)
	}

	return nil
}

// Eval executes a JavaScript expression in the page and returns the result.
// Uses CDP Runtime.evaluate directly (like the Chrome console) so any
// expression works, not just function bodies.
func (b *Browser) Eval(ctx context.Context, js string) (string, error) {
	if err := b.Connect(); err != nil {
		return "", err
	}

	page, err := b.activePage()
	if err != nil {
		return "", err
	}

	result, err := proto.RuntimeEvaluate{
		Expression:   js,
		ReplMode:     true,
		AwaitPromise: true,
	}.Call(page.Context(ctx))
	if err != nil {
		return "", fmt.Errorf("evaluating JS: %w", err)
	}

	if result.ExceptionDetails != nil {
		if result.ExceptionDetails.Exception != nil && result.ExceptionDetails.Exception.Description != "" {
			return "", fmt.Errorf("JS error: %s", result.ExceptionDetails.Exception.Description)
		}
		return "", fmt.Errorf("JS error: %s", result.ExceptionDetails.Text)
	}

	r := result.Result
	if v := r.Value.String(); v != "" {
		return v, nil
	}
	if r.Description != "" {
		return r.Description, nil
	}
	return string(r.Type), nil
}

// Back navigates back in browser history and waits for the page to load.
func (b *Browser) Back(ctx context.Context) error {
	if err := b.Connect(); err != nil {
		return err
	}

	page, err := b.activePage()
	if err != nil {
		return err
	}

	if err := page.Context(ctx).NavigateBack(); err != nil {
		return fmt.Errorf("navigating back: %w", err)
	}

	if err := page.Context(ctx).WaitLoad(); err != nil {
		return fmt.Errorf("waiting for page load: %w", err)
	}

	return nil
}

// Forward navigates forward in browser history and waits for the page to load.
func (b *Browser) Forward(ctx context.Context) error {
	if err := b.Connect(); err != nil {
		return err
	}

	page, err := b.activePage()
	if err != nil {
		return err
	}

	if err := page.Context(ctx).NavigateForward(); err != nil {
		return fmt.Errorf("navigating forward: %w", err)
	}

	if err := page.Context(ctx).WaitLoad(); err != nil {
		return fmt.Errorf("waiting for page load: %w", err)
	}

	return nil
}

// Reload reloads the current page and waits for it to load.
func (b *Browser) Reload(ctx context.Context) error {
	if err := b.Connect(); err != nil {
		return err
	}

	page, err := b.activePage()
	if err != nil {
		return err
	}

	if err := page.Context(ctx).Reload(); err != nil {
		return fmt.Errorf("reloading page: %w", err)
	}

	if err := page.Context(ctx).WaitLoad(); err != nil {
		return fmt.Errorf("waiting for page load: %w", err)
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
