package errors

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInternal          = errors.New("internal error")
	ErrInvalidInput      = errors.New("invalid input parameter")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrUnauthorized      = errors.New("unauthorized")
)

type AppError struct {
	Err     error
	Message string
	Code    int
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(err error, msg string) *AppError {
	return &AppError{Err: err, Message: msg}
}
