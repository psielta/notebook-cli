package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelpers(t *testing.T) {
	db := NewTestDB(t)
	notebook := MakeNotebook(t, db, "erp")
	note := MakeNote(t, db, notebook, "x")

	require.NotZero(t, notebook.ID)
	require.Equal(t, 1, note.LocalID)

	a := NewTestApp(t)
	require.NotNil(t, a.DB)
	require.NotNil(t, a.NotebookService)
	require.NotNil(t, a.NoteService)
	require.NotNil(t, a.CurrentStore)
}
