package auth

// ? ----+-----+-----enums-----+-----+-----
// enums are made to be equal to the database level enums for compatibility and maintainability

// used to describe account status
type accountStatus string

const (
	statusUnverified accountStatus = "unverified"
	statusVerified   accountStatus = "verified"
	statusBanned     accountStatus = "banned"
)

// authChannel defines WHAT medium is being used.
//
// use the .String() method to get the string equivalent which can be sent to the user response
type authChannel string

const (
	channelEmail    authChannel = "email"
	channelPhone    authChannel = "phone"    // [UPDATE]: when switched to sms update this value
	channelUsername authChannel = "username" // Used for DB lookup only/ login
	channelTOTP     authChannel = "totp"     // Used for 2FA validation only
)

func (ac authChannel) String() string {
	switch ac {
	case channelPhone:
		return "telegram"
	case channelTOTP:
		return "authenticator"
	default:
		return string(ac)
	}
}

// authPurpose defines WHY the OTP or action is happening.
type authPurpose string

const (
	purposeRegistration authPurpose = "registration"
	purpose2FA          authPurpose = "2fa"
	purposeResetPass    authPurpose = "reset_password"
)

// ? ----+-----+-----Common-----+-----+-----

// Some service functions require this metadata for session storage
type requestMetadata struct {
	IP        string
	userAgent string
}

// ? ----+-----+-----Request Body JSON-----+-----+-----

// ! important:
// make sure the startswith value is always correct to the config prefix value

type registerRequest struct {
	Username string `json:"username" validate:"required,min=3,max=15"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
	Dob      string `json:"dob" validate:"required"`

	Password string `json:"password" validate:"required,min=8,max=20"`
}

type loginRequest struct {
	Username string `json:"username" validate:"optional,min=3,max=15"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`

	Password string `json:"password" validate:"required,min=8,max=20"`
}

type verifyOTPRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,startswith=otp_"` // make sure the startswith value is always correct to the config prefix value
	OTP         string `json:"otp" validate:"required,numeric,len=6"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" validate:"optional,email"`
	Phone string `json:"phone" validate:"optional,e.164"`
}

type resetPasswordRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,startswith=pwd_"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=20"`
}
