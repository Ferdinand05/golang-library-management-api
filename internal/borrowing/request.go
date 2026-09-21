package borrowing

type CreateBorrowingRequest struct {
	MemberID int64 `json:"member_id" binding:"required"`
	BookID   int64 `json:"book_id" binding:"required"`
}
