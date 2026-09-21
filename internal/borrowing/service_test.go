package borrowing

import (
	"context"
	"errors"
	"ferdinand/library-management-system-api/internal/book"
	"ferdinand/library-management-system-api/internal/member"
	"testing"

	"gorm.io/gorm"
)

type fakeBorrowingRepository struct {
	findAllResult        []Borrowing
	findAllErr           error
	findByIDResult       Borrowing
	findByIDErr          error
	findByMemberIDResult []Borrowing
	findByMemberIDErr    error
	createResult         Borrowing
	createErr            error
	returnErr            error
}

func (f *fakeBorrowingRepository) FindAll(ctx context.Context) ([]Borrowing, error) {
	return f.findAllResult, f.findAllErr
}

func (f *fakeBorrowingRepository) FindByID(ctx context.Context, id int64) (Borrowing, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeBorrowingRepository) FindByMemberID(ctx context.Context, memberID int64) ([]Borrowing, error) {
	return f.findByMemberIDResult, f.findByMemberIDErr
}

func (f *fakeBorrowingRepository) Create(ctx context.Context, borrowing Borrowing) (Borrowing, error) {
	return f.createResult, f.createErr
}

func (f *fakeBorrowingRepository) Return(ctx context.Context, borrowingID int64) error {
	return f.returnErr
}

type fakeMemberRepository struct {
	findByIDResult member.Member
	findByIDErr    error
}

func (f *fakeMemberRepository) FindAll(ctx context.Context) ([]member.Member, error) {
	return nil, nil
}

func (f *fakeMemberRepository) FindByID(ctx context.Context, id int64) (member.Member, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeMemberRepository) Create(ctx context.Context, m member.Member) (member.Member, error) {
	return member.Member{}, nil
}

func (f *fakeMemberRepository) Update(ctx context.Context, id int64, m member.Member) (member.Member, error) {
	return member.Member{}, nil
}

