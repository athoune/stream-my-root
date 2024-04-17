package blocks

import (
	"io"
)

type Block struct {
	Seek   int
	Hash   string
	Reader io.Reader
}

const DEFAULT_BLOCK_SIZE = 512 * 1024
