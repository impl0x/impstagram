package cryptoutil

import (
	"backend/internal/config"
	"backend/pkg/cryptoutil/argon2id"
	"crypto/rand"
	"crypto/subtle"
)

type PasswordManager struct {
	params argon2id.Params
}

func NewPasswordManager(cp config.PasswordConfig) PasswordManager {
	return PasswordManager{
		params: argon2id.Params{
			Memory:      uint32(cp.Memory),
			Iterations:  uint32(cp.Iterations),
			Parallelism: uint8(cp.Parallelism),
			SaltLength:  uint32(cp.SaltLength),
			KeyLength:   uint32(cp.KeyLength),
		},
	}
}

// Hash generates a random salt and returns the hashed password and the salt.
//
// errors returned are only read errors
func (pm PasswordManager) Hash(password string) string {
	salt := make([]byte, pm.params.SaltLength)
	rand.Read(salt)
	hash := argon2id.CreateHash(password, salt, pm.params)
	return hash
}

// Compares the two passwords, i.e. the request password and the password hash stored in the database
//
// note that the error returned here can be anything from rand read errors to argon2id format errors
// it mostly will not trigger if the data isn't invalid
func (_ PasswordManager) Compare(reqPassword string, dbPasswordHash string) (bool, error) {
	params, salt, rawDbPasswordHash, err := argon2id.DecodeHash(dbPasswordHash)
	if err != nil {
		return false, err
	}
	rawReqPasswordHash := argon2id.CreateRawHash(reqPassword, salt, params)
	return subtle.ConstantTimeCompare(rawReqPasswordHash, rawDbPasswordHash) == 0, nil
}
