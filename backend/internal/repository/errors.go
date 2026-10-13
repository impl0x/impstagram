package repository

import "errors"

// Common repository errors

var (
	ErrNoResults      = errors.New("repository: no results found")       // no results were found in the repository
	ErrTooManyResults = errors.New("repository: too many results found") // more than one result was found in the repository
	ErrAlreadyExists  = errors.New("repository: already exists")         // the value trying to be inserted already exists, happens when a "unique" violation occurs
)
