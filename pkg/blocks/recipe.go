package blocks

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
)

type Recipe struct {
	Blocks    []*Block
	blockSize int64
}

func NewRecipe(size uint64) *Recipe {
	return &Recipe{
		Blocks:    make([]*Block, 0),
		blockSize: int64(size),
	}
}

func ReadRecipe(f io.Reader) (*Recipe, error) {
	blocks := NewRecipe(DEFAULT_BLOCK_SIZE)
	fileScanner := bufio.NewScanner(f)

	fileScanner.Split(bufio.ScanLines)
	var err error
	i := 0
	for fileScanner.Scan() {
		slugs := strings.Split(fileScanner.Text(), " ")
		block := &Block{
			Hash: slugs[1],
		}
		block.Seek, err = strconv.Atoi(slugs[0])
		if err != nil {
			return nil, fmt.Errorf("can't parse line %d [%s] : %s", i, slugs, err)
		}
		blocks.Blocks = append(blocks.Blocks, block)
		i++
	}
	return blocks, nil
}

func (b *Recipe) whichBlock(offset int64) (int64, int64) {
	return offset / b.blockSize, offset % b.blockSize
}

func (b *Recipe) Get(n int) (*Block, bool) {
	for _, bb := range b.Blocks {
		if bb.Seek == n {
			return bb, true
		}
		if bb.Seek > n {
			return nil, true
		}
	}
	// We don't know the size of the image, it may have trailing 0
	return nil, true
}

// BlocksForReadingAt fetch all Blocks needed for reading the recipe
func (b *Recipe) BlocksForReading(seek, length int64) (*Slice, error) {
	if length == 0 {
		slog.Debug("Blocks.BlocksForReading", "seek", seek, "length", length)
		return &Slice{}, nil
	}
	start_block, start_offset := b.whichBlock(seek)
	end_block, end_offset := b.whichBlock(seek + length)
	end := end_block - start_block
	slice := &Slice{
		start:  start_offset,
		end:    end_offset,
		blocks: make([]*Block, end+1),
	}
	for i := range slice.blocks {
		block, ok := b.Get(int(start_block) + i)
		if !ok {
			return nil, fmt.Errorf("can't get block %d", i)
		}
		slice.blocks[i] = block
	}
	return slice, nil
}

func (b *Recipe) NumberOfBlocks() int {
	return len(b.Blocks)
}
