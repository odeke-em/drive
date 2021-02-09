package amd64

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"code.google.com/p/rsc/c2go/liblink"
)

// Inferno utils/6l/pass.c
// http://code.google.com/p/inferno-os/source/browse/utils/6l/pass.c
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
var zprg = liblink.Prog{
	Back: 2,
	As:   AGOK,
	From: liblink.Addr{
		Typ:   D_NONE,
		Index: D_NONE,
	},
	To: liblink.Addr{
		Typ:   D_NONE,
		Index: D_NONE,
	},
}

func nopout(p *liblink.Prog) {
	p.As = ANOP
	p.From.Typ = D_NONE
	p.To.Typ = D_NONE
}

func symtype(a *liblink.Addr) int {
	var t int
	t = a.Typ
	if t == D_ADDR {
		t = a.Index
	}
	return t
}

func isdata(p *liblink.Prog) int {
	return bool2int(p.As == ADATA || p.As == AGLOBL)
}

func iscall(p *liblink.Prog) int {
	return bool2int(p.As == ACALL)
}

func datasize(p *liblink.Prog) int {
	return int(p.From.Scale)
}

func textflag(p *liblink.Prog) int {
	return int(p.From.Scale)
}

func settextflag(p *liblink.Prog, f int) {
	p.From.Scale = int8(f)
}

func canuselocaltls(ctxt *liblink.Link) bool {
	switch ctxt.Headtype {
	case liblink.Hplan9,
		liblink.Hwindows:
		return false
	}
	return true
}

