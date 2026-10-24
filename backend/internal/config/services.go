package config

import "time"

// parent struct for auth config, not to be used directly in service layer, contains pkg config as-well directly related to auth
type Auth struct {
	JWT      JWTConfig
	OTP      OTPConfig
	Password PasswordConfig
	Service  AuthService
}

// jwt config
type JWTConfig struct {
	Secret string `env:"AUTH_JWT_SECRET,required"`
}

// otp config
type OTPConfig struct {
	OTPLen      int `env:"AUTH_OTP_LEN" validate:"min=1"`
	TOTPLen     int `env:"AUTH_TOTP_LEN" validate:"min=1"`
	TOTPKeySize int `env:"AUTH_TOTP_KEY_SIZE" validate:"min=1"`
}

// password hashing configuration
type PasswordConfig struct {
	Memory      int `env:"AUTH_PASSWORD_MEMORY"`
	Iterations  int `env:"AUTH_PASSWORD_ITERATIONS"`
	Parallelism int `env:"AUTH_PASSWORD_PARALLELISM"`
	SaltLength  int `env:"AUTH_PASSWORD_SALT_LENGTH"`
	KeyLength   int `env:"AUTH_PASSWORD_KEY_LENGTH"`
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

// the default configuration values for [AuthService]
var DefaultAuth = Auth{
	OTP: OTPConfig{
		TOTPKeySize: 20,
		OTPLen:      6,
		TOTPLen:     6,
	},
	Password: PasswordConfig{
		Memory:      19 * 1024,
		Iterations:  2,
		Parallelism: 2,
		SaltLength:  8,
		KeyLength:   32,
	},
	Service: AuthService{
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
