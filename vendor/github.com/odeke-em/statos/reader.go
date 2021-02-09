package statos

import (
	"io"
	"sync"
	"syscall"
)

// ReaderStatos implements io.Reader
var _ io.Reader = &ReaderStatos{}

type ReaderStatos struct {
	sync.RWMutex
	iterator   io.Reader
	commChan   chan int
	commClosed bool

	commOnce sync.Once
}

func (rs *ReaderStatos) closeCommChan() bool {
	alreadyClosed := rs.wasCommClosed()
	if !alreadyClosed {
		rs.commOnce.Do(func() {
			rs.Lock()
			defer rs.Unlock()

			close(rs.commChan)
			rs.commClosed = true
			alreadyClosed = rs.commClosed
		})
	}
	return alreadyClosed
}

func NewReader(r io.Reader) *ReaderStatos {
	return &ReaderStatos{
		iterator:   r,
		commChan:   make(chan int),
		commClosed: false,
	}
}

func (rs *ReaderStatos) wasCommClosed() bool {
	rs.RLock()
	defer rs.RUnlock()

	return rs.commClosed
}

func (rs *ReaderStatos) Read(p []byte) (n int, err error) {
	n, err = rs.iterator.Read(p)

	if err != nil && err != syscall.EINTR {
		rs.closeCommChan()
		return
	}

	if n >= 0 && !rs.wasCommClosed() {
		rs.commChan <- n
	}
	return
}

func (r *ReaderStatos) ProgressChan() chan int {
	return r.commChan
}
