package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunHappyPathAndBusinessError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	t.Setenv("NB_SESSION_ID", "test-x")

	var out bytes.Buffer
	var errOut bytes.Buffer

	require.Equal(t, 0, run([]string{"new", "erp"}, nil, &out, &errOut))
	require.Contains(t, out.String(), "Notebook 'erp' criado.")

	out.Reset()
	require.Equal(t, 0, run([]string{"use", "erp"}, nil, &out, &errOut))
	require.Contains(t, out.String(), "Usando 'erp'.")

	out.Reset()
	require.Equal(t, 0, run([]string{"add", "x"}, nil, &out, &errOut))
	require.Contains(t, out.String(), "Nota 1 adicionada.")

	out.Reset()
	require.Equal(t, 0, run([]string{"show"}, nil, &out, &errOut))
	require.Contains(t, out.String(), "x")

	require.FileExists(t, filepath.Join(home, ".notebook-cli", "notebook.db"))
	require.FileExists(t, filepath.Join(home, ".notebook-cli", "sessions", "test-x.current"))
	require.NoFileExists(t, filepath.Join(home, ".notebook-cli", ".current"))

	errOut.Reset()
	require.Equal(t, 1, run([]string{"use", "missing"}, nil, &out, &errOut))
	require.Contains(t, errOut.String(), "Erro:")
}

func TestRunVersionDoesNotCreateAppData(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	var out bytes.Buffer
	var errOut bytes.Buffer

	require.Equal(t, 0, run([]string{"--version"}, nil, &out, &errOut))
	require.Equal(t, "nb dev\n", out.String())
	require.Empty(t, errOut.String())
	require.NoDirExists(t, filepath.Join(home, ".notebook-cli"))
}
