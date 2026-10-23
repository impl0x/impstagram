package cryptoutil

import (
	"crypto/rand"
	"encoding/base64"
	mathrand "math/rand/v2"
)

// returns random bytes of [size] prefixed with [prefix]
func GenerateToken(prefix string, size int) string {
	b := make([]byte, size)
	rand.Read(b) // ignore error as it is never returned according to the documentation
	return prefix + base64.RawURLEncoding.EncodeToString(b)
}

const hex = "0123456789abcdef"

// Generates a pseudo random 64 bit hex id
func GeneratePseudoHexToken64Bit() string {
	var id [16]byte
	for i := range id {
		id[i] = hex[mathrand.IntN(16)]
	}
	return string(id[:])
}
