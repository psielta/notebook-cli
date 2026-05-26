package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"

	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/domain"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(ctx context.Context, notebookID uint, text string) (*domain.Note, error) {
	var note domain.Note

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var notebook domain.Notebook
		if err := tx.Select("id", "next_note_id").First(&notebook, notebookID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrNotebookNotFound
			}
			return err
		}

		localID := notebook.NextNoteID
		if localID <= 0 {
			localID = 1
		}

		if err := tx.Model(&domain.Notebook{}).
			Where("id = ?", notebookID).
			Update("next_note_id", localID+1).Error; err != nil {
			return err
		}

		note = domain.Note{
			NotebookID: notebookID,
			LocalID:    localID,
			Text:       text,
			CreatedAt:  time.Now().UTC(),
		}
		return tx.Create(&note).Error
	})
	if err != nil {
		return nil, err
	}

	return &note, nil
}

func (r *NoteRepository) ListByNotebook(ctx context.Context, notebookID uint) ([]domain.Note, error) {
	var notes []domain.Note
	err := r.db.WithContext(ctx).
		Where("notebook_id = ?", notebookID).
		Order("local_id ASC").
		Find(&notes).Error
	return notes, err
}

func (r *NoteRepository) ListLastByNotebook(ctx context.Context, notebookID uint, n int) ([]domain.Note, error) {
	var notes []domain.Note
	err := r.db.WithContext(ctx).
		Where("notebook_id = ?", notebookID).
		Order("local_id DESC").
		Limit(n).
		Find(&notes).Error
	if err != nil {
		return nil, err
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].LocalID < notes[j].LocalID
	})
	return notes, nil
}

func (r *NoteRepository) DeleteByLocalID(ctx context.Context, notebookID uint, localID int) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("notebook_id = ? AND local_id = ?", notebookID, localID).
		Delete(&domain.Note{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *NoteRepository) DeleteAllByNotebook(ctx context.Context, notebookID uint) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("notebook_id = ?", notebookID).
		Delete(&domain.Note{})
	return result.RowsAffected, result.Error
}
