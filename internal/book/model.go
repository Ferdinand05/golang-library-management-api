package book

import (
	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/category"
	"time"
)

type Book struct {
	ID          int64  `gorm:"primaryKey"`
	ISBN        string `gorm:"uniqueIndex;not null"`
	Title       string `gorm:"not null"`
	Description *string
	AuthorID    int64 `gorm:"not null"`
	CategoryID  int64 `gorm:"not null"`
	Stock       int   `gorm:"not null;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Author   author.Author
	Category category.Category
}
