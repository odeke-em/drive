package main

import (
	"bufio"
	"go/build"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/rsc/c2go/liblink"
	"code.google.com/p/rsc/c2go/liblink/amd64"
	"code.google.com/p/rsc/c2go/liblink/arm"
	"code.google.com/p/rsc/c2go/liblink/x86"
)

var arch *liblink.LinkArch

func main() {
	switch build.Default.GOARCH {
	case "amd64":
		arch = &amd64.Linkamd64
	case "amd64p32":
		arch = &amd64.Linkamd64p32
	case "386":
		arch = &x86.Link386
	case "arm":
		arch = &arm.Linkarm
	}
	if len(os.Args) == 3 {
		input()
		return
	}
	f, err := os.Create("x.6")
	if err != nil {
		log.Fatal(err)
	}
	ctxt := liblink.Linknew(arch)
	ctxt.Debugasm = 1
	ctxt.Bso = liblink.Binitw(os.Stdout)
	defer liblink.Bflush(ctxt.Bso)
	ctxt.Diag = log.Fatalf
	obuf := liblink.Binitw(f)
	liblink.Bprint(obuf, "go object %s %s %s\n", liblink.Getgoos(), liblink.Getgoarch(), liblink.Getgoversion())
	liblink.Bprint(obuf, "!\n")
	p1 := &liblink.Prog{
		Ctxt:   ctxt,
		As:     amd64.ATEXT,
		Lineno: 1,
		From: liblink.Addr{
			Typ:   amd64.D_EXTERN,
			Index: amd64.D_NONE,
			Sym:   liblink.Linklookup(ctxt, "main.Main", 0),
			Scale: 0,
		},
		To: liblink.Addr{
			Typ:   amd64.D_CONST,
			Index: amd64.D_NONE,
		},
	}
	p2 := &liblink.Prog{
		Ctxt: ctxt,
		As:   amd64.ARET,
		From: liblink.Addr{
			Typ:   amd64.D_NONE,
			Index: amd64.D_NONE,
		},
		To: liblink.Addr{
			Typ:   amd64.D_NONE,
			Index: amd64.D_NONE,
		},
	}
	p3 := &liblink.Prog{
		Ctxt:   ctxt,
		As:     amd64.ATEXT,
		Lineno: 1,
		From: liblink.Addr{
			Typ:   amd64.D_EXTERN,
			Index: amd64.D_NONE,
			Sym:   liblink.Linklookup(ctxt, "main.Init", 0),
			Scale: 0,
		},
		To: liblink.Addr{
			Typ:   amd64.D_CONST,
			Index: amd64.D_NONE,
		},
	}
	p4 := &liblink.Prog{
		Ctxt: ctxt,
		As:   amd64.ARET,
		From: liblink.Addr{
			Typ:   amd64.D_NONE,
			Index: amd64.D_NONE,
		},
		To: liblink.Addr{
			Typ:   amd64.D_NONE,
			Index: amd64.D_NONE,
		},
	}
	pl := liblink.Linknewplist(ctxt)
	pl.Firstpc = p1
	p1.Link = p2
	p2.Link = p3
	p3.Link = p4
	liblink.Writeobj(ctxt, obuf)
	liblink.Bflush(obuf)
}

var (
	ctxt   *liblink.Link
	plists = map[string]*liblink.Plist{}
	syms   = map[string]*liblink.LSym{}
	progs  = map[string]*liblink.Prog{}
	hists  = map[string]*liblink.Hist{}
	undef  = map[interface{}]bool{}
)

func input() {
	ctxt = liblink.Linknew(arch)
	//ctxt.Debugasm = 1
	ctxt.Bso = liblink.Binitw(os.Stdout)
	defer liblink.Bflush(ctxt.Bso)
	ctxt.Diag = log.Fatalf
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	b := bufio.NewReader(f)
	if rdstring(b) != "ctxt" {
		log.Fatal("invalid input - missing ctxt")
	}
	name := rdstring(b)
	if name != ctxt.Arch.Name {
		log.Fatalf("bad arch %s - want %s", name, ctxt.Arch.Name)
	}

	ctxt.Goarm = int(rdint(b))
	ctxt.Debugasm = int(rdint(b))
	ctxt.Trimpath = rdstring(b)
	ctxt.Plist = rdplist(b)
	ctxt.Plast = rdplist(b)
	ctxt.Hist = rdhist(b)
	ctxt.Ehist = rdhist(b)
	for {
		i := rdint(b)
		if i < 0 {
			break
		}
		ctxt.Hash[i] = rdsym(b)
	}
	last := "ctxt"

Loop:
	for {
		s := rdstring(b)
		switch s {
		default:
			log.Fatalf("unexpected input after %s: %v", s, last)
		case "end":
			break Loop
		case "plist":
			readplist(b, rdplist(b))
		case "sym":
			readsym(b, rdsym(b))
		case "prog":
			readprog(b, rdprog(b))
		case "hist":
			readhist(b, rdhist(b))
		}
		last = s
	}

	if len(undef) > 0 {
		panic("missing definitions")
	}

	ff, err := os.Create(os.Args[2])
	obuf := liblink.Binitw(ff)
	liblink.Writeobj(ctxt, obuf)
	liblink.Bflush(obuf)
}

func rdstring(b *bufio.Reader) string {
	s, err := b.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	s = strings.TrimSpace(s)
	if s == "<nil>" {
		s = ""
	}
	return s
}

func rdplist(b *bufio.Reader) *liblink.Plist {
	id := rdstring(b)
	if id == "0" || id == "0x0" {
		return nil
	}
	pl := plists[id]
	if pl == nil {
		pl = new(liblink.Plist)
		plists[id] = pl
		undef[pl] = true
	}
	return pl
}

