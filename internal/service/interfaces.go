package service

import (
	"context"

	"notebook-cli/internal/domain"
)

type NotebookRepository interface {
	Create(ctx context.Context, notebook *domain.Notebook) error
	GetByName(ctx context.Context, name string) (*domain.Notebook, error)
	List(ctx context.Context) ([]domain.Notebook, error)
}

type NoteRepository interface {
	Create(ctx context.Context, notebookID uint, text string) (*domain.Note, error)
	ListByNotebook(ctx context.Context, notebookID uint) ([]domain.Note, error)
	ListLastByNotebook(ctx context.Context, notebookID uint, n int) ([]domain.Note, error)
	DeleteByLocalID(ctx context.Context, notebookID uint, localID int) (bool, error)
	DeleteAllByNotebook(ctx context.Context, notebookID uint) (int64, error)
}

type CurrentStore interface {
	Get() (string, error)
	Set(name string) error
	Clear() error
}

type NotebookListItem struct {
	Name      string
	NoteCount int
}
