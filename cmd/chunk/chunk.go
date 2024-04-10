package main

import (
	"fmt"
	"os"

	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/athoune/stream-my-root/pkg/chunk"
)

func main() {
	chunker := chunk.NewChunker(blocks.DEFAULT_BLOCK_SIZE)

	for _, img := range os.Args[1:] {
		_, err := os.Stat(fmt.Sprintf("%s.recipe", img))
		if err != nil && os.IsNotExist(err) { // FIXME handle other errors
			err = chunker.ChunkRawFile(img)
			if err != nil {
				fmt.Printf("🐞 %s : %s\n", img, err)
				continue
			}
			fmt.Println("✅ ", img)
		} else {
			fmt.Println("☁️ ", img)
		}
	}
}
