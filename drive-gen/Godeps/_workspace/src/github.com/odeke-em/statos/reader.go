package statos

import (
	"io"
	"syscall"
)

// ReaderStatos implements the Read() interface
type ReaderStatos struct {
	iterator   io.Reader
	commChan   chan int
	commClosed bool
}

func NewReader(rd io.Reader) *ReaderStatos {
	return &ReaderStatos{
		iterator:   rd,
		commChan:   make(chan int),
		commClosed: false,
	}
}

func (r *ReaderStatos) Read(p []byte) (n int, err error) {
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

func (r *ReaderStatos) ProgressChan() chan int {
	return r.commChan
}
