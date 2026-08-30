package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/mileusna/useragent"
)

type user struct {
	ID            uuid.UUID `db:"id"`
	DisplayName   string    `db:"display_name"`
	Username      string    `db:"username"`
	Email         string    `db:"email"`
	Phone         string    `db:"phone"`
	TotpSecretKey string    `db:"totp_secret_key"` // most of the times this will be empty in the database if the user has not enabled authenticator based totp
	PasswordHash  string    `db:"password_hash"`
	Dob           string    `db:"dob"`

	Status accountStatus `db:"status"`  // account status which can either be unverified, verified or banned.
	TwoFAs twoFAs        `db:"two_fas"` // slice of auth channels, if nil means twoFa not enabled, else its enabled on whichever identifiers are in the slice

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func newUser(req registerRequest, passwordHash string) *user {
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
	ID          uuid.UUID `db:"id"`
	JwtID       uuid.UUID `db:"jwt_id"`
	TokenHash   string    `db:"token_hash"`
	UserID      uuid.UUID `db:"user_id"`
	IPAddress   string    `db:"ip_address"`
	OSName      string    `db:"os_name"`
	BrowserName string    `db:"browser_name"`
	DeviceType  string    `db:"device_type"`
	ExpiresAt   time.Time `db:"expires_at"`
	CreatedAt   time.Time `db:"created_at"`
}

func newUserSession(jwtID uuid.UUID, tokenHash, userIP, userAgent string, userID uuid.UUID) *userSession {
	// parsing the user agent and storing the current time
	ua := useragent.Parse(userAgent)

	// creating and returning the userSession struct
	return &userSession{
		JwtID:       jwtID,
		TokenHash:   tokenHash,
		UserID:      userID,
		IPAddress:   userIP,
		OSName:      ua.OS,
		BrowserName: ua.Name,
		DeviceType:  ua.Device,
		ExpiresAt:   time.Now().AddDate(0, 0, ruleExpiryTimeRefreshToken),
	}
}
