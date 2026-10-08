package useragent

import "github.com/mileusna/useragent"

// wrapper over [github.com/mileusna/useragent] to only have required fields.
type Data struct {
	OSName      string
	BrowserName string
	DeviceType  string
	IsBot       bool
}

func Parse(u string) Data {
	ua := useragent.Parse(u)
	return Data{
		OSName:      ua.OS,
		BrowserName: ua.Name,
		DeviceType:  ua.Device,
		IsBot:       ua.Bot,
	}
}
