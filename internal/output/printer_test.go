package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"notebook-cli/internal/domain"
)

func TestPrintNotebooks(t *testing.T) {
	var out bytes.Buffer
	err := NewPrinter(&out).PrintNotebooks([]domain.NotebookListItem{{Name: "erp", NoteCount: 2}})

	require.NoError(t, err)
	require.Contains(t, out.String(), "NOME")
	require.Contains(t, out.String(), "erp")
	require.Contains(t, out.String(), "2")
}

func TestPrintNotes(t *testing.T) {
	var out bytes.Buffer
	err := NewPrinter(&out).PrintNotes([]domain.Note{{
		LocalID:   1,
		Text:      "linha 1\nlinha 2",
		CreatedAt: time.Date(2026, 5, 26, 8, 0, 0, 0, time.Local),
	}})

	require.NoError(t, err)
	require.Contains(t, out.String(), "ID")
	require.Contains(t, out.String(), "linha 1 linha 2")
}

func TestPrintEmpty(t *testing.T) {
	var out bytes.Buffer
	printer := NewPrinter(&out)

	require.NoError(t, printer.PrintNotebooks(nil))
	require.Contains(t, out.String(), "nenhum notebook encontrado")

	out.Reset()
	require.NoError(t, printer.PrintNotes(nil))
	require.Contains(t, out.String(), "nenhuma nota")
}
