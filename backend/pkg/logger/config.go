package logger

import (
	"backend/pkg/logger/color"
	"io"
	"os"
)

type Config struct {
	Output io.Writer
	Time   TimeConfig
	Color  ColorConfig
	Others OtherConfig
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
	Time     color.RGB
	Levels   levelColors
	MsgColor color.RGB
	ExKey    color.RGB
	ExVal    color.RGB
}
type levelColors struct {
	Trace color.RGB
	Debug color.RGB
	Info  color.RGB
	Warn  color.RGB
	Error color.RGB
	Fatal color.RGB
}

func (lc levelColors) fromLevel(l level) color.RGB {
	switch l {
	case levelTrace:
		return lc.Trace
	case levelDebug:
		return lc.Debug
	case levelInfo:
		return lc.Info
	case levelWarn:
		return lc.Warn
	case levelError:
		return lc.Error
	case levelFatal:
		return lc.Fatal
	default:
		panic("invalid level passed")
	}
}

type OtherConfig struct {
	PanicAtFatal bool
}

var DefaultConfig = Config{
	Output: os.Stdout,
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
	Color: ColorConfig{
		Time: color.White[5],
		Levels: levelColors{
			Trace: color.NewRGB(128, 128, 128), // #808080
			Debug: color.NewRGB(0, 192, 255),   // #00c0ff
			Info:  color.NewRGB(0, 200, 0),     // #00c800
			Warn:  color.NewRGB(255, 165, 0),   // #ffa500
			Error: color.NewRGB(200, 0, 0),     // #c80000
			Fatal: color.NewRGB(255, 0, 255),   // #ff00ff
		},
		MsgColor: color.White[1],
		ExKey:    color.NewRGB(10, 117, 119), // #0a7577
		ExVal:    color.White[3],
	},
	Others: OtherConfig{
		PanicAtFatal: true,
	},
}
