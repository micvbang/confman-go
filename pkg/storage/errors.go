package storage

import "errors"

var (
	ErrConfigNotFound = errors.New("config not found")
	ErrTooManyKeys    = errors.New("too many keys")
)
