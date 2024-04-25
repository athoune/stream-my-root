package rpc

import (
	"fmt"
	"net"
	"os"
	"testing"

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
	router.Register(1, func(arg []byte) ([]byte, error) {
		return []byte(fmt.Sprintf("Hello %s", string(arg))), nil
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

	resp, err := client.Query(1, []byte("World"))
	assert.NoError(t, err)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "Hello World", string(resp.Value))
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

	conn, err := net.Dial("unix", s)
	assert.NoError(t, err)

	client, err := NewClient(conn)
	assert.NoError(t, err)

	resp, err := client.Query(1, []byte("World"))
	assert.NoError(t, err)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "Hello World", string(resp.Value))
}
