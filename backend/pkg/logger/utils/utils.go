package utils

import (
	"sync"
	"sync/atomic"
)

type SyncPool[T any] struct {
	p sync.Pool
}

func NewPool[T any](new func() T) SyncPool[T] {
	return SyncPool[T]{
		sync.Pool{
			New: func() any { return new() },
		},
	}
}

func (p *SyncPool[T]) Get() T {
	return p.p.Get().(T)
}

func (p *SyncPool[T]) Put(x T) {
	p.p.Put(x)
}

type AtomicValue[T any] struct {
	v atomic.Value
}

func (v AtomicValue[T]) Store(x T) {
	v.v.Store(x)
}

func (v AtomicValue[T]) Load() T {
	return v.v.Load().(T)
}

func (v AtomicValue[T]) Swap(new T) {
	v.Swap(new)
}
