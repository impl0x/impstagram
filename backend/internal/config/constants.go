package config

import "time"

const (
	ServiceName = "Impstagram"

	MinAge uint8 = 18
	MaxAge uint8 = 120

	EmailID string = Env + "@impstagram.ripgod.xyz"

	OTPExpiryTime = 10 * time.Minute

)
