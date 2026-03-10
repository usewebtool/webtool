package browser

import "fmt"

// BrowserError is the interface for all actionability errors. It follows
// the net.Error convention: callers can type-assert or use errors.As to
// distinguish actionability failures from system errors.
type BrowserError interface {
	error
	Selector() string
}

// ErrStaleNode indicates that a backendNodeId is no longer attached to the DOM.
// The agent should re-snapshot to get fresh node IDs.
type ErrStaleNode struct {
	Sel string
}

func (e *ErrStaleNode) Error() string {
	return fmt.Sprintf("stale node: %s — run snapshot again to get fresh node IDs", e.Sel)
}
func (e *ErrStaleNode) Selector() string { return e.Sel }

// ErrNotFound indicates that no element matched the given CSS or XPath selector.
type ErrNotFound struct {
	Sel string
}

func (e *ErrNotFound) Error() string    { return fmt.Sprintf("element not found: %s", e.Sel) }
func (e *ErrNotFound) Selector() string { return e.Sel }

// ErrNotInteractable indicates that the element exists but cannot be acted on
// (not visible, obscured by overlay, or animation not settled).
type ErrNotInteractable struct {
	Sel    string
	Reason string
}

func (e *ErrNotInteractable) Error() string {
	return fmt.Sprintf("element not interactable: %s (%s)", e.Sel, e.Reason)
}
func (e *ErrNotInteractable) Selector() string { return e.Sel }

// ErrNotEnabled indicates that the element is disabled or readonly.
type ErrNotEnabled struct {
	Sel string
}

func (e *ErrNotEnabled) Error() string    { return fmt.Sprintf("element disabled: %s", e.Sel) }
func (e *ErrNotEnabled) Selector() string { return e.Sel }
