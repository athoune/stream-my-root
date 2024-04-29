package cached

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/athoune/stream-my-root/pkg/lock"
	"github.com/athoune/stream-my-root/pkg/lruk"
	"github.com/athoune/stream-my-root/pkg/rpc"
)

const (
	_ rpc.Method = iota
	Lock
	Read
	Write
)

type Cached struct {
	lru       *lruk.LRUK[string, int]
	locks     map[string]*lock.Lock
	mutex     *sync.RWMutex
	directory string
}

type CachedOpts struct {
	K              uint
	CorrelatedTime time.Duration
	Max            uint
	Directory      string
}

func DefaultCachedOpts() *CachedOpts {
	return &CachedOpts{
		K:              2,
		CorrelatedTime: time.Second,
		Max:            1024,
	}
}

func NewCached(opts *CachedOpts) (*Cached, error) {
	if opts == nil {
		opts = DefaultCachedOpts()
	}
	if opts.Directory == "" {
		slog.Info("No cache directory set")
	} else {
		info, err := os.Stat(opts.Directory)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("not a directory : %s", opts.Directory)
		}
	}
	cached := &Cached{
		locks:     make(map[string]*lock.Lock),
		mutex:     &sync.RWMutex{},
		directory: opts.Directory,
	}
	cached.lru = lruk.New[string, int](opts.K, opts.CorrelatedTime, opts.Max, func(i int) int {
		return i
	}, cached.deleteHandler)
	return cached, nil
}

func (c *Cached) deleteHandler(key string) error {
	if c.directory == "" { // FIXME is it a good idea to accept unset directory ?
		return nil
	}
	logger := slog.Default().With("directory", c.directory, "key", key)
	if strings.ContainsRune(key, '.') {
		err := fmt.Errorf("dangerous key : %s", key)
		logger.Error(err.Error())
		return err
	}
	file := fmt.Sprintf("%s/%s", c.directory, key)
	err := os.Remove(file)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	return nil
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

func (c *Cached) Read(raw []byte) ([]byte, error) {
	key := string(raw)
	_, ok := c.lru.Get(key)
	resp := Bool(ok)
	return resp.MarshalBinary()
}

func (c *Cached) write(arg *SetArg) (bool, error) {
	c.mutex.RLock()
	eviction, err := c.lru.Add(arg.Key, int(arg.Size))
	if err != nil {
		c.mutex.RUnlock()
		return eviction, err
	}
	locker, ok := c.locks[arg.Key]
	c.mutex.RUnlock()
	if ok {
		locker.Release()
	}
	return eviction, nil
}

func (c *Cached) Write(raw []byte) ([]byte, error) {
	arg := &SetArg{}
	err := arg.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	eviction, err := c.write(arg)
	if err != nil {
		return nil, err
	}
	resp, err := Bool(eviction).MarshalBinary()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Cached) RegisterAll(server *rpc.Server) {
	server.Register(Lock, c.Lock)
	server.Register(Read, c.Read)
	server.Register(Write, c.Write)
}
