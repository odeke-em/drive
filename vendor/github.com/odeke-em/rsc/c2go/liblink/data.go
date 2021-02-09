package liblink

import (
	"log"
	"math"
)

// Derived from Inferno utils/6l/obj.c and utils/6l/span.c
// http://code.google.com/p/inferno-os/source/browse/utils/6l/obj.c
// http://code.google.com/p/inferno-os/source/browse/utils/6l/span.c
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
func mangle(file string) {
	log.Fatalf("%s: mangled input file", file)
}

func Symgrow(ctxt *Link, s *LSym, lsiz int64) {
	var siz int
	siz = int(lsiz)
	if int64(siz) != lsiz {
		sysfatal("Symgrow size %d too long", lsiz)
	}
	if len(s.P) >= siz {
		return
	}
	for cap(s.P) < siz {
		s.P = append(s.P[:cap(s.P)], 0)
	}
	s.P = s.P[:siz]
}

func savedata(ctxt *Link, s *LSym, p *Prog, pn string) {
	var off int
	var siz int
	var i int
	var o int64
	var r *Reloc
	off = int(p.From.Offset)
	siz = ctxt.Arch.Datasize(p)
	if off < 0 || siz < 0 || off >= 1<<30 || siz >= 100 {
		mangle(pn)
	}
	Symgrow(ctxt, s, int64(off+siz))
	if p.To.Typ == ctxt.Arch.D_FCONST {
		switch siz {
		default:
		case 4:
			ctxt.Arch.ByteOrder.PutUint32(s.P[off:], math.Float32bits(float32(p.To.U.Dval)))
		case 8:
			ctxt.Arch.ByteOrder.PutUint64(s.P[off:], math.Float64bits(p.To.U.Dval))
		}
	} else {
		if p.To.Typ == ctxt.Arch.D_SCONST {
			for i = 0; i < siz; i++ {
				s.P[off+i] = uint8(p.To.U.Sval[i])
			}
		} else {
			if p.To.Typ == ctxt.Arch.D_CONST {
				if p.To.Sym != nil {
					r = Addrel(s)
					r.Off = int64(off)
					r.Siz = uint8(siz)
					r.Sym = p.To.Sym
					r.Typ = int(R_ADDR)
					r.Add = p.To.Offset
					goto out
				}
				o = p.To.Offset
				switch siz {
				default:
					ctxt.Diag("bad nuxi %d\n%v", siz, p)
					break
				case 1:
					s.P[off] = byte(o)
				case 2:
					ctxt.Arch.ByteOrder.PutUint16(s.P[off:], uint16(o))
				case 4:
					ctxt.Arch.ByteOrder.PutUint32(s.P[off:], uint32(o))
				case 8:
					ctxt.Arch.ByteOrder.PutUint64(s.P[off:], uint64(o))
				}
			} else {
				if p.To.Typ == ctxt.Arch.D_ADDR {
					r = Addrel(s)
					r.Off = int64(off)
					r.Siz = uint8(siz)
					r.Sym = p.To.Sym
					r.Typ = int(R_ADDR)
					r.Add = p.To.Offset
				} else {
					ctxt.Diag("bad data: %v", p)
				}
			}
		out:
		}
	}
}

func Addrel(s *LSym) *Reloc {
	s.R = append(s.R, Reloc{})
	return &s.R[len(s.R)-1]
}

func setuintxx(ctxt *Link, s *LSym, off int64, v uint64, wid int64) int64 {
	if s.Typ == 0 {
		s.Typ = int(SDATA)
	}
	s.Reachable = 1
	if s.Size < off+wid {
		s.Size = off + wid
		Symgrow(ctxt, s, s.Size)
	}
	switch wid {
	case 1:
		s.P[off] = uint8(v)
		break
	case 2:
		ctxt.Arch.ByteOrder.PutUint16(s.P[off:], uint16(v))
	case 4:
		ctxt.Arch.ByteOrder.PutUint32(s.P[off:], uint32(v))
	case 8:
		ctxt.Arch.ByteOrder.PutUint64(s.P[off:], uint64(v))
	}
	return off + wid
}

func adduintxx(ctxt *Link, s *LSym, v uint64, wid int64) int64 {
	var off int64
	off = s.Size
	setuintxx(ctxt, s, off, v, wid)
	return off
}

