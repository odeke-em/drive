package statos

import (
	"io"
	"sync"
	"syscall"
)

// WriterStatos implements io.Writer
var _ io.Writer = &WriterStatos{}

type WriterStatos struct {
	sync.RWMutex
	iterator   io.Writer
	commChan   chan int
	commClosed bool

	commOnce sync.Once
}

func (w *WriterStatos) closeCommChan() bool {
	alreadyClosed := w.wasCommClosed()
	if !alreadyClosed {
		w.commOnce.Do(func() {
			w.Lock()
			defer w.Unlock()

			close(w.commChan)
			w.commClosed = true
			alreadyClosed = w.commClosed
		})
	}
	return w.commClosed
}

func NewWriter(w io.Writer) *WriterStatos {
	return &WriterStatos{
		commChan:   make(chan int),
		iterator:   w,
		commClosed: false,
	}
}

func (w *WriterStatos) Write(p []byte) (n int, err error) {
	n, err = w.iterator.Write(p)

	if err != nil && err != syscall.EINTR {
		w.closeCommChan()
		return
	}

	if n >= 0 && !w.wasCommClosed() {
		w.commChan <- n
	}
	return
}

func (ws *WriterStatos) wasCommClosed() bool {
	ws.RLock()
	defer ws.RUnlock()

	return ws.commClosed
}

func (w *WriterStatos) ProgressChan() chan int {
	return w.commChan
}
