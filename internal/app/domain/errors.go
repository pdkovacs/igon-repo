package domain

import "errors"

var (
	ErrIconNotFound          = errors.New("icon not found")
	ErrIconfileNotFound      = errors.New("iconfile not found")
	ErrTooManyIconsFound     = errors.New("too many icons found")
	ErrIconAlreadyExists     = errors.New("icon already exists")
	ErrIconfileAlreadyExists = errors.New("iconfile already exists")
)
