package auth

import "errors"

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already registered")
	ErrBadAuth     = errors.New("invalid email or password")
)
