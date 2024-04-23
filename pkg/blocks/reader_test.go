package blocks

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DummyHash []byte

func (d DummyHash) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(d)) {
		return 0, io.EOF
	}
	return copy(p, d[off:]), nil
}

type DummyReader map[string]*DummyHash

func (d DummyReader) Get(hash string) (ReadableAt, error) {
	h, ok := d[hash]
	if !ok {
		return nil, fmt.Errorf("Bad key %s", hash)
	}
	return h, nil
}

func (d *DummyReader) BlockSize() int {
	return 8
}

func (d *DummyReader) Name() string {
	return "Dummy"
}

func TestReader(t *testing.T) {
	buffer := make([]byte, 8)
	n, err := (&DummyHash{1, 2, 3, 4, 0, 0, 0, 0}).ReadAt(buffer, 0)
	assert.NoError(t, err)
	assert.Equal(t, 8, n)

	reader := &Reader{
		&DummyReader{
			"aaaa": &DummyHash{1, 2, 3, 4, 0, 0, 0, 0},
		},
	}
	buffer = make([]byte, 8)
	slice := &Slice{
		blocks: []*Block{
			{
				Hash: "aaaa",
			},
		},
		start: 0,
		end:   8,
	}
	n, err = reader.read(buffer, slice)
	assert.NoError(t, err)
	assert.Equal(t, 8, n)
}
