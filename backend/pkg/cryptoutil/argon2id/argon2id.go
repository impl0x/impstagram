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

type Params struct {
	// The amount of memory used by the algorithm (in kilobytes).
	Memory uint32

	// The number of iterations over the memory.
	Iterations uint32

	// The number of threads (or lanes) used by the algorithm.
	// Recommended value is between 1 and runtime.NumCPU().
	Parallelism uint8

	// Length of the random salt. 16 bytes is recommended for password hashing.
	SaltLength uint32

	// Length of the generated key. 16 bytes or more is recommended.
	KeyLength uint32
}

type Argon2ID struct {
	Params           Params
	formattedHashLen int
}

func NewArgon2ID(p Params) Argon2ID {
	l := 23
	v := strconv.Itoa(argon2.Version)
	l += len(v)
	m := strconv.Itoa(int(p.Memory))
	l += len(m)
	i := strconv.Itoa(int(p.Iterations))
	l += len(i)
	pr := strconv.Itoa(int(p.Parallelism))
	l += len(pr)
	l += base64.RawStdEncoding.EncodedLen(int(p.SaltLength))
	l += base64.RawStdEncoding.EncodedLen(int(p.KeyLength))
	return Argon2ID{
		Params:           p,
		formattedHashLen: l,
	}
}

func (a Argon2ID) CreateRawHash(password string, salt []byte, params Params) []byte {
	return argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)
}

func (a Argon2ID) CreateHash(password string, salt []byte) string {
	key := argon2.IDKey([]byte(password), salt, a.Params.Iterations, a.Params.Memory, a.Params.Parallelism, a.Params.KeyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	// we store it in this format: $argon2id$v=19$m=65536,t=1,p=4$ZWFzeXNhbHQ$R01...NDA
	// where, v is version, m for memory, t for time/iterations, p for parallelism/threads
	// var buf strings.Builder

	var buf strings.Builder
	buf.Grow(a.formattedHashLen)
	buf.WriteString("$argon2id$v=")
	buf.WriteString(strconv.Itoa(argon2.Version))
	buf.WriteString("$m=")
	buf.WriteString(strconv.Itoa(int(a.Params.Memory)))
	buf.WriteString(",t=")
	buf.WriteString(strconv.Itoa(int(a.Params.Iterations)))
	buf.WriteString(",p=")
	buf.WriteString(strconv.Itoa(int(a.Params.Parallelism)))
	buf.WriteString("$")
	buf.WriteString(b64Salt)
	buf.WriteString("$")
	buf.WriteString(b64Key)
	// hash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Key)
	return buf.String()
}

// expects format example:
//
// $argon2id$v=19$m=19456,t=2,p=2$LPz5gLBKWtk$cZXDRxSOh1JAImjgmo2vOs+IClCyGQyS+sFM9CKYzXk
//
// returns params, salt, key, error
func DecodeHash(hash string) (Params, string, string, error) {
	if hash[0] != '$' {
		return Params{}, "", "", ErrInvalidHash
	}

	hash = hash[1:]
	idx := strings.IndexByte(hash, '$')
	if hash[:idx] != "argon2id" {
		return Params{}, "", "", ErrIncompatibleVariant
	}

	hash = hash[idx+1:]
	if hash[:2] != "v=" {
		return Params{}, "", "", ErrInvalidHash
	}

	hash = hash[2:]
	idx = strings.IndexByte(hash, '$')
	versionNum, err := strconv.ParseUint(hash[:idx], 10, 32)
	if err != nil {
		return Params{}, "", "", ErrInvalidHash
	}
	if versionNum != argon2.Version {
		return Params{}, "", "", ErrIncompatibleVersion
	}

	hash = hash[idx+1:]
	if hash[:2] != "m=" {
		return Params{}, "", "", ErrInvalidHash
	}

	hash = hash[2:]
	idx = strings.IndexByte(hash, ',')
	memoryNum, err := strconv.ParseUint(hash[:idx], 10, 32)
	if err != nil {
		return Params{}, "", "", ErrInvalidHash
	}

	hash = hash[idx+1:]
	if hash[:2] != "t=" {
		return Params{}, "", "", ErrInvalidHash
	}

	hash = hash[2:]
	idx = strings.IndexByte(hash, ',')
	iterationsNum, err := strconv.ParseUint(hash[:idx], 10, 32)
	if err != nil {
		return Params{}, "", "", ErrInvalidHash
	}

	hash = hash[idx+1:]
	if hash[:2] != "p=" {
		return Params{}, "", "", ErrInvalidHash
	}

	hash = hash[2:]
	idx = strings.IndexByte(hash, '$')
	parallelismNum, err := strconv.ParseUint(hash[:idx], 10, 8)
	if err != nil {
		return Params{}, "", "", ErrInvalidHash
	}

	hash = hash[idx+1:]
	idx = strings.IndexByte(hash, '$')
	salt := hash[:idx]

	saltLen := (len(salt) * 6) / 8 // assumes un padded base64

	key := hash[idx+1:]
	keyLen := (len(key) * 6) / 8

	params := Params{
		Memory:      uint32(memoryNum),
		Iterations:  uint32(iterationsNum),
		Parallelism: uint8(parallelismNum),
		SaltLength:  uint32(saltLen),
		KeyLength:   uint32(keyLen),
	}
	return params, salt, key, nil
}
