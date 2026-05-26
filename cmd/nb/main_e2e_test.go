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

func buildNB(t *testing.T) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "nb")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}

	build := exec.Command("go", "build", "-o", exe, ".")
	build.Env = os.Environ()
	out, err := build.CombinedOutput()
	require.NoError(t, err, string(out))
	return exe
}

func runNB(t *testing.T, exe, home, sessionID string, args ...string) string {
	t.Helper()
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(),
		"USERPROFILE="+home,
		"HOME="+home,
		"NB_SESSION_ID="+sessionID,
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func TestBinaryFlow(t *testing.T) {
	exe := buildNB(t)
	home := t.TempDir()

	run := func(args ...string) string {
		return runNB(t, exe, home, "e2e-a", args...)
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

func TestBinaryIsolatesSessions(t *testing.T) {
	exe := buildNB(t)
	home := t.TempDir()

	runNB(t, exe, home, "e2e-seed", "new", "erp")
	runNB(t, exe, home, "e2e-seed", "new", "compras")

	require.Contains(t, runNB(t, exe, home, "e2e-A", "use", "erp"), "Usando 'erp'.")
	require.Contains(t, runNB(t, exe, home, "e2e-B", "use", "compras"), "Usando 'compras'.")

	require.Equal(t, "erp\n", runNB(t, exe, home, "e2e-A", "current"))
	require.Equal(t, "compras\n", runNB(t, exe, home, "e2e-B", "current"))

	runNB(t, exe, home, "e2e-B", "add", "tarefa B")

	require.NotContains(t, runNB(t, exe, home, "e2e-A", "show"), "tarefa B")
	require.Contains(t, runNB(t, exe, home, "e2e-B", "show"), "tarefa B")
}

func TestBinaryAdoptsLegacyCurrent(t *testing.T) {
	exe := buildNB(t)
	home := t.TempDir()

	runNB(t, exe, home, "e2e-seed", "new", "erp")

	baseDir := filepath.Join(home, ".notebook-cli")
	legacy := filepath.Join(baseDir, ".current")
	require.NoError(t, os.WriteFile(legacy, []byte("erp"), 0o644))

	require.Equal(t, "erp\n", runNB(t, exe, home, "e2e-a", "current"))
	require.NoFileExists(t, legacy)
	require.FileExists(t, filepath.Join(baseDir, "sessions", "e2e-a.current"))
}

func TestBinaryRemovesLegacyEvenWhenSessionFileExists(t *testing.T) {
	exe := buildNB(t)
	home := t.TempDir()

	runNB(t, exe, home, "e2e-seed", "new", "erp")
	runNB(t, exe, home, "e2e-seed", "new", "compras")
	runNB(t, exe, home, "e2e-a", "use", "compras")

	baseDir := filepath.Join(home, ".notebook-cli")
	sessionFile := filepath.Join(baseDir, "sessions", "e2e-a.current")
	require.FileExists(t, sessionFile)

	legacy := filepath.Join(baseDir, ".current")
	require.NoError(t, os.WriteFile(legacy, []byte("erp"), 0o644))

	require.Equal(t, "compras\n", runNB(t, exe, home, "e2e-a", "current"))
	require.NoFileExists(t, legacy)
}
