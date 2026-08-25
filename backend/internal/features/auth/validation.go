package auth

import (
	"errors"

	"github.com/impl0x/mo/validator"
)

// ? This package is compatible with github.com/impl0x/mo/validator
// ? example usage: validator.AddCustomValidation(validateUsername())

// adds all the validation rules defined to global validator package instance
func AddValidations() {
	validator.AddCustomValidation(validateUsername())
}

// Username validation logic:
//   - allowed a-z 0-9 _ .
//   - first!=0-9
// tag: username
func validateUsername() (string, (func(v string) error)) {
	// storing sentinel errors to avoid heap allocation on every call to func 
	errEmptyUsername := errors.New("Empty username")
	errStartsWithPeriod := errors.New("Cannot start username with a period")
	errStartsWithNumber := errors.New("Cannot start username with a number")
	errInvalidUsername := errors.New("Username can only contain lowercase letters, numbers and underscore")
	errConsecutivePeriod := errors.New("Username cannot contain consecutive periods")
	return "username", func(s string) error {
		if s == "" {
			return errEmptyUsername
		}
		if s[0] == '.' {
			return errStartsWithPeriod
		}
		if s[0] >= '0' && s[0] <= '9' {
			return errStartsWithNumber
		}
		prev := '-'
		for _, l := range s {
			if (l < '0' && l != '.') || (l > '9' && l < 'a' && l != '_') || l > 'z' {
				return errInvalidUsername
			}
			if prev == '.' && l == '.' {
				return errConsecutivePeriod
			}
			prev = l
		}
		return nil
	}
}
