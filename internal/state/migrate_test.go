package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrateLegacyAdoptsAndRemoves(t *testing.T) {
	base := t.TempDir()
	legacy := filepath.Join(base, ".current")
	require.NoError(t, os.WriteFile(legacy, []byte("erp"), 0o644))

	require.NoError(t, MigrateLegacy(base, "session-a"))

	require.NoFileExists(t, legacy)

	got, err := os.ReadFile(filepath.Join(base, "sessions", "session-a.current"))
	require.NoError(t, err)
	require.Equal(t, "erp", string(got))
}

func TestMigrateLegacyDoesNotOverwriteSessionFileButRemovesLegacy(t *testing.T) {
	base := t.TempDir()
	sessionsDir := filepath.Join(base, "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))
	sessionPath := filepath.Join(sessionsDir, "session-a.current")
	require.NoError(t, os.WriteFile(sessionPath, []byte("compras"), 0o644))

	legacy := filepath.Join(base, ".current")
	require.NoError(t, os.WriteFile(legacy, []byte("erp"), 0o644))

	require.NoError(t, MigrateLegacy(base, "session-a"))

	require.NoFileExists(t, legacy)

	got, err := os.ReadFile(sessionPath)
	require.NoError(t, err)
	require.Equal(t, "compras", string(got))
}

func TestMigrateLegacyIsIdempotent(t *testing.T) {
	base := t.TempDir()
	legacy := filepath.Join(base, ".current")
	require.NoError(t, os.WriteFile(legacy, []byte("erp"), 0o644))

	require.NoError(t, MigrateLegacy(base, "session-a"))
	require.NoError(t, MigrateLegacy(base, "session-a"))

	got, err := os.ReadFile(filepath.Join(base, "sessions", "session-a.current"))
	require.NoError(t, err)
	require.Equal(t, "erp", string(got))
}

func TestMigrateLegacyNoLegacyFile(t *testing.T) {
	base := t.TempDir()
	require.NoError(t, MigrateLegacy(base, "session-a"))
	require.NoFileExists(t, filepath.Join(base, "sessions", "session-a.current"))
}

func TestMigrateLegacyMissingBaseDir(t *testing.T) {
	base := filepath.Join(t.TempDir(), "does-not-exist")
	require.NoError(t, MigrateLegacy(base, "session-a"))
}
