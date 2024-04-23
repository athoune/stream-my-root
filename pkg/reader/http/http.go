package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/athoune/stream-my-root/pkg/blocks"
	_local "github.com/athoune/stream-my-root/pkg/reader/local"
)

type HttpReader struct {
	local  _local.LocalReader
	client *http.Client
	url    string
}

func New(local *_local.LocalReader, storage string) (*HttpReader, error) {
	_, err := url.Parse(storage)
	if err != nil {
		return nil, err
	}
	return &HttpReader{
		local:  *local,
		client: &http.Client{},
		url:    storage,
	}, nil
}

func (h *HttpReader) Name() string {
	return "HttpReader"
}

func (l *HttpReader) Get(hash string) (blocks.ReadableAt, error) {
	if l.local.Contains(hash) {
		return l.Get(hash)
	}
	// FIXME is there conccurent Get ?
	resp, err := l.client.Get(fmt.Sprintf("%s/%s", l.url, hash))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// FIXME validate the SHA256
	err = os.WriteFile(fmt.Sprintf("%s/%s.zst", l.local.Folder(), hash), raw, 0660)
	if err != nil {
		return nil, err
	}
	return l.Get(hash)
}
