package member

type CreateMemberRequest struct {
	Name    string  `json:"name" binding:"required"`
	Email   string  `json:"email" binding:"required"`
	Phone   string  `json:"phone" binding:"required"`
	Address *string `json:"address" binding:"required"`
}

type UpdateMemberRequest struct {
	Name    string  `json:"name" binding:"required"`
	Email   string  `json:"email" binding:"required"`
	Phone   string  `json:"phone" binding:"required"`
	Address *string `json:"address" binding:"required"`
}
