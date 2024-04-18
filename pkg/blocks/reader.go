package blocks

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/athoune/stream-my-root/pkg/trimmed"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/klauspost/compress/zstd"
)

type ReadableAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

type Getable interface {
	Get(hash string) (ReadableAt, error)
	BlockSize() int
	Name() string
}

type Reader struct {
	Getable
}

// LocalReader has a folder full of zstd compressed blocks
type LocalReader struct {
	folder    string
	blockSize int
	cache     *lru.Cache[string, *trimmed.Trimmed]
	decoder   *zstd.Decoder
}

func NewLocalReader(folder string) (*Reader, error) {
	_, err := os.Stat(folder)
	if err != nil {
		return nil, err
	}
	cache, err := lru.New[string, *trimmed.Trimmed](128)
	if err != nil {
		return nil, err
	}
	var decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	return &Reader{&LocalReader{
		folder:    folder,
		blockSize: DEFAULT_BLOCK_SIZE,
		cache:     cache,
		decoder:   decoder,
	}}, nil
}

func (l *LocalReader) BlockSize() int {
	return l.blockSize
}

func (l *LocalReader) Name() string {
	return "LocalReader"
}

func (l *LocalReader) Get(hash string) (ReadableAt, error) {
	logger := slog.Default().With("hash", hash)
	block, ok := l.cache.Get(hash)
	if ok {
		logger.Debug("LocalReader.Get", "cached", true)
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
	l.cache.Add(hash, block)
	logger.Debug("LocalReader.Get")
	return block, nil
}

func Read(p []byte, off int64, recipe *Recipe, reader *Reader) (int, error) {
	slice, err := recipe.BlocksForReading(off, int64(len(p)))
	slog.Debug("Read", "length", len(p), "off", off, "#blocks", len(slice.blocks),
		"start", slice.start, "end", slice.end)
	if err != nil {
		return 0, err
	}
	return reader.read(p, slice)
}

func (l *Reader) read(p []byte, slice *Slice) (int, error) {
	logger := slog.Default().With(
		"name", fmt.Sprintf("%s.read", l.Name()),
		slog.Group("slice",
			"dest_length", len(p), "hashes", slice.Hashes(),
			"start", slice.start, "end", slice.end))
	logger.Debug("Reader.read")
	total := 0
	var src_start, src_end int64
	var chunk_size int
	var n int
	dest_start := 0
	for i, block := range slice.blocks {
		if i == 0 {
			src_start = slice.start
		} else {
			src_start = 0
		}
		if i == slice.Length()-1 { // last block
			src_end = slice.end
		} else {
			src_end = int64(l.BlockSize())
		}
		chunk_size = int(src_end - src_start)
		ll := logger.With(slog.Group("chunk",
			"i", i,
			"size", chunk_size,
			slog.Group("src", "start", src_start, "end", src_end),
			slog.Group("dest", "start", dest_start, "end", dest_start+chunk_size),
		))
		ll.Debug("Reading")
		if block == nil { // the chunk is full of zero
			n = chunk_size
		} else {
			ll = ll.With("hash", block.Hash)
			chunk, err := l.Get(block.Hash)
			if err != nil {
				ll.Error("can't get hash")
				return 0, err
			}
			n, err = chunk.ReadAt(p[dest_start:dest_start+chunk_size], src_start)
			if err != nil {
				ll.Error("ReadAT", "err", err)
				return 0, err
			}
			if n != chunk_size {
				err = fmt.Errorf("incomplete read %d of %d", n, chunk_size)
				ll.Error("ReadAT", "err", err)
				return 0, err // FIXME
			}
		}
		dest_start += chunk_size
		total += n
	}
	return total, nil
}
