package cryptoutil

// interface to represent hashing passwords and comparing their hashes
type PasswordHasher interface {
	// creates a formatted hash string, ready to be stored, not recommended to be used to compare,
	// use the [PasswordHasher.Compare] function for that
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
