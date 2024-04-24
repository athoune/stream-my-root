package lruk

import (
	"fmt"
	"sync"
	"time"
)

type LRUK[K comparable, V any] struct {
	correlated time.Duration
	k          int
	cache      map[K]V
	history    map[K]*Ring[time.Time]
	lock       *sync.RWMutex
	max        int
	sizeFunc   func(V) int
	deleteFunc func(K) error
	size       int
}

func New[K comparable, V any](k uint, correlated_time time.Duration, max uint, sizeFunc func(V) int, deleteFunc func(K) error) *LRUK[K, V] {
	return &LRUK[K, V]{
		correlated: correlated_time,
		k:          int(k),
		cache:      make(map[K]V),
		history:    make(map[K]*Ring[time.Time]),
		lock:       &sync.RWMutex{},
		max:        int(max),
		sizeFunc:   sizeFunc,
		deleteFunc: deleteFunc,
	}
}

func (l *LRUK[K, V]) Add(key K, value V) error {
	return l.AddAt(key, value, time.Now())
}

func (l *LRUK[K, V]) AddAt(key K, value V, ts time.Time) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	size := l.sizeFunc(value)
	if size > l.max {
		return fmt.Errorf("value is too large %d > %d", size, l.max)
	}
	if size+l.size > l.max { // eviction time
		err := l.eviction(l.sizeFunc(value))
		if err != nil {
			return err
		}
	}
	l.cache[key] = value
	l.history[key] = NewRing[time.Time](l.k, ts)
	l.size += size
	return nil
}

func (l *LRUK[K, V]) eviction(size int) error {
	old := time.Now()
	var key K
	var last time.Time
	for i := 0; i < len(l.cache); i++ {
		for k, history := range l.history {
			last = history.Last()
			if last.Before(old) {
				old = last
				key = k
			}
		}
		l.delete(key)
		if l.size+size < l.max { // enough place is freed
			return nil
		}
	}
	return nil
}

func (l *LRUK[K, V]) Delete(key K) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.delete(key)
}

func (l *LRUK[K, V]) delete(key K) error {
	if l.deleteFunc != nil {
		err := l.deleteFunc(key)
		if err != nil {
			return err
		}
	}
	l.size -= l.sizeFunc(l.cache[key])
	delete(l.cache, key)
	delete(l.history, key)
	return nil
}

func (l *LRUK[K, V]) Get(key K) (V, bool) {
	return l.GetAt(key, time.Now())
}

func (l *LRUK[K, V]) GetAt(key K, ts time.Time) (value V, ok bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	value, ok = l.cache[key]
	if !ok {
		return
	}
	history := l.history[key]
	if ts.Sub(history.Last()) > l.correlated {
		history.Append(ts)
	}
	return
}
