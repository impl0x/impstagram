package response

// A short string sequence with underscores and all caps signifying in short what went wrong or right
type Code string

// Helpful to the frontend application to know these codes

// NounAdjective format

// The error/success codes that are returned in the json

// General codes
const (
	CodeOk      Code = "OK"
	CodeUnknown Code = "UNKNOWN"

	CodeNotFound         Code = "NOT_FOUND"
	CodeMethodNotAllowed Code = "METHOD_NOT_ALLOWED"
	CodeInternal         Code = "INTERNAL_SERVER_ERROR"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeTimeout          Code = "TIMEOUT"
	CodeValidationError  Code = "VALIDATION_ERROR"

	CodeJSONInvalid Code = "JSON_INVALID"
	CodeEOF         Code = "EOF"
)
