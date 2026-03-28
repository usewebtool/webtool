package browser

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Element wraps a rod element with accessibility metadata. The Role and Name
// fields are populated from the accessibility tree at resolution time, before
// any action is taken on the element.
type Element struct {
	el            *rod.Element
	BackendNodeID int
	Role          string // e.g. "button", "textbox", "link"
	Name          string // accessible name, e.g. "Sign in", "Email"
}

// Element returns the underlying rod element for direct rod API calls.
func (e *Element) Element() *rod.Element {
	return e.el
}

// Context sets the context on the underlying rod element and returns
// the same Element, so callers can write el = el.Context(ctx).
func (e *Element) Context(ctx context.Context) *Element {
	e.el = e.el.Context(ctx)
	return e
}

// resolveElement finds an element on the page by selector and returns it
// wrapped with accessibility metadata. The selector is interpreted as a
// backendNodeId (integer), XPath (starts with / or //), or CSS selector
// (anything else). CSS and XPath selectors use rod's built-in auto-retry
// until the element appears or the context deadline expires.
func resolveElement(ctx context.Context, page *rod.Page, selector string) (*Element, error) {
	el, err := findElement(ctx, page, selector)
	if err != nil {
		return nil, err
	}

	node, err := el.Describe(0, false)
	if err != nil {
		return nil, &ErrStaleNode{Sel: selector}
	}
	backendNodeID := int(node.BackendNodeID)

	axNode, err := describeAX(ctx, page, node.BackendNodeID)
	if err != nil {
		return nil, err
	}

	var role, name string
	if axNode != nil {
		role = axStr(axNode.Role)
		name = axStr(axNode.Name)
	}

	return &Element{
		el:            el,
		BackendNodeID: backendNodeID,
		Role:          role,
		Name:          name,
	}, nil
}

// findElement locates a rod element by selector without enriching it with
// accessibility metadata. This is the shared selector dispatch used by
// resolveElement.
func findElement(ctx context.Context, page *rod.Page, selector string) (*rod.Element, error) {
	id, err := strconv.Atoi(selector)
	if err == nil {
		el, err := page.Context(ctx).ElementFromNode(&proto.DOMNode{
			BackendNodeID: proto.DOMBackendNodeID(id),
		})
		if err != nil {
			return nil, &ErrStaleNode{Sel: selector}
		}
		return el, nil
	}

	if strings.HasPrefix(selector, "/") {
		el, err := page.Context(ctx).ElementX(selector)
		if err != nil {
			return nil, &ErrNotFound{Sel: selector}
		}
		return el, nil
	}

	el, err := page.Context(ctx).Element(selector)
	if err != nil {
		return nil, &ErrNotFound{Sel: selector}
	}
	return el, nil
}

// describeAX fetches the accessibility node for a DOM element.
func describeAX(ctx context.Context, page *rod.Page, backendNodeID proto.DOMBackendNodeID) (*proto.AccessibilityAXNode, error) {
	result, err := proto.AccessibilityGetPartialAXTree{
		BackendNodeID:  backendNodeID,
		FetchRelatives: false,
	}.Call(page.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting accessibility node: %w", err)
	}
	if len(result.Nodes) == 0 {
		return nil, nil
	}

	return result.Nodes[0], nil
}
