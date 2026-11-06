package services

import (
	auth "backend/internal/service/auth/config"
	"time"
)

// auth config, directly business related
type Auth struct {
	AgeMin                             uint          `env:"AUTH_SERVICE_AGE_MIN"`
	AgeMax                             uint          `env:"AUTH_SERVICE_AGE_MAX"`
	MaxAttemptsOTP                     uint          `env:"AUTH_SERVICE_MAX_ATTEMPTS_OTP"`
	MaxAttemptsTOTPVerify              uint          `env:"AUTH_SERVICE_MAX_ATTEMPTS_TOTP_VERIFY"`
	ExpiryTimeOTP                      time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_OTP"`
	ExpiryTimeTOTPSetup                time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_TOTP_SETUP"`
	ExpiryTimeAccessToken              time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_ACCESS_TOKEN"`
	ExpiryTimeRefreshToken             time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_REFRESH_TOKEN"`
	ExpiryTimeResetPassword            time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_RESET_PASSWORD"`
	TTLCacheCleanIntervalOTP           time.Duration `env:"AUTH_SERVICE_TTLCACHE_CLEAN_INTERVAL_OTP"`
	TTLCacheCleanIntervalTOTP          time.Duration `env:"AUTH_SERVICE_TTLCACHE_CLEAN_INTERVAL_TOTP"`
	TTLCacheCleanIntervalResetPassword time.Duration `env:"AUTH_SERVICE_TTLCACHE_CLEAN_INTERVAL_RESET_PASSWORD"`
	TTLCacheCleanIntervalJWTBlockList  time.Duration `env:"AUTH_SERVICE_TTLCACHE_CLEAN_INTERVAL_JWT_BLOCKLIST"`
	SizeSessionID                      uint          `env:"AUTH_SERVICE_SIZE_SESSION_ID"`
	SizeRefreshToken                   uint          `env:"AUTH_SERVICE_SIZE_REFRESH_TOKEN"`
	PrefixRefreshToken                 string        `env:"AUTH_SERVICE_PREFIX_REFRESH_TOKEN"`
	PrefixOTPSession                   string        `env:"AUTH_SERVICE_PREFIX_OTP_SESSION"`
	PrefixTOTPSession                  string        `env:"AUTH_SERVICE_PREFIX_TOTP_SESSION"`
	PrefixResetSession                 string        `env:"AUTH_SERVICE_PREFIX_RESET_SESSION"`
}

// maps to [config.Config] where config is [backend/internal/service/auth/config]
//
// this is accepted directly by the auth service constructor
func (a *Auth) Map() *auth.Config {
	return &auth.Config{
		Age: auth.Age{
			Min: a.AgeMin,
			Max: a.AgeMax,
		},
		MaxAttempts: auth.MaxAttempts{
			OTP:        a.MaxAttemptsOTP,
			TOTPVerify: a.MaxAttemptsTOTPVerify,
		},
		ExpiryTime: auth.ExpiryTime{
			OTP:           a.ExpiryTimeOTP,
			TOTPSetup:     a.ExpiryTimeTOTPSetup,
			AccessToken:   a.ExpiryTimeAccessToken,
			RefreshToken:  a.ExpiryTimeRefreshToken,
			ResetPassword: a.ExpiryTimeResetPassword,
		},
		TTLCacheCleanInterval: auth.TTLCacheCleanInterval{
			OTP:           a.TTLCacheCleanIntervalOTP,
			TOTP:          a.TTLCacheCleanIntervalTOTP,
			ResetPassword: a.TTLCacheCleanIntervalResetPassword,
			JWTBlockList:  a.TTLCacheCleanIntervalJWTBlockList,
		},
		Size: auth.Size{
			SessionID:    a.SizeSessionID,
			RefreshToken: a.SizeRefreshToken,
		},
		Prefix: auth.Prefix{
			RefreshToken: a.PrefixRefreshToken,
			OTPSession:   a.PrefixOTPSession,
			TOTPSession:  a.PrefixTOTPSession,
			ResetSession: a.PrefixResetSession,
		},
	}
}

func fromAuthConfig(a auth.Config) Auth {
	return Auth{
		AgeMin:                             a.Age.Min,
		AgeMax:                             a.Age.Max,
		MaxAttemptsOTP:                     a.MaxAttempts.OTP,
		MaxAttemptsTOTPVerify:              a.MaxAttempts.TOTPVerify,
		ExpiryTimeOTP:                      a.ExpiryTime.OTP,
		ExpiryTimeTOTPSetup:                a.ExpiryTime.TOTPSetup,
		ExpiryTimeAccessToken:              a.ExpiryTime.AccessToken,
		ExpiryTimeRefreshToken:             a.ExpiryTime.RefreshToken,
		ExpiryTimeResetPassword:            a.ExpiryTime.ResetPassword,
		TTLCacheCleanIntervalOTP:           a.TTLCacheCleanInterval.OTP,
		TTLCacheCleanIntervalTOTP:          a.TTLCacheCleanInterval.TOTP,
		TTLCacheCleanIntervalResetPassword: a.TTLCacheCleanInterval.ResetPassword,
		TTLCacheCleanIntervalJWTBlockList:  a.TTLCacheCleanInterval.JWTBlockList,
		SizeSessionID:                      a.Size.SessionID,
		SizeRefreshToken:                   a.Size.RefreshToken,
		PrefixRefreshToken:                 a.Prefix.RefreshToken,
		PrefixOTPSession:                   a.Prefix.OTPSession,
		PrefixTOTPSession:                  a.Prefix.TOTPSession,
		PrefixResetSession:                 a.Prefix.ResetSession,
	}
}

var DefaultAuth = fromAuthConfig(auth.Default)
