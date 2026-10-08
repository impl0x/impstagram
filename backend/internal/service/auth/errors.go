package auth

import (
	"backend/pkg/apperr"
	"backend/pkg/response"
	"errors"
)

// ? INFO
// Contains errors which later get handled by global error handler
// ! Extra
// all codes are prefixed with "code" and all errors with "err", for consistency reasons.
// all sub divisions of codes and errors are also prefixed with the category that it belongs to.
// examples, where Name is the var name for the code/name
// eg: common errors have prefix of "err" + "Common" + Name
// all categories for errors are based on the service function that it belongs to, such as middleware errors are of category middleware and have that as its prefix

// Common codes, prefix: common
const (
	codeCommonUserNotFound      response.Code = "USER_NOT_FOUND"
	codeCommonSessionExpired    response.Code = "SESSION_EXPIRED"
	codeCommonAttemptsExhausted response.Code = "ATTEMPTS_EXHAUSTED"
	codeCommonTokenInvalid      response.Code = "TOKEN_INVALID"
	codeCommonTokenExpired      response.Code = "TOKEN_EXPIRED"
)

// Common errors, prefix: common
var (
	errCommonUserNotFound = apperr.NewNotFound(
		codeCommonUserNotFound,
		"User not found",
	)
	errCommonUnauthorized = apperr.NewUnauthorized(
		response.CodeUnauthorized,
		"Unauthorized user",
	)
	errCommonAttemptsExhausted = apperr.NewUnauthorized(
		"ATTEMPTS_EXHAUSTED",
		"No attempts remaining",
	)
)

// Middleware errors, prefix: Middleware
var (
	errAuthTokenInvalidPayload     = errors.New("auth.Middleware: Auth token has invalid payload")
	errMiddlewareAuthHeaderMissing = apperr.NewUnauthorized(
		response.CodeUnauthorized,
		"Authorization header missing or empty",
	)
	errMiddlewareAuthTokenInvalid = apperr.NewUnauthorized(
		codeCommonTokenInvalid,
		"Invalid authorization token")
	// errMiddlewareInvalidAuthTokenPayload = apperr.New(
	// apperr.KindInternal,
	// response.CodeInternal,
	// "Invalid data in the JSON payload for authorization token",
	// ) // internal error because server should not issue wrong tokens
	errMiddlewareAuthTokenExpired = apperr.NewUnauthorized(
		codeCommonTokenExpired,
		"Token has expired, please refresh it using the refresh token",
	)
)

// Register errors, prefix: Register
var (
	errRegisterMissingIdentifier = apperr.NewValidation(
		"Need at least email or phone to register",
	)
	errRegisterNotOldEnough = apperr.NewForbidden(
		"USER_NOT_OLD_ENOUGH",
		"User not old enough to use this service",
	)
	errRegisterTooOld = apperr.NewForbidden(
		"USER_TOO_OLD",
		"User too old, cannot create account",
	)
	errRegisterAlreadyExistingUser = apperr.NewConflict(
		"USER_ALREADY_EXISTS",
		"This user already exists, please try to log in",
	)
	errRegisterUsernameAlreadyExists = apperr.NewConflict(
		"USERNAME_ALREADY_EXISTS",
		"This username is already registered, please try a different username",
	)
)

// Login errors, prefix: Login
var (
	errLoginCredentialsInvalid = apperr.NewUnauthorized(
		"CREDENTIALS_INVALID",
		"Invalid identifier or password",
	)
	errLoginUserBanned = apperr.NewForbidden(
		"USER_BANNED",
		"User is permanently banned from accessing this Service",
	)
	errLoginUserUnverified = apperr.NewForbidden(
		"USER_UNVERIFIED",
		"User account is unverified. Please verify with your account identifier, i.e. email or phone whichever you used to register",
	)
)

// Resend otp errors, prefix: Resend
var (
	errResendInvalidIdentifier = apperr.NewValidation(
		"Invalid identifier for resending OTP",
	)
	errResend2FANotEnabled = apperr.NewForbidden(
		"2FA_NOT_ENABLED",
		"User does not have 2FA enabled, cannot know where to send otp",
	)
)

// Reset password errors, prefix: Reset
var (
	errResetSessionNotFound = apperr.NewNotFound(
		response.CodeNotFound,
		"Session not found, please try to raise a reset password request again",
	)
	errResetSessionExpired = apperr.NewUnauthorized(
		codeCommonSessionExpired,
		"Session has expired, please try to raise a reset password request again",
	)
)

// Verify otp errors, prefix: Verify
var (
	errVerifyRefIDNotFound = apperr.NewNotFound(
		response.CodeNotFound,
		"Reference ID for this verify request not found, please request a new otp again",
	)
	errVerifyOTPExpired = apperr.NewUnauthorized(
		"OTP_EXPIRED",
		"The OTP provided is expired",
	)
	errVerifyOTPIncorrect = apperr.NewUnauthorized(
		"OTP_INCORRECT",
		"The OTP provided is incorrect",
	)
)

// Refresh errors, prefix: Refresh
var (
	errRefreshTokenInvalid = apperr.NewUnauthorized(
		codeCommonTokenInvalid,
		"Refresh token is invalid, please login again",
	)
	errRefreshTokenExpired = apperr.NewUnauthorized(
		codeCommonTokenExpired,
		"Refresh token has expired, please login again",
	)
)

// Add 2FA errors, prefix: Add2fa
var (
	errAdd2FAChannelExists = apperr.NewConflict(
		"CHANNEL_EXISTS",
		"This channel already has 2FA enabled",
	)
	errAdd2FAChannelEmpty = apperr.NewNotFound(
		"CHANNEL_EMPTY",
		"The channel provided is not added to your account and cannot be set as two factor verification method",
	)
)

// Remove 2FA errors, prefix: Remove2FA
var (
	errRemove2FAChannelNotFound = apperr.NewNotFound(
		response.CodeNotFound,
		"The channel is not present in the user's 2fa to be removed",
	)
	errRemove2FANotEnabled = apperr.NewForbidden(
		"2FA_NOT_ENABLED",
		"2FA is not enabled for user",
	)
)

// Totp setup errors, prefix: TotpSetup
var (
	errTotpSetupAlreadyEnabled = apperr.NewConflict(
		"TOTP_ALREADY_ENABLED",
		"User already has TOTP enabled",
	)
)

// Totp verify errors, prefix: TotpVerify
var (
	errTotpVerifySessionNotFound = apperr.NewNotFound(
		response.CodeNotFound,
		"Session not found, please try to raise a TOTP setup request again",
	)
	errTotpVerifySessionExpired = apperr.NewUnauthorized(
		codeCommonSessionExpired,
		"Session has expired, please try to raise a TOTP setup request again",
	)
	errTotpVerifyTOTPIncorrect = apperr.NewForbidden(
		"TOTP_INCORRECT",
		"Incorrect OTP",
	)
)
