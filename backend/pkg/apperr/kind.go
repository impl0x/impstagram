package apperr

type Kind uint

const (
	KindNotFound     Kind = iota // example: Resource doesn't exist
	KindValidation               // example: Malformed JSON, validation failure
	KindUnauthorized             // example: Token missing, invalid credentials
	KindForbidden                // example: Token valid, but user cannot access this resource
	KindConflict                 // example: Username already taken
	KindRateLimit                // example: Too many login attempts, OTP exhausted
)
