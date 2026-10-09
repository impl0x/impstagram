package repository

import "errors"

var (
	ErrNoResults      = errors.New("repository: no results found")
	ErrTooManyResults = errors.New("repository: too many results found")
	ErrAlreadyExists  = errors.New("repository: already exists")
)
