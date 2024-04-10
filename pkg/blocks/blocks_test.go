package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrim(t *testing.T) {
	assert.Equal(t, Rtrim([]byte{0, 0, 1, 0, 2, 0, 0}), []byte{0, 0, 1, 0, 2})
	assert.Empty(t, Rtrim([]byte{0, 0, 0, 0}))
}

func TestCurrentBlock(t *testing.T) {
	block, offset := currentBlock(10, 5)
	assert.Equal(t, int64(0), block)
	assert.Equal(t, int64(5), offset)
	block, offset = currentBlock(10, 11)
	assert.Equal(t, int64(1), block)
	assert.Equal(t, int64(1), offset)
}

func TestBlocksForReadingAt(t *testing.T) {
	b := &Blocks{
		blockSize: 8,
		Blocks: []*Block{
			{
				Seek: 0,
				Hash: "aaaa",
			},
			{
				Seek: 2,
				Hash: "bbbb",
			},
		},
	}
	buff := make([]byte, 12)
	blocks, off, err := b.BlocksForReadingAt(buff, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), off)
	assert.Len(t, blocks, 2)
	assert.Equal(t, "aaaa", blocks[0].Hash)
	assert.Nil(t, blocks[1])
}
