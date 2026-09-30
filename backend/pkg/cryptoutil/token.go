package cryptoutil

import (
	"crypto/rand"
	"encoding/base64"
)

// returns random bytes of [size] prefixed with [prefix]
func GenerateToken(prefix string, size int) string {
	b := make([]byte, size)
	rand.Read(b) // ignore error as it is never returned according to the documentation
	return prefix + base64.RawURLEncoding.EncodeToString(b)
}
