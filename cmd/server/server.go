package main

import (
	"cmp"
	"log/slog"
	"os"
	"time"

	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/athoune/stream-my-root/pkg/nbd"
	"github.com/athoune/stream-my-root/pkg/reader/http"
	"github.com/lmittmann/tint"
)

func main() {
	opts := &tint.Options{
		AddSource:  true,
		Level:      slog.LevelInfo,
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

	reader, err := http.NewHttpReader(&http.HttpReaderOpts{
		CacheDirectory: "smr",
		CachedUrl:      cmp.Or(os.Getenv("CAHED_URL"), "/tmp/cached.sock"),
		SourceUrl:      cmp.Or(os.Getenv("SOURCE_URL"), "http://localhost:8080"),
	})

	if err != nil {
		panic(err)
	}

	err = nbd.Serve(recipe, reader, "tcp://0.0.0.0:10809")
	if err != nil {
		panic(err)
	}
}
