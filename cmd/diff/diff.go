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
	fmt.Println(len(aa), len(bb), aa.Diff(bb))
}
