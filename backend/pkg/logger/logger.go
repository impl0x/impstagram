package logger

import (
	"backend/pkg/logger/utils"
	"bytes"
	"math"
)

type Logger struct {
	engine engine
}

type engine struct {
	config Config
	pools  [5]_pool // 5 pools, each having a max cap, 32, 64, 128, 256, inf, panics if capacity exceeds limit of any buffer in the pool
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

func (e *engine) run()
