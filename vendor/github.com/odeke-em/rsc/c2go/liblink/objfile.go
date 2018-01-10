package liblink

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Writing and reading of Go object files.
//
// Originally, Go object files were Plan 9 object files, but no longer.
// Now they are more like standard object files, in that each symbol is defined
// by an associated memory image (bytes) and a list of relocations to apply
// during linking. We do not (yet?) use a standard file format, however.
// For now, the format is chosen to be as simple as possible to read and write.
// It may change for reasons of efficiency, or we may even switch to a
// standard file format if there are compelling benefits to doing so.
// See golang.org/s/go13linker for more background.
//
// The file format is:
//
//	- magic header: "\x00\x00go13ld"
//	- byte 1 - version number
//	- sequence of strings giving dependencies (imported packages)
//	- empty string (marks end of sequence)
//	- sequence of defined symbols
//	- byte 0xff (marks end of sequence)
//	- magic footer: "\xff\xffgo13ld"
//
// All integers are stored in a zigzag varint format.
// See golang.org/s/go12symtab for a definition.
//
// Data blocks and strings are both stored as an integer
// followed by that many bytes.
//
// A symbol reference is a string name followed by a version.
// An empty name corresponds to a nil LSym* pointer.
//
// Each symbol is laid out as the following fields (taken from LSym*):
//
//	- byte 0xfe (sanity check for synchronization)
//	- type [int]
//	- name [string]
//	- version [int]
//	- dupok [int]
//	- size [int]
//	- gotype [symbol reference]
//	- p [data block]
//	- nr [int]
//	- r [nr relocations, sorted by off]
//
// If type == STEXT, there are a few more fields:
//
//	- args [int]
//	- locals [int]
//	- nosplit [int]
//	- leaf [int]
//	- nlocal [int]
//	- local [nlocal automatics]
//	- pcln [pcln table]
//
// Each relocation has the encoding:
//
//	- off [int]
//	- siz [int]
//	- type [int]
//	- add [int]
//	- xadd [int]
//	- sym [symbol reference]
//	- xsym [symbol reference]
//
// Each local has the encoding:
//
//	- asym [symbol reference]
//	- offset [int]
//	- type [int]
//	- gotype [symbol reference]
//
// The pcln table has the encoding:
//
//	- pcsp [data block]
//	- pcfile [data block]
//	- pcline [data block]
//	- npcdata [int]
//	- pcdata [npcdata data blocks]
//	- nfuncdata [int]
//	- funcdata [nfuncdata symbol references]
//	- funcdatasym [nfuncdata ints]
//	- nfile [int]
//	- file [nfile symbol references]
//
// The file layout and meaning of type integers are architecture-independent.
//
// TODO(rsc): The file format is good for a first pass but needs work.
//	- There are SymID in the object file that should really just be strings.
//	- The actual symbol memory images are interlaced with the symbol
//	  metadata. They should be separated, to reduce the I/O required to
//	  load just the metadata.
//	- The symbol references should be shortened, either with a symbol
//	  table or by using a simple backward index to an earlier mentioned symbol.

