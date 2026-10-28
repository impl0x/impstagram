package httpserver

import (
	"backend/internal/validation"

	"github.com/impl0x/mo/validator/v3"
)

type tagName = string

const UsernameTag tagName = "username"
const DobTag tagName = "dob"

// adds all the validation rules defined to global validator package instance
//
// Run this if using custom validations in the models
func AddValidations() {
	validator.AddCustomValidation(UsernameTag, wrap(validation.Username))
	validator.AddCustomValidation(DobTag, wrap(validation.Dob))
}

// wraps a normal validation function without parameter to a [validator.CustomValidatorFunc] type
func wrap(fn func(s string) error) validator.CustomValidatorFunc {
	return func(v any, _ string) error {
		return fn(v.(string))
	}
}
