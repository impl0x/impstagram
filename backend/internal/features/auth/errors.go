package auth

import "backend/internal/pkg/response"

// auth codes
const (
	codeCredentialsInvalid    response.Code = "CREDENTIALS_INVALID"
	codeUsernameAlreadyExists response.Code = "USERNAME_ALREADY_EXISTS"

	codeUserNotFound      response.Code = "USER_NOT_FOUND"
	codeUserBanned        response.Code = "USER_BANNED"
	codeUserUnverified    response.Code = "USER_UNVERIFIED"
	codeUserAlreadyExists response.Code = "USER_ALREADY_EXISTS"
	codeUserNotOldEnough  response.Code = "USER_NOT_OLD_ENOUGH"
	codeUserTooOld        response.Code = "USER_TOO_OLD"

	codeOTPExpired   response.Code = "OTP_EXPIRED"
	codeOTPIncorrect response.Code = "OTP_INCORRECT"

	codeAccessTokenInvalid  response.Code = "ACCESS_TOKEN_INVALID"
	codeAccessTokenExpired  response.Code = "ACCESS_TOKEN_EXPIRED"
	codeRefreshTokenInvalid response.Code = "REFRESH_TOKEN_INVALID"
	codeRefreshTokenExpired response.Code = "REFRESH_TOKEN_EXPIRED"

	codeResetSessionExpired response.Code = "RESET_SESSION_EXPIRED"

	codeLoginSuccess    response.Code = "LOGIN_SUCCESS"
	codeTwoFARequired   response.Code = "TWO_FACTOR_REQUIRED"
	codeRegisterSuccess response.Code = "REGISTER_SUCCESS"
	codeRefreshSuccess  response.Code = "REFRESH_SUCCESS"
)
