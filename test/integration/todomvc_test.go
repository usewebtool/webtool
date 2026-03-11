//go:build integration

package integration

import (
	"strings"
	"testing"
)

const todoMVCURL = "https://demo.playwright.dev/todomvc"

func TestTodoMVC_AddAndComplete(t *testing.T) {
	webtoolOK(t, "open", todoMVCURL)

	// Snapshot the empty page — should have the input field.
	snap := webtoolOK(t, "snapshot")
	if !strings.Contains(snap, "textbox") {
		t.Fatalf("expected textbox in snapshot, got:\n%s", snap)
	}

	// Find the input element ID from the snapshot.
	inputID := findElement(t, snap, "textbox")

	// Add first todo.
	webtoolOK(t, "type", inputID, "Buy groceries")
	webtoolOK(t, "key", "Enter")

	// Snapshot should now show the todo item.
	snap = webtoolOK(t, "snapshot")
	if !strings.Contains(snap, "Buy groceries") {
		t.Fatalf("expected 'Buy groceries' in snapshot after adding, got:\n%s", snap)
	}

	// Add second todo.
	inputID = findElement(t, snap, "textbox")
	webtoolOK(t, "type", inputID, "Walk the dog")
	webtoolOK(t, "key", "Enter")

	snap = webtoolOK(t, "snapshot")
	if !strings.Contains(snap, "Walk the dog") {
		t.Fatalf("expected 'Walk the dog' in snapshot after adding, got:\n%s", snap)
	}

	// Complete the first todo by clicking its checkbox.
	checkboxID := findElement(t, snap, "checkbox")
	webtoolOK(t, "click", checkboxID)

	// Snapshot should still show both items after the SPA re-render.
	snap = webtoolOK(t, "snapshot")
	if !strings.Contains(snap, "Buy groceries") {
		t.Fatalf("expected 'Buy groceries' in snapshot after completing, got:\n%s", snap)
	}
	if !strings.Contains(snap, "Walk the dog") {
		t.Fatalf("expected 'Walk the dog' in snapshot after completing, got:\n%s", snap)
	}
}

// findElement returns the backendNodeId for the first element matching the given role in a snapshot.
// Snapshot lines look like: [12345] role "name"
func findElement(t *testing.T, snapshot, role string) string {
	t.Helper()
	for _, line := range strings.Split(snapshot, "\n") {
		if strings.Contains(line, role) {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "[") {
				continue
			}
			end := strings.Index(line, "]")
			if end < 0 {
				continue
			}
			return line[1:end]
		}
	}
	t.Fatalf("no element with role %q found in snapshot:\n%s", role, snapshot)
	return ""
}
