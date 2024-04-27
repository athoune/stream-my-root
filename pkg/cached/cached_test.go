package cached

import (
	"bytes"
	"sync"
	"testing"

	"github.com/athoune/stream-my-root/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

func TestMethod(t *testing.T) {
	assert.Equal(t, rpc.Method(1), Lock)
	assert.Equal(t, rpc.Method(2), Get)
	assert.Equal(t, rpc.Method(3), Set)
}
func TestCached(t *testing.T) {
	cache := NewCached(1024 * 1024)
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
	_, err := cache.Set(buff.Bytes())
	assert.NoError(t, err)
	done.Wait()
}

func TestGet(t *testing.T) {
	cache := NewCached(1024 * 1024)
	resp, err := cache.Get([]byte("plop"))
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, byte(0), resp[0])
}
