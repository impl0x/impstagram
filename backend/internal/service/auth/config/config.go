package config

import "time"

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
	Min uint `validate:"min=1"`
	Max uint `validate:"max=120"`
}

type MaxAttempts struct {
	OTP        uint
	TOTPVerify uint
}

type ExpiryTime struct {
	OTP           time.Duration
	TOTPSetup     time.Duration
	AccessToken   time.Duration
	RefreshToken  time.Duration
	ResetPassword time.Duration
}

type TTLCacheCleanInterval struct {
	OTP           time.Duration
	TOTP          time.Duration
	ResetPassword time.Duration
	JWTBlockList  time.Duration
}

type Size struct {
	SessionID    uint
	RefreshToken uint
}

type Prefix struct {
	RefreshToken string
	OTPSession   string
	TOTPSession  string
	ResetSession string
}

// Default values for the configuration, to be used when not provided explicit
// configuration, but not to be used directly as well
var Default = Config{
	Age: Age{
		Min: 13,
		Max: 120,
	},
	MaxAttempts: MaxAttempts{
		OTP:        5,
		TOTPVerify: 5,
	},
	ExpiryTime: ExpiryTime{
		OTP:           10 * time.Minute,
		TOTPSetup:     30 * time.Minute,
		AccessToken:   30 * time.Minute,
		RefreshToken:  7 * 24 * time.Hour,
		ResetPassword: 30 * time.Minute,
	},
	TTLCacheCleanInterval: TTLCacheCleanInterval{
		OTP:           10 * time.Minute,
		TOTP:          10 * time.Minute,
		ResetPassword: 10 * time.Minute,
		JWTBlockList:  15 * time.Minute, // half the time of access time expiry token
	},
	Size: Size{
		SessionID:    24,
		RefreshToken: 32,
	},
	Prefix: Prefix{
		RefreshToken: "ref_",
		OTPSession:   "otp_",
		TOTPSession:  "totp_",
		ResetSession: "pwd_",
	},
}
