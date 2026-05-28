package commands

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"notebook-cli/internal/app"
	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/testutil"
)

func TestCommandsHappyPath(t *testing.T) {
	a := testutil.NewTestApp(t)

	out, err := execute(t, a, []string{"new", "erp"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "Notebook 'erp' criado.")

	out, err = execute(t, a, []string{"use", "erp"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "Usando 'erp'.")

	out, err = execute(t, a, []string{"current"}, "")
	require.NoError(t, err)
	require.Equal(t, "erp\n", out)

	out, err = execute(t, a, []string{"add", "corrigir problema x"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "Nota 1 adicionada.")

	out, err = execute(t, a, []string{"add", "testar importacao"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "Nota 2 adicionada.")

	out, err = execute(t, a, []string{"show"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "ID")
	require.Contains(t, out, "corrigir problema x")
	require.Contains(t, out, "testar importacao")

	out, err = execute(t, a, []string{"last"}, "")
	require.NoError(t, err)
	require.NotContains(t, out, "corrigir problema x")
	require.Contains(t, out, "testar importacao")

	out, err = execute(t, a, []string{"list"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "erp")
	require.Contains(t, out, "2")

	out, err = execute(t, a, []string{"remove", "1"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "Nota 1 removida.")

	out, err = execute(t, a, []string{"clear", "--yes"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "1 nota removida.")

	out, err = execute(t, a, []string{"show"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "nenhuma nota")
}

func TestNewTaskCreatesAndUsesDailyNotebook(t *testing.T) {
	a := testutil.NewTestApp(t)

	out, err := execute(t, a, []string{"new", "task"}, "")
	require.NoError(t, err)

	matches := regexp.MustCompile(`Notebook '(T001-\d{4}-\d{2}-\d{2})' criado e selecionado\.`).FindStringSubmatch(out)
	require.Len(t, matches, 2)

	out, err = execute(t, a, []string{"current"}, "")
	require.NoError(t, err)
	require.Equal(t, matches[1]+"\n", out)

	out, err = execute(t, a, []string{"new", "task"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "T002-")
}

func TestAddResolvesAgentShortcuts(t *testing.T) {
	a := testutil.NewTestApp(t)
	_, err := execute(t, a, []string{"new", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"use", "erp"}, "")
	require.NoError(t, err)

	shortcuts := []struct {
		input    string
		expected string
	}{
		{"clau-p", "Claude planejando..."},
		{"clau-vp", "Claude validando plano do Codex..."},
		{"clau-cp", "Claude corrigindo plano com melhorias do Codex..."},
		{"dex-vp", "Codex validando plano do Claude..."},
		{"dex-cp", "Codex corrigindo plano com melhorias do Claude..."},
		{"dex-to-clau", "Codex gerando prompt para Claude..."},
	}

	var out string
	for i, shortcut := range shortcuts {
		out, err = execute(t, a, []string{"add", shortcut.input}, "")
		require.NoError(t, err)
		require.Contains(t, out, fmt.Sprintf("Nota %d adicionada.", i+1))
		require.Contains(t, out, shortcut.expected)
		require.NotContains(t, out, shortcut.input)
	}

	out, err = execute(t, a, []string{"show"}, "")
	require.NoError(t, err)
	for _, shortcut := range shortcuts {
		require.Contains(t, out, shortcut.expected)
		require.NotContains(t, out, shortcut.input)
	}
}

func TestAddKeepsFreeText(t *testing.T) {
	a := testutil.NewTestApp(t)
	_, err := execute(t, a, []string{"new", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"use", "erp"}, "")
	require.NoError(t, err)

	out, err := execute(t, a, []string{"add", "texto livre"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "texto livre")
}

func TestClearPrompt(t *testing.T) {
	a := testutil.NewTestApp(t)
	_, err := execute(t, a, []string{"new", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"use", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"add", "x"}, "")
	require.NoError(t, err)

	out, err := execute(t, a, []string{"clear"}, "y\n")

	require.NoError(t, err)
	require.Contains(t, out, "1 nota removida.")
}

func TestClearPromptCancel(t *testing.T) {
	a := testutil.NewTestApp(t)
	_, err := execute(t, a, []string{"new", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"use", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"add", "x"}, "")
	require.NoError(t, err)

	out, err := execute(t, a, []string{"clear"}, "n\n")

	require.NoError(t, err)
	require.Contains(t, out, "Operacao cancelada.")

	out, err = execute(t, a, []string{"show"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "x")
}

func TestClearPromptEOFIsCancel(t *testing.T) {
	a := testutil.NewTestApp(t)
	_, err := execute(t, a, []string{"new", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"use", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"add", "x"}, "")
	require.NoError(t, err)

	out, err := execute(t, a, []string{"clear"}, "")

	require.NoError(t, err)
	require.Contains(t, out, "Operacao cancelada.")

	out, err = execute(t, a, []string{"show"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "x")
}

func TestCommandErrors(t *testing.T) {
	a := testutil.NewTestApp(t)

	_, err := execute(t, a, []string{"add", "x"}, "")
	require.ErrorIs(t, err, apperrors.ErrNoCurrentNotebook)
	require.True(t, IsBusinessError(err))

	_, err = execute(t, a, []string{"new", "nome invalido"}, "")
	require.ErrorIs(t, err, apperrors.ErrInvalidName)

	_, err = execute(t, a, []string{"remove", "abc"}, "")
	require.ErrorIs(t, err, apperrors.ErrInvalidID)

	_, err = execute(t, a, []string{"last", "0"}, "")
	require.ErrorIs(t, err, apperrors.ErrInvalidCount)
}

func TestFormatError(t *testing.T) {
	require.Equal(t, "Erro: x", FormatError(errors.New("x")))
	require.Empty(t, FormatError(nil))
}

func TestListCurrentAndClearEmpty(t *testing.T) {
	a := testutil.NewTestApp(t)

	out, err := execute(t, a, []string{"list"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "nenhum notebook encontrado")

	_, err = execute(t, a, []string{"current"}, "")
	require.ErrorIs(t, err, apperrors.ErrNoCurrentNotebook)

	_, err = execute(t, a, []string{"new", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, a, []string{"use", "erp"}, "")
	require.NoError(t, err)

	out, err = execute(t, a, []string{"clear"}, "")
	require.NoError(t, err)
	require.Contains(t, out, "0 notas removidas.")
}

func TestSessionIsolationBetweenApps(t *testing.T) {
	base := t.TempDir()
	appA := testutil.NewTestAppForSession(t, base, "session-a")
	appB := testutil.NewTestAppForSession(t, base, "session-b")

	_, err := execute(t, appA, []string{"new", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, appA, []string{"new", "compras"}, "")
	require.NoError(t, err)

	_, err = execute(t, appA, []string{"use", "erp"}, "")
	require.NoError(t, err)
	_, err = execute(t, appB, []string{"use", "compras"}, "")
	require.NoError(t, err)

	out, err := execute(t, appA, []string{"current"}, "")
	require.NoError(t, err)
	require.Equal(t, "erp\n", out)

	out, err = execute(t, appB, []string{"current"}, "")
	require.NoError(t, err)
	require.Equal(t, "compras\n", out)

	_, err = execute(t, appB, []string{"add", "tarefa B"}, "")
	require.NoError(t, err)

	out, err = execute(t, appA, []string{"show"}, "")
	require.NoError(t, err)
	require.NotContains(t, out, "tarefa B")
}

func TestRootVersionFlag(t *testing.T) {
	a := testutil.NewTestApp(t)

	out, err := execute(t, a, []string{"--version"}, "")

	require.NoError(t, err)
	require.Equal(t, "nb dev\n", out)
}

func execute(t *testing.T, a *app.App, args []string, stdin string) (string, error) {
	t.Helper()

	var out bytes.Buffer
	var errOut bytes.Buffer
	root := NewRootCmd(a)
	root.SetArgs(args)
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetIn(strings.NewReader(stdin))

	err := root.Execute()
	return out.String() + errOut.String(), err
}
