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

func (s *Trimmed) ReadAt(p []byte, start int64) (n int, err error) {
	if start >= int64(s.size) { // too far
		return 0, io.EOF
	}
	end := int64(len(p))
	if end > int64(s.size) {
		end = int64(s.size)
	}
	non_zero := int64(len(s.values))

	if start <= non_zero { // real values are available
		n = copy(p, s.values[start:non_zero])
	}

	if end <= non_zero { // enough values are wrote
		return
	}

	// tacit zeros
	for i := int64(n); i < end; i++ {
		if int(i) == len(p) {
			panic(fmt.Sprintln("array panic",
				"i", i, "start", start, "len(p)", len(p), "non_zero", non_zero, "end", end, "n", n))
		}
		p[i] = 0
	}
	n += int(end - non_zero)
	return
}
