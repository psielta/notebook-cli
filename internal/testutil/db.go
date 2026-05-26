package testutil

import (
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"notebook-cli/internal/database"
)

func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "notebook.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}
