package apperr

import "backend/internal/pkg/response"

type AppErr struct {
	Kind     Kind
	Code response.Code
	Message  string
	Err      error
}

func New(kind Kind, code response.Code, message string) AppErr {
	return AppErr{
		Kind:     kind,
		Code: code,
		Message:  message,
	}
}

func (ae AppErr) Wrap(err error) AppErr {
	ae.Err = err
	return ae
}

func (ae AppErr) Error() string {
	return ae.Message
}

func NewValidation(code response.Code, message string) AppErr {
	return New(KindValidation, code, message)
}

func NewNotFound(code response.Code, message string) AppErr {
	return New(KindNotFound, code, message)
}

func NewConflict(code response.Code, message string) AppErr {
	return New(KindConflict, code, message)
}

func NewUnauthorized(code response.Code, message string) AppErr {
	return New(KindUnauthorized, code, message)
}

func NewForbidden(code response.Code, message string) AppErr {
	return New(KindForbidden, code, message)
}

func NewInternal(code response.Code, message string) AppErr {
	return New(KindInternal, code, message)
}
