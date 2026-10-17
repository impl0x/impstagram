package logger

import (
	"backend/pkg/logger/utils"
	"bytes"
	"math"
	"time"
)

type Logger struct {
	Engine engine
}

func NewLogger() Logger {
	return Logger{
		*newEngine(DefaultConfig),
	}
}

type engine struct {
	config Config
	pools  [5]_pool // 5 pools, each having a max cap, 64, 128, 256, 512, inf, panics if capacity exceeds limit of any buffer in the pool
	tb     timeBuf
}

var bufSizes = [5]int{64, 128, 256, 512, math.MaxInt}

func newEngine(config Config) *engine {
	n := func() *bytes.Buffer {
		return &bytes.Buffer{}
	}
	e := engine{
		config: config,
		tb: timeBuf{
			lastTime:    utils.NewAtomicValue(time.Time{}),
			lastTimeStr: utils.NewAtomicValue(""),
		},
	}
	for i, bs := range bufSizes {
		e.pools[i] = _pool{
			bs,
			utils.NewPool(n),
		}
	}
	return &e
}

func (e *engine) run() {

}