// The Go and C compilers, and the assembler, call writeobj to write
// out a Go object file.  The linker does not call this; the linker
// does not write out object files.
func Writeobj(ctxt *Link, b *Biobuf) {
	var flag int
	var h *Hist
	var s *LSym
	var text *LSym
	var etext *LSym
	var curtext *LSym
	var data *LSym
	var edata *LSym
	var pl *Plist
	var p *Prog
	var plink *Prog
	var a *Auto
	// Build list of symbols, and assign instructions to lists.
	// Ignore ctxt->plist boundaries. There are no guarantees there,
	// and the C compilers and assemblers just use one big list.
	text = nil
	curtext = nil
	data = nil
	etext = nil
	edata = nil
	for pl = ctxt.Plist; pl != nil; pl = pl.Link {
		for p = pl.Firstpc; p != nil; p = plink {
			plink = p.Link
			p.Link = nil
			if p.As == ctxt.Arch.AEND {
				continue
			}
			if p.As == ctxt.Arch.ATYPE {
				// Assume each TYPE instruction describes
				// a different local variable or parameter,
				// so no dedup.
				// Using only the TYPE instructions means
				// that we discard location information about local variables
				// in C and assembly functions; that information is inferred
				// from ordinary references, because there are no TYPE
				// instructions there. Without the type information, gdb can't
				// use the locations, so we don't bother to save them.
				// If something else could use them, we could arrange to
				// preserve them.
				if curtext == nil {
					continue
				}
				a = new(Auto)
				a.Asym = p.From.Sym
				a.Aoffset = int(p.From.Offset)
				a.Typ = ctxt.Arch.Symtype(&p.From)
				a.Gotype = p.From.Gotype
				a.Link = curtext.Autom
				curtext.Autom = a
				continue
			}
			if p.As == ctxt.Arch.AGLOBL {
				s = p.From.Sym
				tmp6 := s.Seenglobl
				s.Seenglobl++
				if tmp6 != 0 {
					fmt.Printf("duplicate %v\n", p)
				}
				if s.Onlist != 0 {
					log.Fatalf("symbol %s listed multiple times", s.Name)
				}
				s.Onlist = 1
				if data == nil {
					data = s
				} else {
					edata.Next = s
				}
				s.Next = nil
				s.Size = p.To.Offset
				if s.Typ == 0 || s.Typ == SXREF {
					s.Typ = SBSS
				}
				if ctxt.Arch.Thechar == '5' {
					flag = p.Reg
				} else {
					flag = int(p.From.Scale)
				}
				if flag&DUPOK != 0 {
					s.Dupok = 1
				}
				if flag&RODATA != 0 {
					s.Typ = SRODATA
				} else if flag&NOPTR != 0 {
					s.Typ = SNOPTRBSS
				}
				edata = s
				continue
			}
			if p.As == ctxt.Arch.ADATA {
				savedata(ctxt, p.From.Sym, p, "<input>")
				continue
			}
			if p.As == ctxt.Arch.ATEXT {
				s = p.From.Sym
				if s == nil {
					// func _() { }
					curtext = nil
					continue
				}
				if s.Text != nil {
					log.Fatalf("duplicate TEXT for %s", s.Name)
				}
				if s.Onlist != 0 {
					log.Fatalf("symbol %s listed multiple times", s.Name)
				}
				s.Onlist = 1
				if text == nil {
					text = s
				} else {
					etext.Next = s
				}
				etext = s
				if ctxt.Arch.Thechar == '5' {
					flag = p.Reg
				} else {
					flag = int(p.From.Scale)
				}
				if flag&DUPOK != 0 {
					s.Dupok = 1
				}
				if flag&NOSPLIT != 0 {
					s.Nosplit = 1
				}
				s.Next = nil
				s.Typ = STEXT
				s.Text = p
				s.Etext = p
				curtext = s
				continue
			}
			if curtext == nil {
				continue
			}
			s = curtext
			s.Etext.Link = p
			s.Etext = p
		}
	}
	// Turn functions into machine code images.
	for s = text; s != nil; s = s.Next {
		mkfwd(s)
		linkpatch(ctxt, s)
		ctxt.Arch.Follow(ctxt, s)
		ctxt.Arch.Addstacksplit(ctxt, s)
		ctxt.Arch.Assemble(ctxt, s)
		linkpcln(ctxt, s)
	}
	// Emit header.
	Bputc(b, 0)
	Bputc(b, 0)
	fmt.Fprintf(b, "go13ld")
	Bputc(b, 1) // version
	// Emit autolib.
	for h = ctxt.Hist; h != nil; h = h.Link {
		if h.Offset < 0 {
			wrstring(b, h.Name)
		}
	}
	wrstring(b, "")
	// Emit symbols.
	for s = text; s != nil; s = s.Next {
		writesym(ctxt, b, s)
	}
	for s = data; s != nil; s = s.Next {
		writesym(ctxt, b, s)
	}
	// Emit footer.
	Bputc(b, 0xff)
	Bputc(b, 0xff)
	fmt.Fprintf(b, "go13ld")
	Bflush(b)
}

