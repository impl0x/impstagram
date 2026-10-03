package useragent

import "github.com/mileusna/useragent"

// wrapper over "github.com/mileusna/useragent" to only have required fields.
type UserAgentData struct {
	OSName      string
	BrowserName string
	DeviceType  string
	IsBot       bool
}

func Parse(u string) UserAgentData {
	ua := useragent.Parse(u)
	var ud UserAgentData
	if ua.OS != "" {
		ud.OSName = ua.OS
	}
	if ua.Name != "" {
		ud.BrowserName = ua.Name
	}
	if ua.Device != "" {
		ud.DeviceType = ua.Name
	}
	ud.IsBot = ua.Bot
	return ud
}
