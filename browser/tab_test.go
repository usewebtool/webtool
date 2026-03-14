package browser

import (
	"errors"
	"fmt"
	"testing"
)

func newTestTab() *tab {
	return &tab{asyncErr: make(chan error, 1)}
}

func TestTabErr_EmptyReturnsNil(t *testing.T) {
	tab := newTestTab()
	if err := tab.Err(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestTabErr_DrainsError(t *testing.T) {
	tab := newTestTab()
	tab.sendErr(fmt.Errorf("test error"))

	err := tab.Err()
	if err == nil || err.Error() != "test error" {
		t.Fatalf("expected 'test error', got %v", err)
	}

	// Second call returns nil — drained.
	if err := tab.Err(); err != nil {
		t.Fatalf("expected nil after drain, got %v", err)
	}
}

func TestTabSendErr_NonBlocking(t *testing.T) {
	tab := newTestTab()

	// First send succeeds.
	tab.sendErr(fmt.Errorf("first"))

	// Second send is dropped (buffer full).
	tab.sendErr(fmt.Errorf("second"))

	err := tab.Err()
	if err == nil || err.Error() != "first" {
		t.Fatalf("expected 'first', got %v", err)
	}

	// Channel is now empty.
	if err := tab.Err(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestErrBlocked_Format(t *testing.T) {
	err := &ErrBlocked{
		Method: "POST",
		URL:    "https://api.example.com/sync",
		Rule:   "url=*api.example.com/sync* body=danger",
	}

	want := "request blocked by policy: POST https://api.example.com/sync (rule: url=*api.example.com/sync* body=danger)"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestErrBlocked_ErrorsAs(t *testing.T) {
	tab := newTestTab()
	tab.sendErr(&ErrBlocked{Method: "DELETE", URL: "https://example.com", Rule: "method=DELETE"})

	err := tab.Err()
	var blocked *ErrBlocked
	if !errors.As(err, &blocked) {
		t.Fatalf("expected errors.As to match ErrBlocked, got %T", err)
	}
	if blocked.Method != "DELETE" {
		t.Errorf("method: got %q, want %q", blocked.Method, "DELETE")
	}
}
