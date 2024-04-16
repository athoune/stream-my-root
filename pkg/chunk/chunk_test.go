package chunk

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrim(t *testing.T) {
	assert.Equal(t, Rtrim([]byte{0, 0, 1, 0, 2, 0, 0}), []byte{0, 0, 1, 0, 2})
	assert.Empty(t, Rtrim([]byte{0, 0, 0, 0}))
}

func TestChunker(t *testing.T) {
	f, err := os.MkdirTemp("/tmp", "smr")
	assert.NoError(t, err)
	defer os.Remove(f)

	r := make([]byte, 16)
	for i := 0; i < len(r); i++ {
		r[i] = byte(i)
	}
	raw := fmt.Sprintf("%s/raw", f)
	err = os.WriteFile(raw, r, 0660)
	assert.NoError(t, err)
	c := NewChunker(8)
	c.folder = f
	err = c.ChunkRawFile(raw)
	assert.NoError(t, err)
}
