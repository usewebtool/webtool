package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

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

	if err := el.WaitStable(300 * time.Millisecond); err != nil {
		return fmt.Errorf("waiting for element stability: %w", err)
	}

	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking element: %w", err)
	}

	return nil
}
