package main

import (
	"cmp"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/athoune/stream-my-root/pkg/cached"
	"github.com/athoune/stream-my-root/pkg/rpc"
)

func main() {
	opts := cached.DefaultCachedOpts()
	r_max := os.Getenv("MAX")
	if r_max != "" {
		max, err := strconv.ParseUint(r_max, 10, 64)
		if err != nil {
			panic(err)
		}
		opts.Max = uint(max)
	}
	r_correlated_time := os.Getenv("CORRELATED_TIME")
	if r_correlated_time != "" {
		correlated_time, err := time.ParseDuration(r_correlated_time)
		if err != nil {
			panic(err)
		}
		opts.CorrelatedTime = correlated_time
	}
	slog.Default().Info("Cached", "opts", opts)
	c := cached.NewCached(opts)
	socket := cmp.Or[string](os.Getenv("SOCKET"), "/tmp/smr.sock")
	server := rpc.New(socket)
	c.RegisterAll(server)
	server.Listen()
	server.Serve()
}
