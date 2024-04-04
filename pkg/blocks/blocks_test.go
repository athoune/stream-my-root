package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmpty(t *testing.T) {
	assert.True(t, isEmpty([]byte{0, 0, 0, 0}))
	assert.False(t, isEmpty([]byte{0, 0, 1, 0}))
}

func TestTrim(t *testing.T) {
	assert.Equal(t, Rtrim([]byte{0, 0, 1, 0, 2, 0, 0}), []byte{0, 0, 1, 0, 2})
	assert.Empty(t, Rtrim([]byte{0, 0, 0, 0}))
}
