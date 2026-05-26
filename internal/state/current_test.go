package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrentStoreGetSetAndClear(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state")
	store := New(dir, "sid")

	name, err := store.Get()
	require.NoError(t, err)
	require.Empty(t, name)

	require.NoError(t, store.Set("erp"))
	name, err = store.Get()
	require.NoError(t, err)
	require.Equal(t, "erp", name)

	require.NoError(t, store.Set("compras"))
	name, err = store.Get()
	require.NoError(t, err)
	require.Equal(t, "compras", name)

	require.NoError(t, store.Clear())
	name, err = store.Get()
	require.NoError(t, err)
	require.Empty(t, name)
	require.NoError(t, store.Clear())
}

func TestCurrentStoreSetFailsWhenParentIsFile(t *testing.T) {
	baseDir := t.TempDir()
	sessionsAsFile := filepath.Join(baseDir, "sessions")
	require.NoError(t, os.WriteFile(sessionsAsFile, []byte("x"), 0o644))

	store := New(baseDir, "sid")

	require.Error(t, store.Set("erp"))
}

func TestCurrentStoreIsolationBetweenSessions(t *testing.T) {
	baseDir := t.TempDir()
	a := New(baseDir, "session-a")
	b := New(baseDir, "session-b")

	require.NoError(t, a.Set("erp"))
	require.NoError(t, b.Set("compras"))

	got, err := a.Get()
	require.NoError(t, err)
	require.Equal(t, "erp", got)

	got, err = b.Get()
	require.NoError(t, err)
	require.Equal(t, "compras", got)

	require.NoError(t, a.Clear())
	got, err = a.Get()
	require.NoError(t, err)
	require.Empty(t, got)

	got, err = b.Get()
	require.NoError(t, err)
	require.Equal(t, "compras", got)
}
