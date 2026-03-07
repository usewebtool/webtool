package agent

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/machinae/webtool/browser"
)

// startTestServer starts a server in a temp directory and returns the socket path and a cleanup function.
func startTestServer(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	srv := NewServer(browser.New())
	srv.dir = dir

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Wait for socket to appear.
	sock := filepath.Join(dir, "agent.sock")
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sock); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Cleanup(func() {
		srv.Shutdown(nil)
		<-errCh
	})

	return sock
}

func httpClient(sock string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sock)
			},
		},
	}
}

func TestServerHealth(t *testing.T) {
	sock := startTestServer(t)
	client := httpClient(sock)

	resp, err := client.Get("http://localhost/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Err() != nil {
		t.Errorf("expected nil error, got %v", r.Err())
	}
}

func TestServerPIDFile(t *testing.T) {
	sock := startTestServer(t)
	dir := filepath.Dir(sock)

	data, err := os.ReadFile(filepath.Join(dir, "agent.pid"))
	if err != nil {
		t.Fatalf("reading PID file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("PID file is empty")
	}
}

func TestOpenMissingURL(t *testing.T) {
	sock := startTestServer(t)
	client := httpClient(sock)

	resp, err := client.Post("http://localhost/open", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST /open: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Err() == nil {
		t.Error("expected error for missing URL")
	}
}

func TestServerStopCleansUp(t *testing.T) {
	dir := t.TempDir()

	srv := NewServer(browser.New())
	srv.dir = dir

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	sock := filepath.Join(dir, "agent.sock")
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sock); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Send stop request.
	client := httpClient(sock)
	resp, err := client.Post("http://localhost/stop", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /stop: %v", err)
	}
	resp.Body.Close()

	// Wait for server to exit.
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server exited with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not exit after stop")
	}

	// Verify cleanup.
	if _, err := os.Stat(sock); !os.IsNotExist(err) {
		t.Error("socket file should be removed after stop")
	}
	if _, err := os.Stat(filepath.Join(dir, "agent.pid")); !os.IsNotExist(err) {
		t.Error("PID file should be removed after stop")
	}
}
