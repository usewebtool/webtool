package browser

import (
	"context"
	"fmt"
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
			return nil, fmt.Errorf("node %d not found: %w", id, err)
		}
		return el, nil
	}

	if strings.HasPrefix(selector, "/") {
		return page.Context(ctx).ElementX(selector)
	}

	return page.Context(ctx).Element(selector)
}
