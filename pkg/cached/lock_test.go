package cached

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	l := NewLocker()
	ok := l.Wait()
	assert.True(t, ok)
	waiting := &sync.WaitGroup{}
	waiting.Add(10)
	done := &sync.WaitGroup{}
	done.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			waiting.Done()
			ok := l.Wait()
			assert.False(t, ok)
			done.Done()
		}(i)
	}
	waiting.Wait()
	l.Release()
	done.Wait()
}
