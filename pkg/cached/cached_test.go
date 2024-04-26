package cached

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
