package blocks

import (
	"fmt"
	"log/slog"
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
