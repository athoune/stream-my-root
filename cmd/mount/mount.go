package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/athoune/stream-my-root/pkg/imgfs"
	"github.com/athoune/stream-my-root/pkg/reader/local"
	"github.com/jacobsa/fuse"
	"github.com/lmittmann/tint"
)

var fMountPoint = flag.String("mount_point", "", "Path to mount point.")
var fDebug = flag.Bool("debug", false, "Enable debug logging.")

func main() {
	opts := &tint.Options{
		AddSource:  true,
		Level:      slog.LevelInfo,
		TimeFormat: time.Kitchen,
	}
	var handler slog.Handler = tint.NewHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Println("args", flag.Args())
		panic(err)
	}
	recipe, err := blocks.ReadRecipe(r)
	if err != nil {
		panic(err)
	}
	slog.Info("Image", "image", os.Args[1], "blocks", recipe.NumberOfBlocks())
	var tainted bool
	if os.Getenv("TAINTED") != "" {
		tainted = true
	}
	reader, err := local.NewLocalReader(&local.LocalReaderOpts{
		CacheDirectory: "smr",
		Tainted:        tainted,
	})
	if err != nil {
		panic(err)
	}

	server, err := imgfs.New(recipe, reader)
	if err != nil {
		panic(err)
	}

	// Mount the file system.
	if *fMountPoint == "" {
		log.Fatalf("You must set --mount_point.")
	}

	cfg := &fuse.MountConfig{
		ReadOnly: true,
	}

	if *fDebug {
		cfg.DebugLogger = log.New(os.Stderr, "fuse: ", 0)
	}

	mfs, err := fuse.Mount(*fMountPoint, server, cfg)
	if err != nil {
		log.Fatalf("Mount: %v", err)
	}

	// Wait for it to be unmounted.
	if err = mfs.Join(context.Background()); err != nil {
		log.Fatalf("Join: %v", err)
	}

}
