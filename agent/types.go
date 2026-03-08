package agent

import (
	"errors"

	"github.com/machinae/webtool/browser"
)

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

// SnapshotResponse is the response for the /snapshot endpoint.
type SnapshotResponse struct {
	Response
	Snapshot string `json:"snapshot,omitempty"`
}
