package entity

import "backend/pkg/useragent"

type ClientMetadata struct {
	IPAddress string
	UserAgent useragent.Data
}
