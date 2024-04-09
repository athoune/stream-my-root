package main

import (
	"fmt"
	"os"

	"github.com/athoune/stream-my-root/pkg/blocks"
)

func main() {
	for _, img := range os.Args[1:] {
		_, err := os.Stat(fmt.Sprintf("%s.recipe", img))
		if err != nil && os.IsNotExist(err) { // FIXME handle other errors
			err = blocks.ChunkRawFile(img)
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
