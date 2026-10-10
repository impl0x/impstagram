package config

import "time"

// parent struct for auth config, not to be used directly in service layer, contains pkg config
type Auth struct {
	JWT  JWTConfig
	OTP  OTPConfig
	Auth AuthService
}

// config values for auth service
type AuthService struct {
	// age limits
	AgeMin int `env:"AUTH_AGE_MIN" validate:"min=1"`
	AgeMax int `env:"AUTH_AGE_MAX" validate:"min=120"`

	// attempt limits
	AttemptsOTP        int `env:"AUTH_ATTEMPTS_OTP" validate:"min=0"`
	AttemptsTOTPVerify int `en:"AUTH_ATTEMPTS_TOTP_VERIFY" validate:"min=0"`

	// expiry times
	ExpiryTimeOTP           time.Duration `env:"AUTH_EXPIRY_TIME_OTP"`
	ExpiryTimeAccessToken   time.Duration `env:"AUTH_EXPIRY_TIME_ACCESS_TOKEN"`
	ExpiryTimeRefreshToken  time.Duration `env:"AUTH_EXPIRY_TIME_REFRESH_TOKEN"`
	ExpiryTimeResetPassword time.Duration `env:"AUTH_EXPIRY_TIME_RESET_PASSWORD"`

	// cache.TTLCache clean intervals
	TTLCacheCleanIntervalOTP          time.Duration `env:"AUTH_TTLCACHE_CLEAN_INTERVAL_OTP"`
	TTLCacheCleanIntervalTOTP         time.Duration `env:"AUTH_TTLCACHE_CLEAN_INTERVAL_TOTP"`
	TTLCacheCleanIntervalReset        time.Duration `env:"AUTH_TTLCACHE_CLEAN_INTERVAL_RESET"`
	TTLCacheCleanIntervalJWTBlockList time.Duration `env:"AUTH_TTLCACHE_CLEAN_INTERVAL_JWT_BLOCKLIST"`

	// byte sizes
	SizeSessionID    int `env:"AUTH_SIZE_SESSION_ID" validate:"min=1"`
	SizeRefreshToken int `env:"AUTH_SIZE_REFRESH_TOKEN" validate:"min=1"`

	// session id prefixes
	PrefixRefreshToken string `env:"AUTH_PREFIX_REFRESH_TOKEN"`
	PrefixOTPSession   string `env:"AUTH_PREFIX_OTP_SESSION"`
	PrefixTOTPSession  string `env:"AUTH_PREFIX_TOTP_SESSION"`
	PrefixResetSession string `env:"AUTH_PREFIX_RESET_SESSION"`
}

// jwt config
type JWTConfig struct {
	JwtSecret string `env:"AUTH_JWT_SECRET,required"`
}

// otp config
type OTPConfig struct {
	LenOTP      int `env:"AUTH_LEN_OTP" validate:"min=1"`
	LenTOTP     int `env:"AUTH_LEN_TOTP" validate:"min=1"`
	SizeTOTPKey int `env:"AUTH_SIZE_TOTP_KEY" validate:"min=1"`
}

// the default configuration values for [authConfig]
var DefaultAuth = Auth{
	OTP: OTPConfig{
		SizeTOTPKey: 20,
		LenOTP:      6,
		LenTOTP:     6,
	},
	Auth: AuthService{
		AgeMin: 13,
		AgeMax: 120,

		AttemptsOTP:        5,
		AttemptsTOTPVerify: 5,

		ExpiryTimeOTP:           10 * time.Minute,
		ExpiryTimeAccessToken:   30 * time.Minute,
		ExpiryTimeRefreshToken:  7 * 24 * time.Hour,
		ExpiryTimeResetPassword: 30 * time.Minute,

		TTLCacheCleanIntervalOTP:          10 * time.Minute,
		TTLCacheCleanIntervalTOTP:         10 * time.Minute,
		TTLCacheCleanIntervalReset:        10 * time.Minute,
		TTLCacheCleanIntervalJWTBlockList: 15 * time.Minute,

		SizeSessionID:    24,
		SizeRefreshToken: 32,

		PrefixRefreshToken: "ref_",
		PrefixOTPSession:   "otp_",
		PrefixTOTPSession:  "totp_",
		PrefixResetSession: "pwd_",
	},
}
