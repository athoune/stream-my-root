package blocks

import (
	"fmt"
	"io"
)

type Block struct {
	Seek   int
	Hash   string
	Reader io.Reader
}

type Blocks struct {
	Blocks    []*Block
	blockSize int64
}

const DEFAULT_BLOCK_SIZE = 512 * 1024

func NewBlocks(size uint64) *Blocks {
	return &Blocks{
		Blocks:    make([]*Block, 0),
		blockSize: int64(size),
	}
}

func currentBlock(blockSize, offset int64) (int64, int64) {
	return offset / blockSize, offset % blockSize
}

func (b *Blocks) Get(n int) (*Block, bool) {
	for _, bb := range b.Blocks {
		if bb.Seek == n {
			return bb, true
		}
		if bb.Seek > n {
			return nil, true
		}
	}
	return nil, false
}

// BlocksForReadingAt fetch all Blocks needed for reading the recipe
func (b *Blocks) BlocksForReadingAt(p []byte, off int64) ([]*Block, int64, error) {
	start_block, start_offset := currentBlock(b.blockSize, off)
	end_block, _ := currentBlock(b.blockSize, off+int64(len(p)))
	end := end_block - start_block
	r := make([]*Block, end+1)
	for i := 0; i <= int(end); i++ {
		block, ok := b.Get(int(start_block) + i)
		if !ok {
			return nil, 0, fmt.Errorf("can't get block %d", i)
		}
		r[i] = block
	}
	return r, start_offset, nil
}

func (b *Blocks) NumberOfBlocks() int {
	return len(b.Blocks)
}

func Rtrim(buffer []byte) []byte {
	for i := len(buffer); i > 0; i-- {
		if buffer[i-1] != 0 {
			return buffer[:i]
		}
	}
	return []byte{}
}
