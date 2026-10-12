package handler

import (
	"backend/internal/entity"
	"backend/pkg/useragent"
	"net/http"
)

// Function to return the IP address of the client
// Change this if behind a reverse proxy
func GetIpFromRequest(r *http.Request) string {
	return r.Host
}

func GetClientMetadata(r *http.Request) entity.ClientMetadata {
	ua := useragent.Parse(r.UserAgent())
	return entity.ClientMetadata{
		IPAddress: GetIpFromRequest(r),
		OSName: ua.OSName,
		BrowserName: ua.BrowserName,
		DeviceType: ua.DeviceType,
		IsBot: ua.IsBot,
	}
}
