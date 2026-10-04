package validation

import "github.com/impl0x/mo/validator/v3"

// adds all the validation rules defined to global validator package instance
//
// Run this if using custom validations in the models
func AddValidations() {
	validator.AddCustomValidation(validateUsername())
}

// store keys are the keys used to store items in the context storage
type storeKey = string