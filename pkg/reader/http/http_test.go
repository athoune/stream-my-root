package http

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttp(t *testing.T) {

	server := httptest.NewServer(http.FileServer(http.Dir("../../../fixtures/chunks")))
	defer server.Close()

	tmp, err := os.MkdirTemp("/tmp", "reader-http")
	assert.NoError(t, err)
	slog.Info("path", "tmp", tmp)
	defer os.RemoveAll(tmp)

	h, err := New(&HttpReaderOpts{
		CacheDirectory: tmp,
		SourceUrl:      server.URL,
	})
	assert.NoError(t, err)

	_, err = h.Get("nope")
	assert.Error(t, err)
	r, err := h.Get("03e62706ea71be374789eb985ea8260825ba707c79cfc3b0434d8632cb53eabc")
	assert.NoError(t, err)
	buffer := make([]byte, 10)
	n, err := r.ReadAt(buffer, 0)
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.True(t, h.local.Contains("03e62706ea71be374789eb985ea8260825ba707c79cfc3b0434d8632cb53eabc"))
}
