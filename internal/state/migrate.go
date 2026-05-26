package state

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// MigrateLegacy moves a pre-session ".current" file inside baseDir into the
// per-session storage layout. If the active session already has its own
// file, the session file is preserved. The legacy file is always removed at
// the end (best-effort), even when the session already had its own value,
// to prevent a later session from "adopting" stale global state.
func MigrateLegacy(baseDir, sessionID string) error {
	legacy := filepath.Join(baseDir, ".current")
	data, err := os.ReadFile(legacy)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	sessionPath := filepath.Join(baseDir, "sessions", sessionID+".current")
	if _, statErr := os.Stat(sessionPath); errors.Is(statErr, fs.ErrNotExist) {
		if writeErr := writeAtomic(sessionPath, data); writeErr != nil {
			_ = os.Remove(legacy)
			return writeErr
		}
	}

	_ = os.Remove(legacy)
	return nil
}

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".current.*.tmp")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
