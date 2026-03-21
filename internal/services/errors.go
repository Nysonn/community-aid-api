package services

import "errors"

// ErrNotFound is returned by any service method when the requested record does not exist.
var ErrNotFound = errors.New("record not found")
