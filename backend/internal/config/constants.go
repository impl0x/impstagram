package config

import "time"

const (
	ServiceName = "Impstagram"

	MinAge uint8 = 18
	MaxAge uint8 = 120

	EmailID string = Env + "@impstagram.ripgod.xyz"

	OTPExpiryTime = 10 * time.Minute

	AccessTokenExpiryTime  = 30 * time.Minute
	RefreshTokenExpiryTime = 7 // Days, use this with time.AddDate method only
)