func progedit(ctxt *liblink.Link, p *liblink.Prog) {
	var literal string
	var s *liblink.LSym
	var q *liblink.Prog
	// Thread-local storage references use the TLS pseudo-register.
	// As a register, TLS refers to the thread-local storage base, and it
	// can only be loaded into another register:
	//
	//         MOVQ TLS, AX
	//
	// An offset from the thread-local storage base is written off(reg)(TLS*1).
	// Semantically it is off(reg), but the (TLS*1) annotation marks this as
	// indexing from the loaded TLS base. This emits a relocation so that
	// if the linker needs to adjust the offset, it can. For example:
	//
	//         MOVQ TLS, AX
	//         MOVQ 8(AX)(TLS*1), CX // load m into CX
	//
	// On systems that support direct access to the TLS memory, this
	// pair of instructions can be reduced to a direct TLS memory reference:
	//
	//         MOVQ 8(TLS), CX // load m into CX
	//
	// The 2-instruction and 1-instruction forms correspond roughly to
	// ELF TLS initial exec mode and ELF TLS local exec mode, respectively.
	//
	// We applies this rewrite on systems that support the 1-instruction form.
	// The decision is made using only the operating system (and probably
	// the -shared flag, eventually), not the link mode. If some link modes
	// on a particular operating system require the 2-instruction form,
	// then all builds for that operating system will use the 2-instruction
	// form, so that the link mode decision can be delayed to link time.
	//
	// In this way, all supported systems use identical instructions to
	// access TLS, and they are rewritten appropriately first here in
	// liblink and then finally using relocations in the linker.
	if canuselocaltls(ctxt) {
		// Reduce TLS initial exec model to TLS local exec model.
		// Sequences like
		//	MOVQ TLS, BX
		//	... off(BX)(TLS*1) ...
		// become
		//	NOP
		//	... off(TLS) ...
		//
		// TODO(rsc): Remove the Hsolaris special case. It exists only to
		// guarantee we are producing byte-identical binaries as before this code.
		// But it should be unnecessary.
		if (p.As == AMOVQ || p.As == AMOVL) && p.From.Typ == D_TLS && D_AX <= p.To.Typ && p.To.Typ <= D_R15 && ctxt.Headtype != liblink.Hsolaris {
			nopout(p)
		}
		if p.From.Index == D_TLS && D_INDIR+D_AX <= p.From.Typ && p.From.Typ <= D_INDIR+D_R15 {
			p.From.Typ = D_INDIR + D_TLS
			p.From.Scale = 0
			p.From.Index = D_NONE
		}
		if p.To.Index == D_TLS && D_INDIR+D_AX <= p.To.Typ && p.To.Typ <= D_INDIR+D_R15 {
			p.To.Typ = D_INDIR + D_TLS
			p.To.Scale = 0
			p.To.Index = D_NONE
		}
	} else {
		// As a courtesy to the C compilers, rewrite TLS local exec load as TLS initial exec load.
		// The instruction
		//	MOVQ off(TLS), BX
		// becomes the sequence
		//	MOVQ TLS, BX
		//	MOVQ off(BX)(TLS*1), BX
		// This allows the C compilers to emit references to m and g using the direct off(TLS) form.
		if (p.As == AMOVQ || p.As == AMOVL) && p.From.Typ == D_INDIR+D_TLS && D_AX <= p.To.Typ && p.To.Typ <= D_R15 {
			q = liblink.Appendp(ctxt, p)
			q.As = p.As
			q.From = p.From
			q.From.Typ = D_INDIR + p.To.Typ
			q.From.Index = D_TLS
			q.From.Scale = 2 // TODO: use 1
			q.To = p.To
			p.From.Typ = D_TLS
			p.From.Index = D_NONE
			p.From.Offset = 0
		}
	}
	// TODO: Remove.
	if ctxt.Headtype == liblink.Hwindows || ctxt.Headtype == liblink.Hplan9 {
		if p.From.Scale == 1 && p.From.Index == D_TLS {
			p.From.Scale = 2
		}
		if p.To.Scale == 1 && p.To.Index == D_TLS {
			p.To.Scale = 2
		}
	}
	if ctxt.Headtype == liblink.Hnacl {
		nacladdr(ctxt, p, &p.From)
		nacladdr(ctxt, p, &p.To)
	}
	// Maintain information about code generation mode.
	if ctxt.Mode == 0 {
		ctxt.Mode = 64
	}
	p.Mode = ctxt.Mode
	switch p.As {
	case AMODE:
		if p.From.Typ == D_CONST || p.From.Typ == D_INDIR+D_NONE {
			switch int(p.From.Offset) {
			case 16,
				32,
				64:
				ctxt.Mode = int(p.From.Offset)
				break
			}
		}
		nopout(p)
		break
	}
	// Rewrite CALL/JMP/RET to symbol as D_BRANCH.
	switch p.As {
	case ACALL,
		AJMP,
		ARET:
		if (p.To.Typ == D_EXTERN || p.To.Typ == D_STATIC) && p.To.Sym != nil {
			p.To.Typ = D_BRANCH
		}
		break
	}
	// Rewrite float constants to values stored in memory.
	switch p.As {
	case AFMOVF,
		AFADDF,
		AFSUBF,
		AFSUBRF,
		AFMULF,
		AFDIVF,
		AFDIVRF,
		AFCOMF,
		AFCOMFP,
		AMOVSS,
		AADDSS,
		ASUBSS,
		AMULSS,
		ADIVSS,
		ACOMISS,
		AUCOMISS:
		if p.From.Typ == D_FCONST {
			var i32 uint32
			var f32 float32
			f32 = float32(p.From.U.Dval)
			i32 = math.Float32bits(f32)
			literal = fmt.Sprintf("$f32.%08x", uint32(i32))
			s = liblink.Linklookup(ctxt, literal, 0)
			if s.Typ == 0 {
				s.Typ = liblink.SRODATA
				liblink.Adduint32(ctxt, s, i32)
				s.Reachable = 0
			}
			p.From.Typ = D_EXTERN
			p.From.Sym = s
			p.From.Offset = 0
		}
	case AFMOVD,
		AFADDD,
		AFSUBD,
		AFSUBRD,
		AFMULD,
		AFDIVD,
		AFDIVRD,
		AFCOMD,
		AFCOMDP,
		AMOVSD,
		AADDSD,
		ASUBSD,
		AMULSD,
		ADIVSD,
		ACOMISD,
		AUCOMISD:
		if p.From.Typ == D_FCONST {
			var i64 uint64
			i64 = math.Float64bits(p.From.U.Dval)
			literal = fmt.Sprintf("$f64.%016x", uint64(i64))
			s = liblink.Linklookup(ctxt, literal, 0)
			if s.Typ == 0 {
				s.Typ = liblink.SRODATA
				liblink.Adduint64(ctxt, s, i64)
				s.Reachable = 0
			}
			p.From.Typ = D_EXTERN
			p.From.Sym = s
			p.From.Offset = 0
		}
		break
	}
}

