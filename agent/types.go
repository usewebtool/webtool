package agent

import (
	"encoding/json"
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
	URL    string `json:"url"`
	NewTab bool   `json:"new_tab,omitempty"`
}

// TabsResponse is the response for the /tabs endpoint.
type TabsResponse struct {
	Response
	Tabs []browser.TabInfo `json:"tabs,omitempty"`
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

// EvalRequest is the request body for the /eval endpoint.
type EvalRequest struct {
	JS string `json:"js"`
}

// EvalResponse is the response for the /eval endpoint.
type EvalResponse struct {
	Response
	Result string `json:"result,omitempty"`
}

// SelectRequest is the request body for the /select endpoint.
type SelectRequest struct {
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

// UploadRequest is the request body for the /upload endpoint.
type UploadRequest struct {
	Selector string   `json:"selector"`
	Files    []string `json:"files"`
}

// WaitRequest is the request body for the /wait endpoint.
type WaitRequest struct {
	Target string `json:"target"`
}

// HoverRequest is the request body for the /hover endpoint.
type HoverRequest struct {
	Selector string `json:"selector"`
}

// SwitchRequest is the request body for the /switch endpoint.
type SwitchRequest struct {
	Index int `json:"index"`
}

// SnapshotResponse is the response for the /snapshot endpoint.
type SnapshotResponse struct {
	Response
	Snapshot string `json:"snapshot,omitempty"`
}

// CDPRequest is the request body for the /cdp endpoint.
type CDPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// CDPResponse is the response for the /cdp endpoint.
type CDPResponse struct {
	Response
	Result json.RawMessage `json:"result,omitempty"`
}
