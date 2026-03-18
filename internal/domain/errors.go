package domain

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInternalServer     = errors.New("internal server error")
	ErrNotFound           = errors.New("not found")
)
