package logger

import (
	"backend/pkg/logger/utils"
	"bytes"
	"math"
	"strconv"
	"sync"
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

// writes the parsed time according to config values to the buffer provided,
// it should approximately have an available cap of about 32 bytes at max
func (tc TimeConfig) parse(t time.Time, b *bytes.Buffer) {
	if tc.ShowDate {
		y, m, d := t.Date()
		if !tc.ShowFullYear {
			y = y % 100
		}
		b.WriteString(strconv.Itoa(y))
		b.WriteString(tc.DateSep)
		b.WriteString(strconv.Itoa(int(m)))
		b.WriteString(tc.DateSep)
		b.WriteString(strconv.Itoa(d))
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
		b.WriteString(strconv.Itoa(h))
		b.WriteString(tc.TimeSep)
		b.WriteString(strconv.Itoa(m))
		b.WriteString(tc.TimeSep)
		b.WriteString(strconv.Itoa(s))
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
	// Time
}

type Logger struct {
	engine engine
}

type engine struct {
	config Config
	pools  [5]_pool // 5 pools, each having a max cap, 32, 64, 128, 256, inf
	tb     timeBuf
}

var bufSizes = [5]int{32, 64, 128, 256, math.MaxInt}

func newEngine(config Config) *engine {
	n := func() *bytes.Buffer {
		return &bytes.Buffer{}
	}
	e := engine{
		config: config,
	}
	for i, bs := range bufSizes {
		e.pools[i] = _pool{
			bs,
			utils.NewPool(n),
		}
	}
	return &e
}

type _pool struct {
	maxCap int
	utils.SyncPool[*bytes.Buffer]
}

func (p *_pool) Get() *bytes.Buffer {
	b := p.SyncPool.Get()
	if b.Cap() > p.maxCap {
		panic("pool get: buffer overflow, max cap - " + strconv.Itoa(p.maxCap))
	}
	return b
}

func (p *_pool) Put(b *bytes.Buffer) {
	if b.Cap() > p.maxCap {
		panic("pool put: buffer overflow, max cap - " + strconv.Itoa(p.maxCap))
	}
	p.SyncPool.Put(b)
}

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

func (e *engine) time() string {
	now := time.Now()
	if tEq(now, e.tb.lastTime.Load()) {
		return e.tb.lastTimeStr.Load()
	}

	// getting a buffer from the lowest pool size
	buf := e.pools[0].Get()
	e.config.Time.parse(now, buf)
	str := buf.String()

	// setting the lastTime to current
	e.tb.mu.Lock()
	e.tb.lastTime.Store(now)
	e.tb.lastTimeStr.Store(str)
	e.tb.mu.Unlock()

	// reset and put buf to pool
	buf.Reset()
	e.pools[0].Put(buf)

	return str
}
