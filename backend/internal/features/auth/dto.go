package auth

import "github.com/google/uuid"

type AccountStatus uint8

const (
	StatusUnverified AccountStatus = iota
	StatusVerified
	StatusBanned
)

type Identifiers string

const (
	IdentifierEmail Identifiers = "email"
	IdentifierPhone Identifiers = "phone"
)

type TwoFAs []Identifiers

type User struct { // not complete, will do after i setup database properly
	Id           uuid.UUID
	Username     string
	Email        string
	Phone        string
	PasswordHash string
	Dob          string

	Status AccountStatus // account status which can either be unverified, verified or banned.
	TwoFA  TwoFAs         // slice of identifiers, if null means twoFa not enabled, else its enabled on whichever identifiers are in the slice
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
