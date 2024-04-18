package trimmed

import (
	"io"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRTrim(t *testing.T) {
	a := Rtrim([]byte{0, 0, 0, 0})
	assert.Len(t, a, 0)
}

func TestTrimmed(t *testing.T) {
	a := New([]byte{1, 2, 3, 0, 0, 0, 0, 0})
	assert.Equal(t, 8, a.size)
	assert.Len(t, a.values, 3)

	n, err := a.ReadAt(nil, 8)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)

	buffer := []byte{1, 1, 1} // the buffer is dirty
	n, err = a.ReadAt(buffer, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{1, 2, 3}, buffer)

	buffer = []byte{1, 1, 1, 1, 1, 1, 1, 1}
	n, err = a.ReadAt(buffer, 0)
	assert.NoError(t, err)
	assert.Equal(t, 8, n)
	assert.Equal(t, []byte{1, 2, 3, 0, 0, 0, 0, 0}, buffer)

	buffer = []byte{1, 1, 1, 1, 1, 1, 1, 1}
	n, err = a.ReadAt(buffer, 2)
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte{3, 0, 0, 0, 0, 0, 1, 1}, buffer)

	buffer = []byte{1, 1}
	n, err = a.ReadAt(buffer, 6)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, []byte{0, 0}, buffer)
}

func randomArray(n uint64) []byte {
	a := make([]byte, n)
	a_v := rand.Intn(len(a))
	for i := 0; i < a_v; i++ {
		a[i] = byte(rand.Intn(8))
	}
	return a
}
func FuzzTrimmed(f *testing.F) {
	f.Add(uint64(2048), uint64(2048), uint64(2048), uint64(512))
	f.Fuzz(func(t *testing.T, trimmed_size, buffer_size, off_size uint64, zero uint64) {
		raw := randomArray(trimmed_size)
		a, err := NewTrimmed(raw, len(raw)+int(zero))
		assert.NoError(t, err)
		buffer := make([]byte, buffer_size)
		_, err = a.ReadAt(buffer, int64(off_size))
		if err != nil {
			assert.Equal(t, io.EOF, err)
		}
	})
}

func TestNew(t *testing.T) {
	a, err := NewTrimmed([]byte{1, 2, 3}, 8)
	assert.NoError(t, err)
	assert.Equal(t, 8, a.size)

	a, err = NewTrimmed([]byte{1, 2, 3}, 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, a.size)

	a, err = NewTrimmed([]byte{1, 2, 3}, 2)
	assert.Error(t, err)
	assert.Nil(t, a)
}
