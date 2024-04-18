package trimmed

import (
	"fmt"
	"io"
)

// Trimmed is an array of bytes without ending 0
type Trimmed struct {
	values []byte
	size   int
}

func New(v []byte) *Trimmed {
	return &Trimmed{
		values: Rtrim(v),
		size:   len(v),
	}
}

func NewTrimmed(v []byte, size int) (*Trimmed, error) {
	if size < len(v) {
		return nil, fmt.Errorf("array length must shorter than size : len(%s) < %d", v, size)
	}
	return &Trimmed{
		values: Rtrim(v), // FIXME don't trim data from the pool
		size:   size,
	}, nil
}

func Rtrim(buffer []byte) []byte {
	for i := len(buffer); i > 0; i-- {
		if buffer[i-1] != 0 {
			return buffer[:i]
		}
	}
	return []byte{}
}

func (s *Trimmed) ReadAt(p []byte, src_start int64) (n int, err error) {
	if src_start >= int64(s.size) { // too far
		return 0, io.EOF
	}
	size := s.size - int(src_start)
	if size > len(p) {
		size = len(p)
	}
	src_end := src_start + int64(size)
	non_zero := int64(len(s.values))

	if src_start < non_zero { // real values are available
		n = copy(p, s.values[src_start:non_zero])
	}

	if src_end <= non_zero { // enough values are wrote
		return
	}

	zero_start := int64(n)
	// tacit zeros
	for i := zero_start; i < int64(size); i++ {
		p[i] = 0
	}
	n += size - int(zero_start)
	return
}
