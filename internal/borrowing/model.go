package borrowing

import (
	"ferdinand/library-management-system-api/internal/book"
	"ferdinand/library-management-system-api/internal/member"
	"time"
)

type Borrowing struct {
	ID         uint      `gorm:"primaryKey"`
	MemberID   int64     `gorm:"not null"`
	BookID     int64     `gorm:"not null"`
	BorrowedAt time.Time `gorm:"not null"`
	DueAt      time.Time `gorm:"not null"`
	ReturnedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Member member.Member
	Book   book.Book
}
