package apperr

import (
	"backend/internal/pkg/response"
)

type AppErr struct {
	Kind    Kind
	Code    response.Code
	Message string
	Err     error
}

func New(kind Kind, code response.Code, message string) AppErr {
	return AppErr{
		Kind:    kind,
		Code:    code,
		Message: message,
	}
}

func (ae AppErr) Wrap(err error) AppErr {
	ae.Err = err
	return ae
}

// Returns the AppErr to a status code and response.Response struct
//
// data is optional and can be set to nil
func (ae AppErr) ToHttp(data any) (int, response.Response) {
	return ae.Kind.ToStatusCode(), response.Response{
		Code:    ae.Code,
		Message: ae.Message,
		Data:    data,
	}
}

func (ae AppErr) Error() string {
	return ae.Message
}

func NewValidation(message string) AppErr {
	return New(kindValidation, response.CodeValidationError, message)
}

func NewNotFound(code response.Code, message string) AppErr {
	return New(kindNotFound, code, message)
}

func NewConflict(code response.Code, message string) AppErr {
	return New(kindConflict, code, message)
}

func NewUnauthorized(code response.Code, message string) AppErr {
	return New(kindUnauthorized, code, message)
}

func NewForbidden(code response.Code, message string) AppErr {
	return New(kindForbidden, code, message)
}
