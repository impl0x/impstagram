package cryptoutil

import (
	"backend/internal/config"
	"backend/pkg/cryptoutil/argon2id"
	"crypto/rand"
	"crypto/subtle"
)

// interface to represent hashing passwords and comparing their hashes
type PasswordHasher interface {
	// creates a formatted hash string, ready to be stored, not recommended to be used to compare, use the [PasswordHasher.Compare] function for that
	Hash(password string) string

	// Compares the two passwords, i.e. the current password and the password hash stored in the repository
	//
	// the compare function expects storedHash to be hashed by the same implementation used to compare it
	//
	// therefore the error that can be returned here can be hash format errors.
	// which occur if the format of hash string is incorrect,
	// i.e. if a different hash implementation has been used to hash it and has been
	// formatted not according to expectations by this implementation
	Compare(currentPassword string, storedHash string) (bool, error)
}

// argon2id implementation for [PasswordHasher]
type Argon2ID struct {
	argon2id argon2id.Argon2ID
	saltLen  uint
}

func NewArgon2ID(c config.PasswordConfig) Argon2ID {
	return Argon2ID{
		argon2id: argon2id.NewArgon2ID(
			argon2id.Params{
				Memory:      uint32(c.Memory),
				Iterations:  uint32(c.Iterations),
				Parallelism: uint8(c.Parallelism),
				KeyLength:   uint32(c.KeyLength),
			},
			c.SaltLength,
		),
		saltLen: uint(c.SaltLength),
	}
}

// Hash generates a random salt and returns the hashed password and the salt.
//
// errors returned are only read errors
func (pm Argon2ID) Hash(password string) string {
	salt := make([]byte, pm.saltLen)
	rand.Read(salt)
	hash := pm.argon2id.CreateHash(password, salt)
	return hash
}

// Compares the two passwords, i.e. the request password and the password hash stored in the database
//
// note that the error returned here can be argon2id format errors
// it mostly will not trigger if the data isn't invalid
func (pm Argon2ID) Compare(currentPassword string, storedHash string) (bool, error) {
	params, salt, rawStoredHash, err := argon2id.DecodeHash(storedHash)
	if err != nil {
		return false, err
	}
	rawCurrentPasswordHash := argon2id.CreateRawHashWithParams(currentPassword, []byte(salt), params)
	return subtle.ConstantTimeCompare(rawCurrentPasswordHash, []byte(rawStoredHash)) == 0, nil
}
