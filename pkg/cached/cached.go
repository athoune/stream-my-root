package cached

import (
	"sync"
	"time"

	"github.com/athoune/stream-my-root/pkg/lock"
	"github.com/athoune/stream-my-root/pkg/lruk"
	"github.com/athoune/stream-my-root/pkg/rpc"
)

const (
	_ rpc.Method = iota
	Lock
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
	ok = locker.Wait()
	resp := Bool(ok)
	return resp.MarshalBinary()
}

func (c *Cached) Get(raw []byte) ([]byte, error) {
	key := string(raw)
	_, ok := c.lru.Get(key)
	resp := Bool(ok)
	return resp.MarshalBinary()
}

func (c *Cached) Set(raw []byte) ([]byte, error) {
	arg := &SetArg{}
	err := arg.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	c.mutex.RLock()
	c.lru.Add(arg.Key, int(arg.Size))
	locker, ok := c.locks[arg.Key]
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
