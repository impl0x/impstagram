package token

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateReferenceID() string {
	b := make([]byte, 24) // 192 bits of entropy
	rand.Read(b)          // ignore error as it is never returned according to the documentation
	// Use RawURLEncoding to avoid issues with standard base64 characters like '/'
	return "ref_" + base64.RawURLEncoding.EncodeToString(b)
}
