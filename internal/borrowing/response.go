package borrowing

import "time"

type BorrowingResponse struct {
	ID         int64          `json:"id"`
	Member     MemberResponse `json:"member"`
	Book       BookResponse   `json:"book"`
	BorrowedAt time.Time      `json:"borrowed_at"`
	DueAt      time.Time      `json:"due_at"`
	ReturnedAt *time.Time     `json:"returned_at"`
}

type MemberResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"member_code"`
	Name string `json:"name"`
}

type BookResponse struct {
	ID    int64  `json:"id"`
	ISBN  string `json:"isbn"`
	Title string `json:"title"`
}
