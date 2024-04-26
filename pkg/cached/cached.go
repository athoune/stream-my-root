package cached

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/athoune/stream-my-root/pkg/lock"
	"github.com/athoune/stream-my-root/pkg/lruk"
	"github.com/athoune/stream-my-root/pkg/rpc"
)

const (
	Lock rpc.Method = 1
	Get
	Set
)

type Cached struct {
	lru   *lruk.LRUK[string, int]
	locks map[string]*lock.Lock
	mutex *sync.RWMutex
}

func NewCached(max uint) *Cached {
	return &Cached{
		lru: lruk.New[string, int](2, time.Second, max, func(i int) int {
			return i
		}, nil),
		locks: make(map[string]*lock.Lock),
		mutex: &sync.RWMutex{},
	}
}

func (c *Cached) Lock(raw []byte) ([]byte, error) {
	key := string(raw)
	c.mutex.Lock()
	locker, ok := c.locks[key]
	if !ok {
		locker = lock.NewLock()
		c.locks[key] = locker
	}
	c.mutex.Unlock()
	locker.Wait()
	return nil, nil
}

func (c *Cached) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (c *Cached) Set(raw []byte) ([]byte, error) {
	size := binary.BigEndian.Uint32(raw[0:4])
	key := string(raw[4:])
	c.mutex.RLock()
	c.lru.Add(key, int(size))
	locker, ok := c.locks[key]
	c.mutex.RUnlock()
	if ok {
		locker.Release()
	}
	return nil, nil
}

func (c *Cached) RegisterAll(server *rpc.Server) {
	server.Register(Lock, c.Lock)
	server.Register(Get, c.Get)
	server.Register(Set, c.Set)
}