func nacladdr(ctxt *liblink.Link, p *liblink.Prog, a *liblink.Addr) {
	if p.As == ALEAL || p.As == ALEAQ {
		return
	}
	if a.Typ == D_BP || a.Typ == D_INDIR+D_BP {
		ctxt.Diag("invalid address: %P", p)
		return
	}
	if a.Typ == D_INDIR+D_TLS {
		a.Typ = D_INDIR + D_BP
	} else if a.Typ == D_TLS {
		a.Typ = D_BP
	}
	if D_INDIR <= a.Typ && a.Typ <= D_INDIR+D_INDIR {
		switch a.Typ {
		// all ok
		case D_INDIR + D_BP,
			D_INDIR + D_SP,
			D_INDIR + D_R15:
			break
		default:
			if a.Index != D_NONE {
				ctxt.Diag("invalid address %P", p)
			}
			a.Index = a.Typ - D_INDIR
			if a.Index != D_NONE {
				a.Scale = 1
			}
			a.Typ = D_INDIR + D_R15
			break
		}
	}
}

var morename = []string{
	"runtime.morestack00",
	"runtime.morestack00_noctxt",
	"runtime.morestack10",
	"runtime.morestack10_noctxt",
	"runtime.morestack01",
	"runtime.morestack01_noctxt",
	"runtime.morestack11",
	"runtime.morestack11_noctxt",
	"runtime.morestack8",
	"runtime.morestack8_noctxt",
	"runtime.morestack16",
	"runtime.morestack16_noctxt",
	"runtime.morestack24",
	"runtime.morestack24_noctxt",
	"runtime.morestack32",
	"runtime.morestack32_noctxt",
	"runtime.morestack40",
	"runtime.morestack40_noctxt",
	"runtime.morestack48",
	"runtime.morestack48_noctxt",
}

func parsetextconst(arg int64, textstksiz *int64, textarg *int64) {
	*textstksiz = arg & 0xffffffff
	if *textstksiz&0x80000000 != 0 {
		*textstksiz = -(-*textstksiz & 0xffffffff)
	}
	*textarg = (arg >> 32) & 0xffffffff
	if *textarg&0x80000000 != 0 {
		*textarg = 0
	}
	*textarg = (*textarg + 7) &^ 7
}

