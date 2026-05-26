package database

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenCreatesDirectoryAndEnablesForeignKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "notebook.db")

	db, err := Open(path)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	var foreignKeys int
	require.NoError(t, db.Raw("PRAGMA foreign_keys").Scan(&foreignKeys).Error)
	require.Equal(t, 1, foreignKeys)
	require.FileExists(t, path)
}
