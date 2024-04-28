package http

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/athoune/stream-my-root/pkg/blocks"
	_local "github.com/athoune/stream-my-root/pkg/reader/local"
)

type HttpReaderOpts struct {
	SourceUrl      string
	CacheDirectory string
	CacheSize      uint
}

type HttpReader struct {
	local  *_local.LocalReader
	client *http.Client
	url    string
}

func New(opts *HttpReaderOpts) (*HttpReader, error) {
	_, err := url.Parse(opts.SourceUrl)
	if err != nil {
		return nil, err
	}
	local, err := _local.New(&_local.LocalReaderOpts{
		CacheDirectory: opts.CacheDirectory,
	})
	if err != nil {
		return nil, err
	}
	return &HttpReader{
		local:  local,
		client: &http.Client{},
		url:    opts.SourceUrl,
	}, nil
}

func (h *HttpReader) Name() string {
	return "HttpReader"
}

func (l *HttpReader) Get(hash string) (blocks.ReadableAt, error) {
	logger := slog.Default().With("hash", hash)
	if l.local.Contains(hash) {
		logger.Debug("Cached")
		return l.Get(hash)
	}
	// FIXME is there concurrent Get ?
	resp, err := l.client.Get(fmt.Sprintf("%s/%s.zst", l.url, hash))
	if err != nil {
		return nil, err
	}
	logger = logger.With("status", resp.StatusCode)
	if resp.StatusCode >= 400 {
		logger.Error("")
		return nil, fmt.Errorf("bad status %s", resp.Status)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("can't read response", "err", err)
		return nil, err
	}
	// FIXME validate the SHA256
	err = os.WriteFile(fmt.Sprintf("%s/%s.zst", l.local.Folder(), hash), raw, 0660)
	if err != nil {
		logger.Error("can't write cache", "err", err)
		return nil, err
	}
	logger.Debug("Get")
	return l.local.Get(hash)
}
