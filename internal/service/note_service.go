package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/domain"
)

type NoteService struct {
	notes     NoteRepository
	notebooks NotebookRepository
	current   CurrentStore
}

func NewNoteService(notes NoteRepository, notebooks NotebookRepository, current CurrentStore) *NoteService {
	return &NoteService{notes: notes, notebooks: notebooks, current: current}
}

func (s *NoteService) Add(ctx context.Context, text string) (*domain.Note, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, apperrors.ErrEmptyText
	}

	notebook, err := s.currentNotebook(ctx)
	if err != nil {
		return nil, err
	}

	return s.notes.Create(ctx, notebook.ID, text)
}

func (s *NoteService) List(ctx context.Context) ([]domain.Note, error) {
	notebook, err := s.currentNotebook(ctx)
	if err != nil {
		return nil, err
	}
	return s.notes.ListByNotebook(ctx, notebook.ID)
}

func (s *NoteService) Last(ctx context.Context, n int) ([]domain.Note, error) {
	if n <= 0 {
		return nil, apperrors.ErrInvalidCount
	}

	notebook, err := s.currentNotebook(ctx)
	if err != nil {
		return nil, err
	}
	return s.notes.ListLastByNotebook(ctx, notebook.ID, n)
}

func (s *NoteService) Remove(ctx context.Context, localID int) error {
	if localID <= 0 {
		return apperrors.ErrInvalidID
	}

	notebook, err := s.currentNotebook(ctx)
	if err != nil {
		return err
	}

	ok, err := s.notes.DeleteByLocalID(ctx, notebook.ID, localID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: %d", apperrors.ErrNoteNotFound, localID)
	}
	return nil
}

func (s *NoteService) Clear(ctx context.Context) (int64, error) {
	notebook, err := s.currentNotebook(ctx)
	if err != nil {
		return 0, err
	}
	return s.notes.DeleteAllByNotebook(ctx, notebook.ID)
}

func (s *NoteService) currentNotebook(ctx context.Context) (*domain.Notebook, error) {
	name, err := s.current.Get()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, apperrors.ErrNoCurrentNotebook
	}

	notebook, err := s.notebooks.GetByName(ctx, name)
	if errors.Is(err, apperrors.ErrNotebookNotFound) {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrNotebookNotFound, name)
	}
	return notebook, err
}
