package cached

import (
	"fmt"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/athoune/stream-my-root/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	temp, err := os.MkdirTemp("/tmp", "test_sock")
	assert.NoError(t, err)
	defer os.RemoveAll(temp)

	s := fmt.Sprintf("%s/test.sock", temp)

	server := rpc.New(s)

	cache, err := NewCached(nil)
	assert.NoError(t, err)
	cache.RegisterAll(server)
	err = server.Listen()
	assert.NoError(t, err)
	go server.Serve()

	conn, err := net.Dial("unix", s)
	assert.NoError(t, err)
	client, err := NewClient(conn)
	assert.NoError(t, err)

	ok, err := client.Get("plop")
	assert.NoError(t, err)
	assert.False(t, ok)

	waiting := &sync.WaitGroup{}
	waiting.Add(10)
	done := &sync.WaitGroup{}
	done.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			waiting.Done()
			client.Lock("plop")
			done.Done()
		}()
	}
	waiting.Wait()
	err = client.Set("plop", 512)
	assert.NoError(t, err)

}