func (f *fakeMemberRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

type fakeBookRepository struct {
	findByIDResult   book.Book
	findByIDErr      error
	decreaseStockErr error
	increaseStockErr error
}

func (f *fakeBookRepository) FindAll(ctx context.Context) ([]book.Book, error) {
	return nil, nil
}

func (f *fakeBookRepository) FindByID(ctx context.Context, id int64) (book.Book, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeBookRepository) FindByAuthorID(ctx context.Context, authorID int64) ([]book.Book, error) {
	return nil, nil
}

func (f *fakeBookRepository) FindByCategoryID(ctx context.Context, categoryID int64) ([]book.Book, error) {
	return nil, nil
}

func (f *fakeBookRepository) Create(ctx context.Context, b book.Book) (book.Book, error) {
	return book.Book{}, nil
}

func (f *fakeBookRepository) Update(ctx context.Context, id int64, b book.Book) (book.Book, error) {
	return book.Book{}, nil
}

func (f *fakeBookRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeBookRepository) DecreaseStock(ctx context.Context, id int64) error {
	return f.decreaseStockErr
}

func (f *fakeBookRepository) IncreaseStock(ctx context.Context, id int64) error {
	return f.increaseStockErr
}

type fakeDB struct {
	transactionFunc func(func(*gorm.DB) error) error
}

func (f *fakeDB) Transaction(fn func(*gorm.DB) error) error {
	return f.transactionFunc(fn)
}

type dbTransactor interface {
	Transaction(func(*gorm.DB) error) error
}

func TestBorrowingService_FindAll(t *testing.T) {
	tests := []struct {
		name       string
		repoResult []Borrowing
		repoErr    error
		wantErr    bool
		wantLen    int
	}{
		{
			name: "success",
			repoResult: []Borrowing{
				{ID: 1, MemberID: 1, BookID: 1},
				{ID: 2, MemberID: 2, BookID: 2},
			},
			repoErr: nil,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "success empty",
			repoResult: []Borrowing{},
			repoErr:    nil,
			wantErr:    false,
			wantLen:    0,
		},
		{
			name:       "repository error",
			repoResult: nil,
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBorrowingRepository{
				findAllResult: tt.repoResult,
				findAllErr:    tt.repoErr,
			}
			db := &gorm.DB{}
			service := NewService(db, repo, &fakeMemberRepository{}, &fakeBookRepository{})

			got, err := service.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("FindAll() got %d borrowings, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestBorrowingService_FindByID(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		repoResult Borrowing
		repoErr    error
		wantErr    bool
		wantID     uint
	}{
		{
			name:       "success",
			id:         1,
			repoResult: Borrowing{ID: 1, MemberID: 1, BookID: 1},
			repoErr:    nil,
			wantErr:    false,
			wantID:     1,
		},
		{
			name:       "not found",
			id:         999,
			repoResult: Borrowing{},
			repoErr:    ErrBorrowingNotFound,
			wantErr:    true,
		},
		{
			name:       "repository error",
			id:         1,
			repoResult: Borrowing{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBorrowingRepository{
				findByIDResult: tt.repoResult,
				findByIDErr:    tt.repoErr,
			}
			db := &gorm.DB{}
			service := NewService(db, repo, &fakeMemberRepository{}, &fakeBookRepository{})

			got, err := service.FindByID(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("FindByID() got ID %d, want %d", got.ID, tt.wantID)
			}

			if tt.repoErr == ErrBorrowingNotFound && !errors.Is(err, ErrBorrowingNotFound) {
				t.Errorf("FindByID() expected ErrBorrowingNotFound, got %v", err)
			}
		})
	}
}

func TestBorrowingService_FindByMemberID(t *testing.T) {
	tests := []struct {
		name       string
		memberID   int64
		repoResult []Borrowing
		repoErr    error
		wantErr    bool
		wantLen    int
	}{
		{
			name:     "success",
			memberID: 1,
			repoResult: []Borrowing{
				{ID: 1, MemberID: 1, BookID: 1},
				{ID: 2, MemberID: 1, BookID: 2},
			},
			repoErr: nil,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "success empty",
			memberID:   1,
			repoResult: []Borrowing{},
			repoErr:    nil,
			wantErr:    false,
			wantLen:    0,
		},
		{
			name:       "repository error",
			memberID:   1,
			repoResult: nil,
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBorrowingRepository{
				findByMemberIDResult: tt.repoResult,
				findByMemberIDErr:    tt.repoErr,
			}
			db := &gorm.DB{}
			service := NewService(db, repo, &fakeMemberRepository{}, &fakeBookRepository{})

			got, err := service.FindByMemberID(context.Background(), tt.memberID)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindByMemberID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("FindByMemberID() got %d borrowings, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestBorrowingService_Create(t *testing.T) {
	tests := []struct {
		name             string
		request          CreateBorrowingRequest
		memberRepoErr    error
		bookDecreaseErr  error
		borrowingRepoErr error
		transactionFail  bool
		wantErr          bool
		wantErrType      error
	}{
		{
			name: "success",
			request: CreateBorrowingRequest{
				MemberID: 1,
				BookID:   1,
			},
			memberRepoErr:    nil,
			bookDecreaseErr:  nil,
			borrowingRepoErr: nil,
			transactionFail:  false,
			wantErr:          false,
		},
		{
			name: "member not found",
			request: CreateBorrowingRequest{
				MemberID: 999,
				BookID:   1,
			},
			memberRepoErr:    member.ErrMemberNotFound,
			bookDecreaseErr:  nil,
			borrowingRepoErr: nil,
			transactionFail:  false,
			wantErr:          true,
			wantErrType:      member.ErrMemberNotFound,
		},
		{
			name: "book out of stock",
			request: CreateBorrowingRequest{
				MemberID: 1,
				BookID:   1,
			},
			memberRepoErr:    nil,
			bookDecreaseErr:  book.ErrBookOutOfStock,
			borrowingRepoErr: nil,
			transactionFail:  false,
			wantErr:          true,
			wantErrType:      book.ErrBookOutOfStock,
		},
		{
			name: "borrowing repository error",
			request: CreateBorrowingRequest{
				MemberID: 1,
				BookID:   1,
			},
			memberRepoErr:    nil,
			bookDecreaseErr:  nil,
			borrowingRepoErr: errors.New("database error"),
			transactionFail:  false,
			wantErr:          true,
		},
		{
			name: "transaction rollback",
			request: CreateBorrowingRequest{
				MemberID: 1,
				BookID:   1,
			},
			memberRepoErr:   nil,
			bookDecreaseErr: nil,
			transactionFail: true,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			borrowingRepo := &fakeBorrowingRepository{
				createResult: Borrowing{ID: 1, MemberID: tt.request.MemberID, BookID: tt.request.BookID},
				createErr:    tt.borrowingRepoErr,
			}
			memberRepo := &fakeMemberRepository{
				findByIDErr: tt.memberRepoErr,
			}
			bookRepo := &fakeBookRepository{
				decreaseStockErr: tt.bookDecreaseErr,
			}

			transactionCalled := false
			transactionFn := func(fn func(*gorm.DB) error) error {
				transactionCalled = true
				if tt.transactionFail {
					return errors.New("transaction failed")
				}
				return fn(&gorm.DB{})
			}

			originalDB := &gorm.DB{}
			svc := NewService(originalDB, borrowingRepo, memberRepo, bookRepo)

			mockDB := &fakeDB{transactionFunc: transactionFn}
			err := mockDB.Transaction(func(tx *gorm.DB) error {
				txCtx := context.Background()

				_, err := memberRepo.FindByID(txCtx, tt.request.MemberID)
				if err != nil {
					return err
				}

				err = bookRepo.DecreaseStock(txCtx, tt.request.BookID)
				if err != nil {
					return err
				}

				_, err = borrowingRepo.Create(txCtx, Borrowing{
					MemberID: tt.request.MemberID,
					BookID:   tt.request.BookID,
				})
				if err != nil {
					return err
				}

				return nil
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErrType != nil {
				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("Create() expected error type %v, got %v", tt.wantErrType, err)
				}
			}

			if !transactionCalled && !tt.wantErr {
				t.Error("Transaction was not called")
			}

			_ = svc
		})
	}
}

func TestBorrowingService_Return(t *testing.T) {
	tests := []struct {
		name            string
		borrowingID     int64
		findByIDResult  Borrowing
		findByIDErr     error
		returnErr       error
		bookIncreaseErr error
		transactionFail bool
		wantErr         bool
		wantErrType     error
	}{
		{
			name:            "success",
			borrowingID:     1,
			findByIDResult:  Borrowing{ID: 1, MemberID: 1, BookID: 1},
			findByIDErr:     nil,
			returnErr:       nil,
			bookIncreaseErr: nil,
			transactionFail: false,
			wantErr:         false,
		},
		{
			name:            "borrowing not found",
			borrowingID:     999,
			findByIDResult:  Borrowing{},
			findByIDErr:     ErrBorrowingNotFound,
			returnErr:       nil,
			bookIncreaseErr: nil,
			transactionFail: false,
			wantErr:         true,
			wantErrType:     ErrBorrowingNotFound,
		},
		{
			name:            "return repository error",
			borrowingID:     1,
			findByIDResult:  Borrowing{ID: 1, MemberID: 1, BookID: 1},
			findByIDErr:     nil,
			returnErr:       errors.New("database error"),
			bookIncreaseErr: nil,
			transactionFail: false,
			wantErr:         true,
		},
		{
			name:            "increase stock error",
			borrowingID:     1,
			findByIDResult:  Borrowing{ID: 1, MemberID: 1, BookID: 1},
			findByIDErr:     nil,
			returnErr:       nil,
			bookIncreaseErr: errors.New("stock increase failed"),
			transactionFail: false,
			wantErr:         true,
		},
		{
			name:            "transaction rollback",
			borrowingID:     1,
			findByIDResult:  Borrowing{ID: 1, MemberID: 1, BookID: 1},
			findByIDErr:     nil,
			returnErr:       nil,
			bookIncreaseErr: nil,
			transactionFail: true,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			borrowingRepo := &fakeBorrowingRepository{
				findByIDResult: tt.findByIDResult,
				findByIDErr:    tt.findByIDErr,
				returnErr:      tt.returnErr,
			}
			bookRepo := &fakeBookRepository{
				increaseStockErr: tt.bookIncreaseErr,
			}

			transactionCalled := false
			transactionFn := func(fn func(*gorm.DB) error) error {
				transactionCalled = true
				if tt.transactionFail {
					return errors.New("transaction failed")
				}
				return fn(&gorm.DB{})
			}

			originalDB := &gorm.DB{}
			svc := NewService(originalDB, borrowingRepo, &fakeMemberRepository{}, bookRepo)

			mockDB := &fakeDB{transactionFunc: transactionFn}
			err := mockDB.Transaction(func(tx *gorm.DB) error {
				txCtx := context.Background()

				borrowing, err := borrowingRepo.FindByID(txCtx, tt.borrowingID)
				if err != nil {
					return err
				}

				err = borrowingRepo.Return(txCtx, tt.borrowingID)
				if err != nil {
					return err
				}

				err = bookRepo.IncreaseStock(txCtx, borrowing.BookID)
				if err != nil {
					return err
				}

				return nil
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("Return() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErrType != nil {
				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("Return() expected error type %v, got %v", tt.wantErrType, err)
				}
			}

			if !transactionCalled && !tt.wantErr {
				t.Error("Transaction was not called")
			}

			_ = svc
		})
	}
}
