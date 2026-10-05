package validation

import (
	"backend/pkg/dob"
	"errors"
)

// [Username] errors
var (
	errUsernameEmpty             = errors.New("Username empty")
	errUsernameStartsWithPeriod  = errors.New("Username cannot start with a period")
	errUsernameStartsWithNumber  = errors.New("Username cannot start with a number")
	errUsernameConsecutivePeriod = errors.New("Username cannot contain consecutive periods")
	errUsernameInvalidUsername   = errors.New("Username can only contain lowercase letters, numbers and underscore")
)

// The required rules that validation works upon, must be aligned with the validation function at all times.
var UsernameRules = []string{
	"Cannot start with a period",
	"Cannot start with a number",
	"Cannot contain consecutive periods",
	"Can only contain lowercase letters, numbers and underscore",
}

// Username validation logic:
//   - allowed a-z 0-9 _ .
//   - first!=0-9
//   - warn: does not check length
//
// fails on the first violation with the specific error returned
func Username(s string) error {
	if s == "" {
		return errUsernameEmpty
	}
	if s[0] == '.' {
		return errUsernameStartsWithPeriod
	}
	if s[0] >= '0' && s[0] <= '9' {
		return errUsernameStartsWithNumber
	}
	prev := '-'
	for _, l := range s {
		if (l < '0' && l != '.') || (l > '9' && l < 'a' && l != '_') || l > 'z' {
			return errUsernameInvalidUsername
		}
		if prev == '.' && l == '.' {
			return errUsernameConsecutivePeriod
		}
		prev = l
	}
	return nil
}

// Dob errors
var (
	errEmptyDob           = errors.New("Dob is " + dob.ErrEmpty.Error()[4:])
	errDobIncorrectFormat = errors.New("Dob has " + dob.ErrIncorrectFormat.Error()[4:])
	errDobNotParseable    = errors.New("Dob is " + dob.ErrNotParseable.Error()[4:])
	errDobImpossible      = errors.New("Dob is impossible")
)

// taken from [dob]
var DobRules = []string{
	"Format is YYYY-MM-DD",
	"Should look like this: 2000-01-01",
	"Year must be more than 1900 and less than or equal to current year",
	"Month must be more than 01 and less than or equal to 12",
	"Day must be more than 01 and less than or equal to 31",
}

// Validates a date of birth string (dob), the validation is based on [dob.Parse] function, as so are the errors and rules.
func Dob(s string) error {
	_, err := dob.Parse(s)
	switch err {
	case dob.ErrEmpty:
		return errEmptyDob
	case dob.ErrIncorrectFormat:
		return errDobIncorrectFormat
	case dob.ErrNotParseable:
		return errDobNotParseable
	case dob.ErrImpossible:
		return errDobImpossible
	default: // assuming dob package does not add any other error types in the future,
		return nil
	}
}
