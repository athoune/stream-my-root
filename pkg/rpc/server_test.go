package rpc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryAnswer(t *testing.T) {
	up := &bytes.Buffer{}
	down := &bytes.Buffer{}
	s := &ServerSide{&Stream{}}
	s.readWriter = bufio.NewReadWriter(bufio.NewReader(up), bufio.NewWriter(down))
	up.Write([]byte{2, 0, 0, 0, 0})
	method, arg, err := s.Query()
	assert.NoError(t, err)
	assert.Equal(t, Method(2), method)
	assert.Nil(t, arg)

	up.Write([]byte{3, 0, 0, 0, 4})
	up.Write([]byte("popo"))
	method, arg, err = s.Query()
	assert.NoError(t, err)
	assert.Equal(t, Method(3), method)
	assert.Equal(t, "popo", string(arg))

	err = s.Answer(nil, errors.New("hop"))
	assert.NoError(t, err)
	fmt.Println(down.Bytes())
	resp := make([]byte, 7)
	n, err := down.Read(resp)
	assert.NoError(t, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, []byte{0, 0, 0, 3}, resp[:4])
	assert.Equal(t, []byte("hop"), resp[4:])

	err = s.Answer([]byte("plop"), nil)
	assert.NoError(t, err)
	fmt.Println(down.Bytes())
	resp = make([]byte, 12)
	n, err = down.Read(resp)
	assert.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, []byte{0, 0, 0, 0}, resp[:4])
	assert.Equal(t, []byte{0, 0, 0, 4}, resp[4:8])
	assert.Equal(t, []byte("plop"), resp[8:])
}
