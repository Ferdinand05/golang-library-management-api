package book

import (
	"context"
	"errors"
	"testing"
	"time"

	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/category"
	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type fakeCacheClient struct{}

func (f *fakeCacheClient) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, redisCache.ErrCacheMiss
}

func (f *fakeCacheClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (f *fakeCacheClient) Delete(ctx context.Context, keys ...string) error {
	return nil
}

type fakeBookRepository struct {
	findAllResult          []Book
	findAllErr             error
	findByIDResult         Book
	findByIDErr            error
	findByAuthorIDResult   []Book
	findByAuthorIDErr      error
	findByCategoryIDResult []Book
	findByCategoryIDErr    error
	createResult           Book
	createErr              error
	updateResult           Book
	updateErr              error
	deleteErr              error
}

func (f *fakeBookRepository) FindAll(ctx context.Context) ([]Book, error) {
	return f.findAllResult, f.findAllErr
}

func (f *fakeBookRepository) FindByID(ctx context.Context, id int64) (Book, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeBookRepository) FindByAuthorID(ctx context.Context, authorID int64) ([]Book, error) {
	return f.findByAuthorIDResult, f.findByAuthorIDErr
}

func (f *fakeBookRepository) FindByCategoryID(ctx context.Context, categoryID int64) ([]Book, error) {
	return f.findByCategoryIDResult, f.findByCategoryIDErr
}

func (f *fakeBookRepository) Create(ctx context.Context, book Book) (Book, error) {
	return f.createResult, f.createErr
}

func (f *fakeBookRepository) Update(ctx context.Context, id int64, book Book) (Book, error) {
	return f.updateResult, f.updateErr
}

func (f *fakeBookRepository) Delete(ctx context.Context, id int64) error {
	return f.deleteErr
}

func (f *fakeBookRepository) DecreaseStock(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeBookRepository) IncreaseStock(ctx context.Context, id int64) error {
	return nil
}

type fakeAuthorRepository struct {
	findByIDResult author.Author
	findByIDErr    error
}

func (f *fakeAuthorRepository) FindAll(ctx context.Context) ([]author.Author, error) {
	return nil, nil
}

func (f *fakeAuthorRepository) FindByID(ctx context.Context, id int64) (author.Author, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeAuthorRepository) Create(ctx context.Context, a author.Author) (author.Author, error) {
	return author.Author{}, nil
}

func (f *fakeAuthorRepository) Update(ctx context.Context, id int64, a author.Author) (author.Author, error) {
	return author.Author{}, nil
}

func (f *fakeAuthorRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

type fakeCategoryRepository struct {
	findByIDResult category.Category
	findByIDErr    error
}

func (f *fakeCategoryRepository) FindAll(ctx context.Context) ([]category.Category, error) {
	return nil, nil
}

func (f *fakeCategoryRepository) FindByID(ctx context.Context, id int64) (category.Category, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeCategoryRepository) Create(ctx context.Context, c category.Category) (category.Category, error) {
	return category.Category{}, nil
}

func (f *fakeCategoryRepository) Update(ctx context.Context, id int64, c category.Category) (category.Category, error) {
	return category.Category{}, nil
}

func (f *fakeCategoryRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func TestBookService_FindAll(t *testing.T) {
	tests := []struct {
		name       string
		repoResult []Book
		repoErr    error
		wantErr    bool
		wantLen    int
	}{
		{
			name: "success",
			repoResult: []Book{
				{ID: 1, Title: "Book 1"},
				{ID: 2, Title: "Book 2"},
			},
			repoErr: nil,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "success empty",
			repoResult: []Book{},
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
			repo := &fakeBookRepository{
				findAllResult: tt.repoResult,
				findAllErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeAuthorRepository{}, &fakeCategoryRepository{}, &fakeCacheClient{})

			got, err := service.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("FindAll() got %d books, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestBookService_FindByID(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		repoResult Book
		repoErr    error
		wantErr    bool
		wantID     int64
	}{
		{
			name:       "success",
			id:         1,
			repoResult: Book{ID: 1, Title: "Test Book"},
			repoErr:    nil,
			wantErr:    false,
			wantID:     1,
		},
		{
			name:       "not found",
			id:         999,
			repoResult: Book{},
			repoErr:    ErrBookNotFound,
			wantErr:    true,
		},
		{
			name:       "repository error",
			id:         1,
			repoResult: Book{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookRepository{
				findByIDResult: tt.repoResult,
				findByIDErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeAuthorRepository{}, &fakeCategoryRepository{}, &fakeCacheClient{})

			got, err := service.FindByID(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("FindByID() got ID %d, want %d", got.ID, tt.wantID)
			}

			if tt.repoErr == ErrBookNotFound && !errors.Is(err, ErrBookNotFound) {
				t.Errorf("FindByID() expected ErrBookNotFound, got %v", err)
			}
		})
	}
}

func TestBookService_FindByAuthorID(t *testing.T) {
	tests := []struct {
		name           string
		authorID       int64
		authorRepoErr  error
		bookRepoResult []Book
		bookRepoErr    error
		wantErr        error
		wantLen        int
	}{
		{
			name:          "success",
			authorID:      1,
			authorRepoErr: nil,
			bookRepoResult: []Book{
				{ID: 1, Title: "Book 1", AuthorID: 1},
				{ID: 2, Title: "Book 2", AuthorID: 1},
			},
			bookRepoErr: nil,
			wantErr:     nil,
			wantLen:     2,
		},
		{
			name:           "success empty",
			authorID:       1,
			authorRepoErr:  nil,
			bookRepoResult: []Book{},
			bookRepoErr:    nil,
			wantErr:        nil,
			wantLen:        0,
		},
		{
			name:           "author not found",
			authorID:       999,
			authorRepoErr:  author.ErrAuthorNotFound,
			bookRepoResult: nil,
			bookRepoErr:    nil,
			wantErr:        author.ErrAuthorNotFound,
		},
		{
			name:           "author repository error",
			authorID:       1,
			authorRepoErr:  errors.New("database error"),
			bookRepoResult: nil,
			bookRepoErr:    nil,
			wantErr:        errors.New("database error"),
		},
		{
			name:           "book repository error",
			authorID:       1,
			authorRepoErr:  nil,
			bookRepoResult: nil,
			bookRepoErr:    errors.New("database error"),
			wantErr:        errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookRepo := &fakeBookRepository{
				findByAuthorIDResult: tt.bookRepoResult,
				findByAuthorIDErr:    tt.bookRepoErr,
			}
			authorRepo := &fakeAuthorRepository{
				findByIDErr: tt.authorRepoErr,
			}
			service := NewService(bookRepo, authorRepo, &fakeCategoryRepository{}, &fakeCacheClient{})

			got, err := service.FindByAuthorID(context.Background(), tt.authorID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("FindByAuthorID() error = nil, wantErr %v", tt.wantErr)
				}
				if tt.wantErr == author.ErrAuthorNotFound && !errors.Is(err, author.ErrAuthorNotFound) {
					t.Errorf("FindByAuthorID() expected ErrAuthorNotFound, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("FindByAuthorID() unexpected error = %v", err)
				}
				if len(got) != tt.wantLen {
					t.Errorf("FindByAuthorID() got %d books, want %d", len(got), tt.wantLen)
				}
			}
		})
	}
}

func TestBookService_FindByCategoryID(t *testing.T) {
	tests := []struct {
		name            string
		categoryID      int64
		categoryRepoErr error
		bookRepoResult  []Book
		bookRepoErr     error
		wantErr         error
		wantLen         int
	}{
		{
			name:            "success",
			categoryID:      1,
			categoryRepoErr: nil,
			bookRepoResult: []Book{
				{ID: 1, Title: "Book 1", CategoryID: 1},
				{ID: 2, Title: "Book 2", CategoryID: 1},
			},
			bookRepoErr: nil,
			wantErr:     nil,
			wantLen:     2,
		},
		{
			name:            "success empty",
			categoryID:      1,
			categoryRepoErr: nil,
			bookRepoResult:  []Book{},
			bookRepoErr:     nil,
			wantErr:         nil,
			wantLen:         0,
		},
		{
			name:            "category not found",
			categoryID:      999,
			categoryRepoErr: category.ErrCategoryNotFound,
			bookRepoResult:  nil,
			bookRepoErr:     nil,
			wantErr:         category.ErrCategoryNotFound,
		},
		{
			name:            "category repository error",
			categoryID:      1,
			categoryRepoErr: errors.New("database error"),
			bookRepoResult:  nil,
			bookRepoErr:     nil,
			wantErr:         errors.New("database error"),
		},
		{
			name:            "book repository error",
			categoryID:      1,
			categoryRepoErr: nil,
			bookRepoResult:  nil,
			bookRepoErr:     errors.New("database error"),
			wantErr:         errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookRepo := &fakeBookRepository{
				findByCategoryIDResult: tt.bookRepoResult,
				findByCategoryIDErr:    tt.bookRepoErr,
			}
			categoryRepo := &fakeCategoryRepository{
				findByIDErr: tt.categoryRepoErr,
			}
			service := NewService(bookRepo, &fakeAuthorRepository{}, categoryRepo, &fakeCacheClient{})

			got, err := service.FindByCategoryID(context.Background(), tt.categoryID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("FindByCategoryID() error = nil, wantErr %v", tt.wantErr)
				}
				if tt.wantErr == category.ErrCategoryNotFound && !errors.Is(err, category.ErrCategoryNotFound) {
					t.Errorf("FindByCategoryID() expected ErrCategoryNotFound, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("FindByCategoryID() unexpected error = %v", err)
				}
				if len(got) != tt.wantLen {
					t.Errorf("FindByCategoryID() got %d books, want %d", len(got), tt.wantLen)
				}
			}
		})
	}
}

func TestBookService_Create(t *testing.T) {
	desc := "Test description"
	tests := []struct {
		name            string
		request         CreateBookRequest
		authorRepoErr   error
		categoryRepoErr error
		bookRepoResult  Book
		bookRepoErr     error
		wantErr         error
		wantTitle       string
	}{
		{
			name: "success",
			request: CreateBookRequest{
				ISBN:        "123-456",
				Title:       "New Book",
				Description: &desc,
				Stock:       10,
				AuthorID:    1,
				CategoryID:  1,
			},
			authorRepoErr:   nil,
			categoryRepoErr: nil,
			bookRepoResult:  Book{ID: 1, Title: "New Book", ISBN: "123-456"},
			bookRepoErr:     nil,
			wantErr:         nil,
			wantTitle:       "New Book",
		},
		{
			name: "author not found",
			request: CreateBookRequest{
				ISBN:        "123-456",
				Title:       "New Book",
				Description: &desc,
				Stock:       10,
				AuthorID:    999,
				CategoryID:  1,
			},
			authorRepoErr:   author.ErrAuthorNotFound,
			categoryRepoErr: nil,
			bookRepoResult:  Book{},
			bookRepoErr:     nil,
			wantErr:         author.ErrAuthorNotFound,
		},
		{
			name: "category not found",
			request: CreateBookRequest{
				ISBN:        "123-456",
				Title:       "New Book",
				Description: &desc,
				Stock:       10,
				AuthorID:    1,
				CategoryID:  999,
			},
			authorRepoErr:   nil,
			categoryRepoErr: category.ErrCategoryNotFound,
			bookRepoResult:  Book{},
			bookRepoErr:     nil,
			wantErr:         category.ErrCategoryNotFound,
		},
		{
			name: "book repository error",
			request: CreateBookRequest{
				ISBN:        "123-456",
				Title:       "New Book",
				Description: &desc,
				Stock:       10,
				AuthorID:    1,
				CategoryID:  1,
			},
			authorRepoErr:   nil,
			categoryRepoErr: nil,
			bookRepoResult:  Book{},
			bookRepoErr:     errors.New("database error"),
			wantErr:         errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookRepo := &fakeBookRepository{
				createResult: tt.bookRepoResult,
				createErr:    tt.bookRepoErr,
			}
			authorRepo := &fakeAuthorRepository{
				findByIDErr: tt.authorRepoErr,
			}
			categoryRepo := &fakeCategoryRepository{
				findByIDErr: tt.categoryRepoErr,
			}
			service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})

			got, err := service.Create(context.Background(), tt.request)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Create() error = nil, wantErr %v", tt.wantErr)
				}
				if tt.wantErr == author.ErrAuthorNotFound && !errors.Is(err, author.ErrAuthorNotFound) {
					t.Errorf("Create() expected ErrAuthorNotFound, got %v", err)
				}
				if tt.wantErr == category.ErrCategoryNotFound && !errors.Is(err, category.ErrCategoryNotFound) {
					t.Errorf("Create() expected ErrCategoryNotFound, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Create() unexpected error = %v", err)
				}
				if got.Title != tt.wantTitle {
					t.Errorf("Create() got title %s, want %s", got.Title, tt.wantTitle)
				}
			}
		})
	}
}