func addstacksplit(ctxt *liblink.Link, cursym *liblink.LSym) {
	var p *liblink.Prog
	var q *liblink.Prog
	var q1 *liblink.Prog
	var autoffset int64
	var deltasp int64
	var a int
	var pcsize int
	var i uint32
	var textstksiz int64
	var textarg int64
	if ctxt.Tlsg == nil {
		ctxt.Tlsg = liblink.Linklookup(ctxt, "runtime.tlsg", 0)
	}
	if ctxt.Symmorestack[0] == nil {
		if len(morename) > len(ctxt.Symmorestack) {
			log.Fatalf("Link.symmorestack needs at least %d elements", len(morename))
		}
		for i = 0; i < uint32(len(morename)); i++ {
			ctxt.Symmorestack[i] = liblink.Linklookup(ctxt, morename[i], 0)
		}
	}
	if ctxt.Headtype == liblink.Hplan9 && ctxt.Plan9privates == nil {
		ctxt.Plan9privates = liblink.Linklookup(ctxt, "_privates", 0)
	}
	ctxt.Cursym = cursym
	if cursym.Text == nil || cursym.Text.Link == nil {
		return
	}
	p = cursym.Text
	parsetextconst(p.To.Offset, &textstksiz, &textarg)
	autoffset = textstksiz
	if autoffset < 0 {
		autoffset = 0
	}
	cursym.Args = int(p.To.Offset >> 32)
	cursym.Locals = textstksiz
	if autoffset < liblink.StackSmall && p.From.Scale&liblink.NOSPLIT == 0 {
		for q = p; q != nil; q = q.Link {
			if q.As == ACALL {
				goto noleaf
			}
			if (q.As == ADUFFCOPY || q.As == ADUFFZERO) && autoffset >= liblink.StackSmall-8 {
				goto noleaf
			}
		}
		p.From.Scale |= liblink.NOSPLIT
	noleaf:
	}
	q = nil
	if p.From.Scale&liblink.NOSPLIT == 0 || (p.From.Scale&liblink.WRAPPER != 0) {
		p = liblink.Appendp(ctxt, p)
		p = load_g_cx(ctxt, p) // load g into CX
	}
	if cursym.Text.From.Scale&liblink.NOSPLIT == 0 {
		p = stacksplit(ctxt, p, autoffset, textarg, bool2int(cursym.Text.From.Scale&liblink.NEEDCTXT == 0), &q) // emit split check
	}
	if autoffset != 0 {
		if autoffset%int64(ctxt.Arch.Regsize) != 0 {
			ctxt.Diag("unaligned stack size %d", autoffset)
		}
		p = liblink.Appendp(ctxt, p)
		p.As = AADJSP
		p.From.Typ = D_CONST
		p.From.Offset = autoffset
		p.Spadj = autoffset
	} else {
		// zero-byte stack adjustment.
		// Insert a fake non-zero adjustment so that stkcheck can
		// recognize the end of the stack-splitting prolog.
		p = liblink.Appendp(ctxt, p)
		p.As = ANOP
		p.Spadj = -ctxt.Arch.Ptrsize
		p = liblink.Appendp(ctxt, p)
		p.As = ANOP
		p.Spadj = ctxt.Arch.Ptrsize
	}
	if q != nil {
		q.Pcond = p
	}
	deltasp = autoffset
	if cursym.Text.From.Scale&liblink.WRAPPER != 0 {
		// g->panicwrap += autoffset + ctxt->arch->regsize;
		p = liblink.Appendp(ctxt, p)
		p.As = AADDL
		p.From.Typ = D_CONST
		p.From.Offset = autoffset + int64(ctxt.Arch.Regsize)
		indir_cx(ctxt, &p.To)
		p.To.Offset = 2 * ctxt.Arch.Ptrsize
	}
	if ctxt.Debugstack > 1 && autoffset != 0 {
		// 6l -K -K means double-check for stack overflow
		// even after calling morestack and even if the
		// function is marked as nosplit.
		p = liblink.Appendp(ctxt, p)
		p.As = AMOVQ
		indir_cx(ctxt, &p.From)
		p.From.Offset = 0
		p.To.Typ = D_BX
		p = liblink.Appendp(ctxt, p)
		p.As = ASUBQ
		p.From.Typ = D_CONST
		p.From.Offset = liblink.StackSmall + 32
		p.To.Typ = D_BX
		p = liblink.Appendp(ctxt, p)
		p.As = ACMPQ
		p.From.Typ = D_SP
		p.To.Typ = D_BX
		p = liblink.Appendp(ctxt, p)
		p.As = AJHI
		p.To.Typ = D_BRANCH
		q1 = p
		p = liblink.Appendp(ctxt, p)
		p.As = AINT
		p.From.Typ = D_CONST
		p.From.Offset = 3
		p = liblink.Appendp(ctxt, p)
		p.As = ANOP
		q1.Pcond = p
	}
	if ctxt.Debugzerostack != 0 && autoffset != 0 && cursym.Text.From.Scale&liblink.NOSPLIT == 0 {
		// 6l -Z means zero the stack frame on entry.
		// This slows down function calls but can help avoid
		// false positives in garbage collection.
		p = liblink.Appendp(ctxt, p)
		p.As = AMOVQ
		p.From.Typ = D_SP
		p.To.Typ = D_DI
		p = liblink.Appendp(ctxt, p)
		p.As = AMOVQ
		p.From.Typ = D_CONST
		p.From.Offset = autoffset / 8
		p.To.Typ = D_CX
		p = liblink.Appendp(ctxt, p)
		p.As = AMOVQ
		p.From.Typ = D_CONST
		p.From.Offset = 0
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = AREP
		p = liblink.Appendp(ctxt, p)
		p.As = ASTOSQ
	}
	for ; p != nil; p = p.Link {
		pcsize = p.Mode / 8
		a = p.From.Typ
		if a == D_AUTO {
			p.From.Offset += deltasp
		}
		if a == D_PARAM {
			p.From.Offset += deltasp + int64(pcsize)
		}
		a = p.To.Typ
		if a == D_AUTO {
			p.To.Offset += deltasp
		}
		if a == D_PARAM {
			p.To.Offset += deltasp + int64(pcsize)
		}
		switch p.As {
		default:
			continue
		case APUSHL,
			APUSHFL:
			deltasp += 4
			p.Spadj = 4
			continue
		case APUSHQ,
			APUSHFQ:
			deltasp += 8
			p.Spadj = 8
			continue
		case APUSHW,
			APUSHFW:
			deltasp += 2
			p.Spadj = 2
			continue
		case APOPL,
			APOPFL:
			deltasp -= 4
			p.Spadj = -4
			continue
		case APOPQ,
			APOPFQ:
			deltasp -= 8
			p.Spadj = -8
			continue
		case APOPW,
			APOPFW:
			deltasp -= 2
			p.Spadj = -2
			continue
		case ARET:
			break
		}
		if autoffset != deltasp {
			ctxt.Diag("unbalanced PUSH/POP")
		}
		if cursym.Text.From.Scale&liblink.WRAPPER != 0 {
			p = load_g_cx(ctxt, p)
			p = liblink.Appendp(ctxt, p)
			// g->panicwrap -= autoffset + ctxt->arch->regsize;
			p.As = ASUBL
			p.From.Typ = D_CONST
			p.From.Offset = autoffset + int64(ctxt.Arch.Regsize)
			indir_cx(ctxt, &p.To)
			p.To.Offset = 2 * ctxt.Arch.Ptrsize
			p = liblink.Appendp(ctxt, p)
			p.As = ARET
		}
		if autoffset != 0 {
			p.As = AADJSP
			p.From.Typ = D_CONST
			p.From.Offset = -autoffset
			p.Spadj = -autoffset
			p = liblink.Appendp(ctxt, p)
			p.As = ARET
			// If there are instructions following
			// this ARET, they come from a branch
			// with the same stackframe, so undo
			// the cleanup.
			p.Spadj = +autoffset
		}
		if p.To.Sym != nil { // retjmp
			p.As = AJMP
		}
	}
}

