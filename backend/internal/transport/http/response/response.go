package response

import (
	"backend/pkg/apperr"
	"net/http"
)

// Http response struct
type Response struct {
	Code    apperr.Code `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    any         `json:"data,omitempty"`
}

// returns [Response] with the data
func Success(code apperr.Code, message string, data any) Response {
	return Response{code, message, data}
}

// returns [Response] without the data
func Error(code apperr.Code, message string) Response {
	return Response{code, message, nil}
}

func FromAppErr(a apperr.AppErr, data any) Response {
	return Response{
		Code:    a.Code,
		Message: a.Message,
		Data:    data,
	}
}

func HttpStatuscodeFromAppErrKind(k apperr.Kind) int {
	switch k {
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindValidation:
		return http.StatusBadRequest
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindConflict:
		return http.StatusConflict
	case apperr.KindRateLimit:
		return http.StatusTooManyRequests
	default:
		panic("invalid app error kind")
	}
}
