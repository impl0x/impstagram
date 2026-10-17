package logger

import (
	"backend/pkg/logger/utils"
	"bytes"
	"sync"
	"time"
)

type timeBuf struct {
	lastTime    utils.AtomicValue[time.Time]
	lastTimeStr utils.AtomicValue[string]
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

func (e *engine) writeTime(b *bytes.Buffer) {
	now := time.Now()
	if tEq(now, e.tb.lastTime.Load()) {
		str := e.tb.lastTimeStr.Load()
		b.WriteString(str)
		return
	}
	// getting a buffer from the lowest pool size
	parseTime(now, b, e.config.Time)
	str := b.String() // allocates

	// setting the lastTime to current
	e.tb.mu.Lock()
	e.tb.lastTime.Store(now)
	e.tb.lastTimeStr.Store(str)
	e.tb.mu.Unlock()
}

// writes the parsed time according to config values to the buffer provided,
// it should approximately have an available cap of about 32 bytes at max
func parseTime(t time.Time, b *bytes.Buffer, tc TimeConfig) {
	if tc.ShowDate {
		y, m, d := t.Date()
		if !tc.ShowFullYear {
			y = y % 100
			if y < 10 && tc.AddZerosForSingleDigitDate {
				b.WriteRune('0')
			}
		}
		utils.WriteInteger(b, y)
		b.WriteString(tc.DateSep)
		if m < 10 && tc.AddZerosForSingleDigitDate {
			b.WriteRune('0')
		}
		utils.WriteInteger(b, m)
		b.WriteString(tc.DateSep)
		if d < 10 && tc.AddZerosForSingleDigitDate {
			b.WriteRune('0')
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
			b.WriteRune('0')
		}
		utils.WriteInteger(b, h)
		b.WriteString(tc.TimeSep)
		if m < 10 && tc.AddZerosForSingleDigitTime {
			b.WriteRune('0')
		}
		utils.WriteInteger(b, m)
		b.WriteString(tc.TimeSep)
		if s < 10 && tc.AddZerosForSingleDigitTime {
			b.WriteRune('0')
		}
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
