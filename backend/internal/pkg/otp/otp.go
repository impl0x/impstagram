package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateOTP() (string, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	formatter := fmt.Sprintf("%%0%dd", 6)
	return fmt.Sprintf(formatter, num.Int64()), nil
}