package main

import (
	"fmt"
	"os"

	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/dustin/go-humanize"
)

func main() {
	a, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	aa, err := blocks.ReadRecipe(a)
	if err != nil {
		panic(err)
	}
	b, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	bb, err := blocks.ReadRecipe(b)
	if err != nil {
		panic(err)
	}
	common_size, err := aa.DiffSize(bb)
	if err != nil {
		panic(err)
	}
	fmt.Printf(`A: %d chunks, %d unique chunks
B: %d chunks, %d unique chunks
B has %d chunks in common with A, %s
`,
		aa.NumberOfBlocks(), aa.Distinct(),
		bb.NumberOfBlocks(), bb.Distinct(),
		aa.Diff(bb),
		humanize.Bytes(uint64(common_size)),
	)
}
