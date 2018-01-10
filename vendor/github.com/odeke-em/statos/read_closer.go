package statos

import (
	"io"
	"sync"
)

// ReadCloserStatos implements io.ReadCloser
var _ io.ReadCloser = &ReadCloserStatos{}

type ReadCloserStatos struct {
	*ReaderStatos

	c io.Closer

	closerOnce sync.Once
}

func NewReadCloser(rc io.ReadCloser) *ReadCloserStatos {
	return &ReadCloserStatos{NewReader(rc), rc, sync.Once{}}
}

func (rcs *ReadCloserStatos) Close() error {
	var err error
	rcs.closerOnce.Do(func() {
		err = rcs.c.Close()
		rcs.closeCommChan()
	})
	return err
}
