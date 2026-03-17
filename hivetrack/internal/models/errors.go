package models

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrForbidden        = errors.New("forbidden")
	ErrConflict         = errors.New("conflict")
	ErrBadRequest       = errors.New("bad request")
	ErrConcurrentUpdate = errors.New("concurrent update")
)
