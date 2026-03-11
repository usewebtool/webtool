package browser

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// resolveElement finds an element on the page by selector. The selector is
// interpreted as a backendNodeId (integer), XPath (starts with / or //),
// or CSS selector (anything else). CSS and XPath selectors use rod's built-in
// auto-retry until the element appears or the context deadline expires.
func resolveElement(ctx context.Context, page *rod.Page, selector string) (*rod.Element, error) {
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

// resolveElementNow finds an element immediately without retrying. Returns nil
// if the element does not exist right now. Uses rod's Has/HasX which check once
// and return, unlike Element/ElementX which retry until the context expires.
func resolveElementNow(ctx context.Context, page *rod.Page, selector string) (*rod.Element, error) {
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
		has, el, err := page.Context(ctx).HasX(selector)
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, &ErrNotFound{Sel: selector}
		}
		return el, nil
	}

	has, el, err := page.Context(ctx).Has(selector)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, &ErrNotFound{Sel: selector}
	}
	return el, nil
}
