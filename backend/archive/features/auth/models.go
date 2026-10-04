package auth

import (
	"errors"
	"time"
	"uuid"

	"github.com/mileusna/useragent"
)

// ? INFO
// helper file containing enums and models
// ! Ownership and usage
// owned and used by many files simultaneously,
// giving it a shared responsibility where whatever part of this file is
// used by any other file it also owns that particular part.
// ! Extra
// this file contains more or less enums, request models, jwt models, etc.

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

func (at accessTokenJwt) ToPayload() accessTokenJwtPayload {
	return accessTokenJwtPayload{
		UserID:    at.userID.String(),
		IssuedAt:  uint(at.issuedAt.Unix()),
		ExpiresAt: uint(at.expiresAt.Unix()),
		JwtID:     at.jwtID.String(),
	}
}

// ? ----+-----+-----DB Models-----+-----+-----

type userModel struct {
	// user data
	ID           uuid.UUID     `db:"id"`            // primary key default gen_random_uuid()
	Email        *string       `db:"email"`         // unique
	Phone        *string       `db:"phone"`         // unique
	PasswordHash string        `db:"password_hash"` // not null
	Dob          string        `db:"dob"`           // not null
	Status       accountStatus `db:"status"`        // not null default 'unverified'
	// 2fa related
	TotpSecretKey *string `db:"totp_secret_key"`
	TwoFAs        twoFAs  `db:"two_fas"` // slice of auth channels, if nil means twoFa not enabled, else its enabled on whichever identifiers are in the slice
	// timestamps
	CreatedAt time.Time `db:"created_at"` // not null default current_timestamp
	UpdatedAt time.Time `db:"updated_at"` // not null default current_timestamp
}

func newUserModel(req registerRequest, passwordHash string) *userModel {
	model := &userModel{
		PasswordHash: passwordHash,
		Dob:          req.Dob,
	}
	if req.Email != "" {
		model.Email = &req.Email
	}
	if req.Phone != "" {
		model.Phone = &req.Phone
	}
	return model
}

type userSessionModel struct {
	// session info
	ID        uuid.UUID `db:"id"`         // primary key default gen_random_uuid()
	JwtID     uuid.UUID `db:"jwt_id"`     // not null unique
	TokenHash string    `db:"token_hash"` // not null
	UserID    uuid.UUID `db:"user_id"`    // not null references users(id)
	// device info
	IPAddress   *string `db:"ip_address"`
	OSName      *string `db:"os_name"`
	BrowserName *string `db:"browser_name"`
	DeviceType  *string `db:"device_type"`
	// timestamps
	ExpiresAt time.Time `db:"expires_at"` // not null
	CreatedAt time.Time `db:"created_at"` // not null default current_timestamp
}

func newUserSessionModel(jwtID uuid.UUID, tokenHash string, rmd requestMetadata, userID uuid.UUID) *userSessionModel {
	// parsing the user agent and storing the current time
	ua := useragent.Parse(rmd.userAgent)

	// creating the userSession struct
	session := &userSessionModel{
		JwtID:     jwtID,
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().AddDate(0, 0, ruleExpiryTimeRefreshToken),
	}
	if rmd.IP != "" {
		session.IPAddress = &rmd.IP
	}
	if ua.OS != "" {
		session.OSName = &ua.OS
	}
	if ua.Name != "" {
		session.BrowserName = &ua.Name
	}
	if ua.Device != "" {
		session.DeviceType = &ua.Device
	}
	return session
}

type profileModel struct {
	// profile info
	UserID      uuid.UUID `db:"user_id"`  // primary key references users(id)
	Username    string    `db:"username"` // not null unique
	DisplayName *string   `db:"display_name"`
	AvatarUrl   *string   `db:"avatar_url"` //
	IsPrivate   bool      `db:"is_private"` // not null default false
	Bio         *string   `db:"bio"`
	// timestamps
	UpdatedAt time.Time `db:"updated_at"` // not null default current_timestamp
}

func newProfileModel(userID uuid.UUID, username string) *profileModel {
	return &profileModel{
		UserID:   userID,
		Username: username,
	}
}

// ? ----+-----+-----Request Body JSON-----+-----+-----

// ! important: make sure the validate tags are in align to the business rules and code, and custom validation tags are registered

type registerRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
	Dob      string `json:"dob" validate:"required,dob"`
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
