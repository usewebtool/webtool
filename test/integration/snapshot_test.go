//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestSnapshot_ExampleDotCom(t *testing.T) {
	webtoolOK(t, "open", "https://example.com")

	out := webtoolOK(t, "snapshot")

	if !strings.Contains(out, "https://example.com") {
		t.Errorf("snapshot missing URL line, got:\n%s", out)
	}

	if !strings.Contains(out, "link") {
		t.Errorf("snapshot missing link element, got:\n%s", out)
	}
}
