// Note:
//
// This is a custom jwt implementation according to my needs,
// this may or may not follow what the industry does.
//
// I'll be using 2 parts in the jwt instead of 3, as my header values will always be fixed to using HS256
// so it'll just be payload.signature, b64 encoded of course.
package jwt

import (
	"backend/internal/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

var (
	ErrInvalidJsonPayload = errors.New("jwt: invalid json payload")
)

// Payload must be a json compatible struct, do not use maps. It's not efficient, use structs.
//
// error can only be a invalid json error
func GenerateToken(payload any) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", ErrInvalidJsonPayload
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	mac := hmac.New(sha256.New, []byte(config.JwtHMACKey))
	mac.Write([]byte(payloadB64))
	signature := mac.Sum(nil)

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)
	return payloadB64 + "." + signatureB64, nil
}

var (
	ErrInvalidJWTToken   = errors.New("jwt: invalid jwt token")
	ErrIncorrectJWTToken = errors.New("jwt: incorrect jwt token")
)

// If error is nil then token is valid, error can be the jwt errors listed in here or a json decode error.
//
// target must be a valid json compatible struct, we decode and unmarshal the jwt payload into that struct
//
// validation is not done here, so validate it yourself
func VerifyToken(token string, target any) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return ErrInvalidJWTToken
	}

	mac := hmac.New(sha256.New, []byte(config.JwtHMACKey))
	mac.Write([]byte(parts[0]))
	signature := mac.Sum(nil)

	if string(signature) != parts[1] {
		return ErrIncorrectJWTToken
	}

	err := json.Unmarshal([]byte(parts[0]), target)
	if err != nil {
		return err
	}
	return nil
}
