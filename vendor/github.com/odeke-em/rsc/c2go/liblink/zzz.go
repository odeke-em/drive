package liblink

import (
	"bufio"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"runtime"
)

func Cputime() float64 {
	return 0
}

func sysfatal(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Getgoroot() string {
	return build.Default.GOROOT
}

func Getgoos() string {
	return build.Default.GOOS
}

func Getgoarch() string {
	return build.Default.GOARCH
}

func Getgoarm() string {
	p := os.Getenv("GOARM")
	if p == "" {
		p = "6"
	}
	return p
}

type Biobuf struct {
	r       *bufio.Reader
	w       *bufio.Writer
	written int64
}

func Bprint(b *Biobuf, format string, args ...interface{}) {
	n, _ := fmt.Fprintf(b.w, format, args...)
	b.written += int64(n)
}

func print(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func Bflush(b *Biobuf) {
	b.w.Flush()
}

func Boffset(b *Biobuf) int64 {
	return b.written
}

func Bread(b *Biobuf, buf []byte) int {
	n, err := io.ReadFull(b.r, buf)
	if err != nil && n == 0 {
		n = -1
	}
	return n
}

func Bgetc(b *Biobuf) int {
	c, err := b.r.ReadByte()
	if err != nil {
		return -1
	}
	return int(c)
}

func Bungetc(b *Biobuf) {
	b.r.UnreadByte()
}

func Bputc(b *Biobuf, c int) {
	err := b.w.WriteByte(byte(c))
	if err == nil {
		b.written++
	}
}

func Bwrite(b *Biobuf, buf []byte) int {
	n, err := b.w.Write(buf)
	b.written += int64(n)
	if err != nil && n == 0 {
		n = -1
	}
	return n
}

func (b *Biobuf) Write(buf []byte) (int, error) {
	n, err := b.w.Write(buf)
	b.written += int64(n)
	return n, err
}

func Getgoversion() string {
	return runtime.Version()
}

func Binitw(f io.Writer) *Biobuf {
	return &Biobuf{w: bufio.NewWriter(f)}
}

const (
	NOPROF_textflag   = 1
	DUPOK_textflag    = 2
	NOSPLIT_textflag  = 4
	RODATA_textflag   = 8
	NOPTR_textflag    = 16
	WRAPPER_textflag  = 32
	NEEDCTXT_textflag = 64
)
