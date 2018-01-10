// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Editpcln displays and edits the file names stored in a Go object file or archive.
//
// Usage:
//
//	editpcln [-v] [-r from->to] old.a new.a
//
// Editpcln copies old.a to new.a (it can also be invoked on .5, .6, and .8 files),
// rewriting file paths beginning with 'from' into paths beginning with 'to'.
// The -r option can be repeated to apply multiple rewrites.
//
// The -v option prints information about the paths encountered and rewritten.
//
// To print the paths present in an object file without writing a new one, use
//
//	editpcln -v old.a /dev/null
//
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	verbose = flag.Bool("v", false, "print all function and file names")
)

var rewrites [][2]string

type addRewrite struct{}

func (addRewrite) String() string {
	return ""
}

func (addRewrite) Set(s string) error {
	if strings.Count(s, "->") != 1 {
		return fmt.Errorf("-r argument must be of the form 'from -> to'")
	}
	i := strings.Index(s, "->")
	rewrites = append(rewrites, [2]string{strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+2:])})
	return nil
}

func rewritePath(s string) string {
	for _, oldnew := range rewrites {
		old, new := oldnew[0], oldnew[1]
		if s == old || strings.HasPrefix(s, old+"/") {
			return new + s[len(old):]
		}
	}
	return s
}

func main() {
	flag.Var(addRewrite{}, "r", "add 'from -> to' path rewrite (can be repeated)")

	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
	}

	rf, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	wf, err := os.Create(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}

	r := newRewriter(rf, wf, rewritePath)

	r.verbose = *verbose
	if err := r.rewrite(); err != nil {
		log.Println(err)
		os.Exit(2)
	}
	if err := wf.Close(); err != nil {
		log.Println("writing %s: %v", wf.Name(), err)
		os.Exit(2)
	}
}

type rewriter struct {
	fn      func(string) string
	rf      *os.File
	wf      *os.File
	r       *bufio.Reader
	w       *bufio.Writer
	roffset int64
	verbose bool
	tmp     []byte
}

func newRewriter(rf, wf *os.File, fn func(string) string) *rewriter {
	return &rewriter{
		rf: rf,
		wf: wf,
		r:  bufio.NewReader(rf),
		w:  bufio.NewWriter(wf),
		fn: fn,
	}
}

type rewriterError struct {
	err error
}

func (r *rewriter) rewrite() (err error) {
	defer func() {
		e := recover()
		if re, ok := e.(rewriterError); ok {
			err = re.err
			return
		}
		if e != nil {
			panic(e)
		}
	}()

	buf, _ := r.r.Peek(8)
	if bytes.Equal(buf, []byte("!<arch>\n")) {
		r.readArchive()
	} else if bytes.Equal(buf, []byte("go objec")) {
		r.readObject()
	} else {
		panic(rewriterError{fmt.Errorf("reading %s: unrecognized file format", r.rf.Name())})
	}

	r.flush()
	return nil
}

func (r *rewriter) flush() {
	err := r.w.Flush()
	if err != nil {
		panic(rewriterError{fmt.Errorf("writing %s: %v", r.wf.Name(), err)})
	}
}

func (r *rewriter) wseek(delta int64, whence int) int64 {
	r.flush()
	off, err := r.wf.Seek(delta, whence)
	if err != nil {
		panic(rewriterError{fmt.Errorf("seeking in %s: %v", r.wf.Name(), err)})
	}
	return off
}

// trimSpace removes trailing spaces from b and returns the corresponding string.
// This effectively parses the form used in archive headers.
func trimSpace(b []byte) string {
	return string(bytes.TrimRight(b, " "))
}

func (r *rewriter) readArchive() {
	r.readFixed([]byte("!<arch>\n"))

	for {
		_, err := r.r.ReadByte()
		if err == io.EOF {
			return
		}
		r.r.UnreadByte()

		// Each file is preceded by this text header (slice indices in first column):
		//	 0:16	name
		//	16:28 date
		//	28:34 uid
		//	34:40 gid
		//	40:48 mode
		//	48:58 size
		//	58:60 magic - `\n
		buf := make([]byte, 60)
		r.readFull(buf)
		if buf[58] != '`' || buf[59] != '\n' {
			r.corrupt()
		}
		wstart := r.wseek(0, 1)

		name := trimSpace(buf[0:16])
		size, err := strconv.ParseInt(trimSpace(buf[48:58]), 10, 64)
		if err != nil {
			r.corrupt()
		}

		switch name {
		case "__.SYMDEF", "__.GOSYMDEF", "__.PKGDEF":
			r.skip(size + size&1)
			continue
		}

		start := r.roffset
		r.readObject()
		n := r.roffset - start
		if n != int64(size) {
			r.corrupt()
		}
		if size&1 != 0 {
			r.r.ReadByte()
			r.roffset++
		}

		wend := r.wseek(0, 1)
		wsize := wend - wstart
		r.wseek(wstart-60+48, 0)
		r.w.WriteString(fmt.Sprintf("%-10x", wsize))
		r.wseek(wend, 0)
		if wsize&1 != 0 {
			r.w.WriteByte(0)
		}
	}
}

