package borrowing

import (
	"context"
	"ferdinand/library-management-system-api/internal/book"
	"ferdinand/library-management-system-api/internal/database"
	"ferdinand/library-management-system-api/internal/member"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	FindAll(ctx context.Context) ([]Borrowing, error)
	FindByID(ctx context.Context, id int64) (Borrowing, error)
	FindByMemberID(ctx context.Context, memberID int64) ([]Borrowing, error)
	Create(ctx context.Context, request CreateBorrowingRequest) (Borrowing, error)
	Return(ctx context.Context, borrowingID int64) error
}

type service struct {
	db         *gorm.DB
	repo       Repository
	memberRepo member.Repository
	bookRepo   book.Repository
}

func NewService(db *gorm.DB, repo Repository, memberRepo member.Repository, bookRepo book.Repository) *service {
	return &service{
		db:         db,
		repo:       repo,
		memberRepo: memberRepo,
		bookRepo:   bookRepo,
	}
}

func (s *service) FindAll(ctx context.Context) ([]Borrowing, error) {

	borrowings, err := s.repo.FindAll(ctx)

	if err != nil {
		return nil, err
	}

	return borrowings, nil

}

func (s *service) FindByID(ctx context.Context, id int64) (Borrowing, error) {

	borrowing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Borrowing{}, err
	}

	return borrowing, nil

}

func (s *service) FindByMemberID(ctx context.Context, memberID int64) ([]Borrowing, error) {

	borrowings, err := s.repo.FindByMemberID(ctx, memberID)
	if err != nil {

		return nil, err
	}

	return borrowings, nil
}

func (s *service) Create(ctx context.Context, request CreateBorrowingRequest) (Borrowing, error) {

	var result Borrowing

	err := s.db.Transaction(func(tx *gorm.DB) error {

		txCtx := database.InjectTx(ctx, tx)

		// check if member exist
		_, err := s.memberRepo.FindByID(txCtx, request.MemberID)
		if err != nil {
			return err
		}

		// take book
		err = s.bookRepo.DecreaseStock(txCtx, request.BookID)
		if err != nil {
			return err
		}

		// create borrowing
		now := time.Now()

		borrowing := Borrowing{
			MemberID:   request.MemberID,
			BookID:     request.BookID,
			BorrowedAt: now,
			DueAt:      now.AddDate(0, 0, 7),
		}

		newBorrowing, err := s.repo.Create(txCtx, borrowing)
		if err != nil {
			return err
		}

		result = newBorrowing

		return nil
	})

	if err != nil {
		return Borrowing{}, err
	}

	return result, nil

}

func (s *service) Return(ctx context.Context, borrowingID int64) error {

	err := s.db.Transaction(func(tx *gorm.DB) error {

		txCtx := database.InjectTx(ctx, tx)

		// check borrowing
		borrowing, err := s.repo.FindByID(txCtx, borrowingID)
		if err != nil {
			return err
		}

		// return book
		err = s.repo.Return(txCtx, borrowingID)
		if err != nil {
			return err
		}

		// increase book stock
		err = s.bookRepo.IncreaseStock(txCtx, borrowing.BookID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
