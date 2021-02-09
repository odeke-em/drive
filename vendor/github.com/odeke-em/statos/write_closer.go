package statos

import (
	"io"
	"sync"
)

// WriteCloserStatos implements io.WriteCloser
var _ io.WriteCloser = &WriteCloserStatos{}

type WriteCloserStatos struct {
	*WriterStatos

	c io.Closer

	closerOnce sync.Once
}

func NewWriteCloser(wc io.WriteCloser) *WriteCloserStatos {
	return &WriteCloserStatos{NewWriter(wc), wc, sync.Once{}}
}

func (wcs *WriteCloserStatos) Close() error {
	var err error
	wcs.closerOnce.Do(func() {
		err = wcs.c.Close()
		wcs.closeCommChan()
	})
	return err
}
