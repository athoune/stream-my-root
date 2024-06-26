package rpc

import (
	"encoding/binary"
	"io"
	"log/slog"
	"net"

	"github.com/hashicorp/yamux"
)

type Server struct {
	socket   string
	logger   *slog.Logger
	router   *Router
	listener net.Listener
}

func New(socket string) *Server {
	return &Server{
		socket: socket,
		logger: slog.Default().With("socket", socket),
		router: NewRouter(),
	}
}

func (s *Server) Register(method Method, handler Handler) {
	s.router.Register(method, handler)
}

func (s *Server) Listen() error {
	var err error
	s.listener, err = net.Listen("unix", s.socket)
	slog.Default().Info("RPC Server Listen", "socket", s.socket)
	return err
}

func (s *Server) Serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error("Server", "err", err)
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	logger := s.logger.With("remote", conn.RemoteAddr())
	session, err := yamux.Server(conn, nil)
	if err != nil {
		logger.Error("handle", "err", err)
		conn.Close()
		return
	}
	logger.Info("The server has a new connection")
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			logger.Error("handle", "err", err)
			continue
		}
		logger.With("stream id", stream.StreamID()).Debug("New stream")
		go s.router.Loop(NewServerSide(stream))
	}
}

type ServerSide struct {
	*Stream
}

func NewServerSide(conn net.Conn) *ServerSide {
	return &ServerSide{
		NewStream(conn),
	}
}

func (r *ServerSide) Query() (Method, []byte, error) {
	var method Method
	err := binary.Read(r.readWriter, binary.BigEndian, &method)
	if err != nil {
		return 0, nil, err
	}
	var size uint32
	err = binary.Read(r.readWriter, binary.BigEndian, &size)
	if err != nil {
		return 0, nil, err
	}
	if size == 0 {
		return method, nil, nil
	}
	buff := make([]byte, size)
	_, err = io.ReadFull(r.readWriter, buff)
	if err != nil {
		return 0, nil, err
	}
	return method, buff, nil
}

func (s *ServerSide) Answer(resp []byte, a_err error) error {
	if a_err != nil {
		msg := a_err.Error()
		err := binary.Write(s.readWriter, binary.BigEndian, uint32(len(msg)))
		if err != nil {
			return err
		}
		_, err = s.readWriter.Write([]byte(msg))
		if err != nil {
			return err
		}
		return s.readWriter.Flush()
	}
	err := binary.Write(s.readWriter, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}
	if resp == nil {
		err := binary.Write(s.readWriter, binary.BigEndian, uint32(0))
		if err != nil {
			return err
		}
	} else {
		err = binary.Write(s.readWriter, binary.BigEndian, uint32(len(resp)))
		if err != nil {
			return err
		}
		_, err = s.readWriter.Write(resp)
		if err != nil {
			return err
		}
	}
	return s.readWriter.Flush()
}
