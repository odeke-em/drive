package statos

import (
	"io"
	"syscall"
)

// WriteCloserStatos implements the Write() interface
type WriteCloserStatos struct {
	commChan   chan int
	commClosed bool
	iterator   io.WriteCloser
}

func NewWriteCloser(w io.WriteCloser) *WriteCloserStatos {
	return &WriteCloserStatos{
		commChan:   make(chan int),
		iterator:   w,
		commClosed: false,
	}
}

func (w *WriteCloserStatos) Write(p []byte) (n int, err error) {
	n, err = w.iterator.Write(p)

	if err != nil && err != syscall.EINTR {
		if !w.commClosed {
			close(w.commChan)
			w.commClosed = true
		}
	} else if n >= 0 {
		w.commChan <- n
	}
	return
}

func (w *WriteCloserStatos) ProgressChan() chan int {
	return w.commChan
}

func (w *WriteCloserStatos) Close() error {
	err := w.iterator.Close()
	if err == nil {
		if !w.commClosed {
			close(w.commChan)
			w.commClosed = true
		}
	}
	return err
}
