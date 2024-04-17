package blocks

import (
	"fmt"
	"os"
)

func (b Recipe) set() map[string]interface{} {
	// FIXME maybe some caching
	s := make(map[string]interface{})
	for _, block := range b.Blocks {
		s[block.Hash] = new(interface{})
	}
	return s
}

func (b *Recipe) Distinct() int {
	return len(b.set())
}

func (b *Recipe) Diff(other *Recipe) int {
	left := b.set()
	cpt := 0
	for _, block := range other.Blocks {
		_, ok := left[block.Hash]
		if ok {
			cpt++
		}
	}
	return cpt
}

func (b *Recipe) DiffSize(other *Recipe) (int, error) {
	left := b.set()
	common := make(map[string]interface{})
	for _, block := range other.Blocks {
		_, ok := left[block.Hash]
		if ok {
			common[block.Hash] = new(interface{})
		}
	}
	size := 0
	for _, block := range other.Blocks {
		chunk := fmt.Sprintf("smr/%s.zst", block.Hash)
		info, err := os.Stat(chunk)
		if err != nil {
			return 0, err
		}
		size += int(info.Size())
	}

	return size, nil

}