func TestBookService_Update(t *testing.T) {
	desc := "Updated description"
	tests := []struct {
		name            string
		id              int64
		request         UpdateBookRequest
		authorRepoErr   error
		categoryRepoErr error
		bookRepoResult  Book
		bookRepoErr     error
		wantErr         error
		wantTitle       string
	}{
		{
			name: "success",
			id:   1,
			request: UpdateBookRequest{
				Title:       "Updated Book",
				Description: &desc,
				Stock:       5,
				AuthorID:    1,
				CategoryID:  1,
			},
			authorRepoErr:   nil,
			categoryRepoErr: nil,
			bookRepoResult:  Book{ID: 1, Title: "Updated Book"},
			bookRepoErr:     nil,
			wantErr:         nil,
			wantTitle:       "Updated Book",
		},
		{
			name: "author not found",
			id:   1,
			request: UpdateBookRequest{
				Title:       "Updated Book",
				Description: &desc,
				Stock:       5,
				AuthorID:    999,
				CategoryID:  1,
			},
			authorRepoErr:   author.ErrAuthorNotFound,
			categoryRepoErr: nil,
			bookRepoResult:  Book{},
			bookRepoErr:     nil,
			wantErr:         author.ErrAuthorNotFound,
		},
		{
			name: "category not found",
			id:   1,
			request: UpdateBookRequest{
				Title:       "Updated Book",
				Description: &desc,
				Stock:       5,
				AuthorID:    1,
				CategoryID:  999,
			},
			authorRepoErr:   nil,
			categoryRepoErr: category.ErrCategoryNotFound,
			bookRepoResult:  Book{},
			bookRepoErr:     nil,
			wantErr:         category.ErrCategoryNotFound,
		},
		{
			name: "book not found",
			id:   999,
			request: UpdateBookRequest{
				Title:       "Updated Book",
				Description: &desc,
				Stock:       5,
				AuthorID:    1,
				CategoryID:  1,
			},
			authorRepoErr:   nil,
			categoryRepoErr: nil,
			bookRepoResult:  Book{},
			bookRepoErr:     ErrBookNotFound,
			wantErr:         ErrBookNotFound,
		},
		{
			name: "repository error",
			id:   1,
			request: UpdateBookRequest{
				Title:       "Updated Book",
				Description: &desc,
				Stock:       5,
				AuthorID:    1,
				CategoryID:  1,
			},
			authorRepoErr:   nil,
			categoryRepoErr: nil,
			bookRepoResult:  Book{},
			bookRepoErr:     errors.New("database error"),
			wantErr:         errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookRepo := &fakeBookRepository{
				updateResult: tt.bookRepoResult,
				updateErr:    tt.bookRepoErr,
			}
			authorRepo := &fakeAuthorRepository{
				findByIDErr: tt.authorRepoErr,
			}
			categoryRepo := &fakeCategoryRepository{
				findByIDErr: tt.categoryRepoErr,
			}
			service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})

			got, err := service.Update(context.Background(), tt.id, tt.request)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Update() error = nil, wantErr %v", tt.wantErr)
				}
				if tt.wantErr == author.ErrAuthorNotFound && !errors.Is(err, author.ErrAuthorNotFound) {
					t.Errorf("Update() expected ErrAuthorNotFound, got %v", err)
				}
				if tt.wantErr == category.ErrCategoryNotFound && !errors.Is(err, category.ErrCategoryNotFound) {
					t.Errorf("Update() expected ErrCategoryNotFound, got %v", err)
				}
				if tt.wantErr == ErrBookNotFound && !errors.Is(err, ErrBookNotFound) {
					t.Errorf("Update() expected ErrBookNotFound, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Update() unexpected error = %v", err)
				}
				if got.Title != tt.wantTitle {
					t.Errorf("Update() got title %s, want %s", got.Title, tt.wantTitle)
				}
			}
		})
	}
}

func TestBookService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		repoErr error
		wantErr bool
	}{
		{
			name:    "success",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "not found",
			id:      999,
			repoErr: ErrBookNotFound,
			wantErr: true,
		},
		{
			name:    "repository error",
			id:      1,
			repoErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookRepository{
				deleteErr: tt.repoErr,
			}
			service := NewService(repo, &fakeAuthorRepository{}, &fakeCategoryRepository{}, &fakeCacheClient{})

			err := service.Delete(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.repoErr == ErrBookNotFound && !errors.Is(err, ErrBookNotFound) {
				t.Errorf("Delete() expected ErrBookNotFound, got %v", err)
			}
		})
	}
}
