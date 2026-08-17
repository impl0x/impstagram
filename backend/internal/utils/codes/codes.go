package codes

// A short string sequence with underscores and all caps signifying in short what went wrong or right
type Code string

// Helpful to the frontend application to know these codes

// NounAdjective format

// The error codes that are returned in the json
const (
	Unknown Code = "UNKNOWN"

	NotFound            Code = "NOT_FOUND"
	MethodNotAllowed    Code = "METHOD_NOT_ALLOWED"
	InternalServerError Code = "INTERNAL_SERVER_ERROR"
	Unauthorized        Code = "UNAUTHORIZED"
	Timeout Code = "TIMEOUT"

	JSONInvalid Code = "JSON_INVALID"
	EOF         Code = "EOF"

	ValidationError Code = "VALIDATION_ERROR"

	IdentifierMissing  Code = "IDENTIFIER_MISSING"
	CredentialsInvalid Code = "CREDENTIALS_INVALID"

	UserNotFound     Code = "USER_NOT_FOUND"
	UserBanned       Code = "USER_BANNED"
	UserUnverified   Code = "USER_UNVERIFIED"
	UserAlreadyExists       Code = "USER_ALREADY_EXISTS"
	UserNotOldEnough Code = "USER_NOT_OLD_ENOUGH"
	UserTooOld       Code = "USER_TOO_OLD"

	UsernameAlreadyExists Code = "USERNAME_ALREADY_EXISTS"

	OTPExpired   Code = "OTP_EXPIRED"
	OTPIncorrect Code = "OTP_INCORRECT"

	AccessTokenInvalid  Code = "ACCESS_TOKEN_INVALID"
	AccessTokenExpired  Code = "ACCESS_TOKEN_EXPIRED"
	RefreshTokenInvalid Code = "REFRESH_TOKEN_INVALID"
	RefreshTokenExpired Code = "REFRESH_TOKEN_EXPIRED"

	ResetSessionExpired Code = "RESET_SESSION_EXPIRED"
)

// the success codes returned in the json
const (
	Ok              Code = "OK"
	LoginSuccess    Code = "LOGIN_SUCCESS"
	TwoFARequired   Code = "TWO_FACTOR_REQUIRED"
	RegisterSuccess Code = "REGISTER_SUCCESS"
	RefreshSuccess  Code = "REFRESH_SUCCESS"
)
