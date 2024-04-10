package blocks

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ReadRecipe(f io.Reader) (*Blocks, error) {
	blocks := NewBlocks(DEFAULT_BLOCK_SIZE)
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
