package auth

import "github.com/google/uuid"

type accountStatus uint8

const (
	statusUnverified accountStatus = iota
	statusVerified
	statusBanned
)

type identifier uint

const (
	identifierEmail identifier =iota
	identifierPhone 
	identifierUsername
)

type twoFactorType uint

const (
	twoFactorEmail twoFactorType = iota
	twoFactorPhone
	twoFactorTOTP
)

func (tft twoFactorType) string()string{
	switch tft{
	case twoFactorEmail:
		return "email"
	case twoFactorPhone:
		return "telegram" // change if switched to sms
	case twoFactorTOTP:
		return "authenticator"
	default:
		panic("Invalid two factor authentication type, please add more cases in the string method for this enum if added new types")
	}
}

type User struct { // not complete, will do after i setup database properly
	ID            uuid.UUID
	Username      string
	Email         string
	Phone         string
	TotpSecretKey string // most of the times this will be null in the database if the user has not enabled authenticator based totp
	PasswordHash  string
	Dob           string

	Status accountStatus // account status which can either be unverified, verified or banned.
	TwoFAs  []twoFactorType        // slice of 2FA identifiers, if null means twoFa not enabled, else its enabled on whichever identifiers are in the slice
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
