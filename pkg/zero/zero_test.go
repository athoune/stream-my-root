package zero

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZero(t *testing.T) {
	z := NewZero(8)
	b := make([]byte, 8)
	n, err := z.ReadAt(b, 2)
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0}, b[:6])
}
