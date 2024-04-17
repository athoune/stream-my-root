package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/athoune/stream-my-root/pkg/backend/ro"
	"github.com/athoune/stream-my-root/pkg/blocks"
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

	r, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	recipe, err := blocks.ReadRecipe(r)
	if err != nil {
		panic(err)
	}
	slog.Info("Image", "image", os.Args[1], "blocks", recipe.NumberOfBlocks())
	reader, err := blocks.NewLocalReader("smr")
	if err != nil {
		panic(err)
	}
	backend := ro.NewROBackend(recipe, reader)

	for i := 0; i < 10000; i++ {
		buff := make([]byte, rand.Intn(4096))
		n, err := backend.ReadAt(buff, int64(rand.Intn(100)*1024*1024+rand.Intn(1024*1024)))
		fmt.Println("n", n)

		if err != nil {
			panic(err)
		}
	}
}
