package liblink

import (
	"encoding/binary"
	"fmt"
)

// Derived from Inferno utils/6l/l.h and related files.
// http://code.google.com/p/inferno-os/source/browse/utils/6l/l.h
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
type Addr struct {
	Offset int64
	U      struct {
		Sval   string
		Dval   float64
		Branch *Prog
	}
	Sym     *LSym
	Gotype  *LSym
	Typ     int
	Index   int
	Scale   int8
	Reg     int
	Name    int
	Class   int
	Etype   uint8
	Offset2 int
	Node    *struct{}
	Width   int64
}

type Prog struct {
	Ctxt     *Link
	Pc       int64
	Lineno   int
	Link     *Prog
	As       int
	Reg      int
	Scond    int
	From     Addr
	To       Addr
	Opt      *struct{}
	Forwd    *Prog
	Pcond    *Prog
	Comefrom *Prog
	Pcrel    *Prog
	Spadj    int64
	Mark     int
	Back     int
	Ft       uint8
	Tt       uint8
	Optab    int
	Isize    int
	Printed  uint8
	Width    int8
	Mode     int
	TEXTFLAG uint8
}

func (p *Prog) Line() string {
	return linklinefmt(p.Ctxt, int(p.Lineno), false, false)
}

func (p *Prog) String() string {
	if p.Ctxt == nil {
		return fmt.Sprintf("<Prog without ctxt>")
	}
	return p.Ctxt.Arch.Pconv(p)
}

type LSym struct {
	Name        string
	Extname     string
	Typ         int
	Version     uint32
	Dupok       int
	External    uint8
	Nosplit     uint8
	Reachable   uint8
	Cgoexport   uint8
	Special     uint8
	Stkcheck    uint8
	Hide        uint8
	Leaf        uint8
	Fnptr       uint8
	Seenglobl   uint8
	Onlist      uint8
	Printed     uint8
	Symid       int16
	Dynid       int
	Sig         int
	Plt         int
	Got         int
	Align       int
	Elfsym      int
	Args        int
	Locals      int64
	Value       int64
	Size        int64
	Hash        *LSym
	Allsym      *LSym
	Next        *LSym
	Sub         *LSym
	Outer       *LSym
	Gotype      *LSym
	Reachparent *LSym
	Queue       *LSym
	File        string
	Dynimplib   string
	Dynimpvers  string
	Sect        *struct{}
	Autom       *Auto
	Text        *Prog
	Etext       *Prog
	Pcln        *Pcln
	P           []uint8
	R           []Reloc
}

type Reloc struct {
	Off  int64
	Siz  uint8
	Done uint8
	Typ  int
	Add  int64
	Xadd int64
	Sym  *LSym
	Xsym *LSym
}

type Auto struct {
	Asym    *LSym
	Link    *Auto
	Aoffset int
	Typ     int
	Gotype  *LSym
}

type Hist struct {
	Link   *Hist
	Name   string
	Line   int
	Offset int
}

type Link struct {
	Thechar        int
	Thestring      string
	Goarm          int
	Headtype       int
	Arch           *LinkArch
	Ignore         func(string) int
	Debugasm       int
	Debugline      int
	Debughist      int
	Debugread      int
	Debugvlog      int
	Debugstack     int
	Debugzerostack int
	Debugdivmod    int
	Debugfloat     int
	Debugpcln      int
	Flag_shared    int
	Iself          int
	Bso            *Biobuf
	Pathname       string
	Windows        int
	Trimpath       string
	Goroot         string
	Goroot_final   string
	Hash           [LINKHASH]*LSym
	Allsym         *LSym
	Nsymbol        int
	Hist           *Hist
	Ehist          *Hist
	Plist          *Plist
	Plast          *Plist
	Sym_div        *LSym
	Sym_divu       *LSym
	Sym_mod        *LSym
	Sym_modu       *LSym
	Symmorestack   [20]*LSym
	Tlsg           *LSym
	Plan9privates  *LSym
	Curp           *Prog
	Printp         *Prog
	Blitrl         *Prog
	Elitrl         *Prog
	Rexflag        int
	Rep            int
	Repn           int
	Lock           int
	Asmode         int
	Andptr         []uint8
	And            [100]uint8
	Instoffset     int
	Autosize       int
	Armsize        int
	Pc             int64
	Libdir         []string
	Library        []Library
	Tlsoffset      int
	Diag           func(string, ...interface{})
	Mode           int
	Curauto        *Auto
	Curhist        *Auto
	Cursym         *LSym
	Version        uint32
	Textp          *LSym
	Etextp         *LSym
	Histdepth      int
	Nhistfile      int
	Filesyms       *LSym
}

