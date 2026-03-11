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

// ErrNotVisible indicates that the element has no visible shape (zero bounding
// rect) or is outside the viewport.
type ErrNotVisible struct {
	Sel string
}

func (e *ErrNotVisible) Error() string {
	return fmt.Sprintf("element not visible: %s — element has no visible shape or is outside the viewport", e.Sel)
}
func (e *ErrNotVisible) Selector() string { return e.Sel }

// ErrObscured indicates that the element is covered by another element (e.g.
// a modal, cookie banner, or overlay). BlockerID is the backendNodeId of the
// covering element, or empty if it could not be determined.
type ErrObscured struct {
	Sel       string
	BlockerID string
}

func (e *ErrObscured) Error() string {
	if e.BlockerID != "" {
		return fmt.Sprintf("element obscured: %s is covered by element %s — dismiss or click the covering element first", e.Sel, e.BlockerID)
	}
	return fmt.Sprintf("element obscured: %s is covered by another element", e.Sel)
}
func (e *ErrObscured) Selector() string { return e.Sel }

// ErrNoPointerEvents indicates that the element has pointer-events: none set
// in CSS, preventing mouse interaction.
type ErrNoPointerEvents struct {
	Sel string
}

func (e *ErrNoPointerEvents) Error() string {
	return fmt.Sprintf("element not clickable: %s — pointer-events is none", e.Sel)
}
func (e *ErrNoPointerEvents) Selector() string { return e.Sel }

// ErrNotStable indicates that the element's position or size is still changing
// (animation, layout shift, or transition in progress).
type ErrNotStable struct {
	Sel string
}

func (e *ErrNotStable) Error() string {
	return fmt.Sprintf("element not stable: %s — position or size still changing", e.Sel)
}
func (e *ErrNotStable) Selector() string { return e.Sel }

// ErrNotEnabled indicates that the element is disabled.
type ErrNotEnabled struct {
	Sel string
}

func (e *ErrNotEnabled) Error() string    { return fmt.Sprintf("element disabled: %s", e.Sel) }
func (e *ErrNotEnabled) Selector() string { return e.Sel }
