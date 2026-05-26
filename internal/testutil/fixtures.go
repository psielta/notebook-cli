package testutil

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"notebook-cli/internal/domain"
	"notebook-cli/internal/repository"
)

func MakeNotebook(t *testing.T, db *gorm.DB, name string) *domain.Notebook {
	t.Helper()

	notebook := &domain.Notebook{Name: name, NextNoteID: 1}
	if err := db.Create(notebook).Error; err != nil {
		t.Fatalf("create notebook fixture: %v", err)
	}
	return notebook
}

func MakeNote(t *testing.T, db *gorm.DB, notebook *domain.Notebook, text string) *domain.Note {
	t.Helper()

	note, err := repository.NewNoteRepository(db).Create(context.Background(), notebook.ID, text)
	if err != nil {
		t.Fatalf("create note fixture: %v", err)
	}
	return note
}
