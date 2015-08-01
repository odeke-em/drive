package statos

import (
	"io"
	"syscall"
)

// ReadCloserStatos implements the Read() interface
type ReadCloserStatos struct {
	commChan   chan int
	commClosed bool
	iterator   io.ReadCloser
}

func NewReadCloser(rd io.ReadCloser) *ReadCloserStatos {
	return &ReadCloserStatos{
		commChan:   make(chan int),
		commClosed: false,
		iterator:   rd,
	}
}

func (r *ReadCloserStatos) Read(p []byte) (n int, err error) {
	n, err = r.iterator.Read(p)

	if err != nil && err != syscall.EINTR {
		if !r.commClosed {
			close(r.commChan)
			r.commClosed = true
		}
	} else if n >= 0 {
		r.commChan <- n
	}
	return
}

func (r *ReadCloserStatos) ProgressChan() chan int {
	return r.commChan
}

func (r *ReadCloserStatos) Close() error {
	err := r.iterator.Close()
	if err == nil && !r.commClosed {
		close(r.commChan)
		r.commClosed = true
	}
	return err
}
