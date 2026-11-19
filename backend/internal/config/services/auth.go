package services

import (
	auth "backend/internal/service/auth/config"
	"time"
)

// Auth service related config, directly business related
type Auth struct {
	AgeMin                              uint8         `env:"AUTH_SERVICE_AGE_MIN"`
	AgeMax                              uint8         `env:"AUTH_SERVICE_AGE_MAX"`
	MaxAttemptsVerifyOTP                uint8         `env:"AUTH_SERVICE_MAX_ATTEMPTS_VERIFY_OTP"`
	MaxAttemptsVerifyTOTPSetup          uint8         `env:"AUTH_SERVICE_MAX_ATTEMPTS_VERIFY_TOTP_SETUP"`
	MaxAttemptsResendOTP                uint8         `env:"AUTH_SERVICE_MAX_ATTEMPTS_RESEND_OTP"`
	ExpiryTimeOTP                       time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_OTP"`
	ExpiryTimeTOTPSetup                 time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_TOTP_SETUP"`
	ExpiryTimeAccessToken               time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_ACCESS_TOKEN"`
	ExpiryTimeRefreshToken              time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_REFRESH_TOKEN"`
	ExpiryTimeForgotPassword            time.Duration `env:"AUTH_SERVICE_EXPIRY_TIME_FORGOT_PASSWORD"`
	TTLCacheCleanIntervalOTP            time.Duration `env:"AUTH_SERVICE_TTLCACHE_CLEAN_INTERVAL_OTP"`
	TTLCacheCleanIntervalTOTPSetup      time.Duration `env:"AUTH_SERVICE_TTLCACHE_CLEAN_INTERVAL_TOTP_SETUP"`
	TTLCacheCleanIntervalForgotPassword time.Duration `env:"AUTH_SERVICE_TTLCACHE_CLEAN_INTERVAL_FORGOT_PASSWORD"`
	TTLCacheCleanIntervalJWTBlockList   time.Duration `env:"AUTH_SERVICE_TTLCACHE_CLEAN_INTERVAL_JWT_BLOCKLIST"`
	SizeSessionID                       uint16        `env:"AUTH_SERVICE_SIZE_SESSION_ID"`
	SizeRefreshToken                    uint16        `env:"AUTH_SERVICE_SIZE_REFRESH_TOKEN"`
	PrefixRefreshToken                  string        `env:"AUTH_SERVICE_PREFIX_REFRESH_TOKEN"`
	PrefixOTPSession                    string        `env:"AUTH_SERVICE_PREFIX_OTP_SESSION"`
	PrefixTOTPSession                   string        `env:"AUTH_SERVICE_PREFIX_TOTP_SESSION"`
	PrefixForgotPasswordSession         string        `env:"AUTH_SERVICE_PREFIX_FORGOT_PASSWORD_SESSION"`
}

// Map maps to [config.Config] where config is [backend/internal/service/auth/config]
//
// this is accepted directly by the auth service constructor
func (a *Auth) Map() *auth.Config {
	return &auth.Config{
		Age: auth.Age{
			Min: a.AgeMin,
			Max: a.AgeMax,
		},
		MaxAttempts: auth.MaxAttempts{
			Verify: auth.Verify{
				OTP:       a.MaxAttemptsVerifyOTP,
				TOTPSetup: a.MaxAttemptsVerifyTOTPSetup,
			},
			Resend: auth.Resend{
				OTP: a.MaxAttemptsResendOTP,
			},
		},
		ExpiryTime: auth.ExpiryTime{
			OTP:            a.ExpiryTimeOTP,
			TOTPSetup:      a.ExpiryTimeTOTPSetup,
			AccessToken:    a.ExpiryTimeAccessToken,
			RefreshToken:   a.ExpiryTimeRefreshToken,
			ForgotPassword: a.ExpiryTimeForgotPassword,
		},
		TTLCacheCleanInterval: auth.TTLCacheCleanInterval{
			OTP:            a.TTLCacheCleanIntervalOTP,
			TOTPSetup:      a.TTLCacheCleanIntervalTOTPSetup,
			ForgotPassword: a.TTLCacheCleanIntervalForgotPassword,
			JWTBlockList:   a.TTLCacheCleanIntervalJWTBlockList,
		},
		Size: auth.Size{
			SessionID:    a.SizeSessionID,
			RefreshToken: a.SizeRefreshToken,
		},
		Prefix: auth.Prefix{
			RefreshToken:          a.PrefixRefreshToken,
			OTPSession:            a.PrefixOTPSession,
			TOTPSession:           a.PrefixTOTPSession,
			ForgotPasswordSession: a.PrefixForgotPasswordSession,
		},
	}
}

func fromAuthConfig(a auth.Config) Auth {
	return Auth{
		AgeMin:                              a.Age.Min,
		AgeMax:                              a.Age.Max,
		MaxAttemptsVerifyOTP:                a.MaxAttempts.Verify.OTP,
		MaxAttemptsVerifyTOTPSetup:          a.MaxAttempts.Verify.TOTPSetup,
		MaxAttemptsResendOTP:                a.MaxAttempts.Resend.OTP,
		ExpiryTimeOTP:                       a.ExpiryTime.OTP,
		ExpiryTimeTOTPSetup:                 a.ExpiryTime.TOTPSetup,
		ExpiryTimeAccessToken:               a.ExpiryTime.AccessToken,
		ExpiryTimeRefreshToken:              a.ExpiryTime.RefreshToken,
		ExpiryTimeForgotPassword:            a.ExpiryTime.ForgotPassword,
		TTLCacheCleanIntervalOTP:            a.TTLCacheCleanInterval.OTP,
		TTLCacheCleanIntervalTOTPSetup:      a.TTLCacheCleanInterval.TOTPSetup,
		TTLCacheCleanIntervalForgotPassword: a.TTLCacheCleanInterval.ForgotPassword,
		TTLCacheCleanIntervalJWTBlockList:   a.TTLCacheCleanInterval.JWTBlockList,
		SizeSessionID:                       a.Size.SessionID,
		SizeRefreshToken:                    a.Size.RefreshToken,
		PrefixRefreshToken:                  a.Prefix.RefreshToken,
		PrefixOTPSession:                    a.Prefix.OTPSession,
		PrefixTOTPSession:                   a.Prefix.TOTPSession,
		PrefixForgotPasswordSession:         a.Prefix.ForgotPasswordSession,
	}
}

var DefaultAuth = fromAuthConfig(auth.Default)
