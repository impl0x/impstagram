package logger

import (
	"backend/pkg/logger/color"
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
	TimeSep                    string // separator between time values, 12:29:23, here ":" is the separator
	DateSep                    string // separator between date values, 2026-02-02, here "-" is the separator
	DateTimeSep                string // separator between date and time, 2026-02-02 12:29:23, here " " is the separator
	ShowDate                   bool
	ShowFullYear               bool // 2026 / 26
	ShowTime                   bool
	TimeFormat                 TimeFormat
	AddSpaceFor12HrFormat      bool // 5 PM / 5PM
	AddZerosForSingleDigitDate bool // 26-04-01/26-4-1
	AddZerosForSingleDigitTime bool // 04:23 / 4:23
}

type TimeFormat uint8

const (
	Format24Hr TimeFormat = iota
	Format12Hr
)

type ColorConfig struct {
	Time color.RGB
}

var DefaultConfig = Config{
	Time: TimeConfig{
		TimeSep:                    ":",
		DateSep:                    "-",
		DateTimeSep:                " ",
		ShowDate:                   false,
		ShowFullYear:               true,
		ShowTime:                   true,
		TimeFormat:                 Format12Hr,
		AddSpaceFor12HrFormat:      true,
		AddZerosForSingleDigitDate: true,
		AddZerosForSingleDigitTime: true,
	},
}
