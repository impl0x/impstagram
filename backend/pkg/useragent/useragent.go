package useragent

import "github.com/mileusna/useragent"

// wrapper over [github.com/mileusna/useragent] to only have required fields.
type UserAgentData struct {
	OSName      string
	BrowserName string
	DeviceType  string
	IsBot       bool
}

func Parse(u string) UserAgentData {
	ua := useragent.Parse(u)
	return UserAgentData{
		OSName:      ua.OS,
		BrowserName: ua.Name,
		DeviceType:  ua.Device,
		IsBot:       ua.Bot,
	}
}
