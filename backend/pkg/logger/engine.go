package logger

import (
	"backend/pkg/logger/utils"
	"bytes"
	"math"
)

type engine struct {
	config     Config
	pools      [5]_pool // 5 pools, each having a max cap, 64, 128, 256, 512, inf, panics if capacity exceeds limit of any buffer in the pool
	timeWriter timeWriter
}

var bufSizes = [5]int{64, 128, 256, 512, math.MaxInt}

func newEngine(config Config) *engine {
	n := func() *bytes.Buffer {
		return &bytes.Buffer{}
	}
	e := engine{
		config:     config,
		timeWriter: newTimeWriter(config.Time),
	}
	for i, bs := range bufSizes {
		e.pools[i] = _pool{
			bs,
			utils.NewPool(n),
		}
	}
	return &e
}

func (e *engine) log(l level, m string, kv []string) {
	if len(kv)%2 != 0 {
		panic("log: invalid number of key value pairs passed!")
	}
	b := e.pools[4].Get()
	writeColorStart(b, e.config.Color.Time)
	e.timeWriter.write(b)

	b.WriteByte(' ')

	writeColorStart(b, e.config.Color.Levels.fromLevel(l))
	writeLevel(b, l)

	b.WriteByte(' ')

	writeColorStart(b, e.config.Color.MsgColor)
	b.WriteString(m)

	b.WriteByte(' ')

	for i, v := range kv {
		if i%2 == 0 {
			writeColorStart(b, e.config.Color.ExKey)
			b.WriteString(v)
			b.WriteByte('=')
		} else {
			writeColorStart(b, e.config.Color.ExVal)
			b.WriteString(v)
		}
	}
	writeColorEnd(b)

	b.WriteByte('\n')

	b.WriteTo(e.config.Output)
}
