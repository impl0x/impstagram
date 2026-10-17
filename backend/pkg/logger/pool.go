package logger

import (
	"backend/pkg/logger/utils"
	"bytes"
	"strconv"
)

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
