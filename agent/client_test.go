package agent

import (
	"context"
	"testing"
)

func TestClientHealth(t *testing.T) {
	sock := startTestServer(t)

	client := &Client{http: httpClient(sock)}

	if err := client.Health(context.Background()); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestClientStop(t *testing.T) {
	sock := startTestServer(t)

	client := &Client{http: httpClient(sock)}

	if err := client.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestEnsureRunning(t *testing.T) {
	sock := startTestServer(t)

	client := &Client{
		http: httpClient(sock),
		dir:  t.TempDir(),
	}

	if err := client.EnsureRunning(context.Background()); err != nil {
		t.Fatalf("EnsureRunning: %v", err)
	}
}

func TestClientNoServer(t *testing.T) {
	client := NewClientWithDataDir(t.TempDir())

	err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error when no daemon is running")
	}
}
