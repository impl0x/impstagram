package auth

// used to describe account status
type accountStatus int

const (
	statusUnverified accountStatus = iota
	statusVerified
	statusBanned
)

// authChannel defines WHAT medium is being used.
//
// disclaimer: this literal constant value is used for the responses sent to the user in the handler,
// it is directly type casted, so the values must be accurate!
type authChannel string

const (
	channelEmail    authChannel = "email"
	channelPhone    authChannel = "telegram"      // [UPDATE]: when switched to sms update this value
	channelUsername authChannel = "username"      // Used for DB lookup only
	channelTOTP     authChannel = "authenticator" // Used for 2FA validation only
)

// authPurpose defines WHY the OTP or action is happening.
type authPurpose string

const (
	purposeRegistration authPurpose = "registration"
	purpose2FA          authPurpose = "2fa"
	purposeResetPass    authPurpose = "reset_password"
)

type registerRequest struct {
	Username string `json:"username" validate:"required,min=3,max=15"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
	Dob      string `json:"dob" validate:"required"`

	Password string `json:"password" validate:"required,min=8"`
}

type loginRequest struct {
	Username string `json:"username" validate:"optional,min=3,max=15"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`

	Password string `json:"password" validate:"required,min=8"`
}

type verifyOTPRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,startswith=ref_"`
	OTP         string `json:"otp" validate:"required,numeric,len=6"`
}
