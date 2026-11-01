package auth

import (
	"backend/pkg/apperr"
	"strings"
)

// Non error codes, prefix: NonErr
const (
	codeNonErrLoginSuccess    apperr.Code = "LOGIN_SUCCESS"
	codeNonErrTwoFARequired   apperr.Code = "TWO_FACTOR_REQUIRED"
	codeNonErrRegisterSuccess apperr.Code = "REGISTER_SUCCESS"
	codeNonErrRefreshSuccess  apperr.Code = "REFRESH_SUCCESS"
	codeNonErrOTPSent         apperr.Code = "OTP_SENT"
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
