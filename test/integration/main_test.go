//go:build integration

// Package integration contains browser-level integration tests that run against
// a real Chrome instance and a local HTTP server.
//
// Adding a new test:
//  1. Define an HTML constant (e.g. const myPageHTML = `...`) in a new or existing test file.
//  2. Register it in the pages map below (e.g. "/my-page": myPageHTML).
//  3. Write tests that call b.Open(ctx, pageURL("/my-page"), false) and use
//     b.Snapshot, b.Click, etc. directly.
//
// Run with: go test -tags integration ./test/integration/ -v -count=1
package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/usewebtool/webtool/browser"
)

var (
	b      *browser.Browser
	server *httptest.Server
)

// pages maps route paths to HTML content. Add new fixtures here.
var pages = map[string]string{
	"/simple": simpleHTML,
}

func TestMain(m *testing.M) {
	// Start local HTTP server serving embedded HTML fixtures.
	mux := http.NewServeMux()
	for path, content := range pages {
		body := content // capture for closure
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, body)
		})
	}
	server = httptest.NewServer(mux)

	// Connect to Chrome once for all tests.
	b = browser.New()
	if err := b.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to Chrome: %v\n", err)
		server.Close()
		os.Exit(1)
	}

	code := m.Run()

	b.Close()
	server.Close()
	os.Exit(code)
}

// pageURL returns the full URL for a fixture path (e.g. "/simple").
func pageURL(path string) string {
	return server.URL + path
}

// findElement returns the backendNodeId string for the first element matching
// the given substring in a snapshot string. Snapshot lines look like: [12345] role "name"
func findElement(t *testing.T, snapshot, match string) string {
	t.Helper()
	for _, line := range strings.Split(snapshot, "\n") {
		if strings.Contains(line, match) {
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
	t.Fatalf("no element matching %q found in snapshot:\n%s", match, snapshot)
	return ""
}

const simpleHTML = `<!DOCTYPE html>
<html>
<head><title>Simple Test</title></head>
<body>
	<h1>Hello</h1>
	<button id="btn" onclick="document.getElementById('output').textContent = 'clicked'">Click me</button>
	<div id="output"></div>
</body>
</html>`
