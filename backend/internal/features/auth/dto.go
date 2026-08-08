package auth

import "github.com/google/uuid"

type AccountStatus uint8

const (
	StatusUnverified AccountStatus = iota
	StatusVerified
	StatusBanned
)

type Identifier string

const (
	IdentifierEmail Identifier = "email"
	IdentifierPhone Identifier = "phone"
	IdentifierTOTP  Identifier = "totp"
)

type TwoFAs []Identifier

type User struct { // not complete, will do after i setup database properly
	Id            uuid.UUID
	Username      string
	Email         string
	Phone         string
	TotpSecretKey string // most of the times this will be null if the user has not enabled authenticator based totp
	PasswordHash  string
	Dob           string

	Status AccountStatus // account status which can either be unverified, verified or banned.
	TwoFA  TwoFAs        // slice of identifiers, if null means twoFa not enabled, else its enabled on whichever identifiers are in the slice
}

func NewUser(req RegisterRequest, passwordHash string) *User {
	return &User{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		Dob:          req.Dob,
		Status:       StatusUnverified,
	}
}
