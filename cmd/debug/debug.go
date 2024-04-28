package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/athoune/stream-my-root/pkg/backend/ro"
	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/athoune/stream-my-root/pkg/reader/local"
	"github.com/lmittmann/tint"
)

func main() {

	opts := &tint.Options{
		AddSource:  true,
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	}
	var handler slog.Handler = tint.NewHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	img, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	recipe_path := fmt.Sprintf("%s.recipe", os.Args[1])
	r, err := os.Open(recipe_path)
	if err != nil {
		panic(err)
	}
	recipe, err := blocks.ReadRecipe(r)
	if err != nil {
		panic(err)
	}
	slog.Info("Image", "image", recipe_path, "blocks", recipe.NumberOfBlocks())
	reader, err := local.NewLocalReader(&local.LocalReaderOpts{
		CacheDirectory: "smr",
		Tainted:        false,
	})
	if err != nil {
		panic(err)
	}
	backend := ro.NewROBackend(recipe, reader)
	var offset int64

	for i := 0; i < 100000; i++ {
		buff := make([]byte, rand.Intn(4096))
		offset = int64(rand.Intn(100)*1024*1024 + rand.Intn(1024*1024))
		n, err := backend.ReadAt(buff, offset)
		fmt.Println("n", n)

		if err != nil {
			panic(err)
		}
		img.Seek(offset, 0)
		buff2 := make([]byte, len(buff))
		n2, err := img.Read(buff2)
		if err != nil {
			panic(err)
		}
		if n2 != n {
			panic("!= n")
		}
		if !bytes.Equal(buff, buff2) {
			panic("!= buff")
		}
	}
}
