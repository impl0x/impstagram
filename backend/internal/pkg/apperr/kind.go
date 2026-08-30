package apperr

import (
	"net/http"
)

type Kind int

const (
	kindValidation   Kind = iota // http: 400
	kindNotFound                 // http: 404
	kindConflict                 // http: 409
	kindUnauthorized             // http: 401
	kindForbidden                // http: 403
	// KindInternal                 // http: 500, internal errors should preferably be directly returned to the error handler instead.
)

func (k Kind) ToStatusCode() int {
	switch k {
	case kindValidation:
		return http.StatusBadRequest
	case kindNotFound:
		return http.StatusNotFound
	case kindConflict:
		return http.StatusConflict
	case kindUnauthorized:
		return http.StatusUnauthorized
	case kindForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
