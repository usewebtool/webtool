package agent

import (
	"encoding/json"
	"testing"

	"github.com/machinae/webtool/browser"
)

func TestResponseErrEmpty(t *testing.T) {
	r := Response{}
	if err := r.Err(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestResponseErrNonEmpty(t *testing.T) {
	r := Response{Error: "something broke"}
	err := r.Err()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "something broke" {
		t.Errorf("expected %q, got %q", "something broke", err.Error())
	}
}

func TestTabsResponseJSON(t *testing.T) {
	resp := TabsResponse{
		Tabs: []browser.Tab{
			{ID: "abc", Title: "Example", URL: "https://example.com"},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded TabsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Tabs) != 1 {
		t.Fatalf("expected 1 tab, got %d", len(decoded.Tabs))
	}
	if decoded.Tabs[0].ID != "abc" {
		t.Errorf("tab ID: got %q, want %q", decoded.Tabs[0].ID, "abc")
	}
	if decoded.Err() != nil {
		t.Errorf("expected nil error, got %v", decoded.Err())
	}
}

func TestTabsResponseJSONWithError(t *testing.T) {
	resp := TabsResponse{
		Response: Response{Error: "chrome disconnected"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded TabsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Err() == nil {
		t.Fatal("expected error, got nil")
	}
	if decoded.Error != "chrome disconnected" {
		t.Errorf("error: got %q, want %q", decoded.Error, "chrome disconnected")
	}
}
