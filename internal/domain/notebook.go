package domain

import "time"

type Notebook struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"uniqueIndex;size:100;not null"`
	NextNoteID int    `gorm:"not null;default:1"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Notes      []Note `gorm:"foreignKey:NotebookID;constraint:OnDelete:CASCADE"`
}
