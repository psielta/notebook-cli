package state

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const (
	pruneCooldown      = 24 * time.Hour
	pruneFileThreshold = 25
	lastPruneFile      = ".last_prune"
)

var processSessionPattern = regexp.MustCompile(`^sh-(\d+)(?:-([0-9a-f]+))?\.current$`)

// MaybePrune removes session files belonging to processes that are no longer
// alive. Only files matching the canonical "sh-<pid>[-<creationHex>].current"
// naming are considered; ids supplied via NB_SESSION_ID (no "sh-" prefix)
// are left untouched. A cooldown of 24h and a file-count threshold avoid
// running on every invocation. All errors are best-effort — the caller is
// never blocked by a prune failure.
func MaybePrune(baseDir string, alive func(pid int, startHex string) bool) error {
	sessionsDir := filepath.Join(baseDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	if shouldSkipPrune(sessionsDir, entries) {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := processSessionPattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		pid, convErr := strconv.Atoi(match[1])
		if convErr != nil {
			continue
		}
		if alive(pid, match[2]) {
			continue
		}
		_ = os.Remove(filepath.Join(sessionsDir, entry.Name()))
	}

	_ = os.WriteFile(
		filepath.Join(sessionsDir, lastPruneFile),
		[]byte(strconv.FormatInt(time.Now().Unix(), 10)),
		0o644,
	)
	return nil
}

func shouldSkipPrune(sessionsDir string, entries []fs.DirEntry) bool {
	fileCount := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == lastPruneFile {
			continue
		}
		fileCount++
	}
	if fileCount > pruneFileThreshold {
		return false
	}
	data, err := os.ReadFile(filepath.Join(sessionsDir, lastPruneFile))
	if err != nil {
		return false
	}
	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return false
	}
	return time.Since(time.Unix(ts, 0)) < pruneCooldown
}
