package domain

import "errors"

var (
	ErrIconNotFound      = errors.New("icon not found")
	ErrTooManyIconsFound = errors.New("too many icons found")
)
