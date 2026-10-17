package logger

import (
	"backend/pkg/logger/utils"
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
