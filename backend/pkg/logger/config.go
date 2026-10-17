package logger

import (
	"backend/pkg/logger/color"
	"backend/pkg/logger/utils"
	"bytes"
	"time"
)

type Config struct {
	Time  TimeConfig
	Color ColorConfig
}

// config for the time part of the log,
// the time part is allocated 32 bytes only, so make sure the separators do not exceed too many bytes.
//
//   - date takes 8 or 6 bytes depending on if full year is displayed + date sep bytes
//   - datetime sep takes 1 according to default or whatever you choose
//   - time takes 6 + time sep bytes
//   - time format 12 hr takes extra 3 bytes due to space and AM PM, if space enabled else 2.
//     therefore it takes 17 bytes except all the separators, so make sure they take less than 15 bytes in *total*
type TimeConfig struct {
	TimeSep               string // separator between time values, 12:29:23, here ":" is the separator
	DateSep               string // separator between date values, 2026-02-02, here "-" is the separator
	DateTimeSep           string // separator between date and time, 2026-02-02 12:29:23, here " " is the separator
	ShowDate              bool
	ShowFullYear          bool // 2026 / 26
	ShowTime              bool
	TimeFormat            TimeFormat
	AddSpaceFor12HrFormat bool // 5 PM / 5PM
}

type TimeFormat uint8

const (
	Format24Hr TimeFormat = iota
	Format12Hr
)

// writes the parsed time according to config values to the buffer provided,
// it should approximately have an available cap of about 32 bytes at max
func (tc TimeConfig) parse(t time.Time, b *bytes.Buffer) {
	if tc.ShowDate {
		y, m, d := t.Date()
		if !tc.ShowFullYear {
			y = y % 100
		}
		utils.WriteInteger(b, y)
		b.WriteString(tc.DateSep)
		utils.WriteInteger(b, m)
		b.WriteString(tc.DateSep)
		utils.WriteInteger(b, d)
		b.WriteString(tc.DateTimeSep)
	}
	if tc.ShowTime {
		h, m, s := t.Clock()
		is12hr := tc.TimeFormat == Format12Hr
		var isPm bool
		if is12hr && h > 12 {
			h = h - 12
			isPm = true
		}
		utils.WriteInteger(b, h)
		b.WriteString(tc.TimeSep)
		utils.WriteInteger(b, m)
		b.WriteString(tc.TimeSep)
		utils.WriteInteger(b, s)
		if is12hr {
			if tc.AddSpaceFor12HrFormat {
				b.WriteRune(' ')
			}
			if isPm {
				b.WriteString("PM")
			} else {
				b.WriteString("AM")
			}
		}
	}
}

type ColorConfig struct {
	Time color.RGB
}

var DefaultConfig = Config{
	Time: TimeConfig{
		TimeSep:               ":",
		DateSep:               "-",
		DateTimeSep:           " ",
		ShowDate:              false,
		ShowFullYear:          true,
		ShowTime:              true,
		TimeFormat:            Format12Hr,
		AddSpaceFor12HrFormat: true,
	},
}
