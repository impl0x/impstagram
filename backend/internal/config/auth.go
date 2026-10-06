package config

import "time"

// Most of the fields in this config are not required and are optional,
// the auth service uses its own default version of this config for the maximum optional fields.
// However the fields having required tag in the env tag are to be set from environment otherwise the app will not start
// Any other fields which are provided will also be replaced in the default config instance to the value provided
//
//
// LLM generated overview below:

// This Go code defines an `AuthConfig` struct with various fields related to
// authentication settings, such as expiry times, cache clean intervals, OTP lengths,
// byte sizes, and session ID prefixes. The struct is designed to be populated with
// values from environment variables, as indicated by the `env` struct tags.

// The `NewAuthConfig` function returns a new instance of the `AuthConfig` struct.
// This struct can be used to configure authentication settings for an application,
// such as the duration for which an OTP (One-Time Password) is valid, the interval
// at which the cache is cleaned, and the length of the OTP.

// Here is a brief explanation of some of the fields in the `AuthConfig` struct:

// - `ExpiryTimeOTP`: The duration for which an OTP is valid.
// - `ExpiryTimeAccessToken`: The duration for which an access token is valid.
// - `ExpiryTimeRefreshToken`: The duration for which a refresh token is valid.
// - `ExpiryTimeResetPassword`: The duration for which a password reset token is valid.
// - `TTLCacheCleanIntervalOTP`: The interval at which the OTP cache is cleaned.
// - `TTLCacheCleanIntervalTOTP`: The interval at which the TOTP (Time-based One-Time Password) cache is cleaned.
// - `TTLCacheCleanIntervalReset`: The interval at which the password reset cache is cleaned.
// - `TTLCacheCleanIntervalJWTBlockList`: The interval at which the JWT (JSON Web Token) blocklist cache is cleaned.
// - `LenOTP`: The length of the OTP.
// - `LenTOTP`: The length of the TOTP.
// - `SizeSessionID`: The size of the session ID in bytes.
// - `SizeRefreshToken`: The size of the refresh token in bytes.
// - `SizeTOTPKey`: The size of the TOTP key in bytes.
// - `PrefixRefreshToken`: The prefix for the refresh token session ID.
// - `PrefixOTPSession`: The prefix for the OTP session ID.
// - `PrefixTOTPSession`: The prefix for the TOTP session ID.
// - `PrefixResetSession`: The prefix for the password reset session ID.

// The `NewAuthConfig` function can be used to create a new instance of the `AuthConfig` struct, which can then be populated with values from environment variables using a library such as `github.com/kelseyhightower/envconfig`.`

// Auth Service configuration values
type AuthConfig struct {
	// age limits
	AgeMin int `env:"AUTH_AGE_MIN"`
	AgeMax int `env:"AUTH_AGE_MAX"`

	// attempt limits
	AttemptsOTP        int `env:"AUTH_ATTEMPTS_OTP"`
	AttemptsTOTPVerify int `en:"AUTH_ATTEMPTS_TOTP_VERIFY"`

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

	// otp lengths
	LenOTP  int `env:"AUTH_LEN_OTP"`
	LenTOTP int `env:"AUTH_LEN_TOTP"`

	// byte sizes
	SizeSessionID    int `env:"AUTH_SIZE_SESSION_ID"`
	SizeRefreshToken int `env:"AUTH_SIZE_REFRESH_TOKEN"`
	SizeTOTPKey      int `env:"AUTH_SIZE_TOTP_KEY"`

	// session id prefixes
	PrefixRefreshToken string `env:"AUTH_PREFIX_REFRESH_TOKEN"`
	PrefixOTPSession   string `env:"AUTH_PREFIX_OTP_SESSION"`
	PrefixTOTPSession  string `env:"AUTH_PREFIX_TOTP_SESSION"`
	PrefixResetSession string `env:"AUTH_PREFIX_RESET_SESSION"`

	JwtSecret string `env:"AUTH_JWT_SECRET,required"`
}

// the default configuration values for AuthConfig 
var defaultAuthConfig = AuthConfig{
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

	LenOTP:  6,
	LenTOTP: 6,

	SizeSessionID:    24,
	SizeRefreshToken: 32,
	SizeTOTPKey:      20,

	PrefixRefreshToken: "ref_",
	PrefixOTPSession:   "otp_",
	PrefixTOTPSession:  "totp_",
	PrefixResetSession: "pwd_",
}
