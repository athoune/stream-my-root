package zero

import "io"

type Zero struct {
	size int64
}

func NewZero(size int64) *Zero {
	return &Zero{
		size: size,
	}
}

func (z *Zero) ReadAt(p []byte, off int64) (n int, err error) {
	if off > z.size {
		return 0, io.EOF
	}
	end := z.size - off
	if end > int64(len(p)) {
		end = int64(len(p))
	}
	for i := off; i < end; i++ {
		p[i-off] = 0
	}
	return int(end), nil
}