func (r *rewriter) readObject() {
	var c1, c2, c3 byte
	for {
		c1, c2, c3 = c2, c3, r.readByte()
		if c1 == '\n' && c2 == '!' && c3 == '\n' {
			break
		}
	}

	// Header + version.
	r.readFixed([]byte("\x00\x00go13ld\x01"))

	// Package imports.
	for r.readString() != "" {
	}

	discard := bufio.NewWriter(ioutil.Discard)

	// Symbols.
	for {
		if b := r.readByte(); b != 0xfe {
			if b != 0xff {
				r.corrupt()
			}
			break
		}

		kind := r.readInt()
		name, _ := r.readSymID()
		r.readInt()              // dupok
		r.readInt()              // size
		r.readSymID()            // type
		r.readData()             // data
		n := r.readInt()         // nreloc
		for i := 0; i < n; i++ { // reloc
			r.readInt()   // offset
			r.readInt()   // size
			r.readInt()   // type
			r.readInt()   // add
			r.readInt()   // xadd
			r.readSymID() // sym
			r.readSymID() // xsym
		}

		const STEXT = 1
		if kind == STEXT {
			if r.verbose {
				fmt.Printf("%s\n", name)
			}
			r.readInt()              // args
			r.readInt()              // frame
			n = r.readInt()          // nvar
			for i := 0; i < n; i++ { // var
				r.readSymID() // name
				r.readInt()   // offset
				r.readInt()   // kind
				r.readSymID() // type
			}

			r.readData()             // pcsp
			r.readData()             // pcfile
			r.readData()             // pcline
			n = r.readInt()          // npcdata
			for i := 0; i < n; i++ { // pcdata
				r.readData()
			}
			n = r.readInt()          // nfuncdata
			for i := 0; i < n; i++ { // funcdata syms
				r.readSymID()
			}
			for i := 0; i < n; i++ { // funcdata offsets
				r.readInt()
			}
			n = r.readInt() // nfile
			w := r.w
			r.w = discard
			for i := 0; i < n; i++ {
				file, vers := r.readSymID() // file names
				wfile := file
				if r.fn != nil {
					wfile = r.fn(file)
				}
				if r.verbose {
					if wfile != file {
						fmt.Printf("\t%s -> %s\n", file, wfile)
					} else {
						fmt.Printf("\t%s\n", file)
					}
				}
				writeString(w, wfile)
				writeInt(w, vers)
			}
			r.w = w
		}
	}

	r.readFixed([]byte("\xffgo13ld")) // first ff already read
}

func (r *rewriter) readError(err error) {
	panic(rewriterError{fmt.Errorf("reading %s: %v", r.rf.Name(), err)})
}

func (r *rewriter) corrupt() {
	panic(rewriterError{fmt.Errorf("reading %s: corrupt file", r.rf.Name())})
}

func (r *rewriter) readFull(buf []byte) {
	_, err := io.ReadFull(r.r, buf)
	if err != nil {
		r.readError(err)
	}
	r.roffset += int64(len(buf))
	r.w.Write(buf)
}

func (r *rewriter) readFixed(match []byte) {
	buf := make([]byte, len(match))
	r.readFull(buf)
	if !bytes.Equal(buf, match) {
		r.corrupt()
	}
}

func (r *rewriter) skip(n int64) {
	const chunk = 1 << 16
	if r.tmp == nil {
		r.tmp = make([]byte, chunk)
	}
	for n >= chunk {
		r.readFull(r.tmp)
		n -= chunk
	}
	if n > 0 {
		r.readFull(r.tmp[:n])
	}
}

func (r *rewriter) readByte() byte {
	b, err := r.r.ReadByte()
	if err != nil {
		r.readError(err)
	}
	r.roffset++
	r.w.WriteByte(b)
	return b
}

func (r *rewriter) readInt() int {
	var u uint64
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			r.corrupt()
		}
		c := r.readByte()
		u |= uint64(c&0x7F) << shift
		if c&0x80 == 0 {
			break
		}
	}
	v := int64(u>>1) ^ (int64(u) << 63 >> 63)
	if int64(int(v)) != v {
		r.corrupt()
	}
	return int(v)
}

func (r *rewriter) readString() string {
	n := r.readInt()
	buf := make([]byte, n)
	r.readFull(buf)
	return string(buf)
}

func (r *rewriter) readData() {
	n := r.readInt()
	buf := make([]byte, n)
	r.readFull(buf)
}

func (r *rewriter) readSymID() (string, int) {
	return r.readString(), r.readInt()
}

func writeInt(w *bufio.Writer, n int) {
	v := int64(n)
	u := uint64(v<<1) ^ uint64(v>>63)
	for u >= 0x80 {
		w.WriteByte(byte(u) | 0x80)
		u >>= 7
	}
	w.WriteByte(byte(u))
}

func writeString(w *bufio.Writer, s string) {
	writeInt(w, len(s))
	w.WriteString(s)
}
