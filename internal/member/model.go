package member

import "time"

type Member struct {
	ID         int64  `gorm:"primaryKey"`
	MemberCode string `gorm:"uniqueIndex;not null"`
	Name       string `gorm:"not null"`
	Email      string `gorm:"uniqueIndex;not null"`
	Phone      string
	Address    *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
