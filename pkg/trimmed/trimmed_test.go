package trimmed

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRTrim(t *testing.T) {
	a := Rtrim([]byte{0, 0, 0, 0})
	assert.Len(t, a, 0)
}

func TestMinus(t *testing.T) {
	assert.Equal(t, int64(42), min(47, 42))
}

func TestTrimmed(t *testing.T) {
	a := New([]byte{1, 2, 3, 0, 0, 0, 0, 0})
	assert.Equal(t, 8, a.size)
	assert.Len(t, a.values, 3)

	_, err := a.ReadAt(nil, 8)
	assert.Equal(t, io.EOF, err)
	buffer := make([]byte, 8)
	n, err := a.ReadAt(buffer, 2)
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte{3, 0, 0, 0, 0, 0, 0, 0}, buffer)
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
