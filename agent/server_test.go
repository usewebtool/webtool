package agent

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/usewebtool/webtool/browser"
)

func TestHealthReturnsConnectError(t *testing.T) {
	// Browser pointed at a port nothing listens on — Connect() fails immediately.
	b := browser.New().WithURL("ws://127.0.0.1:1/devtools/browser/fake")
	s := NewServer(b)
	s.dir = t.TempDir()

	go func() {
		if err := s.Start(); err != nil {
			t.Logf("server exited: %v", err)
		}
	}()

	// Wait for socket to appear.
	sock := s.dir + "/agent.sock"
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := net.Dial("unix", sock); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Health should return the connect error, not hang.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sock)
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestFormatParams_RedactsSensitiveFields(t *testing.T) {
	body := []byte(`{"selector":"12345","text":"my-secret-password","js":"document.cookie"}`)
	got := formatParams(body)

	if strings.Contains(got, "my-secret-password") {
		t.Error("text value should be redacted")
	}
	if strings.Contains(got, "document.cookie") {
		t.Error("js value should be redacted")
	}
	if !strings.Contains(got, "text=[REDACTED]") {
		t.Error("text should show [REDACTED]")
	}
	if !strings.Contains(got, "js=[REDACTED]") {
		t.Error("js should show [REDACTED]")
	}
	if !strings.Contains(got, "selector=12345") {
		t.Error("selector should be logged normally")
	}
}
