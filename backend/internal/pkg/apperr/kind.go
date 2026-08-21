package apperr

import (
	"net/http"
)

type Kind int

const (
	KindValidation   Kind = iota // http: 400
	KindNotFound                 // http: 404
	KindConflict                 // http: 409
	KindUnauthorized             // http: 401
	KindForbidden                // http: 403
	KindInternal                 // http: 500
)

func (k Kind) ToStatusCode() int {
	switch k {
	case KindValidation:
		return http.StatusBadRequest
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
