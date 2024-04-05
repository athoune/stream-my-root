package main

import (
	"fmt"
	"os"

	"github.com/athoune/stream-my-root/pkg/blocks"
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
	fmt.Printf(`A: %d chunks, %d unique chunks
B: %d chunks, %d unique chunks
B has %d chunks in common with A
`,
		len(aa), aa.Distinct(),
		len(bb), bb.Distinct(),
		aa.Diff(bb))
}
