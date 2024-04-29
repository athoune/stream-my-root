package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	temp, err := os.MkdirTemp("/tmp", "test_sock")
	assert.NoError(t, err)
	defer os.RemoveAll(temp)

	s := fmt.Sprintf("%s/test.sock", temp)

	s_listen, err := net.Listen("unix", s)
	assert.NoError(t, err)
	go func() {
		conn, err := s_listen.Accept()
		assert.NoError(t, err)
		server := NewServerSide(conn)
		m, arg, err := server.Query()
		assert.NoError(t, err)
		assert.Equal(t, Method(1), m)
		assert.Equal(t, []byte("World"), arg)
		err = server.Answer([]byte(fmt.Sprintf("Hello %s", string(arg))), nil)
		assert.NoError(t, err)
	}()
	conn, err := net.Dial("unix", s)
	assert.NoError(t, err)
	client := NewClientSide(conn)

	err = client.query(1, []byte("World"))
	assert.NoError(t, err)

	resp, err := client.answer()
	assert.NoError(t, err)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "Hello World", string(resp.Value))
}

func TestRouter(t *testing.T) {
	router := NewRouter()
	// rcp: Hello world
	router.Register(1, func(arg []byte) ([]byte, error) {
		return []byte(fmt.Sprintf("Hello %s", string(arg))), nil
	})
	// rpc: error
	router.Register(2, func(arg []byte) ([]byte, error) {
		return nil, errors.New("oups")
	})
	// rpc: late
	router.Register(3, func(arg []byte) ([]byte, error) {
		time.Sleep(3 * time.Second)
		return []byte("too late"), nil
	})
	temp, err := os.MkdirTemp("/tmp", "test_sock")
	assert.NoError(t, err)
	defer os.RemoveAll(temp)

	s := fmt.Sprintf("%s/test.sock", temp)

	listen, err := net.Listen("unix", s)
	assert.NoError(t, err)

	go func() {
		conn, err := listen.Accept()
		assert.NoError(t, err)
		router.Loop(NewServerSide(conn))
	}()

	conn, err := net.Dial("unix", s)
	assert.NoError(t, err)
	client := NewClientSide(conn)

	resp, err := client.Query(context.TODO(), 1, []byte("World"))
	assert.NoError(t, err)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "Hello World", string(resp.Value))

	resp, err = client.Query(context.TODO(), 2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "oups", resp.Error.Error())

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	resp, err = client.Query(ctx, 3, nil)
	assert.NotNil(t, err)
	assert.Nil(t, resp)

}

func TestServer(t *testing.T) {
	temp, err := os.MkdirTemp("/tmp", "test_sock")
	assert.NoError(t, err)
	defer os.RemoveAll(temp)

	s := fmt.Sprintf("%s/test.sock", temp)

	server := New(s)
	server.Register(1, func(arg []byte) ([]byte, error) {
		return []byte(fmt.Sprintf("Hello %s", string(arg))), nil
	})
	err = server.Listen()
	assert.NoError(t, err)
	go server.Serve()

	client, err := NewClient(fmt.Sprintf("unix://%s", s))
	assert.NoError(t, err)

	resp, err := client.Query(context.TODO(), 1, []byte("World"))
	assert.NoError(t, err)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "Hello World", string(resp.Value))
}
