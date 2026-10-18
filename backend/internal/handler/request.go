package handler

import (
	"backend/internal/entity"
	"backend/internal/service"
	"math/rand/v2"

	"net/http"

	"github.com/mileusna/useragent"
)

// Function to return the IP address of the client
// Change this if behind a reverse proxy
func GetIpFromRequest(r *http.Request) string {
	return r.Host
}

func GetClientMetadata(r *http.Request) entity.ClientMetadata {
	ua := useragent.Parse(r.UserAgent())
	return entity.ClientMetadata{
		IPAddress:   GetIpFromRequest(r),
		OSName:      ua.OS,
		BrowserName: ua.Name,
		DeviceType:  ua.Device,
		IsBot:       ua.Bot,
	}
}

type ContextGenerator func(r *http.Request) service.Context

func DefaultContextGenerator() ContextGenerator {
	return func(r *http.Request) service.Context {
		const hex = "0123456789abcdef"
		var id [16]byte
		for i := range id {
			id[i] = hex[rand.IntN(16)]
		}
		return service.Context{
			ReqId:   string(id[:]),
			Context: r.Context(),
		}
	}
}
