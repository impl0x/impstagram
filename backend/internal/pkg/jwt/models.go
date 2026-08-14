package jwt

import (
	"backend/internal/config"
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

// Generates a new payload with the Access Token expiry time in it using the [BasicPayload]
func NewAccessTokenPayload(userID uuid.UUID, jwtID uuid.UUID) AccessTokenPayload {
	now := time.Now()
	return AccessTokenPayload{
		UserID:    userID.String(),
		IssuedAt:  uint(now.Unix()),
		ExpiresAt: uint(now.Add(config.AccessTokenExpiryTime).Unix()),
		JwtID:     jwtID.String(),
	}
}
