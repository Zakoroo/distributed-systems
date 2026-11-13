package pubsubapi

import (
	"io"
	"sync"
)

type pollKey struct {
	topic string
	id    uint
}

// Keep the HTTP poll exclusive until its response body closes, not just until
// the service returns: another response must not overtake one still being written.
type pollBody struct {
	body    io.ReadCloser
	release func()
	once    sync.Once
	err     error
}

func (b *pollBody) Read(p []byte) (int, error) { return b.body.Read(p) }
func (b *pollBody) Close() error {
	b.once.Do(func() { b.err = b.body.Close(); b.release() })
	return b.err
}
