package ro

import (
	"errors"
	"log/slog"

	_blocks "github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/pojntfx/go-nbd/pkg/backend"
)

/*
type Backend interface {
	io.ReaderAt
	io.WriterAt

	Size() (int64, error)
	Sync() error
}
*/

var _ backend.Backend = (*ROBackend)(nil)

type ROBackend struct {
	recipe *_blocks.Recipe // the recipe of a disk image
	reader *_blocks.Reader
}

func NewROBackend(recipe *_blocks.Recipe, reader *_blocks.Reader) *ROBackend {
	return &ROBackend{
		recipe: recipe,
		reader: reader,
	}
}

func (r *ROBackend) ReadAt(p []byte, off int64) (n int, err error) {
	logger := slog.Default().With("length", len(p), "off", off)
	logger = logger.With(slog.Group("blocks", "offset", off, "#blocks", r.recipe.NumberOfBlocks()))
	n, err = _blocks.Read(p, off, r.recipe, r.reader)
	if err != nil {
		logger.Error("Can't read", "n", n, "error", err)
		return n, err
	}
	logger.Debug("ReadAt", "n", n)
	return
}

const IMAGE_SIZE = 1024 * 1024 * 1024

func (r *ROBackend) Size() (int64, error) {
	// FIXME it's hardcoded to 1GiB
	slog.Info("Size", "size", IMAGE_SIZE)
	return IMAGE_SIZE, nil
}

func (r *ROBackend) Sync() error {
	// This backend is read only, no sync here
	slog.Info("Sync")
	return nil
}

func (r *ROBackend) WriteAt([]byte, int64) (int, error) {
	return 0, errors.New("read only")
}
