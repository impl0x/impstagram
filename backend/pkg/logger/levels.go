package logger

import "bytes"

type Level uint

const (
	LevelTrace = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

type level string

const (
	levelTrace = "TRC"
	levelDebug = "DBG"
	levelInfo  = "INF"
	levelWarn  = "WRN"
	levelError = "ERR"
	levelFatal = "FTL"
)

func writeLevel(b *bytes.Buffer, l level) {
	b.WriteString(string(l))
}
