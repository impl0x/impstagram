package token

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
)

type OtpGenerator struct{
	otpLimit *big.Int
}

func NewOtpGenerator(otpLength int) OtpGenerator {
	return OtpGenerator{
		big.NewInt(int64(math.Pow(10,float64(otpLength)))),
	}
}

func (og OtpGenerator) Generate() (string, error) {
	num, err := rand.Int(rand.Reader, og.otpLimit)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", num.Int64()), nil
}

func GenerateTOTP(secretKey string) (string, error) {
	// todo
	return "",nil
}
