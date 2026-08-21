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
	kindInternal                 // http: 500
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
	case kindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
