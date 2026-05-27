package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/domain"
)

var notebookNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,100}$`)

const taskTemplateName = "task"

type NotebookService struct {
	notebooks NotebookRepository
	current   CurrentStore
	now       func() time.Time
}

func NewNotebookService(notebooks NotebookRepository, current CurrentStore) *NotebookService {
	return newNotebookService(notebooks, current, time.Now)
}

func newNotebookService(notebooks NotebookRepository, current CurrentStore, now func() time.Time) *NotebookService {
	return &NotebookService{notebooks: notebooks, current: current, now: now}
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

func (s *NotebookService) CreateTask(ctx context.Context) (*domain.Notebook, error) {
	items, err := s.notebooks.List(ctx)
	if err != nil {
		return nil, err
	}

	today := s.now().Local().Format("2006-01-02")
	next, err := nextTaskNumber(items, today)
	if err != nil {
		return nil, err
	}

	notebook := &domain.Notebook{
		Name:       fmt.Sprintf("T%03d-%s", next, today),
		NextNoteID: 1,
	}
	if err := s.notebooks.Create(ctx, notebook); err != nil {
		return nil, err
	}
	if err := s.current.Set(notebook.Name); err != nil {
		return nil, fmt.Errorf("falha ao gravar arquivo da sessao: %w", err)
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

func IsTaskTemplateName(name string) bool {
	return strings.TrimSpace(name) == taskTemplateName
}

func nextTaskNumber(items []domain.NotebookListItem, date string) (int, error) {
	maxNumber := 0
	suffix := "-" + date
	for _, item := range items {
		name := item.Name
		if len(name) != len("T001-2006-01-02") || name[0] != 'T' || !strings.HasSuffix(name, suffix) {
			continue
		}

		number, err := strconv.Atoi(name[1:4])
		if err != nil {
			continue
		}
		if number > maxNumber {
			maxNumber = number
		}
	}

	if maxNumber >= 999 {
		return 0, apperrors.ErrTaskLimitReached
	}
	return maxNumber + 1, nil
}
