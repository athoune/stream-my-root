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

// LocalReader has a folder full of zstd compressed blocks
type LocalReader struct {
	folder    string
	blockSize int
	cache     *lru.Cache[string, *trimmed.Trimmed]
	decoder   *zstd.Decoder
	tainter   *blocks.Tainter
}

func NewLocalReader(folder string, tainted bool) (*blocks.Reader, error) {
	_, err := os.Stat(folder)
	if err != nil {
		return nil, err
	}
	cache, err := lru.New[string, *trimmed.Trimmed](128)
	if err != nil {
		return nil, err
	}
	var decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	var tainter *blocks.Tainter
	if tainted {
		tainter = blocks.NewTainter()
	}
	return &blocks.Reader{&LocalReader{
		folder:    folder,
		blockSize: blocks.DEFAULT_BLOCK_SIZE,
		cache:     cache,
		decoder:   decoder,
		tainter:   tainter,
	}}, nil
}

func (l *LocalReader) BlockSize() int {
	return l.blockSize
}

func (l *LocalReader) Name() string {
	return "LocalReader"
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
