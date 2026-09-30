package cryptoutil

import (
	"backend/internal/config"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type OtpGenerator struct {
	otpLimit *big.Int
	totp     struct {
		length        int
		secretKeySize int
	}
}

func NewOtpGenerator(otpLength, totpLen, totpKeyLen int) OtpGenerator {
	return OtpGenerator{
		big.NewInt(int64(math.Pow(10, float64(otpLength)))),
		struct {
			length        int
			secretKeySize int
		}{totpLen, totpKeyLen},
	}
}

func (og OtpGenerator) Generate() (string, error) {
	num, err := rand.Int(rand.Reader, og.otpLimit)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", num.Int64()), nil
}

var ErrInvalidSecretKey = errors.New("invalid secret key")

// Generates a 6 digit time based otp with the given secret key
//
// possible errors are only ErrInvalidSecretKey which occurs if the key cannot be decoded
func (og OtpGenerator) GenerateTOTP(secret string) (string, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", ErrInvalidSecretKey
	}

	epoch := time.Now().Unix()
	timeStep := epoch / 30

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(timeStep))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0xf
	binaryCode := (int32(hash[offset])&0x7f)<<24 |
		(int32(hash[offset+1])&0xff)<<16 |
		(int32(hash[offset+2])&0xff)<<8 |
		(int32(hash[offset+3]) & 0xff)

	mod := int32(math.Pow10(og.totp.length))
	code := binaryCode % mod

	return fmt.Sprintf("%0*d", og.totp.length, code), nil
}

func (og OtpGenerator) SetupTOTP(userIdentifier string) (secretKey string, uri string) {
	keyBytes := make([]byte, og.totp.secretKeySize)
	rand.Read(keyBytes)
	secretKey = strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(keyBytes))
	uri = "otpauth://totp/" + config.ServiceName + ":" + userIdentifier + "?secret=" + secretKey + "&issuer=" + config.ServiceName + "&digits=" + strconv.Itoa(og.totp.length)
	return
}
