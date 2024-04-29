package cached

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/athoune/stream-my-root/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

func TestMethod(t *testing.T) {
	assert.Equal(t, rpc.Method(1), Lock)
	assert.Equal(t, rpc.Method(2), Read)
	assert.Equal(t, rpc.Method(3), Write)
}
func TestCached(t *testing.T) {
	cache, err := NewCached(nil)
	assert.NoError(t, err)
	waiting := &sync.WaitGroup{}
	waiting.Add(10)
	done := &sync.WaitGroup{}
	done.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			waiting.Done()
			_, err := cache.Lock([]byte("a"))
			assert.NoError(t, err)
			done.Done()
		}()
	}
	waiting.Wait()
	buff := &bytes.Buffer{}
	buff.Write([]byte{0, 0, 1, 4})
	buff.WriteString("a")
	_, err = cache.Write(buff.Bytes())
	assert.NoError(t, err)
	done.Wait()
}

func TestGet(t *testing.T) {
	cache, err := NewCached(nil)
	assert.NoError(t, err)
	resp, err := cache.Read([]byte("plop"))
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, byte(0), resp[0])
}

func TestFull(t *testing.T) {
	temp, err := os.MkdirTemp("/tmp", "smr-cached")
	assert.NoError(t, err)
	defer os.RemoveAll(temp)

	opts := DefaultCachedOpts()
	opts.Directory = temp
	opts.Max = 3
	cache, err := NewCached(opts)
	assert.NoError(t, err)
	eviction, err := cache.write(&SetArg{
		Key:  "a",
		Size: 1,
	})
	assert.NoError(t, err)
	assert.False(t, eviction)
	eviction, err = cache.write(&SetArg{
		Key:  "a",
		Size: 2,
	})
	assert.NoError(t, err)
	assert.False(t, eviction)
	eviction, err = cache.write(&SetArg{
		Key:  "b",
		Size: 2,
	})
	assert.Error(t, err) // the key is not found
	assert.False(t, eviction)
	os.WriteFile(fmt.Sprintf("%s/a", temp), []byte{}, 0640) // lets fake the key
	eviction, err = cache.write(&SetArg{
		Key:  "b",
		Size: 2,
	})
	assert.NoError(t, err)
	assert.True(t, eviction)
}
