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
			lastSec:     atomic.Int64{}, // doesn't matter if we copy lock, ignore warning
			lastTimeStr: atomic.Pointer[bytes.Buffer]{},
			pool: _pool{
				minCap:   64, // cannot use 32 as 64 is the minimum limit for a buffer, and its automatically set to that every time.
				maxCap:   64,
				SyncPool: utils.NewPool(func() *bytes.Buffer { return &bytes.Buffer{} }),
			},
		},
	}
	tw.tb.lastTimeStr.Store(&bytes.Buffer{})
	return tw // does not matter if we copy lock, this function is inlined anyways.
}

type timeBuf struct {
	lastSec     atomic.Int64
	lastTimeStr atomic.Pointer[bytes.Buffer]
	mu          sync.Mutex
	pool        _pool
}

func (tw *timeWriter) write(b *bytes.Buffer) {
	nowUnix := time.Now().Unix()
	if nowUnix == tw.tb.lastSec.Load() {
		b.Write(tw.tb.lastTimeStr.Load().Bytes())
		return
	}
	tw.tb.mu.Lock()
	defer tw.tb.mu.Unlock()
	// if some goroutine has changed it while we were waiting for the lock, double check
	if nowUnix == tw.tb.lastSec.Load() {
		b.Write(tw.tb.lastTimeStr.Load().Bytes())
		return
	}
	// get a buffer from the pool and use it
	buf := tw.tb.pool.Get()
	now := time.Now()               // reason we dont use the previous time.Now call is due to allocation, on hot path it would get allocated every call
	tw._parse(buf, now, tw.config)  // we write the time in the buffer we are going to store
	tw.tb.lastSec.Store(now.Unix()) // store the current unix

	// putting the last buffer we had back into the pool, after resetting it obviously.
	lastBuf := tw.tb.lastTimeStr.Load()
	lastBuf.Reset()
	tw.tb.pool.Put(lastBuf)
	// store the current buffer
	tw.tb.lastTimeStr.Store(buf)
	b.Write(buf.Bytes()) // write the current parsed time

}

// writes the parsed time according to config values to the buffer provided,
// it should approximately have an available cap of about 32 bytes at max
func (tw *timeWriter) _parse(b *bytes.Buffer, t time.Time, tc TimeConfig) {
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
