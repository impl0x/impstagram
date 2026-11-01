package apperr

type AppErr struct {
	Kind    Kind
	Code    Code
	Message string
	Err     error
}

func New(kind Kind, code Code, message string) AppErr {
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

func (ae AppErr) Error() string {
	return ae.Message
}

func NewNotFound(code Code, message string) AppErr {
	return New(KindNotFound, code, message)
}

func NewValidation(message string) AppErr {
	return New(KindValidation, CodeValidationError, message)
}

func NewUnauthorized(code Code, message string) AppErr {
	return New(KindUnauthorized, code, message)
}

func NewForbidden(code Code, message string) AppErr {
	return New(KindForbidden, code, message)
}

func NewConflict(code Code, message string) AppErr {
	return New(KindConflict, code, message)
}

func NewRatelimit(code Code, message string) AppErr {
	return New(KindRateLimit, code, message)
}
