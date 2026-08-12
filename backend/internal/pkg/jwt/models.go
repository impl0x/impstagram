package jwt

import (
	"backend/internal/config"
	"time"

	"github.com/google/uuid"
)

type BasicPayload struct {
	UserID string `json:"sub" validate:"len=36"`
	Iat    uint   `json:"iat"`
	Exp    uint   `json:"exp"`
}

// Generates a new payload with the Access Token expiry time in it using the [BasicPayload]
func NewAccessTokenPayload(userID uuid.UUID) BasicPayload {
	now := time.Now()
	return BasicPayload{
		UserID: userID.String(),
		Iat:    uint(now.Unix()),
		Exp:    uint(now.Add(config.AccessTokenExpiryTime).Unix()),
	}
}
