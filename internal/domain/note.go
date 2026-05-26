package domain

import "time"

type Note struct {
	ID         uint   `gorm:"primaryKey"`
	NotebookID uint   `gorm:"not null;index;uniqueIndex:idx_nb_local"`
	LocalID    int    `gorm:"not null;uniqueIndex:idx_nb_local"`
	Text       string `gorm:"type:text;not null"`
	CreatedAt  time.Time
}
