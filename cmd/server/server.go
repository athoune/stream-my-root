package main

import (
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/athoune/stream-my-root/pkg/backend/ro"
	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/lmittmann/tint"
	"github.com/pojntfx/go-nbd/pkg/server"
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
	var tainted bool
	if os.Getenv("TAINTED") != "" {
		tainted = true
	}
	reader, err := blocks.NewLocalReader("smr", tainted)
	if err != nil {
		panic(err)
	}
	backend := ro.NewROBackend(recipe, reader)
	l, err := net.Listen("tcp", ":10809")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	slog.Info("Listen 0.0.0.0:10809")

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		slog.Info("Connection", "client", conn.RemoteAddr())

		go func(c net.Addr) {
			if err := server.Handle(
				conn,
				[]*server.Export{
					{
						Name:        "smr",
						Description: "YOLO",
						Backend:     backend,
					},
				},
				&server.Options{
					ReadOnly:           true,
					MinimumBlockSize:   1,
					PreferredBlockSize: 4096,
					MaximumBlockSize:   4096,
					SupportsMultiConn:  true,
				}); err != nil {
				slog.Error("Handle", "client", c, "Error", err)
				return
			}
		}(conn.RemoteAddr())
	}
}
