package cryptoutil

import (
	"crypto/md5"
	"encoding/base64"
)

type Hasher func(string) string

func GenerateMD5Hash(data string) string {
	m := md5.New()
	m.Write([]byte(data))
	hashBytes := m.Sum(nil)
	hash := base64.RawURLEncoding.EncodeToString(hashBytes)
	return hash
}
