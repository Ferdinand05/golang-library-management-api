package book

type CreateBookRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	AuthorID    int64   `json:"author_id" binding:"required"`
	CategoryID  int64   `json:"category_id" binding:"required"`
	Stock       int     `json:"stock" binding:"required,gt=0"`
	ISBN        string  `json:"isbn" binding:"required"`
}

type UpdateBookRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description" binding:"required"`
	AuthorID    int64   `json:"author_id" binding:"required"`
	CategoryID  int64   `json:"category_id" binding:"required"`
	Stock       int     `json:"stock" binding:"required,gt=0"`
}
