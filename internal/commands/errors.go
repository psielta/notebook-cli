package commands

import (
	"errors"

	"notebook-cli/internal/apperrors"
)

func FormatError(err error) string {
	if err == nil {
		return ""
	}
	return "Erro: " + err.Error()
}

func IsBusinessError(err error) bool {
	return errors.Is(err, apperrors.ErrNotebookNotFound) ||
		errors.Is(err, apperrors.ErrNotebookExists) ||
		errors.Is(err, apperrors.ErrInvalidName) ||
		errors.Is(err, apperrors.ErrNoCurrentNotebook) ||
		errors.Is(err, apperrors.ErrNoteNotFound) ||
		errors.Is(err, apperrors.ErrEmptyText) ||
		errors.Is(err, apperrors.ErrInvalidID) ||
		errors.Is(err, apperrors.ErrInvalidCount) ||
		errors.Is(err, apperrors.ErrTaskLimitReached)
}
