package book

import "errors"

var ErrBookNotFound = errors.New("Book not found")
var ErrInvalidBookDate = errors.New("invalid Book date")
var ErrDatabaseConnectionFail = errors.New("database connection failed")
var ErrBookOutOfStock = errors.New("book out of stock")
