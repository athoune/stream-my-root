package blocks

import (
	"fmt"
	"io"
	"os"

	"github.com/athoune/stream-my-root/pkg/trimmed"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/klauspost/compress/zstd"
)

type ReadableAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

// LocalReader has a folder full of zstd compressed blocks
type LocalReader struct {
	folder    string
	blockSize int
	cache     *lru.Cache[string, *trimmed.Trimmed]
	decoder   *zstd.Decoder
}

func NewLocalReader(folder string) (*LocalReader, error) {
	_, err := os.Stat(folder)
	if err != nil {
		return nil, err
	}
	cache, err := lru.New[string, *trimmed.Trimmed](128)
	if err != nil {
		return nil, err
	}
	var decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	return &LocalReader{
		folder:    folder,
		blockSize: DEFAULT_BLOCK_SIZE,
		cache:     cache,
		decoder:   decoder,
	}, nil
}

func (l *LocalReader) Get(hash string) (ReadableAt, error) {
	block, ok := l.cache.Get(hash)
	if ok {
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
	rawBlock, err := l.decoder.DecodeAll(compressedBlock, nil)
	if err != nil {
		return nil, err
	}
	block, err = trimmed.NewTrimmed(rawBlock, l.blockSize)
	if err != nil {
		return nil, err
	}
	l.cache.Add(hash, block)
	return block, nil
}