func indir_cx(ctxt *liblink.Link, a *liblink.Addr) {
	if ctxt.Headtype == liblink.Hnacl {
		a.Typ = D_INDIR + D_R15
		a.Index = D_CX
		a.Scale = 1
		return
	}
	a.Typ = D_INDIR + D_CX
}

// Append code to p to load g into cx.
// Overwrites p with the first instruction (no first appendp).
// Overwriting p is unusual but it lets use this in both the
// prologue (caller must call appendp first) and in the epilogue.
// Returns last new instruction.
func load_g_cx(ctxt *liblink.Link, p *liblink.Prog) *liblink.Prog {
	var next *liblink.Prog
	p.As = AMOVQ
	if ctxt.Arch.Ptrsize == 4 {
		p.As = AMOVL
	}
	p.From.Typ = D_INDIR + D_TLS
	p.From.Offset = 0
	p.To.Typ = D_CX
	next = p.Link
	progedit(ctxt, p)
	for p.Link != next {
		p = p.Link
	}
	if p.From.Index == D_TLS {
		p.From.Scale = 2
	}
	return p
}

// Append code to p to check for stack split.
// Appends to (does not overwrite) p.
// Assumes g is in CX.
// Returns last new instruction.
// On return, *jmpok is the instruction that should jump
// to the stack frame allocation if no split is needed.
func stacksplit(ctxt *liblink.Link, p *liblink.Prog, framesize int64, textarg int64, noctxt int, jmpok **liblink.Prog) *liblink.Prog {
	var q *liblink.Prog
	var q1 *liblink.Prog
	var moreconst1 int64
	var moreconst2 int64
	var i uint32
	var cmp int
	var lea int
	var mov int
	var sub int
	cmp = ACMPQ
	lea = ALEAQ
	mov = AMOVQ
	sub = ASUBQ
	if ctxt.Headtype == liblink.Hnacl {
		cmp = ACMPL
		lea = ALEAL
		mov = AMOVL
		sub = ASUBL
	}
	if ctxt.Debugstack != 0 {
		// 6l -K means check not only for stack
		// overflow but stack underflow.
		// On underflow, INT 3 (breakpoint).
		// Underflow itself is rare but this also
		// catches out-of-sync stack guard info
		p = liblink.Appendp(ctxt, p)
		p.As = cmp
		indir_cx(ctxt, &p.From)
		p.From.Offset = 8
		p.To.Typ = D_SP
		p = liblink.Appendp(ctxt, p)
		p.As = AJHI
		p.To.Typ = D_BRANCH
		p.To.Offset = 4
		q1 = p
		p = liblink.Appendp(ctxt, p)
		p.As = AINT
		p.From.Typ = D_CONST
		p.From.Offset = 3
		p = liblink.Appendp(ctxt, p)
		p.As = ANOP
		q1.Pcond = p
	}
	q1 = nil
	if framesize <= liblink.StackSmall {
		// small stack: SP <= stackguard
		//	CMPQ SP, stackguard
		p = liblink.Appendp(ctxt, p)
		p.As = cmp
		p.From.Typ = D_SP
		indir_cx(ctxt, &p.To)
	} else if framesize <= liblink.StackBig {
		// large stack: SP-framesize <= stackguard-StackSmall
		//	LEAQ -xxx(SP), AX
		//	CMPQ AX, stackguard
		p = liblink.Appendp(ctxt, p)
		p.As = lea
		p.From.Typ = D_INDIR + D_SP
		p.From.Offset = -(framesize - liblink.StackSmall)
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = cmp
		p.From.Typ = D_AX
		indir_cx(ctxt, &p.To)
	} else {
		// Such a large stack we need to protect against wraparound.
		// If SP is close to zero:
		//	SP-stackguard+StackGuard <= framesize + (StackGuard-StackSmall)
		// The +StackGuard on both sides is required to keep the left side positive:
		// SP is allowed to be slightly below stackguard. See stack.h.
		//
		// Preemption sets stackguard to StackPreempt, a very large value.
		// That breaks the math above, so we have to check for that explicitly.
		//	MOVQ	stackguard, CX
		//	CMPQ	CX, $StackPreempt
		//	JEQ	label-of-call-to-morestack
		//	LEAQ	StackGuard(SP), AX
		//	SUBQ	CX, AX
		//	CMPQ	AX, $(framesize+(StackGuard-StackSmall))
		p = liblink.Appendp(ctxt, p)
		p.As = mov
		indir_cx(ctxt, &p.From)
		p.From.Offset = 0
		p.To.Typ = D_SI
		p = liblink.Appendp(ctxt, p)
		p.As = cmp
		p.From.Typ = D_SI
		p.To.Typ = D_CONST
		p.To.Offset = liblink.StackPreempt
		p = liblink.Appendp(ctxt, p)
		p.As = AJEQ
		p.To.Typ = D_BRANCH
		q1 = p
		p = liblink.Appendp(ctxt, p)
		p.As = lea
		p.From.Typ = D_INDIR + D_SP
		p.From.Offset = liblink.StackGuard
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = sub
		p.From.Typ = D_SI
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = cmp
		p.From.Typ = D_AX
		p.To.Typ = D_CONST
		p.To.Offset = framesize + (liblink.StackGuard - liblink.StackSmall)
	}
	// common
	p = liblink.Appendp(ctxt, p)
	p.As = AJHI
	p.To.Typ = D_BRANCH
	q = p
	// If we ask for more stack, we'll get a minimum of StackMin bytes.
	// We need a stack frame large enough to hold the top-of-stack data,
	// the function arguments+results, our caller's PC, our frame,
	// a word for the return PC of the next call, and then the StackLimit bytes
	// that must be available on entry to any function called from a function
	// that did a stack check.  If StackMin is enough, don't ask for a specific
	// amount: then we can use the custom functions and save a few
	// instructions.
	moreconst1 = 0
	if liblink.StackTop+textarg+ctxt.Arch.Ptrsize+framesize+ctxt.Arch.Ptrsize+liblink.StackLimit >= liblink.StackMin {
		moreconst1 = framesize
	}
	moreconst2 = textarg
	if moreconst2 == 1 { // special marker
		moreconst2 = 0
	}
	if moreconst2&7 != 0 {
		ctxt.Diag("misaligned argument size in stack split")
	}
	// 4 varieties varieties (const1==0 cross const2==0)
	// and 6 subvarieties of (const1==0 and const2!=0)
	p = liblink.Appendp(ctxt, p)
	if moreconst1 == 0 && moreconst2 == 0 {
		p.As = ACALL
		p.To.Typ = D_BRANCH
		p.To.Sym = ctxt.Symmorestack[0*2+noctxt]
	} else if moreconst1 != 0 && moreconst2 == 0 {
		p.As = AMOVL
		p.From.Typ = D_CONST
		p.From.Offset = moreconst1
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = ACALL
		p.To.Typ = D_BRANCH
		p.To.Sym = ctxt.Symmorestack[1*2+noctxt]
	} else if moreconst1 == 0 && moreconst2 <= 48 && moreconst2%8 == 0 {
		i = uint32(moreconst2/8 + 3)
		p.As = ACALL
		p.To.Typ = D_BRANCH
		p.To.Sym = ctxt.Symmorestack[i*2+uint32(noctxt)]
	} else if moreconst1 == 0 && moreconst2 != 0 {
		p.As = AMOVL
		p.From.Typ = D_CONST
		p.From.Offset = moreconst2
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = ACALL
		p.To.Typ = D_BRANCH
		p.To.Sym = ctxt.Symmorestack[2*2+noctxt]
	} else {
		// Pass framesize and argsize.
		p.As = AMOVQ
		p.From.Typ = D_CONST
		p.From.Offset = int64(uint64(moreconst2) << 32)
		p.From.Offset |= moreconst1
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = ACALL
		p.To.Typ = D_BRANCH
		p.To.Sym = ctxt.Symmorestack[3*2+noctxt]
	}
	p = liblink.Appendp(ctxt, p)
	p.As = AJMP
	p.To.Typ = D_BRANCH
	p.Pcond = ctxt.Cursym.Text.Link
	if q != nil {
		q.Pcond = p.Link
	}
	if q1 != nil {
		q1.Pcond = q.Link
	}
	*jmpok = q
	return p
}

