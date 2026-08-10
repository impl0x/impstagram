package codes

type Code string

// The error codes that are returned in the json
//
// Helpful to the frontend application to know these codes
const (
	Unknown Code = "UNKNOWN"

	NotFound            Code = "NOT_FOUND"
	MethodNotAllowed    Code = "METHOD_NOT_ALLOWED"
	InternalServerError Code = "INTERNAL_SERVER_ERROR"

	InvalidJSON Code = "INVALID_JSON"
	EOF         Code = "EOF"

	ValidationError Code = "VALIDATION_ERROR"

	MissingIdentifier  Code = "MISSING_IDENTIFIER"
	InvalidCredentials Code = "INVALID_CREDENTIALS"
	UserBanned         Code = "USER_BANNED"
	UserUnverified     Code = "USER_UNVERIFIED"

	UserExists       Code = "USER_ALREADY_EXISTS"
	UserNotOldEnough Code = "USER_NOT_OLD_ENOUGH"
	UserTooOld       Code = "USER_TOO_OLD"
	UsernameAlreadyExists Code = "USERNAME_ALREADY_EXISTS"

	ReferenceIDNotFound Code = "REFERENCE_ID_NOT_FOUND"
	OTPExpired Code = "OTP_EXPIRED"
	OTPIncorrect Code = "OTP_INCORRECT"
)

// the success codes returned in the json
const (
	LoginSuccess  Code = "LOGIN_SUCCESS"
	TwoFARequired Code = "TWO_FACTOR_REQUIRED"
	RegisterSuccess Code = "REGISTER_SUCCESS"
)
