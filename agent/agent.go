package agent

import (
	"os"
	"path/filepath"
)

func runtimeDir() string {
	return filepath.Join(os.TempDir(), "webtool")
}

func socketPath() string {
	return filepath.Join(runtimeDir(), "agent.sock")
}

func pidPath() string {
	return filepath.Join(runtimeDir(), "agent.pid")
}

func logPath() string {
	return filepath.Join(runtimeDir(), "webtool.log")
}
