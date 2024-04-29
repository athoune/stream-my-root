package http

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/athoune/stream-my-root/pkg/cached"
	_local "github.com/athoune/stream-my-root/pkg/reader/local"
)

type HttpReaderOpts struct {
	SourceUrl      string
	CacheDirectory string
	CacheSize      uint
	CachedUrl      string
	BlockSize      int
}

func (h *HttpReaderOpts) SetDefault() {
	if h.BlockSize == 0 {
		h.BlockSize = blocks.DEFAULT_BLOCK_SIZE
	}
	if h.CacheSize == 0 {
		h.CacheSize = 128
	}
}

type HttpReader struct {
	local        *_local.LocalReader
	client       *http.Client
	cachedClient *cached.Client
	url          string
	blockSize    int
}

func New(opts *HttpReaderOpts) (*HttpReader, error) {
	opts.SetDefault()
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
	cachedClient, err := cached.NewClient(opts.CachedUrl)
	if err != nil {
		return nil, err
	}
	return &HttpReader{
		local:        local,
		client:       &http.Client{},
		cachedClient: cachedClient,
		url:          opts.SourceUrl,
		blockSize:    opts.BlockSize,
	}, nil
}

func NewHttpReader(opts *HttpReaderOpts) (*blocks.Reader, error) {
	h, err := New(opts)
	if err != nil {
		return nil, err
	}
	return &blocks.Reader{h}, nil
}

func (h *HttpReader) BlockSize() int {
	return h.blockSize
}

func (h *HttpReader) Name() string {
	return "HttpReader"
}

func (l *HttpReader) Get(hash string) (blocks.ReadableAt, error) {
	logger := slog.Default().With("hash", hash)
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	_, err := l.cachedClient.Read(ctx, hash) // eviction is not handled here
	cancel()
	if err != nil {
		return nil, err
	}
	if l.local.Contains(hash) {
		logger.Debug("Cached")
		return l.Get(hash)
	}
	ctx, cancel = context.WithTimeout(context.TODO(), 2*time.Second)
	ok, err := l.cachedClient.Lock(ctx, hash)
	cancel()
	if err != nil {
		return nil, err
	}
	if ok { // I'm the first one, and get the lock, I must download the file, for the rest of us
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
	}
	return l.local.Get(hash)
}
