package cached

import (
	"context"

	"github.com/athoune/stream-my-root/pkg/rpc"
)

type Client struct {
	client *rpc.Client
}

func NewClient(address string) (*Client, error) {
	c, err := rpc.NewClient(address)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: c,
	}, nil
}

func (c *Client) Lock(ctx context.Context, key string) (bool, error) {
	r, err := c.client.Query(ctx, Lock, []byte(key))
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

// Read
func (c *Client) Read(ctx context.Context, key string) (bool, error) {
	r, err := c.client.Query(ctx, Read, []byte(key))
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

func (c *Client) Write(ctx context.Context, key string, size uint32) error {
	arg := SetArg{
		Key:  key,
		Size: size,
	}
	raw, err := arg.MarshalBinary()
	if err != nil {
		return err
	}
	r, err := c.client.Query(ctx, Write, raw)
	if err != nil {
		return err
	}
	if r.Error != nil {
		return r.Error
	}
	return nil
}
