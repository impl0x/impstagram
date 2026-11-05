package argon2id

import (
	"crypto/rand"
	"crypto/subtle"
)

// implementation functions

// Hash generates a random salt and returns the hashed password and the salt.
//
// errors returned are only read errors
func (m Manager) Hash(password string) string {
	salt := make([]byte, m.c.SaltLength)
	rand.Read(salt)
	hash := m.createHash(password, salt)
	return hash
}

// Compares the two passwords, i.e. the request password and the password hash stored in the database
//
// note that the error returned here can be argon2id format errors
// it mostly will not trigger if the data isn't invalid
func (m Manager) Compare(currentPassword string, storedHash string) (bool, error) {
	params, salt, rawStoredHash, err := DecodeHash(storedHash)
	if err != nil {
		return false, err
	}
	rawCurrentPasswordHash := CreateRawHashWithParams(currentPassword, []byte(salt), params)
	return subtle.ConstantTimeCompare(rawCurrentPasswordHash, []byte(rawStoredHash)) == 0, nil
}
