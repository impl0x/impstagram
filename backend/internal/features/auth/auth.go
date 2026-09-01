package auth

// ? INFO
// main file for this feature containing Init functions and stuff that is exposed to the app
// ! Ownership and usage
// owned by itself
// used by the external app after import

// This function runs all the necessary internal functions to set up everything
//   - adds validations
//
// This must be run on startup to ensure the feature runs without issues
func Init() {
	AddValidations()
}
