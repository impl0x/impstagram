package auth

import (
	"backend/internal/config"
	"time"

	"github.com/google/uuid"
	"github.com/mileusna/useragent"
)

type user struct { // not complete, will do after i setup database properly
	ID            uuid.UUID
	Username      string
	Email         string
	Phone         string
	TotpSecretKey string // most of the times this will be null in the database if the user has not enabled authenticator based totp
	PasswordHash  string
	Dob           string

	Status accountStatus // account status which can either be unverified, verified or banned.
	TwoFAs []authChannel // slice of 2FA identifiers, if null means twoFa not enabled, else its enabled on whichever identifiers are in the slice
}

func NewUser(req registerRequest, passwordHash string) *user {
	return &user{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		Dob:          req.Dob,
		Status:       statusUnverified,
	}
}

type userSession struct {
	ID          uuid.UUID
	JwtID       uuid.UUID
	TokenHash   string
	UserID      uuid.UUID
	IPAddress   string
	OSName      string
	BrowserName string
	DeviceType  string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

func newUserSession(jwtID uuid.UUID, tokenHash, userIP, userAgent string, userID uuid.UUID) *userSession {
	// parsing the user agent and storing the current time
	ua := useragent.Parse(userAgent)
	now := time.Now()

	// creating and returning the userSession struct
	return &userSession{
		JwtID:       jwtID,
		TokenHash:   tokenHash,
		UserID:      userID,
		IPAddress:   userIP,
		OSName:      ua.OS,
		BrowserName: ua.Name,
		DeviceType:  ua.Device,
		ExpiresAt:   now.AddDate(0, 0, config.RefreshTokenExpiryTime),
		CreatedAt:   now,
	}
}
