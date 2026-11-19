package cryptoutil

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

type Hasher func(string) string

// wrapper to generate hash using a [hash.Hash] instance
func generateHash(h hash.Hash, d string) string {
	h.Write([]byte(d))
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}

// Generates a MD5 hash
func GenerateMD5Hash(data string) string {
	return generateHash(md5.New(), data)
}

// Generates a SHA256 hash
func GenerateSHA256Hash(data string) string {
	return generateHash(sha256.New(), data)
}
