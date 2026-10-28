package httpserver

import (
	"backend/internal/entity"

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
