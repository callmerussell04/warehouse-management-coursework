package errors

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInternal          = errors.New("internal error")
	ErrInvalidInput      = errors.New("invalid input parameter")
	ErrInsufficientStock = errors.New("insufficient stock")
)
