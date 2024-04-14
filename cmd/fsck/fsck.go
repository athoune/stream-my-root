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
	for _, block := range aa.Blocks {
		_, err := os.Stat(fmt.Sprintf("smr/%s.zst", block.Hash))
		if err != nil {
			fmt.Println("🐞 ", block.Hash, err)
		}
	}
	fmt.Println(len(aa.Blocks), "chunks exist")
}