func follow(ctxt *liblink.Link, s *liblink.LSym) {
	var firstp *liblink.Prog
	var lastp *liblink.Prog
	ctxt.Cursym = s
	firstp = ctxt.Prg()
	lastp = firstp
	xfol(ctxt, s.Text, &lastp)
	lastp.Link = nil
	s.Text = firstp.Link
}

func nofollow(a int) bool {
	switch a {
	case AJMP,
		ARET,
		AIRETL,
		AIRETQ,
		AIRETW,
		ARETFL,
		ARETFQ,
		ARETFW,
		AUNDEF:
		return true
	}
	return false
}

func pushpop(a int) int {
	switch a {
	case APUSHL,
		APUSHFL,
		APUSHQ,
		APUSHFQ,
		APUSHW,
		APUSHFW,
		APOPL,
		APOPFL,
		APOPQ,
		APOPFQ,
		APOPW,
		APOPFW:
		return 1
	}
	return 0
}

func relinv(a int) int {
	switch a {
	case AJEQ:
		return AJNE
	case AJNE:
		return AJEQ
	case AJLE:
		return AJGT
	case AJLS:
		return AJHI
	case AJLT:
		return AJGE
	case AJMI:
		return AJPL
	case AJGE:
		return AJLT
	case AJPL:
		return AJMI
	case AJGT:
		return AJLE
	case AJHI:
		return AJLS
	case AJCS:
		return AJCC
	case AJCC:
		return AJCS
	case AJPS:
		return AJPC
	case AJPC:
		return AJPS
	case AJOS:
		return AJOC
	case AJOC:
		return AJOS
	}
	log.Fatalf("unknown relation: %s", Anames6[a])
	return 0
}

