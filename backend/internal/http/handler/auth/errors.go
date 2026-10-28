package auth

import (
	"backend/pkg/apperr"
	"backend/pkg/response"
	"strings"
)

// Non error codes, prefix: NonErr
const (
	codeNonErrLoginSuccess    response.Code = "LOGIN_SUCCESS"
	codeNonErrTwoFARequired   response.Code = "TWO_FACTOR_REQUIRED"
	codeNonErrRegisterSuccess response.Code = "REGISTER_SUCCESS"
	codeNonErrRefreshSuccess  response.Code = "REFRESH_SUCCESS"
	codeNonErrOTPSent         response.Code = "OTP_SENT"
)

var (
	// extra strings will be joined with a ". "
	errIdentifierNotProvided = func(extra ...string) apperr.AppErr {
		str := "Identifier not provided"
		if extra != nil {
			str += ". "
			str += strings.Join(extra, ". ")
		}
		return apperr.NewValidation(str)
	}
	errInvalidDob    = apperr.NewValidation("Invalid date of birth string")
	errImpossibleDob = apperr.NewValidation("Impossible date of birth")
)
