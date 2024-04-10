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
		values: Rtrim(v),
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

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func (s *Trimmed) ReadAt(p []byte, start int64) (n int, err error) {
	if start >= int64(s.size) {
		return 0, io.EOF
	}
	end := min(int64(s.size), start+int64(len(p)))
	short := int64(len(s.values))
	if start < short {
		n = copy(p, s.values[start:short])
	}

	if end > short {
		for i := short; i < end; i++ {
			p[i] = 0
		}
		n += int(end - short)
	}
	return
}
