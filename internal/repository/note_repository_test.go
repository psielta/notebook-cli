package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"notebook-cli/internal/domain"
	"notebook-cli/internal/repository"
	"notebook-cli/internal/testutil"
)

func TestNoteRepositoryCreateMonotonicLocalID(t *testing.T) {
	db := testutil.NewTestDB(t)
	notebook := testutil.MakeNotebook(t, db, "erp")
	repo := repository.NewNoteRepository(db)

	first, err := repo.Create(context.Background(), notebook.ID, "um")
	require.NoError(t, err)
	second, err := repo.Create(context.Background(), notebook.ID, "dois")
	require.NoError(t, err)
	third, err := repo.Create(context.Background(), notebook.ID, "tres")
	require.NoError(t, err)

	require.Equal(t, []int{1, 2, 3}, []int{first.LocalID, second.LocalID, third.LocalID})

	ok, err := repo.DeleteByLocalID(context.Background(), notebook.ID, 3)
	require.NoError(t, err)
	require.True(t, ok)

	fourth, err := repo.Create(context.Background(), notebook.ID, "quatro")
	require.NoError(t, err)
	require.Equal(t, 4, fourth.LocalID)

	var saved domain.Notebook
	require.NoError(t, db.First(&saved, notebook.ID).Error)
	require.Equal(t, 5, saved.NextNoteID)
}

func TestNoteRepositoryCreateRollsBackCounterOnInsertFailure(t *testing.T) {
	db := testutil.NewTestDB(t)
	notebook := testutil.MakeNotebook(t, db, "erp")
	repo := repository.NewNoteRepository(db)

	require.NoError(t, db.Create(&domain.Note{
		NotebookID: notebook.ID,
		LocalID:    1,
		Text:       "preexistente",
	}).Error)

	_, err := repo.Create(context.Background(), notebook.ID, "vai falhar")
	require.Error(t, err)

	var saved domain.Notebook
	require.NoError(t, db.First(&saved, notebook.ID).Error)
	require.Equal(t, 1, saved.NextNoteID)
}

func TestNoteRepositoryListLastAndClear(t *testing.T) {
	db := testutil.NewTestDB(t)
	notebook := testutil.MakeNotebook(t, db, "erp")
	repo := repository.NewNoteRepository(db)

	_, _ = repo.Create(context.Background(), notebook.ID, "um")
	_, _ = repo.Create(context.Background(), notebook.ID, "dois")
	_, _ = repo.Create(context.Background(), notebook.ID, "tres")

	last, err := repo.ListLastByNotebook(context.Background(), notebook.ID, 2)
	require.NoError(t, err)
	require.Equal(t, []int{2, 3}, []int{last[0].LocalID, last[1].LocalID})

	count, err := repo.DeleteAllByNotebook(context.Background(), notebook.ID)
	require.NoError(t, err)
	require.Equal(t, int64(3), count)

	next, err := repo.Create(context.Background(), notebook.ID, "quatro")
	require.NoError(t, err)
	require.Equal(t, 4, next.LocalID)
}

func TestNoteRepositoryMissingNotebookAndDeleteMissing(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.NewNoteRepository(db)

	_, err := repo.Create(context.Background(), 999, "x")
	require.Error(t, err)

	ok, err := repo.DeleteByLocalID(context.Background(), 999, 1)
	require.NoError(t, err)
	require.False(t, ok)

	notes, err := repo.ListByNotebook(context.Background(), 999)
	require.NoError(t, err)
	require.Empty(t, notes)
}

func TestNoteRepositoryDBErrors(t *testing.T) {
	db := testutil.NewTestDB(t)
	notebook := testutil.MakeNotebook(t, db, "erp")
	repo := repository.NewNoteRepository(db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	_, err = repo.ListByNotebook(context.Background(), notebook.ID)
	require.Error(t, err)

	_, err = repo.ListLastByNotebook(context.Background(), notebook.ID, 1)
	require.Error(t, err)

	_, err = repo.DeleteByLocalID(context.Background(), notebook.ID, 1)
	require.Error(t, err)

	_, err = repo.DeleteAllByNotebook(context.Background(), notebook.ID)
	require.Error(t, err)
}
