package auth

import "github.com/google/uuid"

type accountStatus uint8

const (
	statusUnverified accountStatus = iota
	statusVerified
	statusBanned
)

type identifier uint8

const (
	identifierEmail =iota
	identifierPhone 
	identifierTOTP  
)

func (i identifier) string() string {
	switch i{
	case identifierEmail:
		return "email"
	case identifierPhone:
		return "telegram"	// *CHANGE* if switched to sms
	case identifierTOTP:
		return "authenticator"
	default:
		panic("invalid identifier, please update code if added new identifiers")
	}
}

type twoFAs []identifier

type User struct { // not complete, will do after i setup database properly
	ID            uuid.UUID
	Username      string
	Email         string
	Phone         string
	TotpSecretKey string // most of the times this will be null in the database if the user has not enabled authenticator based totp
	PasswordHash  string
	Dob           string

	Status accountStatus // account status which can either be unverified, verified or banned.
	TwoFA  twoFAs        // slice of identifiers, if null means twoFa not enabled, else its enabled on whichever identifiers are in the slice
}

func NewUser(req registerRequest, passwordHash string) *User {
	return &User{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		Dob:          req.Dob,
		Status:       statusUnverified,
	}
}
