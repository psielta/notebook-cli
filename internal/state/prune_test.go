package state

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func writeSessionFile(t *testing.T, sessionsDir, name string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, name), []byte("x"), 0o644))
}

func TestMaybePruneRemovesOnlyDeadProcessFiles(t *testing.T) {
	base := t.TempDir()
	sessionsDir := filepath.Join(base, "sessions")
	writeSessionFile(t, sessionsDir, "sh-100-abc.current")
	writeSessionFile(t, sessionsDir, "sh-200-def.current")
	writeSessionFile(t, sessionsDir, "sh-300-fff.current")

	alive := func(pid int, _ string) bool { return pid == 200 }
	require.NoError(t, MaybePrune(base, alive))

	require.NoFileExists(t, filepath.Join(sessionsDir, "sh-100-abc.current"))
	require.FileExists(t, filepath.Join(sessionsDir, "sh-200-def.current"))
	require.NoFileExists(t, filepath.Join(sessionsDir, "sh-300-fff.current"))
	require.FileExists(t, filepath.Join(sessionsDir, lastPruneFile))
}

func TestMaybePruneIgnoresCustomSessionIDs(t *testing.T) {
	base := t.TempDir()
	sessionsDir := filepath.Join(base, "sessions")
	writeSessionFile(t, sessionsDir, "shared.current")
	writeSessionFile(t, sessionsDir, "my-session.current")
	writeSessionFile(t, sessionsDir, "ci-job-123.current")

	alive := func(int, string) bool {
		t.Fatalf("alive should not be called for custom ids")
		return false
	}
	require.NoError(t, MaybePrune(base, alive))

	require.FileExists(t, filepath.Join(sessionsDir, "shared.current"))
	require.FileExists(t, filepath.Join(sessionsDir, "my-session.current"))
	require.FileExists(t, filepath.Join(sessionsDir, "ci-job-123.current"))
}

func TestMaybePruneTreatsShPrefixAsProcessManaged(t *testing.T) {
	base := t.TempDir()
	sessionsDir := filepath.Join(base, "sessions")
	writeSessionFile(t, sessionsDir, "sh-123.current")

	require.NoError(t, MaybePrune(base, func(pid int, startHex string) bool {
		require.Equal(t, 123, pid)
		require.Empty(t, startHex)
		return false
	}))

	require.NoFileExists(t, filepath.Join(sessionsDir, "sh-123.current"))
}

func TestMaybePruneHandlesFormatWithoutCreationHex(t *testing.T) {
	base := t.TempDir()
	sessionsDir := filepath.Join(base, "sessions")
	writeSessionFile(t, sessionsDir, "sh-42.current")

	called := false
	alive := func(pid int, startHex string) bool {
		called = true
		require.Equal(t, 42, pid)
		require.Empty(t, startHex)
		return false
	}
	require.NoError(t, MaybePrune(base, alive))

	require.True(t, called)
	require.NoFileExists(t, filepath.Join(sessionsDir, "sh-42.current"))
}

func TestMaybePruneRespectsCooldown(t *testing.T) {
	base := t.TempDir()
	sessionsDir := filepath.Join(base, "sessions")
	writeSessionFile(t, sessionsDir, "sh-100-abc.current")
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionsDir, lastPruneFile),
		[]byte(strconv.FormatInt(time.Now().Unix(), 10)),
		0o644,
	))

	called := false
	alive := func(int, string) bool { called = true; return false }
	require.NoError(t, MaybePrune(base, alive))

	require.False(t, called)
	require.FileExists(t, filepath.Join(sessionsDir, "sh-100-abc.current"))
}

func TestMaybePruneRunsWhenCooldownExpired(t *testing.T) {
	base := t.TempDir()
	sessionsDir := filepath.Join(base, "sessions")
	writeSessionFile(t, sessionsDir, "sh-100-abc.current")
	stale := time.Now().Add(-48 * time.Hour).Unix()
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionsDir, lastPruneFile),
		[]byte(strconv.FormatInt(stale, 10)),
		0o644,
	))

	alive := func(int, string) bool { return false }
	require.NoError(t, MaybePrune(base, alive))

	require.NoFileExists(t, filepath.Join(sessionsDir, "sh-100-abc.current"))
}

func TestMaybePruneRunsWhenAboveThreshold(t *testing.T) {
	base := t.TempDir()
	sessionsDir := filepath.Join(base, "sessions")
	for i := 0; i < pruneFileThreshold+1; i++ {
		writeSessionFile(t, sessionsDir, "sh-"+strconv.Itoa(i)+"-abc.current")
	}
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionsDir, lastPruneFile),
		[]byte(strconv.FormatInt(time.Now().Unix(), 10)),
		0o644,
	))

	alive := func(int, string) bool { return false }
	require.NoError(t, MaybePrune(base, alive))

	for i := 0; i < pruneFileThreshold+1; i++ {
		require.NoFileExists(t, filepath.Join(sessionsDir, "sh-"+strconv.Itoa(i)+"-abc.current"))
	}
}

func TestMaybePruneMissingSessionsDir(t *testing.T) {
	base := t.TempDir()
	alive := func(int, string) bool { return true }
	require.NoError(t, MaybePrune(base, alive))
}
