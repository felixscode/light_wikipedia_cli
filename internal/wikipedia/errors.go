package wikipedia

import "errors"

var (
	ErrNotFound     = errors.New("page not found")
	ErrAPI          = errors.New("API error")
	ErrInvalidTitle = errors.New("invalid title")
)