func rdsym(b *bufio.Reader) *liblink.LSym {
	id := rdstring(b)
	if id == "0" || id == "0x0" {
		return nil
	}
	sym := syms[id]
	if sym == nil {
		sym = new(liblink.LSym)
		syms[id] = sym
		undef[sym] = true
	}
	return sym
}

func rdprog(b *bufio.Reader) *liblink.Prog {
	id := rdstring(b)
	if id == "0" || id == "0x0" {
		return nil
	}
	prog := progs[id]
	if prog == nil {
		prog = new(liblink.Prog)
		prog.Ctxt = ctxt
		progs[id] = prog
		undef[prog] = true
	}
	return prog
}

func rdhist(b *bufio.Reader) *liblink.Hist {
	id := rdstring(b)
	if id == "0" || id == "0x0" {
		return nil
	}
	h := hists[id]
	if h == nil {
		h = new(liblink.Hist)
		hists[id] = h
		undef[h] = true
	}
	return h
}

func readplist(b *bufio.Reader, pl *liblink.Plist) {
	if !undef[pl] {
		panic("double-def")
	}
	delete(undef, pl)
	pl.Recur = int(rdint(b))
	pl.Name = rdsym(b)
	pl.Firstpc = rdprog(b)
	pl.Link = rdplist(b)
}

func readsym(b *bufio.Reader, s *liblink.LSym) {
	if !undef[s] {
		panic("double-def")
	}
	delete(undef, s)
	s.Name = rdstring(b)
	s.Extname = rdstring(b)
	s.Typ = int(rdint(b))
	s.Version = uint32(rdint(b))
	s.Dupok = int(rdint(b))
	s.External = uint8(rdint(b))
	s.Nosplit = uint8(rdint(b))
	s.Reachable = uint8(rdint(b))
	s.Cgoexport = uint8(rdint(b))
	s.Special = uint8(rdint(b))
	s.Stkcheck = uint8(rdint(b))
	s.Hide = uint8(rdint(b))
	s.Leaf = uint8(rdint(b))
	s.Fnptr = uint8(rdint(b))
	s.Seenglobl = uint8(rdint(b))
	s.Onlist = uint8(rdint(b))
	s.Symid = int16(rdint(b))
	s.Dynid = int(rdint(b))
	s.Sig = int(rdint(b))
	s.Plt = int(rdint(b))
	s.Got = int(rdint(b))
	s.Align = int(rdint(b))
	s.Elfsym = int(rdint(b))
	s.Args = int(rdint(b))
	s.Locals = rdint(b)
	s.Value = rdint(b)
	s.Size = rdint(b)
	s.Hash = rdsym(b)
	s.Allsym = rdsym(b)
	s.Next = rdsym(b)
	s.Sub = rdsym(b)
	s.Outer = rdsym(b)
	s.Gotype = rdsym(b)
	s.Reachparent = rdsym(b)
	s.Queue = rdsym(b)
	s.File = rdstring(b)
	s.Dynimplib = rdstring(b)
	s.Dynimpvers = rdstring(b)
	s.Text = rdprog(b)
	s.Etext = rdprog(b)
	n := int(rdint(b))
	if n > 0 {
		s.P = make([]byte, n)
		io.ReadFull(b, s.P)
	}
	s.R = make([]liblink.Reloc, int(rdint(b)))
	for i := range s.R {
		r := &s.R[i]
		r.Off = rdint(b)
		r.Siz = uint8(rdint(b))
		r.Done = uint8(rdint(b))
		r.Typ = int(rdint(b))
		r.Add = rdint(b)
		r.Xadd = rdint(b)
		r.Sym = rdsym(b)
		r.Xsym = rdsym(b)
	}
}

func readprog(b *bufio.Reader, p *liblink.Prog) {
	if !undef[p] {
		panic("double-def")
	}
	delete(undef, p)
	p.Pc = rdint(b)
	p.Lineno = int(rdint(b))
	p.Link = rdprog(b)
	p.As = int(rdint(b))
	p.Reg = int(rdint(b))
	p.Scond = int(rdint(b))
	p.Width = int8(rdint(b))
	readaddr(b, &p.From)
	readaddr(b, &p.To)
}

func readaddr(b *bufio.Reader, a *liblink.Addr) {
	if rdstring(b) != "addr" {
		log.Fatal("out of sync")
	}
	a.Offset = rdint(b)
	a.U.Dval = rdfloat(b)
	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[i] = byte(rdint(b))
	}
	a.U.Sval = string(buf)
	a.U.Branch = rdprog(b)
	a.Sym = rdsym(b)
	a.Gotype = rdsym(b)
	a.Typ = int(rdint(b))
	a.Index = int(rdint(b))
	a.Scale = int8(rdint(b))
	a.Reg = int(rdint(b))
	a.Name = int(rdint(b))
	a.Class = int(rdint(b))
	a.Etype = uint8(rdint(b))
	a.Offset2 = int(rdint(b))
	a.Width = rdint(b)
}

func readhist(b *bufio.Reader, h *liblink.Hist) {
	if !undef[h] {
		panic("double-def")
	}
	delete(undef, h)
	h.Link = rdhist(b)
	h.Name = rdstring(b)
	h.Line = int(rdint(b))
	h.Offset = int(rdint(b))
}

func rdint(b *bufio.Reader) int64 {
	s := rdstring(b)
	x, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		panic(err)
	}
	return x
}

func rdfloat(b *bufio.Reader) float64 {
	s := rdstring(b)
	x, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		panic(err)
	}
	return math.Float64frombits(x)
}
