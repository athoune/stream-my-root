package cached

import (
	"encoding/binary"
	"net"

	"github.com/athoune/stream-my-root/pkg/rpc"
)

type Client struct {
	client *rpc.Client
}

func NewClient(conn net.Conn) (*Client, error) {
	c, err := rpc.NewClient(conn)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: c,
	}, nil
}

func (c *Client) Lock(key string) error {
	r, err := c.client.Query(Lock, []byte(key))
	if err != nil {
		return err
	}
	if r.Error != nil {
		return r.Error
	}
	return nil
}

func (c *Client) Get(key string) (bool, error) {
	r, err := c.client.Query(Get, []byte(key))
	if err != nil {
		return false, err
	}
	if r.Error != nil {
		return false, r.Error
	}
	if r.Value[0] == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func setArg(key string, size uint32) []byte {
	bkey := []byte(key)
	arg := make([]byte, len(bkey)+4)
	binary.BigEndian.PutUint32(arg[0:4], size)
	copy(arg[4:], bkey)
	return arg
}

func (c *Client) Set(key string, size uint32) error {
	r, err := c.client.Query(Set, setArg(key, size))
	if err != nil {
		return err
	}
	if r.Error != nil {
		return r.Error
	}
	return nil
}
