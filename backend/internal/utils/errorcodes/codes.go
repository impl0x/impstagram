package errorcodes

type Code string

// The error codes that are returned in the json
const (
	Unknown Code="UNKNOWN"

	NotFound Code = "NOT_FOUND"
	MethodNotAllowed Code = "METHOD_NOT_ALLOWED"
	InternalServerError Code = "INTERNAL_SERVER_ERROR"

	InvalidJSON Code = "INVALID_JSON"
	EOF Code = "EOF"

	ValidationError Code = "VALIDATION_ERROR"
)
