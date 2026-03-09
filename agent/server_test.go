package agent

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/machinae/webtool/browser"
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
