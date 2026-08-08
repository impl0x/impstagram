package token

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var otpLimit = big.NewInt(1000000) // declaring once to avoid heap allocation on every call

func GenerateOTP() (string, error) {
	num, err := rand.Int(rand.Reader, otpLimit)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", num.Int64()), nil
}
