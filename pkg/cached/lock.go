package cached

import "sync"

type Locker struct {
	subscribers []chan interface{}
	first       bool
	lock        *sync.Mutex
}

func NewLocker() *Locker {
	return &Locker{
		subscribers: make([]chan interface{}, 0),
		first:       true,
		lock:        &sync.Mutex{},
	}
}

func (l *Locker) Wait() bool {
	l.lock.Lock()
	if l.first {
		l.first = false
		l.lock.Unlock()
		return true
	}
	lock := make(chan interface{})
	l.subscribers = append(l.subscribers, lock)
	l.lock.Unlock()
	<-lock
	return false
}

func (l *Locker) Release() {
	for _, subscriber := range l.subscribers {
		subscriber <- nil
	}
}
