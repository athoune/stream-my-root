package blocks

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Block struct {
	Seek int
	Hash string
}

type Blocks []Block

func ReadRecipe(f io.Reader) (Blocks, error) {
	blocks := make(Blocks, 0)
	fileScanner := bufio.NewScanner(f)

	fileScanner.Split(bufio.ScanLines)
	var err error
	i := 0
	for fileScanner.Scan() {
		slugs := strings.Split(fileScanner.Text(), " ")
		block := Block{
			Hash: slugs[1],
		}
		block.Seek, err = strconv.Atoi(slugs[0])
		if err != nil {
			return nil, fmt.Errorf("can't parse line %d [%s] : %s", i, slugs, err)
		}
		blocks = append(blocks, block)
		i++
	}
	return blocks, nil
}

func (b Blocks) set() map[string]interface{} {
	// FIXME maybe some caching
	s := make(map[string]interface{})
	for _, block := range b {
		s[block.Hash] = new(interface{})
	}
	return s
}

func (b Blocks) Distinct() int {
	return len(b.set())
}

func (b Blocks) Diff(other Blocks) int {
	left := b.set()
	cpt := 0
	for _, block := range other {
		_, ok := left[block.Hash]
		if ok {
			cpt++
		}
	}
	return cpt
}
