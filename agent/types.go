package agent

import (
	"errors"

	"github.com/machinae/webtool/browser"
)

// ErrTimeout is returned when an operation exceeds the --timeout deadline.
var ErrTimeout = errors.New("operation timed out or element not found")

// Response is the base response for all daemon endpoints.
type Response struct {
	Error string `json:"error,omitempty"`
}

// Err returns nil if Error is empty, or a Go error wrapping the message.
func (r Response) Err() error {
	if r.Error == "" {
		return nil
	}
	return errors.New(r.Error)
}

// OpenRequest is the request body for the /open endpoint.
type OpenRequest struct {
	URL string `json:"url"`
}

// TabsResponse is the response for the /tabs endpoint.
type TabsResponse struct {
	Response
	Tabs []browser.Tab `json:"tabs,omitempty"`
}

// ClickRequest is the request body for the /click endpoint.
type ClickRequest struct {
	Selector string `json:"selector"`
}

// TypeRequest is the request body for the /type endpoint.
type TypeRequest struct {
	Selector string `json:"selector"`
	Text     string `json:"text"`
}

// KeyRequest is the request body for the /key endpoint.
type KeyRequest struct {
	Name string `json:"name"`
}

// ExtractRequest is the request body for the /extract endpoint.
type ExtractRequest struct {
	Selector string `json:"selector,omitempty"`
	AsHTML   bool   `json:"as_html,omitempty"`
}

// ExtractResponse is the response for the /extract endpoint.
type ExtractResponse struct {
	Response
	Content string `json:"content,omitempty"`
}

// SnapshotResponse is the response for the /snapshot endpoint.
type SnapshotResponse struct {
	Response
	Snapshot string `json:"snapshot,omitempty"`
}
