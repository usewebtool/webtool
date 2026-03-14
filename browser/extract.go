package browser

import (
	"context"
	"fmt"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// Extract returns the content of an element (or the full page) as markdown or HTML.
// An empty selector extracts the full page body. If asHTML is true, raw HTML is
// returned; otherwise the HTML is converted to markdown.
func (b *Browser) Extract(ctx context.Context, selector string, asHTML bool) (string, error) {
	tab, err := b.activeTab()
	if err != nil {
		return "", err
	}
	page := tab.page

	var html string
	if selector == "" {
		html, err = page.Context(ctx).HTML()
	} else {
		el, resolveErr := resolveElement(ctx, page, selector)
		if resolveErr != nil {
			return "", resolveErr
		}
		html, err = el.HTML()
	}
	if err != nil {
		return "", fmt.Errorf("extracting HTML: %w", err)
	}

	if asHTML {
		return html, nil
	}

	return toMarkdown(html)
}

// toMarkdown converts an HTML string to markdown.
func toMarkdown(html string) (string, error) {
	md, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("converting HTML to markdown: %w", err)
	}
	return md, nil
}
