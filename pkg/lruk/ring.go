package lruk

import (
	"fmt"
	"sync"
)

type Ring[V any] struct {
	values []V
	pos    uint32
	size   uint32
	lock   *sync.RWMutex
	last   int
}

func NewRing[V any](size int, values ...V) *Ring[V] {
	r := &Ring[V]{
		values: make([]V, size),
		pos:    0,
		size:   0,
		lock:   &sync.RWMutex{},
	}
	for _, value := range values {
		r.Append(value)
	}
	return r
}

func (r *Ring[V]) Append(value V) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.values[r.pos] = value
	if int(r.size) < len(r.values) {
		r.size += 1
	}
	r.pos += 1
	if int(r.pos) >= len(r.values) {
		r.pos = 0
	}
	r.last = r.rewind(0)
}

func (r *Ring[V]) rewind(n uint32) int {
	return (int(r.pos) + len(r.values) - int(n) - 1) % len(r.values)
}

func (r *Ring[V]) Last() V {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.values[r.last]
}

// Hist fetch previous value, 0 is the last value
func (r *Ring[V]) Hist(h uint32) (value V, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if int(h) >= len(r.values) {
		err = fmt.Errorf("out of bound %d %d", h, len(r.values))
		return
	}
	return r.values[r.rewind(h)], nil
}

func (r *Ring[V]) Length() int {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return int(r.size)
}
