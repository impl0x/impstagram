// parts of this package is taken from github.com/alexedwards/argon2id
package argon2id

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	// ErrInvalidHash in returned by ComparePasswordAndHash if the provided
	// hash isn't in the expected format.
	ErrInvalidHash = errors.New("argon2id: hash is not in the correct format")

	// ErrIncompatibleVariant is returned by ComparePasswordAndHash if the
	// provided hash was created using a unsupported variant of Argon2.
	// Currently only argon2id is supported by this package.
	ErrIncompatibleVariant = errors.New("argon2id: incompatible variant of argon2")

	// ErrIncompatibleVersion is returned by ComparePasswordAndHash if the
	// provided hash was created using a different version of Argon2.
	ErrIncompatibleVersion = errors.New("argon2id: incompatible version of argon2")
)

// password hashing configuration/parameters
type Config struct {
	// The amount of memory used by the algorithm (in kilobytes).
	Memory uint32
	// The number of iterations over the memory.
	Iterations uint32
	// The number of threads (or lanes) used by the algorithm.
	// Recommended value is between 1 and runtime.NumCPU().
	Parallelism uint8
	// Length of the salt used in the hash, 8 bytes or more is recommended.
	SaltLength uint32
	// Length of the generated key. 16 bytes or more is recommended.
	KeyLength uint32
}

// implements [cryptoutil.PasswordHasher]
type Manager struct {
	c                  Config
	formattedFirstPart string
	totalStrLen        int
}

// returns a new argon2id implementation of [cryptoutil.PasswordHasher]
// Manager struct which can be used to manage password hashes
func New(c Config) Manager {
	s := "$argon2id$v=" +
		strconv.Itoa(argon2.Version) +
		"$m=" +
		strconv.Itoa(int(c.Memory)) +
		",t=" +
		strconv.Itoa(int(c.Iterations)) +
		",p=" +
		strconv.Itoa(int(c.Parallelism)) +
		"$"

	l := len(s)
	l += base64.RawStdEncoding.EncodedLen(int(c.SaltLength))
	l += base64.RawStdEncoding.EncodedLen(int(c.KeyLength))
	return Manager{
		c:                  c,
		formattedFirstPart: s,
		totalStrLen:        l,
	}
}

// creates a raw hash with salt and params provided
func CreateRawHashWithParams(password string, salt []byte, params Config) []byte {
	return argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)
}

// creates a raw hash without formatting
func (m Manager) createRawHash(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, m.c.Iterations, m.c.Memory, m.c.Parallelism, m.c.KeyLength)
}

// creates a formatted hash with the params in struct
func (m Manager) createHash(password string, salt []byte) string {
	key := m.createRawHash(password, salt)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	// we store it in this format: $argon2id$v=19$m=65536,t=1,p=4$ZWFzeXNhbHQ$R01...NDA
	// where, v is version, m for memory, t for time/iterations, p for parallelism/threads
	return m.formattedFirstPart + b64Salt + "$" + b64Key
}

// expects format example:
//
// $argon2id$v=19$m=19456,t=2,p=2$LPz5gLBKWtk$cZXDRxSOh1JAImjgmo2vOs+IClCyGQyS+sFM9CKYzXk
//
// returns config, salt, key, error
func DecodeHash(hash string) (Config, string, string, error) {
	if hash[0] != '$' {
		return Config{}, "", "", ErrInvalidHash
	}

	hash = hash[1:]
	idx := strings.IndexByte(hash, '$')
	if hash[:idx] != "argon2id" {
		return Config{}, "", "", ErrIncompatibleVariant
	}

	hash = hash[idx+1:]
	if hash[:2] != "v=" {
		return Config{}, "", "", ErrInvalidHash
	}

	hash = hash[2:]
	idx = strings.IndexByte(hash, '$')
	versionNum, err := strconv.ParseUint(hash[:idx], 10, 32)
	if err != nil {
		return Config{}, "", "", ErrInvalidHash
	}
	if versionNum != argon2.Version {
		return Config{}, "", "", ErrIncompatibleVersion
	}

	hash = hash[idx+1:]
	if hash[:2] != "m=" {
		return Config{}, "", "", ErrInvalidHash
	}

	hash = hash[2:]
	idx = strings.IndexByte(hash, ',')
	memoryNum, err := strconv.ParseUint(hash[:idx], 10, 32)
	if err != nil {
		return Config{}, "", "", ErrInvalidHash
	}

	hash = hash[idx+1:]
	if hash[:2] != "t=" {
		return Config{}, "", "", ErrInvalidHash
	}

	hash = hash[2:]
	idx = strings.IndexByte(hash, ',')
	iterationsNum, err := strconv.ParseUint(hash[:idx], 10, 32)
	if err != nil {
		return Config{}, "", "", ErrInvalidHash
	}

	hash = hash[idx+1:]
	if hash[:2] != "p=" {
		return Config{}, "", "", ErrInvalidHash
	}

	hash = hash[2:]
	idx = strings.IndexByte(hash, '$')
	parallelismNum, err := strconv.ParseUint(hash[:idx], 10, 8)
	if err != nil {
		return Config{}, "", "", ErrInvalidHash
	}

	hash = hash[idx+1:]
	idx = strings.IndexByte(hash, '$')
	salt := hash[:idx]

	saltLen := base64.RawStdEncoding.DecodedLen(len(salt))

	key := hash[idx+1:]
	keyLen := base64.RawStdEncoding.DecodedLen(len(key))

	params := Config{
		Memory:      uint32(memoryNum),
		Iterations:  uint32(iterationsNum),
		Parallelism: uint8(parallelismNum),
		SaltLength:  uint32(saltLen),
		KeyLength:   uint32(keyLen),
	}
	return params, salt, key, nil
}
