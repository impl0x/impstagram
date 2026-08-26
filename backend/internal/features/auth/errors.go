package auth

import (
	"backend/internal/pkg/apperr"
	"backend/internal/pkg/response"
	"strconv"
)

// this file is used only by the service layer

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

	code2FAExistingChannel response.Code = "EXISTING_CHANNEL"
	codeChannelEmpty       response.Code = "CHANNEL_EMPTY"

	codeLoginSuccess    response.Code = "LOGIN_SUCCESS"
	codeTwoFARequired   response.Code = "TWO_FACTOR_REQUIRED"
	codeRegisterSuccess response.Code = "REGISTER_SUCCESS"
	codeRefreshSuccess  response.Code = "REFRESH_SUCCESS"
)

// Common errors
var (
	errUserNotFound = apperr.NewNotFound(codeUserNotFound, "User not found")
)

// Registration errors
var (
	errInvalidDobString      = apperr.NewValidation("Invalid date of birth string")
	errImpossibleDobString   = apperr.NewValidation("Impossible date of birth string")
	errNotOldEnough          = apperr.NewForbidden(codeUserNotOldEnough, "User not old enough, must be minimum of "+strconv.Itoa(int(ruleMinAge))+" years old to use this Service")
	errTooOld                = apperr.NewForbidden(codeUserTooOld, "User too old, cannot create account")
	errMissingIdentifier     = apperr.NewValidation("Need at least email or phone to register")
	errAlreadyExistingUser   = apperr.NewConflict(codeUserAlreadyExists, "This user already exists, please try to log in")
	errUsernameAlreadyExists = apperr.NewConflict(codeUsernameAlreadyExists, "This username is already registered, please try a different username")
)

// Login errors
var (
	errMissingIdentifierLogin = apperr.NewValidation("Need at least email, phone or username to log in")
	errCredentialsInvalid     = apperr.NewUnauthorized(codeCredentialsInvalid, "Invalid identifier or password")
	errUserBanned             = apperr.NewForbidden(codeUserBanned, "User is permanently banned from accessing this Service")
	errUserUnverified         = apperr.NewForbidden(codeUserUnverified, "User account is unverified. Please verify with your account identifier, i.e. email or phone whichever you used to register")
)

// Resend otp errors
var (
	errMissingIdentifierResend = apperr.NewValidation("Need at least email, phone, or username to resend OTP")
	errInvalidIdentifierResend = apperr.NewValidation("Invalid identifier for resending OTP")
	err2FANotEnabled           = apperr.NewValidation("User does not have 2FA enabled, cannot know where to send otp")
)

// Reset password errors
var (
	errResetSessionNotFound = apperr.NewNotFound(response.CodeNotFound, "Session not found, please try to raise a reset password request again")
	errResetSessionExpired  = apperr.NewUnauthorized(codeResetSessionExpired, "Session has expired, please try to raise a reset password request again")
)

// Verify otp errors
var (
	errRefIDNotFound = apperr.NewNotFound(response.CodeNotFound, "Reference ID for this verify request not found, please request a new otp again")
	errOTPExpired    = apperr.NewUnauthorized(codeOTPExpired, "The OTP provided is expired")
	errIncorrectOTP  = apperr.NewUnauthorized(codeOTPIncorrect, "The OTP provided is incorrect")
)

// Refresh errors
var (
	errInvalidRefreshToken = apperr.NewUnauthorized(codeRefreshTokenInvalid, "Refresh token is invalid, please login again")
	errExpiredRefreshToken = apperr.NewUnauthorized(codeRefreshTokenExpired, "Refresh token has expired, please login again")
)

// Add 2FA errors
var (
	err2FAExistingChannel = apperr.NewConflict(code2FAExistingChannel, "This channel already has 2FA enabled")
	errChannelEmpty       = apperr.NewNotFound(codeChannelEmpty, "The channel provided is not added to your account and cannot be set as two factor verification method")
)