func xfol(ctxt *liblink.Link, p *liblink.Prog, last **liblink.Prog) {
	var q *liblink.Prog
	var i int
	var a int
loop:
	if p == nil {
		return
	}
	if p.As == AJMP {
		q = p.Pcond
		if q != nil && q.As != ATEXT {
			/* mark instruction as done and continue layout at target of jump */
			p.Mark = 1
			p = q
			if p.Mark == 0 {
				goto loop
			}
		}
	}
	if p.Mark != 0 {
		/*
		 * p goes here, but already used it elsewhere.
		 * copy up to 4 instructions or else branch to other copy.
		 */
		i = 0
		q = p
		for ; i < 4; (func() { i++; q = q.Link })() {
			if q == nil {
				break
			}
			if q == *last {
				break
			}
			a = q.As
			if a == ANOP {
				i--
				continue
			}
			if nofollow(a) || pushpop(a) != 0 {
				break // NOTE(rsc): arm does goto copy
			}
			if q.Pcond == nil || q.Pcond.Mark != 0 {
				continue
			}
			if a == ACALL || a == ALOOP {
				continue
			}
			for {
				if p.As == ANOP {
					p = p.Link
					continue
				}
				q = liblink.Copyp(ctxt, p)
				p = p.Link
				q.Mark = 1
				(*last).Link = q
				*last = q
				if q.As != a || q.Pcond == nil || q.Pcond.Mark != 0 {
					continue
				}
				q.As = relinv(q.As)
				p = q.Pcond
				q.Pcond = q.Link
				q.Link = p
				xfol(ctxt, q.Link, last)
				p = q.Link
				if p.Mark != 0 {
					return
				}
				goto loop /* */
			}
		}
		q = ctxt.Prg()
		q.As = AJMP
		q.Lineno = p.Lineno
		q.To.Typ = D_BRANCH
		q.To.Offset = p.Pc
		q.Pcond = p
		p = q
	}
	/* emit p */
	p.Mark = 1
	(*last).Link = p
	*last = p
	a = p.As
	/* continue loop with what comes after p */
	if nofollow(a) {
		return
	}
	if p.Pcond != nil && a != ACALL {
		/*
		 * some kind of conditional branch.
		 * recurse to follow one path.
		 * continue loop on the other.
		 */
		q = liblink.Brchain(ctxt, p.Pcond)
		if q != nil {
			p.Pcond = q
		}
		q = liblink.Brchain(ctxt, p.Link)
		if q != nil {
			p.Link = q
		}
		if p.From.Typ == D_CONST {
			if p.From.Offset == 1 {
				/*
				 * expect conditional jump to be taken.
				 * rewrite so that's the fall-through case.
				 */
				p.As = relinv(a)
				q = p.Link
				p.Link = p.Pcond
				p.Pcond = q
			}
		} else {
			q = p.Link
			if q.Mark != 0 {
				if a != ALOOP {
					p.As = relinv(a)
					p.Link = p.Pcond
					p.Pcond = q
				}
			}
		}
		xfol(ctxt, p.Link, last)
		if p.Pcond.Mark != 0 {
			return
		}
		p = p.Pcond
		goto loop
	}
	p = p.Link
	goto loop
}

