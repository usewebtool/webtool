//go:build integration

package integration

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var webtoolBin string

func TestMain(m *testing.M) {
	// Find the built binary relative to this test file.
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..")
	webtoolBin = filepath.Join(root, "dist", "webtool")

	if _, err := os.Stat(webtoolBin); err != nil {
		fmt.Fprintf(os.Stderr, "binary not found at %s — run 'make build' first\n", webtoolBin)
		os.Exit(1)
	}

	// Ensure daemon is running before tests.
	if out, code := webtool("open", "about:blank"); code != 0 {
		fmt.Fprintf(os.Stderr, "failed to start daemon (exit %d): %s\n", code, out)
		os.Exit(1)
	}

	code := m.Run()

	// Stop the daemon so the test process can exit.
	webtool("stop")

	os.Exit(code)
}

// webtool runs the webtool binary with the given args and returns stdout+stderr and the exit code.
func webtool(args ...string) (string, int) {
	cmd := exec.Command(webtoolBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return string(out), exitErr.ExitCode()
		}
		return string(out), 2
	}
	return string(out), 0
}

// webtoolOK runs webtool and fails the test if the exit code is non-zero.
func webtoolOK(t *testing.T, args ...string) string {
	t.Helper()
	out, code := webtool(args...)
	if code != 0 {
		t.Fatalf("webtool %s failed (exit %d):\n%s", strings.Join(args, " "), code, out)
	}
	return out
}
