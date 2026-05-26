package app

import (
	"gorm.io/gorm"

	"notebook-cli/internal/service"
	"notebook-cli/internal/state"
)

type App struct {
	DB              *gorm.DB
	NotebookService *service.NotebookService
	NoteService     *service.NoteService
	CurrentStore    *state.CurrentStore
}
