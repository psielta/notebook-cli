package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".notebook-cli"), cfg.BaseDir)
	require.Equal(t, filepath.Join(home, ".notebook-cli", "notebook.db"), cfg.DBPath)
}
