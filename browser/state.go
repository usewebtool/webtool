package browser

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const stateFileName = "state.json"

// Load reads persisted state from DataDir/state.json into the Browser.
// Returns nil if the file or directory does not exist (no state means first run).
func (b *Browser) Load() error {
	data, err := os.ReadFile(filepath.Join(b.DataDir, stateFileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading state: %w", err)
	}

	return json.Unmarshal(data, b)
}

// Save writes the Browser's persisted state to DataDir/state.json.
// Creates DataDir if it does not exist (first write creates the dir).
func (b *Browser) Save() error {
	if err := os.MkdirAll(b.DataDir, 0o755); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}

	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	tmp := filepath.Join(b.DataDir, stateFileName+".tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}

	dest := filepath.Join(b.DataDir, stateFileName)
	if err := os.Rename(tmp, dest); err != nil {
		return fmt.Errorf("renaming state: %w", err)
	}

	return nil
}
