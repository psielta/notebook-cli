//go:build e2e

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBinaryFlow(t *testing.T) {
	tmp := t.TempDir()
	exe := filepath.Join(tmp, "nb")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}

	build := exec.Command("go", "build", "-o", exe, ".")
	build.Env = os.Environ()
	out, err := build.CombinedOutput()
	require.NoError(t, err, string(out))

	home := filepath.Join(tmp, "home")
	require.NoError(t, os.MkdirAll(home, 0o755))

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(exe, args...)
		cmd.Env = append(os.Environ(), "USERPROFILE="+home, "HOME="+home)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return string(out)
	}

	require.Contains(t, run("new", "erp"), "Notebook 'erp' criado.")
	require.Contains(t, run("use", "erp"), "Usando 'erp'.")
	require.Equal(t, "erp\n", run("current"))
	require.Contains(t, run("add", "corrigir problema x"), "Nota 1 adicionada.")
	require.Contains(t, run("add", "testar importacao"), "Nota 2 adicionada.")
	require.Contains(t, run("show"), "corrigir problema x")
	require.Contains(t, run("remove", "1"), "Nota 1 removida.")
	require.Contains(t, run("clear", "--yes"), "1 nota removida.")
	require.Contains(t, run("show"), "nenhuma nota")
}
