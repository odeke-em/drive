package io_vs_os_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

var (
	_1kB   int64 = 1 << 10
	_1mB   int64 = 1 << 20
	_10mB  int64 = _1mB * 10
	_100mB int64 = _1mB * 100
	_1gB   int64 = 1 << 30
)


func BenchmarkIOPipe1kB(b *testing.B) { benchmarkIt(b, ioPipe, _1kB) }
func BenchmarkOSPipe1kB(b *testing.B) { benchmarkIt(b, osPipe, _1kB) }

func BenchmarkIOPipe1mB(b *testing.B) { benchmarkIt(b, ioPipe, _1mB) }
func BenchmarkOSPipe1mB(b *testing.B) { benchmarkIt(b, osPipe, _1mB) }

func BenchmarkIOPipe10mB(b *testing.B) { benchmarkIt(b, ioPipe, _10mB) }
func BenchmarkOSPipe10mB(b *testing.B) { benchmarkIt(b, osPipe, _10mB) }

func BenchmarkIOPipe100mB(b *testing.B) { benchmarkIt(b, ioPipe, _100mB) }
func BenchmarkOSPipe100mB(b *testing.B) { benchmarkIt(b, osPipe, _100mB) }

func BenchmarkIOPipe1gB(b *testing.B) { benchmarkIt(b, ioPipe, _1gB) }
func BenchmarkOSPipe1gB(b *testing.B) { benchmarkIt(b, osPipe, _1gB) }

func ioPipe(byteCount int64) error {
	prc, pwc := io.Pipe()
	return urandomAndDiscardIt(prc, pwc, byteCount)
}

func osPipe(byteCount int64) error {
	prc, pwc, err := os.Pipe()
	if err != nil {
		return err
	}
	return urandomAndDiscardIt(prc, pwc, byteCount)
}

func benchmarkIt(b *testing.B, fn func(int64) error, byteCount int64) {
	for i := 0; i < b.N; i++ {
		_ = fn(byteCount)
	}
}

func urandomAndDiscardIt(rc io.ReadCloser, wc io.WriteCloser, byteCount int64) error {
	urandom, err := os.Open("/dev/urandom")
	if err != nil {
		return err
	}
	go func() {
		_, _ = io.CopyN(wc, urandom, byteCount)
		_ = wc.Close()
		_ = urandom.Close()

	}()
	_, err = io.Copy(ioutil.Discard, rc)
	return err
}
