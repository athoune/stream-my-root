package nbd

import (
	"cmp"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"regexp"

	"github.com/athoune/stream-my-root/pkg/backend/ro"
	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/pojntfx/go-nbd/pkg/server"
)

var RE_NAME regexp.Regexp

func init() {
	RE_NAME = *regexp.MustCompile("[a-zA-Z0-9]+")
}

func Serve(recipe *blocks.Recipe, reader *blocks.Reader, listen string) error {
	u, err := url.Parse(listen)
	if err != nil {
		return err
	}
	name := cmp.Or[string](u.Path, "smr")
	if !RE_NAME.MatchString(name) {
		return fmt.Errorf("incorrect name : %s", name)
	}

	backend := ro.NewROBackend(recipe, reader)
	l, err := net.Listen(u.Scheme, u.Host)
	if err != nil {
		return err
	}
	slog.Default().Info("Listen", "address", fmt.Sprintf("%s://%s", u.Scheme, u.Host), "name", name)

	defer l.Close()
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
						Name:        name,
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
