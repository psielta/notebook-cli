package repository

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/domain"
)

type NotebookRepository struct {
	db *gorm.DB
}

func NewNotebookRepository(db *gorm.DB) *NotebookRepository {
	return &NotebookRepository{db: db}
}

func (r *NotebookRepository) Create(ctx context.Context, notebook *domain.Notebook) error {
	err := r.db.WithContext(ctx).Create(notebook).Error
	if isUniqueConstraint(err) {
		return apperrors.ErrNotebookExists
	}
	return err
}

func (r *NotebookRepository) GetByName(ctx context.Context, name string) (*domain.Notebook, error) {
	var notebook domain.Notebook
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&notebook).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotebookNotFound
	}
	if err != nil {
		return nil, err
	}
	return &notebook, nil
}

func (r *NotebookRepository) List(ctx context.Context) ([]domain.Notebook, error) {
	var notebooks []domain.Notebook
	err := r.db.WithContext(ctx).Preload("Notes").Order("name ASC").Find(&notebooks).Error
	return notebooks, err
}

func (r *NotebookRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&domain.Notebook{}, id)
	return result.Error
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
