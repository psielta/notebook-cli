package testutil

import (
	"path/filepath"
	"testing"

	"notebook-cli/internal/app"
	"notebook-cli/internal/database"
	"notebook-cli/internal/repository"
	"notebook-cli/internal/service"
	"notebook-cli/internal/state"
)

func NewTestApp(t *testing.T) *app.App {
	return NewTestAppForSession(t, t.TempDir(), "test-session")
}

func NewTestAppForSession(t *testing.T, baseDir, sessionID string) *app.App {
	t.Helper()

	db, err := database.Open(filepath.Join(baseDir, "notebook.db"))
	if err != nil {
		t.Fatalf("open test app db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	notebookRepo := repository.NewNotebookRepository(db)
	noteRepo := repository.NewNoteRepository(db)
	current := state.New(baseDir, sessionID)

	return &app.App{
		DB:              db,
		NotebookService: service.NewNotebookService(notebookRepo, current),
		NoteService:     service.NewNoteService(noteRepo, notebookRepo, current),
		CurrentStore:    current,
	}
}
