package cached

import (
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

func (c *Client) Lock(key string) (bool, error) {
	r, err := c.client.Query(Lock, []byte(key))
	if err != nil {
		return false, err
	}
	if r.Error != nil {
		return false, r.Error
	}
	ok := Bool(false)
	err = ok.UnmarshalBinary(r.Value)
	return bool(ok), err
}

func (c *Client) Get(key string) (bool, error) {
	r, err := c.client.Query(Get, []byte(key))
	if err != nil {
		return false, err
	}
	if r.Error != nil {
		return false, r.Error
	}
	ok := Bool(false)
	err = ok.UnmarshalBinary(r.Value)
	return bool(ok), err
}

func (c *Client) Set(key string, size uint32) error {
	arg := SetArg{
		Key:  key,
		Size: size,
	}
	raw, err := arg.MarshalBinary()
	if err != nil {
		return err
	}
	r, err := c.client.Query(Set, raw)
	if err != nil {
		return err
	}
	if r.Error != nil {
		return r.Error
	}
	return nil
}