func (ctxt *Link) Prg() *Prog {
	p := ctxt.Arch.Prg()
	p.Ctxt = ctxt
	return p
}

type Plist struct {
	Name    *LSym
	Firstpc *Prog
	Recur   int
	Link    *Plist
}

type LinkArch struct {
	Name          string
	Thechar       int
	ByteOrder     binary.ByteOrder
	Pconv         func(*Prog) string
	Addstacksplit func(*Link, *LSym)
	Assemble      func(*Link, *LSym)
	Datasize      func(*Prog) int
	Follow        func(*Link, *LSym)
	Iscall        func(*Prog) int
	Isdata        func(*Prog) int
	Prg           func() *Prog
	Progedit      func(*Link, *Prog)
	Settextflag   func(*Prog, int)
	Symtype       func(*Addr) int
	Textflag      func(*Prog) int
	Minlc         uint32
	Ptrsize       int64
	Regsize       int
	D_ADDR        int
	D_AUTO        int
	D_BRANCH      int
	D_CONST       int
	D_EXTERN      int
	D_FCONST      int
	D_NONE        int
	D_PARAM       int
	D_SCONST      int
	D_STATIC      int
	ACALL         int
	ADATA         int
	AEND          int
	AFUNCDATA     int
	AGLOBL        int
	AJMP          int
	ANOP          int
	APCDATA       int
	ARET          int
	ATEXT         int
	ATYPE         int
	AUSEFIELD     int
}

type Library struct {
	Objref string
	Srcref string
	File   string
	Pkg    string
}

type Pcln struct {
	Pcsp        Pcdata
	Pcfile      Pcdata
	Pcline      Pcdata
	Pcdata      []Pcdata
	Funcdata    []*LSym
	Funcdataoff []int64
	Nfuncdata   int
	File        []*LSym
	Lastfile    *LSym
	Lastindex   int
}

type Pcdata struct {
	P []uint8
}

type Pciter struct {
	d       Pcdata
	p       []uint8
	pc      uint32
	nextpc  uint32
	pcscale uint32
	value   int
	start   int
	done    int
}

// prevent incompatible type signatures between liblink and 8l on Plan 9

// prevent incompatible type signatures between liblink and 8l on Plan 9

// LSym.type
const (
	Sxxx = iota
	STEXT
	SELFRXSECT
	STYPE
	SSTRING
	SGOSTRING
	SGOFUNC
	SRODATA
	SFUNCTAB
	STYPELINK
	SSYMTAB
	SPCLNTAB
	SELFROSECT
	SMACHOPLT
	SELFSECT
	SMACHO
	SMACHOGOT
	SNOPTRDATA
	SINITARR
	SDATA
	SWINDOWS
	SBSS
	SNOPTRBSS
	STLSBSS
	SXREF
	SMACHOSYMSTR
	SMACHOSYMTAB
	SMACHOINDIRECTPLT
	SMACHOINDIRECTGOT
	SFILE
	SFILEPATH
	SCONST
	SDYNIMPORT
	SHOSTOBJ
	SSUB    = 1 << 8
	SMASK   = SSUB - 1
	SHIDDEN = 1 << 9
)

// Reloc.type
const (
	R_ADDR = 1 + iota
	R_SIZE
	R_CALL
	R_CALLARM
	R_CALLIND
	R_CONST
	R_PCREL
	R_TLS
	R_TLS_LE
	R_TLS_IE
	R_GOTOFF
	R_PLT0
	R_PLT1
	R_PLT2
	R_USEFIELD
)

// Auto.type
const (
	A_AUTO = 1 + iota
	A_PARAM
)

const (
	LINKHASH = 100003
)

// Pcdata iterator.
//	for(pciterinit(ctxt, &it, &pcd); !it.done; pciternext(&it)) { it.value holds in [it.pc, it.nextpc) }

// symbol version, incremented each time a file is loaded.
// version==1 is reserved for savehist.
const (
	HistVersion = 1
)

// Link holds the context for writing object code from a compiler
// to be linker input or for reading that input into the linker.

// LinkArch is the definition of a single architecture.

/* executable header types */
const (
	Hunknown = 0 + iota
	Hdarwin
	Hdragonfly
	Helf
	Hfreebsd
	Hlinux
	Hnacl
	Hnetbsd
	Hopenbsd
	Hplan9
	Hsolaris
	Hwindows
)

const (
	LinkAuto = 0 + iota
	LinkInternal
	LinkExternal
)

// asm5.c
// asm6.c
// asm8.c
// data.c
// go.c
// ld.c
