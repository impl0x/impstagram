package config

import "time"

// common
const (
	ServiceName = "Impstagram"

	MinAge uint16 = 18
	MaxAge uint16 = 120

	EmailID string = Env + "@impstagram.ripgod.xyz"
)

// /ttlcache
const (
	TTLCacheCleanInterval             = 10 * time.Minute
	TTLCacheCleanIntervalOTP          = TTLCacheCleanInterval
	TTLCacheCleanIntervalReset        = TTLCacheCleanInterval
	TTLCacheCleanIntervalJWTBlockList = ExpiryTimeAccessToken / 2
)

// expiry times
const (
	ExpiryTimeOTP           = 10 * time.Minute
	ExpiryTimeAccessToken   = 30 * time.Minute
	ExpiryTimeRefreshToken  = 7 // Days, use this with time.AddDate method only
	ExpiryTimeResetPassword = 30 * time.Minute
)

// byte sizes
const (
	SizeSessionID    = 24 // used for otp sessions, reset password sessions, basically every ttl cache session key
	SizeRefreshToken = 32 //
)

// opaque string / id prefixes
const (
	PrefixRefreshToken = ""
	PrefixOTPSession   = "otp_"
	PrefixResetSession = "pwd_"
)
