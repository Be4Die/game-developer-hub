package domain

import "errors"

var (
	// ErrNotFound возвращается при отсутствии запрашиваемого ресурса.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists возвращается при попытке создать дубликат ресурса.
	ErrAlreadyExists = errors.New("already exists")
)
