package application

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input provided")
	ErrNotFound     = errors.New("resource not found")
	ErrInternal     = errors.New("internal error")
)
