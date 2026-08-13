package token

import (
	"crypto/rand"
	"encoding/base64"
)

const refreshTokenSize uint8 = 32 // bytes

func GenerateRefreshToken() string {
	bytes := make([]byte, refreshTokenSize)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)
}
