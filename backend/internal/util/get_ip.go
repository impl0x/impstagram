package util

import "net/http"

// Function to return the IP address of the client
// Change this if behind a reverse proxy
func GetIpFromRequest(r *http.Request) string {
	return r.Host
}
