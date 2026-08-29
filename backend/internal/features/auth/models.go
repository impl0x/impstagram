package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ? ----+-----+-----Store keys-----+-----+-----
// store keys are the keys used to store items in the context storage
type storeKey = string

const (
	keyAccessToken storeKey = "jwt"
)

// ? ----+-----+-----Enums-----+-----+-----
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
	channelPhone    authChannel = "phone"
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

type twoFAs []authChannel

// ? ----+-----+-----Common-----+-----+-----

// Some service functions require this metadata for session storage
type requestMetadata struct {
	IP        string
	userAgent string
}

// ? ----+-----+-----JWT-----+-----+-----
// jwt helper functions for generating access tokens

// The access token jwt payload used in the actual token data after encoding
type accessTokenJwtPayload struct {
	UserID    string `json:"sub" validate:"required,len=36"`
	IssuedAt  uint   `json:"iat" validate:"required"`
	ExpiresAt uint   `json:"exp" validate:"required"`
	JwtID     string `json:"jti" validate:"required,len=36"`
}

// Usable struct for the service with converted data types
type accessTokenJwt struct {
	userID    uuid.UUID
	issuedAt  time.Time
	expiresAt time.Time
	jwtID     uuid.UUID
}

// Generates a new payload with the Access Token expiry time in it using the [BasicPayload]
func newAccessTokenPayload(userID uuid.UUID, jwtID uuid.UUID, expiryTime time.Duration) accessTokenJwtPayload {
	now := time.Now()
	return accessTokenJwtPayload{
		UserID:    userID.String(),
		IssuedAt:  uint(now.Unix()),
		ExpiresAt: uint(now.Add(expiryTime).Unix()),
		JwtID:     jwtID.String(),
	}
}

var errJwtInvalidUUID = errors.New("auth.models: invalid uuid")

// Converts a AccessTokenPayload to a more usable AccessToken type with the values being converted to usable uuid.UUID and time.Time
//
// only returns error of errJwtInvalidUUID if the UUID parsing fails
func (atp accessTokenJwtPayload) Convert() (accessTokenJwt, error) {
	userID, err := uuid.Parse(atp.UserID)
	if err != nil {
		return accessTokenJwt{}, errJwtInvalidUUID
	}
	jwtID := uuid.MustParse(atp.JwtID)
	if err != nil {
		return accessTokenJwt{}, errJwtInvalidUUID
	}
	return accessTokenJwt{
		userID:    userID,
		issuedAt:  time.Unix(int64(atp.IssuedAt), 0),
		expiresAt: time.Unix(int64(atp.ExpiresAt), 0),
		jwtID:     jwtID,
	}, nil
}

// ? ----+-----+-----Request Body JSON-----+-----+-----

// ! important: make sure the validate tags are in align to the business rules and code

type registerRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
	Dob      string `json:"dob" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

type loginRequest struct {
	Username string `json:"username" validate:"optional,min=3,max=30,username"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`

	Password string `json:"password" validate:"required,min=8,max=20"`
}

type resendOTPRequest struct {
	Purpose  string `json:"purpose" validate:"required,oneof=registration 2fa reset_password"`
	Username string `json:"username" validate:"optional,min=3,max=30,username"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
}

type verifyOTPRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,startswith=otp_"`
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

type add2FARequest struct {
	Channel string `json:"channel" validate:"required,oneof=email phone"`
}

type remove2FARequest struct {
	Channel string `json:"channel" validate:"required,oneof=email phone totp"`
}

type totpVerifyRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,startswith=totp_"`
	OTP         string `json:"otp" validate:"required,numeric,len=6"`
}
