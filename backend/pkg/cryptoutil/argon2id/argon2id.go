// this package is stolen from github.com/alexedwards/argon2id
// reason i didn't import it because it was just a one file and i also did some modifications
// all rights and credits to him
package argon2id

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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

func CreateRawHash(password string, salt []byte, params Params) []byte {
	return argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)
}

func CreateHash(password string, salt []byte, params Params) string {
	key := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	// we store it in this format: $argon2id$v=19$m=65536,t=1,p=4$ZWFzeXNhbHQ$R01...NDA
	// where, v is version, m for memory, t for time/iterations, p for parallelism/threads
	hash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Key)
	return hash
}

// DecodeHash expects a hash created from this package, and parses it to return the params used to
// create it, as well as the salt and key (password hash).
func DecodeHash(hash string) (params Params, salt, key []byte, err error) {

	r := strings.NewReader(hash)

	_, err = fmt.Fscanf(r, "$argon2id$")
	if err != nil {
		return Params{}, nil, nil, ErrIncompatibleVariant
	}

	var version int
	_, err = fmt.Fscanf(r, "v=%d$", &version)
	if err != nil {
		return Params{}, nil, nil, err
	}
	if version != argon2.Version {
		return Params{}, nil, nil, ErrIncompatibleVersion
	}

	params = Params{}
	_, err = fmt.Fscanf(r, "m=%d,t=%d,p=%d$", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return Params{}, nil, nil, err
	}

	rest, err := io.ReadAll(r)
	if err != nil {
		return Params{}, nil, nil, err
	}
	if bytes.ContainsAny(rest, "\r\n") { // base64 decoder ignores these
		return Params{}, nil, nil, ErrInvalidHash
	}

	var i int
	if i = bytes.IndexByte(rest, '$'); i == -1 {
		return Params{}, nil, nil, ErrInvalidHash
	}

	b64Enc := base64.RawStdEncoding.Strict()

	salt = make([]byte, b64Enc.DecodedLen(i))
	_, err = b64Enc.Decode(salt, rest[:i])
	if err != nil {
		return Params{}, nil, nil, err
	}
	params.SaltLength = uint32(len(salt))

	key = make([]byte, b64Enc.DecodedLen(len(rest)-i-1))
	_, err = b64Enc.Decode(key, rest[i+1:])
	if err != nil {
		return Params{}, nil, nil, err
	}
	params.KeyLength = uint32(len(key))

	return params, salt, key, nil
}
