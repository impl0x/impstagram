package jwt

import (
	"time"

	"github.com/google/uuid"
)

// The access token jwt payload for this app
type AccessTokenPayload struct {
	UserID    string `json:"sub" validate:"required,len=36"`
	IssuedAt  uint   `json:"iat" validate:"required"`
	ExpiresAt uint   `json:"exp" validate:"required"`
	JwtID     string `json:"jti" validate:"required,len=36"`
}

// Usable struct for the app with converted data types
type AccessToken struct {
	UserID    uuid.UUID
	IssuedAt  time.Time
	ExpiresAt time.Time
	JwtID     uuid.UUID
}

// Generates a new payload with the Access Token expiry time in it using the [BasicPayload]
func NewAccessTokenPayload(userID uuid.UUID, jwtID uuid.UUID, expiryTime time.Duration) AccessTokenPayload {
	now := time.Now()
	return AccessTokenPayload{
		UserID:    userID.String(),
		IssuedAt:  uint(now.Unix()),
		ExpiresAt: uint(now.Add(expiryTime).Unix()),
		JwtID:     jwtID.String(),
	}
}

func (atp AccessTokenPayload) Convert() (AccessToken, error) {
	userID, err := uuid.Parse(atp.UserID)
	if err != nil {
		return AccessToken{}, ErrInvalidJsonPayload
	}
	jwtID, err := uuid.Parse(atp.JwtID)
	if err != nil {
		return AccessToken{}, ErrInvalidJsonPayload
	}
	return AccessToken{
		UserID:    userID,
		IssuedAt:  time.Unix(int64(atp.IssuedAt), 0),
		ExpiresAt: time.Unix(int64(atp.ExpiresAt), 0),
		JwtID:     jwtID,
	}, nil
}
