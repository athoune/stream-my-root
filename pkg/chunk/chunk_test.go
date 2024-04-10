package chunk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrim(t *testing.T) {
	assert.Equal(t, Rtrim([]byte{0, 0, 1, 0, 2, 0, 0}), []byte{0, 0, 1, 0, 2})
	assert.Empty(t, Rtrim([]byte{0, 0, 0, 0}))
}
