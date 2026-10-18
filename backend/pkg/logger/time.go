package logger

import (
	"backend/pkg/logger/utils"
	"bytes"
	"sync"
	"sync/atomic"
	"time"
)

type timeWriter struct {
	config TimeConfig
	tb     timeBuf
}

func newTimeWriter(tc TimeConfig) timeWriter {
	tw := timeWriter{
		config: tc,
		tb: timeBuf{
			lastTime:    atomic.Pointer[time.Time]{}, // doesn't matter if we copy lock, ignore warning
			lastTimeStr: atomic.Pointer[string]{},
		},
	}
	tw.tb.lastTime.Store(&time.Time{})
	s := ""
	tw.tb.lastTimeStr.Store(&s)
	return tw // does not matter if we copy lock, this function is inlined anyways.
}

type timeBuf struct {
	lastTime    atomic.Pointer[time.Time]
	lastTimeStr atomic.Pointer[string]
	mu          sync.Mutex
}

// checks if y,m,d,h,mm,s are equal between t1 and t2
func tEq(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	h1, mm1, s1 := t1.Clock()
	y2, m2, d2 := t2.Date()
	h2, mm2, s2 := t2.Clock()
	return y1 == y2 && m1 == m2 && d1 == d2 && h1 == h2 && mm1 == mm2 && s1 == s2
}

func (tw *timeWriter) write(b *bytes.Buffer) {
	now := time.Now()
	if tEq(now, *tw.tb.lastTime.Load()) {
		str := *tw.tb.lastTimeStr.Load()
		b.WriteString(str)
		return
	}
	tw._storeTime(b, now)
}

//go:noinline this function is used so that "now" doesn't escape in the write call
func (tw *timeWriter) _storeTime(b *bytes.Buffer, now time.Time) {
	// parsing and writing the time
	tw._parse(now, b, tw.config)
	str := b.String() // allocates

	// setting the lastTime to current
	tw.tb.mu.Lock()
	tw.tb.lastTime.Store(&now)
	tw.tb.lastTimeStr.Store(&str)
	tw.tb.mu.Unlock()
}

// writes the parsed time according to config values to the buffer provided,
// it should approximately have an available cap of about 32 bytes at max
func (tw *timeWriter) _parse(t time.Time, b *bytes.Buffer, tc TimeConfig) {
	if tc.ShowDate {
		y, m, d := t.Date()
		if !tc.ShowFullYear {
			y = y % 100
			if y < 10 && tc.AddZerosForSingleDigitDate {
				b.WriteByte('0')
			}
		}
		utils.WriteInteger(b, y)
		b.WriteString(tc.DateSep)
		if m < 10 && tc.AddZerosForSingleDigitDate {
			b.WriteByte('0')
		}
		utils.WriteInteger(b, m)
		b.WriteString(tc.DateSep)
		if d < 10 && tc.AddZerosForSingleDigitDate {
			b.WriteByte('0')
		}
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
		if h < 10 && tc.AddZerosForSingleDigitTime {
			b.WriteByte('0')
		}
		utils.WriteInteger(b, h)
		b.WriteString(tc.TimeSep)
		if m < 10 && tc.AddZerosForSingleDigitTime {
			b.WriteByte('0')
		}
		utils.WriteInteger(b, m)
		b.WriteString(tc.TimeSep)
		if s < 10 && tc.AddZerosForSingleDigitTime {
			b.WriteByte('0')
		}
		utils.WriteInteger(b, s)
		if is12hr {
			if tc.AddSpaceFor12HrFormat {
				b.WriteByte(' ')
			}
			if isPm {
				b.WriteString("PM")
			} else {
				b.WriteString("AM")
			}
		}
	}
}
