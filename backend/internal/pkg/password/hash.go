package password

import (
	"backend/internal/pkg/password/argon2id"
	"crypto/rand"
	"crypto/subtle"
)

// Tuned for ~20-30ms on modern CPUs.
var DefaultArgon2HashSettings = argon2id.Params{
	Memory:      64 * 1024, // 64 MB
	Iterations:  1,         // Time
	Parallelism: 4,         // Threads
	SaltLength:  16,
	KeyLength:   32,
}

// Hash generates a random salt and returns the hashed password and the salt.
//
// errors returned are only read errors
func Hash(password string) (string, error) {
	salt, err := generateRandomBytes(DefaultArgon2HashSettings.SaltLength)
	if err != nil {
		return "", err
	}
	hash := argon2id.CreateHash(password, salt, DefaultArgon2HashSettings)
	return hash, nil
}

// Compares the two passwords, i.e. the request password and the password hash stored in the database
//
// note that the error returned here can be anything from rand read errors to argon2id format errors
// it mostly will not trigger if the data isn't invalid
func Compare(reqPassword string, dbPasswordHash string) (bool, error) {
	params, salt, rawDbPasswordHash, err := argon2id.DecodeHash(dbPasswordHash)
	if err != nil {
		return false, err
	}
	rawReqPasswordHash := argon2id.CreateRawHash(reqPassword, salt, params)

	return subtle.ConstantTimeCompare(rawReqPasswordHash, rawDbPasswordHash) == 0, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