func writesym(ctxt *Link, b *Biobuf, s *LSym) {
	var r *Reloc
	var i int
	var j int
	var c int
	var n int
	var pc *Pcln
	var p *Prog
	var a *Auto
	var name string
	if ctxt.Debugasm != 0 {
		fmt.Fprintf(ctxt.Bso, "%s ", s.Name)
		if s.Version != 0 {
			fmt.Fprintf(ctxt.Bso, "v=%d ", s.Version)
		}
		if s.Typ != 0 {
			fmt.Fprintf(ctxt.Bso, "t=%d ", s.Typ)
		}
		if s.Dupok != 0 {
			fmt.Fprintf(ctxt.Bso, "dupok ")
		}
		if s.Nosplit != 0 {
			fmt.Fprintf(ctxt.Bso, "nosplit ")
		}
		fmt.Fprintf(ctxt.Bso, "size=%d value=%d", int64(s.Size), int64(s.Value))
		if s.Typ == STEXT {
			fmt.Fprintf(ctxt.Bso, " args=%#x locals=%#x", uint64(s.Args), uint64(s.Locals))
			if s.Leaf != 0 {
				fmt.Fprintf(ctxt.Bso, " leaf")
			}
		}
		fmt.Fprintf(ctxt.Bso, "\n")
		for p = s.Text; p != nil; p = p.Link {
			fmt.Fprintf(ctxt.Bso, "\t%#04x %v\n", uint(int(p.Pc)), p)
		}
		for i = 0; i < len(s.P); {
			fmt.Fprintf(ctxt.Bso, "\t%#04x", uint(i))
			for j = i; j < i+16 && j < len(s.P); j++ {
				fmt.Fprintf(ctxt.Bso, " %02x", s.P[j])
			}
			for ; j < i+16; j++ {
				fmt.Fprintf(ctxt.Bso, "   ")
			}
			fmt.Fprintf(ctxt.Bso, "  ")
			for j = i; j < i+16 && j < len(s.P); j++ {
				c = int(s.P[j])
				if ' ' <= c && c <= 0x7e {
					fmt.Fprintf(ctxt.Bso, "%c", c)
				} else {
					fmt.Fprintf(ctxt.Bso, ".")
				}
			}
			fmt.Fprintf(ctxt.Bso, "\n")
			i += 16
		}
		for i = 0; i < len(s.R); i++ {
			r = &s.R[i]
			name = ""
			if r.Sym != nil {
				name = r.Sym.Name
			}
			fmt.Fprintf(ctxt.Bso, "\trel %d+%d t=%d %s+%d\n", int(r.Off), r.Siz, r.Typ, name, int64(r.Add))
		}
	}
	Bputc(b, 0xfe)
	wrint(b, int64(s.Typ))
	wrstring(b, s.Name)
	wrint(b, int64(s.Version))
	wrint(b, int64(s.Dupok))
	wrint(b, s.Size)
	wrsym(b, s.Gotype)
	wrdata(b, s.P)
	wrint(b, int64(len(s.R)))
	for i = 0; i < len(s.R); i++ {
		r = &s.R[i]
		wrint(b, r.Off)
		wrint(b, int64(r.Siz))
		wrint(b, int64(r.Typ))
		wrint(b, r.Add)
		wrint(b, r.Xadd)
		wrsym(b, r.Sym)
		wrsym(b, r.Xsym)
	}
	if s.Typ == STEXT {
		wrint(b, int64(s.Args))
		wrint(b, s.Locals)
		wrint(b, int64(s.Nosplit))
		wrint(b, int64(s.Leaf))
		n = 0
		for a = s.Autom; a != nil; a = a.Link {
			n++
		}
		wrint(b, int64(n))
		for a = s.Autom; a != nil; a = a.Link {
			wrsym(b, a.Asym)
			wrint(b, int64(a.Aoffset))
			if a.Typ == ctxt.Arch.D_AUTO {
				wrint(b, A_AUTO)
			} else if a.Typ == ctxt.Arch.D_PARAM {
				wrint(b, A_PARAM)
			} else {
				log.Fatalf("%s: invalid local variable type %d", s.Name, a.Typ)
			}
			wrsym(b, a.Gotype)
		}
		pc = s.Pcln
		wrdata(b, pc.Pcsp.P)
		wrdata(b, pc.Pcfile.P)
		wrdata(b, pc.Pcline.P)
		wrint(b, int64(len(pc.Pcdata)))
		for i = 0; i < len(pc.Pcdata); i++ {
			wrdata(b, pc.Pcdata[i].P)
		}
		wrint(b, int64(pc.Nfuncdata))
		for i = 0; i < pc.Nfuncdata; i++ {
			wrsym(b, pc.Funcdata[i])
		}
		for i = 0; i < pc.Nfuncdata; i++ {
			wrint(b, pc.Funcdataoff[i])
		}
		wrint(b, int64(len(pc.File)))
		for i = 0; i < len(pc.File); i++ {
			wrpathsym(ctxt, b, pc.File[i])
		}
	}
}

