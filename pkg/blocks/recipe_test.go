package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhichBlock(t *testing.T) {
	b := &Recipe{
		blockSize: 10,
	}
	block, offset := b.whichBlock(5)
	assert.Equal(t, int64(0), block)
	assert.Equal(t, int64(5), offset)
	block, offset = b.whichBlock(11)
	assert.Equal(t, int64(1), block)
	assert.Equal(t, int64(1), offset)

	b = &Recipe{
		blockSize: DEFAULT_BLOCK_SIZE,
	}
	block, offset = b.whichBlock(490861)
	assert.Equal(t, int64(0), block)
	assert.Equal(t, int64(490861), offset)
}

func TestBlocksForReadingAt(t *testing.T) {
	b := &Recipe{
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
	slice, err := b.BlocksForReading(2, 12)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), slice.start)
	assert.Equal(t, int64(6), slice.end)
	assert.Equal(t, 2, slice.Length())
	assert.Equal(t, "aaaa", slice.blocks[0].Hash)
	assert.Nil(t, slice.blocks[1])

	slice, err = b.BlocksForReading(21, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), slice.start)
	assert.Equal(t, int64(7), slice.end)
	assert.Equal(t, 1, slice.Length())
}
