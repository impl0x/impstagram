package cryptoutil

import (
	"backend/internal/config"
	"backend/pkg/cryptoutil/argon2id"
	"crypto/rand"
	"crypto/subtle"
)

type PasswordManager struct {
	argon2id argon2id.Argon2ID
}

func NewPasswordManager(c config.PasswordConfig) PasswordManager {
	return PasswordManager{
		argon2id: argon2id.NewArgon2ID(
			argon2id.Params{
				Memory:      uint32(c.Memory),
				Iterations:  uint32(c.Iterations),
				Parallelism: uint8(c.Parallelism),
				SaltLength:  uint32(c.SaltLength),
				KeyLength:   uint32(c.KeyLength),
			},
		),
	}
}

// Hash generates a random salt and returns the hashed password and the salt.
//
// errors returned are only read errors
func (pm PasswordManager) Hash(password string) string {
	salt := make([]byte, pm.argon2id.Params.SaltLength)
	rand.Read(salt)
	hash := pm.argon2id.CreateHash(password, salt)
	return hash
}

// Compares the two passwords, i.e. the request password and the password hash stored in the database
//
// note that the error returned here can be anything from rand read errors to argon2id format errors
// it mostly will not trigger if the data isn't invalid
func (pm PasswordManager) Compare(reqPassword string, dbPasswordHash string) (bool, error) {
	params, salt, rawDbPasswordHash, err := argon2id.DecodeHash(dbPasswordHash)
	if err != nil {
		return false, err
	}
	rawReqPasswordHash := pm.argon2id.CreateRawHash(reqPassword, []byte(salt), params)
	return subtle.ConstantTimeCompare(rawReqPasswordHash, []byte(rawDbPasswordHash)) == 0, nil
}