func wrint(b *Biobuf, sval int64) {
	var uv uint64
	var v uint64
	var buf [10]uint8
	var p []uint8
	uv = (uint64(sval) << 1) ^ uint64(int64(sval>>63))
	p = buf[:]
	for v = uv; v >= 0x80; v >>= 7 {
		p[0] = uint8(v | 0x80)
		p = p[1:]
	}
	p[0] = uint8(v)
	p = p[1:]
	Bwrite(b, buf[:len(buf)-len(p)])
}

func wrstring(b *Biobuf, s string) {
	wrdata(b, []byte(s))
}

// wrpath writes a path just like a string, but on windows, it
// translates '\\' to '/' in the process.
func wrpath(ctxt *Link, b *Biobuf, p string) {
	var i int
	var n int
	if ctxt.Windows == 0 || !strings.Contains(p, `\`) {
		wrstring(b, p)
		return
	}
	n = len(p)
	wrint(b, int64(n))
	for i = 0; i < n; i++ {
		var tmp int
		if p[i] == '\\' {
			tmp = '/'
		} else {
			tmp = int(p[i])
		}
		Bputc(b, tmp)
	}
}

func wrdata(b *Biobuf, v []byte) {
	wrint(b, int64(len(v)))
	Bwrite(b, v)
}

func wrpathsym(ctxt *Link, b *Biobuf, s *LSym) {
	if s == nil {
		wrint(b, 0)
		wrint(b, 0)
		return
	}
	wrpath(ctxt, b, s.Name)
	wrint(b, int64(s.Version))
}

func wrsym(b *Biobuf, s *LSym) {
	if s == nil {
		wrint(b, 0)
		wrint(b, 0)
		return
	}
	wrstring(b, s.Name)
	wrint(b, int64(s.Version))
}

var startmagic string = "\x00\x00go13ld"

var endmagic string = "\xff\xffgo13ld"

func ldobjfile(ctxt *Link, f *Biobuf, pkg string, len int64, pn string) {
	var c int
	var buf [8]uint8
	var start int64
	var lib string
	start = Boffset(f)
	ctxt.Version++
	buf = [8]uint8{}
	Bread(f, buf[:])
	if string(buf[:]) != startmagic {
		log.Fatalf("%s: invalid file start %x %x %x %x %x %x %x %x", pn, buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7])
	}
	c = Bgetc(f)
	if c != 1 {
		log.Fatalf("%s: invalid file version number %d", pn, c)
	}
	for {
		lib = rdstring(f)
		if lib[0] == 0 {
			break
		}
		addlib(ctxt, pkg, pn, lib)
	}
	for {
		c = Bgetc(f)
		Bungetc(f)
		if c == 0xff {
			break
		}
		readsym(ctxt, f, pkg, pn)
	}
	buf = [8]uint8{}
	Bread(f, buf[:])
	if string(buf[:]) != endmagic {
		log.Fatalf("%s: invalid file end", pn)
	}
	if Boffset(f) != start+len {
		log.Fatalf("%s: unexpected end at %d, want %d", pn, int64(Boffset(f)), int64(start+len))
	}
}

func readsym(ctxt *Link, f *Biobuf, pkg string, pn string) {
	var i int
	var j int
	var c int
	var t int
	var v uint32
	var n int
	var size int64
	var dupok int
	var ndup uint32
	var name string
	var r *Reloc
	var s *LSym
	var dup *LSym
	var pc *Pcln
	var a *Auto
	if Bgetc(f) != 0xfe {
		log.Fatalf("readsym out of sync")
	}
	t = int(rdint(f))
	name = expandpkg(rdstring(f), pkg)
	v = uint32(rdint(f))
	if v != 0 && v != 1 {
		log.Fatalf("invalid symbol version %d", v)
	}
	dupok = int(rdint(f))
	size = rdint(f)
	if v != 0 {
		v = ctxt.Version
	}
	s = Linklookup(ctxt, name, v)
	dup = nil
	if s.Typ != 0 && s.Typ != SXREF {
		if s.Typ != SBSS && s.Typ != SNOPTRBSS && dupok == 0 && s.Dupok == 0 {
			log.Fatalf("duplicate symbol %s (types %d and %d) in %s and %s", s.Name, s.Typ, t, s.File, pn)
		}
		if len(s.P) > 0 {
			dup = s
			s = linknewsym(ctxt, ".dup", ndup)
			ndup++ // scratch
		}
	}
	s.File = pkg
	s.Dupok = dupok
	if t == SXREF {
		log.Fatalf("bad sxref")
	}
	if t == 0 {
		log.Fatalf("missing type for %s in %s", name, pn)
	}
	s.Typ = t
	if s.Size < size {
		s.Size = size
	}
	s.Gotype = rdsym(ctxt, f, pkg)
	rddata(f, &s.P)
	s.P = s.P[:len(s.P)]
	n = int(rdint(f))
	if n > 0 {
		s.R = make([]Reloc, n)
		s.R = s.R[:n]
		s.R = s.R[:n]
		for i = 0; i < n; i++ {
			r = &s.R[i]
			r.Off = rdint(f)
			r.Siz = uint8(rdint(f))
			r.Typ = int(rdint(f))
			r.Add = rdint(f)
			r.Xadd = rdint(f)
			r.Sym = rdsym(ctxt, f, pkg)
			r.Xsym = rdsym(ctxt, f, pkg)
		}
	}
	if len(s.P) > 0 && dup != nil && len(dup.P) > 0 && strings.HasPrefix(s.Name, "gclocals·") {
		// content-addressed garbage collection liveness bitmap symbol.
		// double check for hash collisions.
		if !bytes.Equal(s.P, dup.P) {
			log.Fatalf("dupok hash collision for %s in %s and %s", s.Name, s.File, pn)
		}
	}
	if s.Typ == STEXT {
		s.Args = int(rdint(f))
		s.Locals = rdint(f)
		s.Nosplit = uint8(rdint(f))
		s.Leaf = uint8(rdint(f))
		n = int(rdint(f))
		for i = 0; i < n; i++ {
			a = new(Auto)
			a.Asym = rdsym(ctxt, f, pkg)
			a.Aoffset = int(rdint(f))
			a.Typ = int(rdint(f))
			a.Gotype = rdsym(ctxt, f, pkg)
			a.Link = s.Autom
			s.Autom = a
		}
		s.Pcln = new(Pcln)
		pc = s.Pcln
		rddata(f, &pc.Pcsp.P)
		rddata(f, &pc.Pcfile.P)
		rddata(f, &pc.Pcline.P)
		n = int(rdint(f))
		pc.Pcdata = make([]Pcdata, n)
		pc.Pcdata = pc.Pcdata[:n]
		for i = 0; i < n; i++ {
			rddata(f, &pc.Pcdata[i].P)
		}
		n = int(rdint(f))
		pc.Funcdata = make([]*LSym, n)
		pc.Funcdataoff = make([]int64, n)
		pc.Nfuncdata = n
		for i = 0; i < n; i++ {
			pc.Funcdata[i] = rdsym(ctxt, f, pkg)
		}
		for i = 0; i < n; i++ {
			pc.Funcdataoff[i] = rdint(f)
		}
		n = int(rdint(f))
		pc.File = make([]*LSym, n)
		pc.File = pc.File[:n]
		for i = 0; i < n; i++ {
			pc.File[i] = rdsym(ctxt, f, pkg)
		}
		if dup == nil {
			if s.Onlist != 0 {
				log.Fatalf("symbol %s listed multiple times", s.Name)
			}
			s.Onlist = 1
			if ctxt.Etextp != nil {
				ctxt.Etextp.Next = s
			} else {
				ctxt.Textp = s
			}
			ctxt.Etextp = s
		}
	}
	if ctxt.Debugasm != 0 {
		fmt.Fprintf(ctxt.Bso, "%s ", s.Name)
		if s.Version != 0 {
			fmt.Fprintf(ctxt.Bso, "v=%d ", s.Version)
		}
		if s.Typ != 0 {
			fmt.Fprintf(ctxt.Bso, "t=%d ", s.Typ)
		}
		if s.Dupok != 0 {
			fmt.Fprintf(ctxt.Bso, "dupok ")
		}
		if s.Nosplit != 0 {
			fmt.Fprintf(ctxt.Bso, "nosplit ")
		}
		fmt.Fprintf(ctxt.Bso, "size=%d value=%d", int64(s.Size), int64(s.Value))
		if s.Typ == STEXT {
			fmt.Fprintf(ctxt.Bso, " args=%#x locals=%#x", uint64(s.Args), uint64(s.Locals))
		}
		fmt.Fprintf(ctxt.Bso, "\n")
		for i = 0; i < len(s.P); {
			fmt.Fprintf(ctxt.Bso, "\t%#04x", uint(i))
			for j = i; j < i+16 && j < len(s.P); j++ {
				fmt.Fprintf(ctxt.Bso, " %02x", s.P[j])
			}
			for ; j < i+16; j++ {
				fmt.Fprintf(ctxt.Bso, "   ")
			}
			fmt.Fprintf(ctxt.Bso, "  ")
			for j = i; j < i+16 && j < len(s.P); j++ {
				c = int(s.P[j])
				if ' ' <= c && c <= 0x7e {
					fmt.Fprintf(ctxt.Bso, "%c", c)
				} else {
					fmt.Fprintf(ctxt.Bso, ".")
				}
			}
			fmt.Fprintf(ctxt.Bso, "\n")
			i += 16
		}
		for i = 0; i < len(s.R); i++ {
			r = &s.R[i]
			fmt.Fprintf(ctxt.Bso, "\trel %d+%d t=%d %s+%d\n", int(r.Off), r.Siz, r.Typ, r.Sym.Name, int64(r.Add))
		}
	}
}

func rdint(f *Biobuf) int64 {
	var c int
	var uv uint64
	var shift int
	uv = 0
	for shift = 0; ; shift += 7 {
		if shift >= 64 {
			log.Fatalf("corrupt input")
		}
		c = Bgetc(f)
		uv |= uint64(c&0x7F) << uint(shift)
		if c&0x80 == 0 {
			break
		}
	}
	return int64(uv>>1) ^ (int64(uint64(uv)<<63) >> 63)
}

func rdstring(f *Biobuf) string {
	n := rdint(f)
	p := make([]byte, n)
	Bread(f, p)
	return string(p)
}

func rddata(f *Biobuf, pp *[]byte) {
	*pp = make([]byte, rdint(f))
	Bread(f, *pp)
}

func rdsym(ctxt *Link, f *Biobuf, pkg string) *LSym {
	var n int64
	var v uint32
	var p []byte
	var s *LSym
	n = rdint(f)
	if n == 0 {
		rdint(f)
		return nil
	}
	p = make([]byte, n)
	Bread(f, p)
	v = uint32(rdint(f))
	if v != 0 {
		v = ctxt.Version
	}
	s = Linklookup(ctxt, expandpkg(string(p), pkg), v)
	if v == 0 && s.Name[0] == '$' && s.Typ == 0 {
		if strings.HasPrefix(s.Name, "$f32.") {
			u64, _ := strconv.ParseUint(s.Name[5:], 16, 32)
			u32 := uint32(u64)
			s.Typ = SRODATA
			Adduint32(ctxt, s, u32)
			s.Reachable = 0
		} else if strings.HasPrefix(s.Name, "$f64.") {
			u64, _ := strconv.ParseUint(s.Name[5:], 16, 64)
			s.Typ = SRODATA
			Adduint64(ctxt, s, u64)
			s.Reachable = 0
		}
	}
	return s
}
