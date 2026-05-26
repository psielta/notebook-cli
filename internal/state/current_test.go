package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrentStoreGetSetAndClear(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state")
	store := New(dir)

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
	filePath := filepath.Join(t.TempDir(), "state-file")
	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0o644))

	store := New(filepath.Join(filePath, "child"))

	require.Error(t, store.Set("erp"))
}
