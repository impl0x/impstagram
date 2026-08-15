package token

import (
	"backend/internal/config"
	"crypto/rand"
	"encoding/base64"
)

// returns random bytes of [size] prefixed with [prefix]
func generateToken(prefix string, size uint8) string {
	b := make([]byte, size)
	rand.Read(b) // ignore error as it is never returned according to the documentation
	return prefix + base64.RawURLEncoding.EncodeToString(b)
}

// Generates a new random string with the prefix [config.PrefixRefreshToken] and size [config.SizeRefreshToken]
func GenerateRefreshToken() string {
	return generateToken(config.PrefixRefreshToken, config.SizeRefreshToken)
}

// Generates a new random string with the prefix [config.PrefixOTPSession] and size [config.SizeSessionID]
func GenerateOTPSessionID() string {
	return generateToken(config.PrefixOTPSession, config.SizeSessionID)
}

// Generates a new random string with the prefix [config.PrefixResetSession] and size [config.SizeSessionID]
func GenerateResetSessionID() string {
	return generateToken(config.PrefixResetSession, config.SizeSessionID)
}
