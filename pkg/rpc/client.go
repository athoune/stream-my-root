package rpc

import (
	"cmp"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	_url "net/url"
	"time"

	"github.com/hashicorp/yamux"
)

type Response struct {
	Value []byte
	Error error
}

type ClientSide struct {
	*Stream
}

func NewClientSide(conn net.Conn) *ClientSide {
	return &ClientSide{NewStream(conn)}
}

func (c *ClientSide) query(method Method, arg []byte) error {
	err := c.readWriter.WriteByte(byte(method))
	if err != nil {
		return err
	}
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(len(arg)))
	_, err = c.readWriter.Write(buff)
	if err != nil {
		return err
	}
	n, err := c.readWriter.Write(arg)
	if err != nil {
		return err
	}
	if n != len(arg) {
		return fmt.Errorf("incomplete Write %d of %d", n, len(arg))
	}
	return c.readWriter.Flush()
}

func (c *ClientSide) answer() (*Response, error) {
	var size uint32
	err := binary.Read(c.readWriter, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}
	var buff []byte
	if size != 0 { // it's an error
		buff = make([]byte, size)
		_, err = io.ReadFull(c.readWriter, buff)
		if err != nil {
			return nil, err
		}
		return &Response{
			Error: errors.New(string(buff)),
		}, nil
	}
	err = binary.Read(c.readWriter, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}
	if size == 0 { // empty response
		return &Response{
			Value: nil,
		}, nil
	}
	buff = make([]byte, size)
	_, err = io.ReadFull(c.readWriter, buff)
	if err != nil {
		return nil, err
	}
	return &Response{
		Value: buff,
	}, nil
}

func (c *ClientSide) Query(ctx context.Context, method Method, arg []byte) (*Response, error) {
	done := make(chan interface{})
	var resp *Response
	var err error
	go func() {
		err = c.query(method, arg)
		if err != nil {
			resp = nil
			done <- nil
		}
		resp, err = c.answer()
		done <- nil
	}()
	select {
	case <-done:
		return resp, err
	case <-ctx.Done():
		return nil, errors.New("Timeout")
	}

}

// Client handles the raw network connection
type Client struct {
	network string
	address string
	session *yamux.Session
}

func NewClient(address string) (*Client, error) {
	url, err := _url.Parse(address)
	if err != nil {
		return nil, err
	}

	return &Client{
		network: url.Scheme,
		address: cmp.Or[string](url.Host, url.Path),
	}, nil
}

// getYamuxSession lazily opens a connection, and return a multiplexed channel
func (c *Client) getYamuxSession() (*yamux.Session, error) {
	if c.session == nil {
		var err error
		var conn net.Conn
		chrono := time.Now()
		logger := slog.Default().With("scheme", c.network, "host", c.address)
		// The client retry to connect for 3 seconds
		for range 10 {
			conn, err = net.Dial(c.network, c.address)
			if err != nil {
				time.Sleep(300 * time.Millisecond)
			} else {
				break
			}
		}
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
		logger = logger.With("connection duration", time.Since(chrono))
		session, err := yamux.Client(conn, nil)
		if err != nil {
			return nil, err
		}
		ping, err := session.Ping()
		if err != nil {
			return nil, err
		}
		logger.With("ping", ping).Info("New connection")
		c.session = session
	}

	return c.session, nil
}

func (c *Client) Close() error {
	return c.session.Close()
}

func (c *Client) Ping() (time.Duration, error) {
	session, err := c.Session()
	if err != nil {
		return 0, err
	}
	return session.stream.Session().Ping()
}

type Session struct {
	// a yamux stream
	stream *yamux.Stream
}

// Session returns a Stream from the multiplexed connection
func (c *Client) Session() (*Session, error) {
	y_session, err := c.getYamuxSession()
	if err != nil {
		return nil, err
	}
	stream, err := y_session.OpenStream()
	if err != nil {
		return nil, err
	}
	return &Session{
		stream: stream,
	}, nil
}

// Close the opened yamux stream
func (s *Session) Close() error {
	return s.stream.Close()
}

func (s *Session) Query(ctx context.Context, method Method, arg []byte) (*Response, error) {
	logger := slog.Default().With("stream id", s.stream.StreamID(),
		"local", s.stream.LocalAddr().String(),
		"remote", s.stream.RemoteAddr().String())
	ctx = context.WithValue(ctx, ClientContextKey("logger"), logger)

	resp, err := NewClientSide(s.stream).Query(ctx, method, arg)

	logger = logger.With("method", method, "arg", arg)
	if err != nil {
		logger.Error(err.Error())
	} else {
		logger.With("resp", resp).Debug("Query")
	}

	return resp, err
}

type ClientContextKey string

func (c *Client) Query(ctx context.Context, method Method, arg []byte) (*Response, error) {
	session, err := c.Session()
	if err != nil {
		return nil, err
	}
	return session.Query(ctx, method, arg)
}
