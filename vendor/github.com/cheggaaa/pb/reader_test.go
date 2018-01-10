package pb

import (
	"testing"
)

func TestPBProxyReader(t *testing.T) {
	bar := new(ProgressBar)
	if bar.GetBool(Bytes) {
		t.Errorf("By default bytes must be false")
	}

	testReader := new(testReaderCloser)
	proxyReader := bar.NewProxyReader(testReader)

	if !bar.GetBool(Bytes) {
		t.Errorf("Bytes must be true after call NewProxyReader")
	}

	for i := 0; i < 10; i++ {
		buf := make([]byte, 10)
		n, e := proxyReader.Read(buf)
		if e != nil {
			t.Errorf("Proxy reader return err: %v", e)
		}
		if n != len(buf) {
			t.Errorf("Proxy reader return unexpected N: %d (wand %d)", n, len(buf))
		}
		for _, b := range buf {
			if b != 'f' {
				t.Errorf("Unexpected read value: %v (want %v)", b, 'f')
			}
		}
		if want := int64((i + 1) * len(buf)); bar.Current() != want {
			t.Errorf("Unexpected bar current value: %d (want %d)", bar.Current(), want)
		}
	}
	proxyReader.Close()
	if !testReader.closed {
		t.Errorf("Reader must be closed after call ProxyReader.Close")
	}
	proxyReader.Reader = nil
	proxyReader.Close()
}

type testReaderCloser struct {
	closed bool
}

func (tr *testReaderCloser) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 'f'
	}
	return len(p), nil
}

func (tr *testReaderCloser) Close() (err error) {
	tr.closed = true
	return
}
