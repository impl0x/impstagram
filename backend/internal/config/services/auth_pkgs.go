package services

import (
	"backend/pkg/cryptoutil"
	"backend/pkg/cryptoutil/argon2id"
)

// config for the packages auth uses directly
type AuthPkgs struct {
	JWT      JWTConfig
	OTP      OTPConfig
	TOTP     TOTPConfig
	Password PasswordConfig
}

// jwt config
type JWTConfig struct {
	Secret string `env:"AUTH_JWT_SECRET,required"`
}

// otp config
type OTPConfig struct {
	OTPLen uint8 `env:"AUTH_OTP_LEN" validate:"min=1"`
}

func (o *OTPConfig) Map() *cryptoutil.OTPConfig {
	return &cryptoutil.OTPConfig{
		OTPLen: o.OTPLen,
	}
}

// totp config
type TOTPConfig struct {
	ServiceName string // not populated, to be populated manually as it is a repeated field
	TOTPLen     uint8  `env:"AUTH_TOTP_LEN" validate:"min=1"`
	TOTPKeySize uint8  `env:"AUTH_TOTP_KEY_SIZE" validate:"min=1"`
}

func (t *TOTPConfig) Map() *cryptoutil.TOTPConfig {
	return &cryptoutil.TOTPConfig{
		ServiceName: t.ServiceName,
		TOTPLen:     t.TOTPLen,
		TOTPKeySize: t.TOTPKeySize,
	}
}

// password hashing configuration
type Argon2IDConfig struct {
	Memory      uint32 `env:"AUTH_PASSWORD_MEMORY"`
	Iterations  uint32 `env:"AUTH_PASSWORD_ITERATIONS"`
	Parallelism uint8  `env:"AUTH_PASSWORD_PARALLELISM"`
	SaltLength  uint32 `env:"AUTH_PASSWORD_SALT_LENGTH"`
	KeyLength   uint32 `env:"AUTH_PASSWORD_KEY_LENGTH"`
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
		OTPLen: 6,
	},
	TOTP: TOTPConfig{
		TOTPKeySize: 20,
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
