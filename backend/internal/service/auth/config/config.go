package config

// validations are not done by service and are just present
// as rules that be enforced by the provider providing the config

import (
	"time"
)

// auth config
type Config struct {
	// age limits
	Age Age

	// max attempts for otps
	MaxAttempts MaxAttempts

	// expiry time for sessions/otps, etc.
	ExpiryTime ExpiryTime

	// time to live cache clean interval durations
	TTLCacheCleanInterval TTLCacheCleanInterval

	// token sizes in bytes
	Size Size

	// token prefixes
	Prefix Prefix
}

type Age struct {
	Min uint8 `validate:"min=1"`
	Max uint8 `validate:"max=120"`
}

type MaxAttempts struct {
	Verify Verify
	Resend Resend
}
type Verify struct {
	OTP       uint8
	TOTPSetup uint8
}
type Resend struct {
	OTP uint8
}

type ExpiryTime struct {
	OTP            time.Duration
	TOTPSetup      time.Duration
	AccessToken    time.Duration
	RefreshToken   time.Duration
	ForgotPassword time.Duration
}

type TTLCacheCleanInterval struct {
	OTP            time.Duration
	TOTPSetup      time.Duration
	ForgotPassword time.Duration
	JWTBlockList   time.Duration
}

type Size struct {
	SessionID    uint16
	RefreshToken uint16
}

type Prefix struct {
	RefreshToken          string
	OTPSession            string
	TOTPSession           string
	ForgotPasswordSession string
}

// Default values for the configuration, to be used when not provided explicit
// configuration, but not to be used directly as well
var Default = Config{
	Age: Age{
		Min: 13,
		Max: 120,
	},
	MaxAttempts: MaxAttempts{
		Verify: Verify{
			OTP:       5,
			TOTPSetup: 10,
		},
		Resend: Resend{
			OTP: 3,
		},
	},
	ExpiryTime: ExpiryTime{
		OTP:            10 * time.Minute,
		TOTPSetup:      30 * time.Minute,
		AccessToken:    30 * time.Minute,
		RefreshToken:   7 * 24 * time.Hour,
		ForgotPassword: 30 * time.Minute,
	},
	TTLCacheCleanInterval: TTLCacheCleanInterval{
		OTP:            10 * time.Minute,
		TOTPSetup:      10 * time.Minute,
		ForgotPassword: 10 * time.Minute,
		JWTBlockList:   15 * time.Minute, // half the time of access time expiry token
	},
	Size: Size{
		SessionID:    24,
		RefreshToken: 32,
	},
	Prefix: Prefix{
		RefreshToken:          "ref_",
		OTPSession:            "otp_",
		TOTPSession:           "totp_",
		ForgotPasswordSession: "pwd_",
	},
}
