package validation

import "errors"

const Username = "username"

// Username validation logic:
//   - allowed a-z 0-9 _ .
//   - first!=0-9
//   - warn: does not check length
//
// tag: username
func validateUsername() (string, func(any, string) error) {
	// storing sentinel errors to avoid heap allocation on every call to func
	errEmptyUsername := errors.New("Empty username")
	errStartsWithPeriod := errors.New("Cannot start username with a period")
	errStartsWithNumber := errors.New("Cannot start username with a number")
	errInvalidUsername := errors.New("Username can only contain lowercase letters, numbers and underscore")
	errConsecutivePeriod := errors.New("Username cannot contain consecutive periods")
	return Username, func(v any, _ string) error {
		s := v.(string) // will panic if the tag is used on a field that isn't string.
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
