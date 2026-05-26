package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/domain"
	"notebook-cli/internal/repository"
	"notebook-cli/internal/testutil"
)

func TestNotebookRepositoryCreateListAndUnique(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.NewNotebookRepository(db)

	err := repo.Create(context.Background(), &domain.Notebook{Name: "erp", NextNoteID: 1})
	require.NoError(t, err)

	err = repo.Create(context.Background(), &domain.Notebook{Name: "erp", NextNoteID: 1})
	require.ErrorIs(t, err, apperrors.ErrNotebookExists)

	notebooks, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, notebooks, 1)
	require.Equal(t, "erp", notebooks[0].Name)
}

func TestNotebookRepositoryGetByNameMissing(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.NewNotebookRepository(db)

	_, err := repo.GetByName(context.Background(), "missing")

	require.ErrorIs(t, err, apperrors.ErrNotebookNotFound)
}

func TestNotebookRepositoryGetByNameSuccessAndDBError(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.NewNotebookRepository(db)
	notebook := &domain.Notebook{Name: "erp", NextNoteID: 1}
	require.NoError(t, repo.Create(context.Background(), notebook))

	found, err := repo.GetByName(context.Background(), "erp")
	require.NoError(t, err)
	require.Equal(t, notebook.ID, found.ID)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	_, err = repo.GetByName(context.Background(), "erp")
	require.Error(t, err)
}

func TestNotebookRepositoryDeleteCascadesNotes(t *testing.T) {
	db := testutil.NewTestDB(t)
	notebookRepo := repository.NewNotebookRepository(db)
	noteRepo := repository.NewNoteRepository(db)

	notebook := &domain.Notebook{Name: "erp", NextNoteID: 1}
	require.NoError(t, notebookRepo.Create(context.Background(), notebook))
	_, err := noteRepo.Create(context.Background(), notebook.ID, "x")
	require.NoError(t, err)

	require.NoError(t, notebookRepo.Delete(context.Background(), notebook.ID))

	notes, err := noteRepo.ListByNotebook(context.Background(), notebook.ID)
	require.NoError(t, err)
	require.Empty(t, notes)
}
