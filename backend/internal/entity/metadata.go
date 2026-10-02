package entity

import "github.com/mileusna/useragent"

type ClientMetadata struct {
	IPAddress string
	UserAgent useragent.UserAgent
}
