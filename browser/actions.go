package browser

import (
	"context"
	"errors"
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

// PageSettleTimeout is the maximum time to wait for the DOM to settle
// after a mutation action (click, type, select). If the page is still changing
// after this duration, we return success anyway — the action already happened,
// and the agent's next snapshot will reflect whatever state the page is in.
var PageSettleTimeout = 3 * time.Second

// pageSettleTick is the interval between DOM snapshot comparisons
// during post-action stabilization.
const pageSettleTick = 500 * time.Millisecond

// pageSettleDiff is the maximum percentage of DOM change (0.0–1.0)
// considered "stable." 0.01 = 1% change tolerance, which ignores noise like
// timestamps and cursor blinks while catching meaningful re-renders.
const pageSettleDiff = 0.01

// waitPageSettle waits for the DOM to stabilize after a mutation action.
// Timeout errors are silently ignored — they mean the page is still busy,
// not that the action failed.
func waitPageSettle(ctx context.Context, page *rod.Page) {
	waitCtx, cancel := context.WithTimeout(ctx, PageSettleTimeout)
	defer cancel()
	_ = page.Context(waitCtx).WaitDOMStable(pageSettleTick, pageSettleDiff)
}

// waitPageLoad subscribes to the CDP Page.frameStoppedLoading event before
// running the navigation action, then blocks until the event fires. This
// uses a pure CDP event — no JavaScript injection — so it works reliably
// even when the page's execution context is destroyed during navigation.
// After the load event, it waits for the DOM to settle via WaitDOMStable.
func waitPageLoad(ctx context.Context, page *rod.Page, action func() error) error {
	wait := page.Context(ctx).WaitNavigation(proto.PageLifecycleEventNameLoad)
	if err := action(); err != nil {
		return err
	}
	wait()
	waitPageSettle(ctx, page)
	return nil
}

// Click finds an element by selector and clicks it. The element is resolved
// via resolveElement (backendNodeId, XPath, or CSS). Actionability checks run
// before the click: WaitStable (animations settled), Disabled (not disabled),
// and Interactable (visible, not obscured, pointer-events ok).
func (b *Browser) Click(ctx context.Context, selector string) (*Element, error) {
	tab, err := b.activeTab()
	if err != nil {
		return nil, err
	}
	page := tab.page

	el, err := resolveElement(ctx, page, selector)
	if err != nil {
		return nil, err
	}

	el = el.Context(ctx)

	if err := el.Element().WaitStable(stableQuietPeriod); err != nil {
		return nil, &ErrNotStable{Sel: selector}
	}

	disabled, err := el.Element().Disabled()
	if err != nil {
		return nil, fmt.Errorf("checking disabled state: %w", err)
	}
	if disabled {
		return nil, &ErrNotEnabled{Sel: selector}
	}

	if err := el.Element().ScrollIntoView(); err != nil {
		return nil, fmt.Errorf("scrolling element into view: %w", err)
	}

	pt, err := el.Element().Interactable()
	if err != nil {
		return nil, translateInteractableErr(err, selector)
	}

	if err := page.Mouse.MoveTo(*pt); err != nil {
		return nil, fmt.Errorf("moving mouse to element: %w", err)
	}

	if err := page.Mouse.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return nil, fmt.Errorf("clicking element: %w", err)
	}

	waitPageSettle(ctx, page)
	return el, nil
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
func (b *Browser) Type(ctx context.Context, selector string, text string) (*Element, error) {
	tab, err := b.activeTab()
	if err != nil {
		return nil, err
	}
	page := tab.page

	el, err := resolveElement(ctx, page, selector)
	if err != nil {
		return nil, err
	}

	el = el.Context(ctx)

	if err := el.Element().WaitStable(stableQuietPeriod); err != nil {
		return nil, &ErrNotStable{Sel: selector}
	}

	disabled, err := el.Element().Disabled()
	if err != nil {
		return nil, fmt.Errorf("checking disabled state: %w", err)
	}
	if disabled {
		return nil, &ErrNotEnabled{Sel: selector}
	}

	if err := el.Element().ScrollIntoView(); err != nil {
		return nil, fmt.Errorf("scrolling element into view: %w", err)
	}

	if _, err := el.Element().Interactable(); err != nil {
		return nil, translateInteractableErr(err, selector)
	}

	// Select existing text so new text replaces it.
	// Fails silently on contenteditable elements (no .select() method).
	_ = el.Element().SelectAllText()

	if err := el.Element().Input(text); err != nil {
		return nil, fmt.Errorf("typing text: %w", err)
	}

	waitPageSettle(ctx, page)
	return el, nil
}

// Select finds a <select> element by selector and selects the option matching
// the given visible text. Uses rod's built-in Element.Select which handles
// scrolling into view, waiting for visibility, and dispatching change events.
func (b *Browser) Select(ctx context.Context, selector string, value string) (*Element, error) {
	tab, err := b.activeTab()
	if err != nil {
		return nil, err
	}
	page := tab.page

	el, err := resolveElement(ctx, page, selector)
	if err != nil {
		return nil, err
	}

	el = el.Context(ctx)

	if err := el.Element().Select([]string{value}, true, rod.SelectorTypeText); err != nil {
		var notFound *rod.ElementNotFoundError
		if errors.As(err, &notFound) {
			return nil, &ErrOptionNotFound{Sel: selector, Value: value}
		}
		return nil, fmt.Errorf("selecting option %q: %w", value, err)
	}

	waitPageSettle(ctx, page)
	return el, nil
}

// Eval executes a JavaScript expression in the page and returns the result.
// Uses CDP Runtime.evaluate directly (like the Chrome console) so any
// expression works, not just function bodies.
func (b *Browser) Eval(ctx context.Context, js string) (string, error) {
	tab, err := b.activeTab()
	if err != nil {
		return "", err
	}
	page := tab.page

	// Rod expects a function definition — it calls .apply() on the expression.
	// Wrap in an async arrow function so arbitrary expressions work and
	// top-level await is supported. Rod handles context cancellation and
	// promise awaiting natively via ByPromise().
	wrapped := fmt.Sprintf("async () => { return (%s); }", js)
	result, err := page.Context(ctx).Eval(wrapped)
	if err != nil {
		return "", fmt.Errorf("evaluating JS: %w", err)
	}

	return result.Value.String(), nil
}

// Back navigates back in browser history and waits for the DOM to settle.
// Uses waitPageSettle instead of waitPageLoad because SPA routers handle
// back navigation via popstate events without a full page load.
func (b *Browser) Back(ctx context.Context) error {
	tab, err := b.activeTab()
	if err != nil {
		return err
	}
	page := tab.page

	if err := page.Context(ctx).NavigateBack(); err != nil {
		return fmt.Errorf("navigating back: %w", err)
	}

	waitPageSettle(ctx, page)
	return nil
}

// Forward navigates forward in browser history and waits for the DOM to settle.
// Uses waitPageSettle instead of waitPageLoad because SPA routers handle
// forward navigation via popstate events without a full page load.
func (b *Browser) Forward(ctx context.Context) error {
	tab, err := b.activeTab()
	if err != nil {
		return err
	}
	page := tab.page

	if err := page.Context(ctx).NavigateForward(); err != nil {
		return fmt.Errorf("navigating forward: %w", err)
	}

	waitPageSettle(ctx, page)
	return nil
}

// Reload reloads the current page and waits for the DOM to settle.
// Uses waitPageSettle instead of waitPageLoad because cached pages and
// service workers may not fire the full navigation lifecycle events.
func (b *Browser) Reload(ctx context.Context) error {
	tab, err := b.activeTab()
	if err != nil {
		return err
	}
	page := tab.page

	if err := page.Context(ctx).Reload(); err != nil {
		return fmt.Errorf("reloading page: %w", err)
	}

	waitPageSettle(ctx, page)
	return nil
}

// Upload sets one or more files on a <input type="file"> element. Each path
// must be absolute and accessible to Chrome. Uses Rod's Element.SetFiles
// which calls CDP DOM.setFileInputFiles.
func (b *Browser) Upload(ctx context.Context, selector string, files []string) (*Element, error) {
	tab, err := b.activeTab()
	if err != nil {
		return nil, err
	}
	page := tab.page

	el, err := resolveElement(ctx, page, selector)
	if err != nil {
		return nil, err
	}

	el = el.Context(ctx)

	if err := el.Element().SetFiles(files); err != nil {
		return nil, fmt.Errorf("setting files: %w", err)
	}

	waitPageSettle(ctx, page)
	return el, nil
}

// Hover moves the mouse over an element without clicking. Triggers CSS :hover
// states and JS mouseenter/mouseover events. Rod's Hover handles scroll-into-view
// and wait-for-interactable internally.
func (b *Browser) Hover(ctx context.Context, selector string) (*Element, error) {
	tab, err := b.activeTab()
	if err != nil {
		return nil, err
	}
	page := tab.page

	el, err := resolveElement(ctx, page, selector)
	if err != nil {
		return nil, err
	}

	if err := el.Context(ctx).Element().Hover(); err != nil {
		return nil, fmt.Errorf("hovering element: %w", err)
	}

	waitPageSettle(ctx, page)
	return el, nil
}

// Wait waits for a duration or for an element to appear.
// If the argument parses as a duration (e.g. "2s", "500ms"), it sleeps.
// Otherwise it treats the argument as a selector and polls until the element exists.
func (b *Browser) Wait(ctx context.Context, target string) error {
	// Try parsing as duration first.
	d, err := time.ParseDuration(target)
	if err == nil {
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Treat as selector — rod's auto-retry handles polling until found or timeout.
	tab, err := b.activeTab()
	if err != nil {
		return err
	}
	page := tab.page

	_, err = findElement(ctx, page, target)
	return err
}

// translateInteractableErr converts Rod's interactability errors into our typed
// errors with backendNodeId context for the agent. Falls back to a generic
// fmt.Errorf if the error is not a recognized Rod interactability type.
func translateInteractableErr(err error, selector string) error {
	var covered *rod.CoveredError
	if errors.As(err, &covered) {
		return &ErrObscured{
			Sel:       selector,
			BlockerID: blockerNodeID(covered.Element),
		}
	}

	var invisible *rod.InvisibleShapeError
	if errors.As(err, &invisible) {
		return &ErrNotVisible{Sel: selector}
	}

	var noPointer *rod.NoPointerEventsError
	if errors.As(err, &noPointer) {
		return &ErrNoPointerEvents{Sel: selector}
	}

	return fmt.Errorf("element not interactable: %s: %w", selector, err)
}

// blockerNodeID extracts the backendNodeId from a Rod element, typically the
// covering element in a CoveredError. Returns "" if the element is nil or
// Describe fails.
func blockerNodeID(el *rod.Element) string {
	if el == nil {
		return ""
	}
	node, err := el.Describe(0, false)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", node.BackendNodeID)
}

// Key sends a single key press to the active page. The key name is
// case-insensitive and follows Playwright/W3C naming (e.g. "Enter",
// "ArrowDown", "Escape"). This dispatches real keyDown/keyUp CDP events,
// unlike Type which uses insertText.
func (b *Browser) Key(ctx context.Context, name string) error {
	key, ok := keyMap[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("unknown key %q", name)
	}

	tab, err := b.activeTab()
	if err != nil {
		return err
	}
	page := tab.page

	if err := page.Context(ctx).Keyboard.Press(key); err != nil {
		return fmt.Errorf("pressing key %q: %w", name, err)
	}

	waitPageSettle(ctx, page)
	return nil
}
