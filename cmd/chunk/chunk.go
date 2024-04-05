package main

import (
	"os"

	"github.com/athoune/stream-my-root/pkg/blocks"
)

func main() {
	err := blocks.ChunkRawFile(os.Args[1])
	if err != nil {
		panic(err)
	}
}
