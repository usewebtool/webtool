package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/machinae/webtool/browser"
)

// Client communicates with the daemon over a Unix socket.
type Client struct {
	http *http.Client
	dir  string // runtime directory for this daemon instance
}

// NewClient creates a client that connects to the daemon for the default Chrome data directory.
func NewClient() *Client {
	dir, err := browser.DefaultChromeUserDataDir()
	if err != nil {
		panic(fmt.Sprintf("unsupported OS: %v", err))
	}
	return NewClientWithDataDir(dir)
}

// NewClientWithDataDir creates a client that connects to the daemon for a specific Chrome data directory.
func NewClientWithDataDir(chromeDataDir string) *Client {
	dir := runtimeDir(chromeDataDir)
	sock := dir + "/agent.sock"
	dialer := &net.Dialer{}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", sock)
		},
	}
	return &Client{
		dir:  dir,
		http: &http.Client{Transport: transport},
	}
}

// EnsureRunning starts the daemon if it is not already running.
// Returns nil if the daemon is healthy (either already running or just started).
func (c *Client) EnsureRunning(ctx context.Context) error {
	if c.Health(ctx) == nil {
		return nil
	}

	if err := c.spawn(); err != nil {
		return fmt.Errorf("starting daemon: %w", err)
	}

	// Poll until the daemon is ready or the caller cancels (e.g. Ctrl+C).
	for {
		if c.Health(ctx) == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// spawn starts the daemon as a detached background process.
func (c *Client) spawn() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}

	logPath := c.dir + "/webtool.log"
	fmt.Fprintf(os.Stderr, "daemon started. logging to %s\n", logPath)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, "_serve")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = sysProcAttr()

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return err
	}

	logFile.Close()
	return nil
}

// Health checks if the daemon is running and responsive.
func (c *Client) Health(ctx context.Context) error {
	var resp Response
	if err := c.do(ctx, "GET", "/health", nil, &resp); err != nil {
		return err
	}
	return resp.Err()
}

// Open navigates the browser to the given URL.
func (c *Client) Open(ctx context.Context, url string) error {
	var resp Response
	if err := c.do(ctx, "POST", "/open", OpenRequest{URL: url}, &resp); err != nil {
		return err
	}
	return resp.Err()
}

// Tabs returns all open browser tabs.
func (c *Client) Tabs(ctx context.Context) ([]browser.Tab, error) {
	var resp TabsResponse
	if err := c.do(ctx, "GET", "/tabs", nil, &resp); err != nil {
		return nil, err
	}
	if err := resp.Err(); err != nil {
		return nil, err
	}
	return resp.Tabs, nil
}

// Stop sends a shutdown request to the daemon.
func (c *Client) Stop(ctx context.Context) error {
	var resp Response
	if err := c.do(ctx, "POST", "/stop", nil, &resp); err != nil {
		return err
	}
	return resp.Err()
}

// do sends an HTTP request to the daemon and decodes the JSON response.
func (c *Client) do(ctx context.Context, method, path string, reqBody, respBody any) error {
	var req *http.Request
	var err error

	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}
		req, err = http.NewRequestWithContext(ctx, method, "http://localhost"+path, bytes.NewReader(data))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, "http://localhost"+path, nil)
		if err != nil {
			return err
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}
