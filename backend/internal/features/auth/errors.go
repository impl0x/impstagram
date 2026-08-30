package auth

import (
	"backend/internal/pkg/apperr"
	"backend/internal/pkg/response"
	"errors"
	"strconv"
)

// Internal errors
var (
	errInternalInvalidChannel = errors.New("invalid channel found") // used when a invalid/unsupported channel is found
)

// ? info:
// This file contains error codes and apperr sentinel errors used by the service layer
// only the below non error codes are used by the handler layer
// the service layer does not access these codes directly but instead returns the sentinel errors instead which uses these codes

// Non error codes, prefix: NonErr
const (
	codeNonErrLoginSuccess    response.Code = "LOGIN_SUCCESS"
	codeNonErrTwoFARequired   response.Code = "TWO_FACTOR_REQUIRED"
	codeNonErrRegisterSuccess response.Code = "REGISTER_SUCCESS"
	codeNonErrRefreshSuccess  response.Code = "REFRESH_SUCCESS"
	codeNonErrOTPSent         response.Code = "OTP_SENT"
)

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
	errRegisterInvalidDobString = apperr.NewValidation(
		"Invalid date of birth string",
	)
	errRegisterImpossibleDobString = apperr.NewValidation(
		"Impossible date of birth string",
	)
	errRegisterMissingIdentifier = apperr.NewValidation(
		"Need at least email or phone to register",
	)
	errRegisterNotOldEnough = apperr.NewForbidden(
		"USER_NOT_OLD_ENOUGH",
		"User not old enough, must be minimum of "+strconv.Itoa(ruleMinAge)+" years old to use this Service",
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
	errLoginMissingIdentifier = apperr.NewValidation(
		"Need at least email, phone or username to log in",
	)
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
	errResendMissingIdentifier = apperr.NewValidation(
		"Need at least email, phone, or username to resend OTP",
	)
	errResendInvalidIdentifier = apperr.NewValidation(
		"Invalid identifier for resending OTP",
	)
	errResend2FANotEnabled = apperr.NewForbidden(
		"2FA_NOT_ENABLED",
		"User does not have 2FA enabled, cannot know where to send otp",
	)
)

// Forgot password errors, prefix: Forgot
var(
	errForgotMissingIdentifier = apperr.NewValidation(
		"Need at least email or phone to verify account ownership",
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