func prg() *liblink.Prog {
	var p *liblink.Prog
	p = new(liblink.Prog)
	*p = zprg
	return p
}

var Linkamd64 = liblink.LinkArch{
	Name:          "amd64",
	Thechar:       '6',
	ByteOrder:     binary.LittleEndian,
	Pconv:         Pconv,
	Addstacksplit: addstacksplit,
	Assemble:      span6,
	Datasize:      datasize,
	Follow:        follow,
	Iscall:        iscall,
	Isdata:        isdata,
	Prg:           prg,
	Progedit:      progedit,
	Settextflag:   settextflag,
	Symtype:       symtype,
	Textflag:      textflag,
	Minlc:         1,
	Ptrsize:       8,
	Regsize:       8,
	D_ADDR:        D_ADDR,
	D_AUTO:        D_AUTO,
	D_BRANCH:      D_BRANCH,
	D_CONST:       D_CONST,
	D_EXTERN:      D_EXTERN,
	D_FCONST:      D_FCONST,
	D_NONE:        D_NONE,
	D_PARAM:       D_PARAM,
	D_SCONST:      D_SCONST,
	D_STATIC:      D_STATIC,
	ACALL:         ACALL,
	ADATA:         ADATA,
	AEND:          AEND,
	AFUNCDATA:     AFUNCDATA,
	AGLOBL:        AGLOBL,
	AJMP:          AJMP,
	ANOP:          ANOP,
	APCDATA:       APCDATA,
	ARET:          ARET,
	ATEXT:         ATEXT,
	ATYPE:         ATYPE,
	AUSEFIELD:     AUSEFIELD,
}

var Linkamd64p32 = liblink.LinkArch{
	Name:          "amd64p32",
	Thechar:       '6',
	ByteOrder:     binary.LittleEndian,
	Pconv:         Pconv,
	Addstacksplit: addstacksplit,
	Assemble:      span6,
	Datasize:      datasize,
	Follow:        follow,
	Iscall:        iscall,
	Isdata:        isdata,
	Prg:           prg,
	Progedit:      progedit,
	Settextflag:   settextflag,
	Symtype:       symtype,
	Textflag:      textflag,
	Minlc:         1,
	Ptrsize:       4,
	Regsize:       8,
	D_ADDR:        D_ADDR,
	D_AUTO:        D_AUTO,
	D_BRANCH:      D_BRANCH,
	D_CONST:       D_CONST,
	D_EXTERN:      D_EXTERN,
	D_FCONST:      D_FCONST,
	D_NONE:        D_NONE,
	D_PARAM:       D_PARAM,
	D_SCONST:      D_SCONST,
	D_STATIC:      D_STATIC,
	ACALL:         ACALL,
	ADATA:         ADATA,
	AEND:          AEND,
	AFUNCDATA:     AFUNCDATA,
	AGLOBL:        AGLOBL,
	AJMP:          AJMP,
	ANOP:          ANOP,
	APCDATA:       APCDATA,
	ARET:          ARET,
	ATEXT:         ATEXT,
	ATYPE:         ATYPE,
	AUSEFIELD:     AUSEFIELD,
}
