package rpc

import (
	"fmt"
	"io"
	"log/slog"
)

type Handler func([]byte) ([]byte, error)
type Router struct {
	methods map[Method]Handler
}

func NewRouter() *Router {
	return &Router{
		methods: make(map[Method]Handler),
	}
}

func (r *Router) Register(method Method, handler Handler) {
	r.methods[method] = handler
}

func (s *Router) do(stream *ServerSide) error {
	logger := slog.Default()
	method, arg, err := stream.Query()
	if err != nil {
		stream.Close()
		logger.Error("Do", "err", err)
		return err
	}
	logger = logger.With("method", method)
	m, ok := s.methods[method]
	if !ok {
		err = stream.Answer(nil, fmt.Errorf("unknown method %d", method))
		if err != nil {
			logger.Error("Do", "err", err)
			return err
		}
		logger.Info("Unknown method")
		return nil
	}
	err = stream.Answer(m(arg))
	if err != nil {
		stream.Close()
		logger.Error("Do", "err", err)
		return err
	}
	return nil
}

func (s *Router) Loop(stream *ServerSide) error {
	for {
		err := s.do(stream)
		if err != nil {
			if err == io.EOF { // end of stream
				return nil
			}
			return err
		}
	}

}
