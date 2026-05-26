package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/domain"
)

var notebookNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,100}$`)

type NotebookService struct {
	notebooks NotebookRepository
	current   CurrentStore
}

func NewNotebookService(notebooks NotebookRepository, current CurrentStore) *NotebookService {
	return &NotebookService{notebooks: notebooks, current: current}
}

func (s *NotebookService) Create(ctx context.Context, name string) (*domain.Notebook, error) {
	name = strings.TrimSpace(name)
	if !notebookNamePattern.MatchString(name) {
		return nil, apperrors.ErrInvalidName
	}

	notebook := &domain.Notebook{Name: name, NextNoteID: 1}
	if err := s.notebooks.Create(ctx, notebook); err != nil {
		return nil, err
	}
	return notebook, nil
}

func (s *NotebookService) Use(ctx context.Context, name string) (*domain.Notebook, error) {
	name = strings.TrimSpace(name)
	if !notebookNamePattern.MatchString(name) {
		return nil, apperrors.ErrInvalidName
	}

	notebook, err := s.notebooks.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := s.current.Set(notebook.Name); err != nil {
		return nil, fmt.Errorf("falha ao gravar arquivo da sessao: %w", err)
	}
	return notebook, nil
}

func (s *NotebookService) Current(ctx context.Context) (string, error) {
	name, err := s.current.Get()
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", apperrors.ErrNoCurrentNotebook
	}
	if _, err := s.notebooks.GetByName(ctx, name); err != nil {
		return "", err
	}
	return name, nil
}

func (s *NotebookService) List(ctx context.Context) ([]domain.NotebookListItem, error) {
	return s.notebooks.List(ctx)
}
