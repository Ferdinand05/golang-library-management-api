package author

type CreateAuthorRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateAuthorRequest struct {
	Name string `json:"name" binding:"required"`
}
