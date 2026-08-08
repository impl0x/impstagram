package errs

import "errors"

var (
	InternalServerError = errors.New("An unexpected error occurred")
)