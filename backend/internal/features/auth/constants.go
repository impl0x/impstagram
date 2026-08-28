package auth

import (
	"time"
)

// contains business rule constant values
// the prefix "rule" to all variables here does not sound or make sense sometimes.
// But i wanted consistency and clear meaning about what these are

// common
const (
	ruleMinAge = 18
	ruleMaxAge = 120
)

// attempt limits
const (
	ruleAttemptsOTP        = 5
	ruleAttemptsTOTPVerify = 5
)

// expiry times
const (
	ruleExpiryTimeOTP           = 10 * time.Minute
	ruleExpiryTimeAccessToken   = 30 * time.Minute
	ruleExpiryTimeRefreshToken  = 7 // Days, use this with time.AddDate method only
	ruleExpiryTimeResetPassword = 30 * time.Minute
)

// ttlcache clean interval
const (
	ruleTTLCacheCleanIntervalOTP          = 10 * time.Minute
	ruleTTLCacheCleanIntervalReset        = 10 * time.Minute
	ruleTTLCacheCleanIntervalTOTP         = 10 * time.Minute
	ruleTTLCacheCleanIntervalJWTBlockList = ruleExpiryTimeAccessToken / 2
)

// otp lengths
const (
	ruleOTPLen  = 6
	ruleTOTPLen = 6
)

// byte sizes
const (
	ruleSizeSessionID    = 24 // used for otp sessions, reset password sessions, basically every ttl cache session key
	ruleSizeRefreshToken = 32
	ruleSizeTOTPKey      = 20 // matches the sha1 output
)

// opaque string / id prefixes
const (
	rulePrefixRefreshToken = ""
	rulePrefixOTPSession   = "otp_"
	rulePrefixResetSession = "pwd_"
	rulePrefixTOTPSession  = "totp_"
)
