package rpc

import (
	"bufio"
	"io"
)

type Method byte

type Stream struct {
	readWriter *bufio.ReadWriter
	stream     io.ReadWriteCloser
}

func NewStream(conn io.ReadWriteCloser) *Stream {
	return &Stream{
		readWriter: bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		stream:     conn,
	}
}

func (s *Stream) Close() error {
	return s.stream.Close()
}
