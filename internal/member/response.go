package member

type MemberResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	MemberCode string `json:"member_code"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
}
