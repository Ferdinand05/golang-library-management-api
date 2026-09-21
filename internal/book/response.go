package book

import (
	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/category"
)

type BookResponse struct {
	ID          int64                     `json:"id"`
	ISBN        string                    `json:"isbn"`
	Title       string                    `json:"title"`
	Description *string                   `json:"description"`
	Stock       int                       `json:"stock"`
	Author      author.AuthorResponse     `json:"author"`
	Category    category.CategoryResponse `json:"category"`
}
