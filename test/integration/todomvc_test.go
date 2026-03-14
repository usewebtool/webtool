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

	// Clear completed todos.
	clearID := findElement(t, snap, "Clear completed")
	webtoolOK(t, "click", clearID)

	// Verify the completed todo is gone.
	snap = webtoolOK(t, "snapshot")
	if strings.Contains(snap, "Buy groceries") {
		t.Fatalf("expected 'Buy groceries' to be cleared, got:\n%s", snap)
	}
}

func TestTodoMVC_HoverAndDelete(t *testing.T) {
	webtoolOK(t, "open", todoMVCURL)

	// Add a todo.
	snap := webtoolOK(t, "snapshot")
	inputID := findElement(t, snap, "textbox")
	webtoolOK(t, "type", inputID, "Delete me")
	webtoolOK(t, "key", "Enter")

	snap = webtoolOK(t, "snapshot")
	if !strings.Contains(snap, "Delete me") {
		t.Fatalf("expected 'Delete me' in snapshot, got:\n%s", snap)
	}

	// The delete button is hidden until hover. Verify it's not visible.
	if strings.Contains(snap, "Delete") && !strings.Contains(snap, "Delete me") {
		t.Fatalf("expected no Delete button before hover, got:\n%s", snap)
	}

	// Hover over the task label to reveal the delete button.
	labelID := findElement(t, snap, `label "Delete me"`)
	webtoolOK(t, "hover", labelID)

	// Snapshot should now show the delete button.
	snap = webtoolOK(t, "snapshot")
	deleteID := findElement(t, snap, `button "Delete"`)
	webtoolOK(t, "click", deleteID)

	// Verify the task is gone.
	snap = webtoolOK(t, "snapshot")
	if strings.Contains(snap, "Delete me") {
		t.Fatalf("expected 'Delete me' to be removed, got:\n%s", snap)
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
