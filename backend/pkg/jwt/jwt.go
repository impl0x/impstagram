// Note:
//
// This is a custom jwt implementation according to my needs,
// this may or may not follow what the industry does.
//
// I'll be using 2 parts in the jwt instead of 3, as my header values will always be fixed to using HS256
// so it'll just be payload.signature, b64 encoded of course.
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"slices"
	"strings"
)

type Manager struct {
	secretKey []byte
}

// returns a new jwt manager using the secret key provided,
// by default only generates hmac sha256 tokens without a header
func NewManager(secretKey string) Manager {
	return Manager{[]byte(secretKey)}
}

// Generates a new jwt token string with the claim fields populated in the [Claims] struct
func (j Manager) GenerateToken(claims *Claims) string {
	// marshal body
	payloadBytes, err := json.Marshal(claims)
	if err != nil { // should never happen till we are using a known deterministic struct
		panic("jwt: json marshalling panic")
	}

	// encode body
	payloadB64 := make([]byte, base64.RawURLEncoding.EncodedLen(len(payloadBytes)))
	base64.RawURLEncoding.Encode(payloadB64, payloadBytes)

	// sign body
	mac := hmac.New(sha256.New, j.secretKey)
	mac.Write(payloadB64)
	signature := mac.Sum(nil)

	// encode signature
	signatureB64 := make([]byte, base64.RawStdEncoding.EncodedLen(len(signature)))
	base64.RawURLEncoding.Encode(signatureB64, signature)

	// join everything together for the token
	final := make([]byte, 0, len(payloadB64)+len(signatureB64)+1)
	final = append(final, payloadB64...)
	final = append(final, '.')
	final = append(final, signatureB64...)

	// return the string version
	return string(final)
}

var (
	ErrInvalidJWTToken   = errors.New("jwt: invalid jwt token")
	ErrIncorrectJWTToken = errors.New("jwt: incorrect jwt token") // only returned if the signature does not match, it means the token has been tampered with
)

// Verifies a token by checking its format and signature
//
// error can be [ErrInvalidJWTToken] or [ErrIncorrectJWTToken],
// read them for more information on when they are returned
func (j Manager) VerifyToken(token string) (*Claims, error) {
	// check format
	idx := strings.IndexByte(token, '.')
	if idx == 0 {
		return nil, ErrInvalidJWTToken
	}

	// separate parts
	bt := []byte(token)
	bodyEncoded := bt[:idx]
	signatureEncoded := bt[idx+1:]

	// decode base64 for signature
	signatureBytes := make([]byte, base64.RawURLEncoding.DecodedLen(len(signatureEncoded)))
	_, err := base64.RawURLEncoding.Decode(signatureBytes, signatureEncoded)
	if err != nil {
		return nil, ErrInvalidJWTToken
	}

	// calculate signature
	mac := hmac.New(sha256.New, j.secretKey)
	mac.Write(bodyEncoded)
	signature := mac.Sum(nil)

	// compare signature
	if slices.Compare(signature, signatureBytes) != 0 {
		return nil, ErrIncorrectJWTToken
	}

	// decode base64 for body
	bodyJsonStr := make([]byte, base64.RawURLEncoding.DecodedLen(len(bodyEncoded)))
	_, err = base64.RawURLEncoding.Decode(bodyJsonStr, bodyEncoded)
	if err != nil {
		return nil, ErrInvalidJWTToken
	}

	// unmarshal to struct
	var claims Claims
	err = json.Unmarshal(bodyJsonStr, &claims)
	if err != nil {
		return nil, ErrInvalidJWTToken
	}

	return &claims, nil
}
