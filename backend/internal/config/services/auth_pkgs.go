package services

import (
	"backend/pkg/cryptoutil"
	"backend/pkg/cryptoutil/argon2id"
)

// config for the packages auth uses directly
type AuthPkgs struct {
	JWT      JWTConfig
	OTP      OTPConfig
	Password PasswordConfig
}

// jwt config
type JWTConfig struct {
	Secret string `env:"AUTH_JWT_SECRET,required"`
}

// otp config
type OTPConfig struct {
	OTPLen      uint `env:"AUTH_OTP_LEN" validate:"min=1"`
	TOTPLen     uint `env:"AUTH_TOTP_LEN" validate:"min=1"`
	TOTPKeySize uint `env:"AUTH_TOTP_KEY_SIZE" validate:"min=1"`
}

func (o *OTPConfig) Map() *cryptoutil.OTPConfig {
	return &cryptoutil.OTPConfig{
		OTPLen:      o.OTPLen,
		TOTPLen:     o.TOTPLen,
		TOTPKeySize: o.TOTPKeySize,
	}
}

// password hashing configuration
type Argon2IDConfig struct {
	Memory      uint `env:"AUTH_PASSWORD_MEMORY"`
	Iterations  uint `env:"AUTH_PASSWORD_ITERATIONS"`
	Parallelism uint `env:"AUTH_PASSWORD_PARALLELISM"`
	SaltLength  uint `env:"AUTH_PASSWORD_SALT_LENGTH"`
	KeyLength   uint `env:"AUTH_PASSWORD_KEY_LENGTH"`
}

func (o *Argon2IDConfig) Map() *argon2id.Config {
	return &argon2id.Config{
		Memory:      uint32(o.Memory),
		Iterations:  uint32(o.Iterations),
		Parallelism: uint8(o.Parallelism),
		SaltLength:  uint32(o.SaltLength),
		KeyLength:   uint32(o.SaltLength),
	}
}

type PasswordConfig struct {
	Argon2ID Argon2IDConfig
	// bcrypt, will add someday.
}

// the default configuration values for [AuthPkgs]
var DefaultAuthPkgs = AuthPkgs{
	OTP: OTPConfig{
		TOTPKeySize: 20,
		OTPLen:      6,
		TOTPLen:     6,
	},
	Password: PasswordConfig{
		Argon2ID: Argon2IDConfig{
			Memory:      19 * 1024,
			Iterations:  2,
			Parallelism: 2,
			SaltLength:  8,
			KeyLength:   32,
		},
	},
}
