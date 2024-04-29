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

func (c *ClientSide) Query(method Method, arg []byte) (*Response, error) {
	err := c.query(method, arg)
	if err != nil {
		return nil, err
	}
	return c.answer()
}

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

func (c *Client) getSession() (*yamux.Session, error) {
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
		logger.Debug("Open connection", "duration", time.Since(chrono))
		return yamux.Client(conn, nil)
	}

	return c.session, nil
}

func (c *Client) Close() error {
	return c.session.Close()
}

func (c *Client) Session() (*Session, error) {
	y_session, err := c.getSession()
	if err != nil {
		return nil, err
	}
	conn, err := y_session.Open()
	if err != nil {
		return nil, err
	}
	return &Session{
		conn: conn,
	}, nil
}

type Session struct {
	conn net.Conn
}

func (s *Session) Close() error {
	return s.conn.Close()
}

func (s *Session) Query(method Method, arg []byte) (*Response, error) {
	return NewClientSide(s.conn).Query(method, arg)
}

func (c *Client) Query(method Method, arg []byte) (*Response, error) {
	session, err := c.Session()
	if err != nil {
		return nil, err
	}
	return session.Query(method, arg)
}
