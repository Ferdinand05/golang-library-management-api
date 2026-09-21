package borrowing

import "errors"

var ErrBorrowingNotFound = errors.New("borrowing not found")
var ErrBookOutOfStock = errors.New("book out of stock")
var ErrBorrowingAlreadyReturned = errors.New("borrowing already returned")
