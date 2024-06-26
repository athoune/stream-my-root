package local

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/athoune/stream-my-root/pkg/trimmed"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/klauspost/compress/zstd"
)

type LocalReaderOpts struct {
	CacheDirectory string
	BlockSize      int
	Tainted        bool
	CacheSize      int
}

func (l *LocalReaderOpts) SetDefault() {
	if l.BlockSize == 0 {
		l.BlockSize = blocks.DEFAULT_BLOCK_SIZE
	}
	if l.CacheSize == 0 {
		l.CacheSize = 128
	}
}

// LocalReader has a folder full of zstd compressed blocks
type LocalReader struct {
	folder    string
	blockSize int
	cache     *lru.Cache[string, *trimmed.Trimmed]
	decoder   *zstd.Decoder
	tainter   *blocks.Tainter
}

func New(opts *LocalReaderOpts) (*LocalReader, error) {
	opts.SetDefault()
	_, err := os.Stat(opts.CacheDirectory)
	if err != nil {
		return nil, err
	}
	cache, err := lru.New[string, *trimmed.Trimmed](128)
	if err != nil {
		return nil, err
	}
	var decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	var tainter *blocks.Tainter
	if opts.Tainted {
		tainter = blocks.NewTainter()
	}
	return &LocalReader{
		folder:    opts.CacheDirectory,
		blockSize: opts.BlockSize,
		cache:     cache,
		decoder:   decoder,
		tainter:   tainter,
	}, nil
}

func NewLocalReader(opts *LocalReaderOpts) (*blocks.Reader, error) {
	l, err := New(opts)
	if err != nil {
		return nil, err
	}
	return &blocks.Reader{l}, nil
}

func (l *LocalReader) BlockSize() int {
	return l.blockSize
}

func (l *LocalReader) Name() string {
	return "LocalReader"
}

func (l *LocalReader) Contains(hash string) bool {
	ok := l.cache.Contains(hash)
	if ok {
		return true
	}
	_, err := os.Stat(fmt.Sprintf("%s/%s.zst", l.folder, hash))
	if err == nil {
		return true
	} else {
		if os.IsNotExist(err) {
			return false
		}
		slog.Error("Contains", err)
		panic(err)
	}
}

func (l *LocalReader) Folder() string {
	return l.folder
}

func (l *LocalReader) Get(hash string) (blocks.ReadableAt, error) {
	logger := slog.Default().With("hash", hash)
	block, ok := l.cache.Get(hash)
	if ok {
		logger.Debug("LocalReader.Get", "cached", true)
		if l.tainter != nil {
			l.tainter.Taint(hash)
		}
		return block, nil
	}
	fBlock, err := os.Open(fmt.Sprintf("%s/%s.zst", l.folder, hash))
	if err != nil {
		return nil, err
	}
	compressedBlock, err := io.ReadAll(fBlock)
	if err != nil {
		return nil, err
	}
	logger = logger.With("compressed_size", len(compressedBlock))
	rawBlock, err := l.decoder.DecodeAll(compressedBlock, nil)
	if err != nil {
		return nil, err
	}
	logger = logger.With("trimmed_size", len(rawBlock))
	block, err = trimmed.NewTrimmed(rawBlock, l.blockSize)
	if err != nil {
		return nil, err
	}
	if l.tainter != nil {
		l.tainter.Taint(hash)
	}
	l.cache.Add(hash, block)
	logger.Debug("LocalReader.Get")
	return block, nil
}
