package models

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrForbidden        = errors.New("forbidden")
	ErrConflict         = errors.New("conflict")
	ErrBadRequest       = errors.New("bad request")
	ErrConcurrentUpdate = errors.New("concurrent update")
)

// DomainError is a typed error that carries a machine-readable code alongside a
// wrapped sentinel error (e.g. ErrBadRequest).
type DomainError struct {
	Code    string
	wrapped error
}

// NewDomainError creates a DomainError with the given code wrapping wrapped.
func NewDomainError(code string, wrapped error) *DomainError {
	return &DomainError{Code: code, wrapped: wrapped}
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.wrapped)
}

func (e *DomainError) Unwrap() error {
	return e.wrapped
}
