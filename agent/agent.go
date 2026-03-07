package agent

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

func runtimeDir(chromeDataDir string) string {
	h := sha256.Sum256([]byte(chromeDataDir))
	return filepath.Join(os.TempDir(), "webtool", fmt.Sprintf("%x", h[:3]))
}