func adduint8(ctxt *Link, s *LSym, v uint8) int64 {
	return adduintxx(ctxt, s, uint64(v), 1)
}

func adduint16(ctxt *Link, s *LSym, v uint16) int64 {
	return adduintxx(ctxt, s, uint64(v), 2)
}

func Adduint32(ctxt *Link, s *LSym, v uint32) int64 {
	return adduintxx(ctxt, s, uint64(v), 4)
}

func Adduint64(ctxt *Link, s *LSym, v uint64) int64 {
	return adduintxx(ctxt, s, v, 8)
}

func setuint8(ctxt *Link, s *LSym, r int64, v uint64) int64 {
	return setuintxx(ctxt, s, r, v, 1)
}

func setuint16(ctxt *Link, s *LSym, r int64, v uint64) int64 {
	return setuintxx(ctxt, s, r, v, 2)
}

func setuint32(ctxt *Link, s *LSym, r int64, v uint64) int64 {
	return setuintxx(ctxt, s, r, v, 4)
}

func setuint64(ctxt *Link, s *LSym, r int64, v uint64) int64 {
	return setuintxx(ctxt, s, r, v, 8)
}

func addaddrplus(ctxt *Link, s *LSym, t *LSym, add int64) int64 {
	var i int64
	var r *Reloc
	if s.Typ == 0 {
		s.Typ = SDATA
	}
	s.Reachable = 1
	i = s.Size
	s.Size += ctxt.Arch.Ptrsize
	Symgrow(ctxt, s, s.Size)
	r = Addrel(s)
	r.Sym = t
	r.Off = i
	r.Siz = uint8(ctxt.Arch.Ptrsize)
	r.Typ = R_ADDR
	r.Add = add
	return i + int64(r.Siz)
}

func addpcrelplus(ctxt *Link, s *LSym, t *LSym, add int64) int64 {
	var i int64
	var r *Reloc
	if s.Typ == 0 {
		s.Typ = SDATA
	}
	s.Reachable = 1
	i = s.Size
	s.Size += 4
	Symgrow(ctxt, s, s.Size)
	r = Addrel(s)
	r.Sym = t
	r.Off = i
	r.Add = add
	r.Typ = R_PCREL
	r.Siz = 4
	return i + int64(r.Siz)
}

func addaddr(ctxt *Link, s *LSym, t *LSym) int64 {
	return addaddrplus(ctxt, s, t, 0)
}

func setaddrplus(ctxt *Link, s *LSym, off int64, t *LSym, add int64) int64 {
	var r *Reloc
	if s.Typ == 0 {
		s.Typ = SDATA
	}
	s.Reachable = 1
	if off+ctxt.Arch.Ptrsize > s.Size {
		s.Size = off + ctxt.Arch.Ptrsize
		Symgrow(ctxt, s, s.Size)
	}
	r = Addrel(s)
	r.Sym = t
	r.Off = off
	r.Siz = uint8(ctxt.Arch.Ptrsize)
	r.Typ = R_ADDR
	r.Add = add
	return off + int64(r.Siz)
}

func setaddr(ctxt *Link, s *LSym, off int64, t *LSym) int64 {
	return setaddrplus(ctxt, s, off, t, 0)
}

func addsize(ctxt *Link, s *LSym, t *LSym) int64 {
	var i int64
	var r *Reloc
	if s.Typ == 0 {
		s.Typ = SDATA
	}
	s.Reachable = 1
	i = s.Size
	s.Size += ctxt.Arch.Ptrsize
	Symgrow(ctxt, s, s.Size)
	r = Addrel(s)
	r.Sym = t
	r.Off = i
	r.Siz = uint8(ctxt.Arch.Ptrsize)
	r.Typ = R_SIZE
	return i + int64(r.Siz)
}

func addaddrplus4(ctxt *Link, s *LSym, t *LSym, add int64) int64 {
	var i int64
	var r *Reloc
	if s.Typ == 0 {
		s.Typ = SDATA
	}
	s.Reachable = 1
	i = s.Size
	s.Size += 4
	Symgrow(ctxt, s, s.Size)
	r = Addrel(s)
	r.Sym = t
	r.Off = i
	r.Siz = 4
	r.Typ = R_ADDR
	r.Add = add
	return i + int64(r.Siz)
}
