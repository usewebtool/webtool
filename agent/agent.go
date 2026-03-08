package agent

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
)

// HomeDir is the base directory for all webtool runtime files.
// Must be set by cmd/ before creating clients or servers.
var HomeDir string

func runtimeDir(chromeDataDir string) string {
	h := sha256.Sum256([]byte(chromeDataDir))
	return filepath.Join(HomeDir, fmt.Sprintf("%x", h[:3]))
}
