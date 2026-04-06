package services

import "errors"

// ErrNotFound is returned by any service method when the requested record does not exist.
var ErrNotFound = errors.New("record not found")

// BadRequestError indicates the caller supplied data that cannot be processed.
type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}
