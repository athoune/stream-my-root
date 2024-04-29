package lock

import "sync"

type Lock struct {
	cond  chan interface{}
	first bool
	lock  *sync.Mutex
}

func NewLock() *Lock {
	return &Lock{
		first: true,
		cond:  make(chan interface{}),
		lock:  &sync.Mutex{},
	}
}

func (l *Lock) Wait() bool {
	l.lock.Lock()
	if l.first {
		l.first = false
		l.lock.Unlock()
		return true
	}
	l.lock.Unlock()
	<-l.cond
	return false
}

func (l *Lock) Release() {
	if !l.first {
		close(l.cond) // release all listeners
	}
}
