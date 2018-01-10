package x86

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"code.google.com/p/rsc/c2go/liblink"
)

// Inferno utils/8l/pass.c
// http://code.google.com/p/inferno-os/source/browse/utils/8l/pass.c
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
		Scale: 1,
	},
	To: liblink.Addr{
		Typ:   D_NONE,
		Index: D_NONE,
		Scale: 1,
	},
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
	case liblink.Hlinux,
		liblink.Hnacl,
		liblink.Hplan9,
		liblink.Hwindows:
		return false
	}
	return true
}

func progedit(ctxt *liblink.Link, p *liblink.Prog) {
	var literal string
	var s *liblink.LSym
	var q *liblink.Prog
	// See obj6.c for discussion of TLS.
	if canuselocaltls(ctxt) {
		// Reduce TLS initial exec model to TLS local exec model.
		// Sequences like
		//	MOVL TLS, BX
		//	... off(BX)(TLS*1) ...
		// become
		//	NOP
		//	... off(TLS) ...
		if p.As == AMOVL && p.From.Typ == D_TLS && D_AX <= p.To.Typ && p.To.Typ <= D_DI {
			p.As = ANOP
			p.From.Typ = D_NONE
			p.To.Typ = D_NONE
		}
		if p.From.Index == D_TLS && D_INDIR+D_AX <= p.From.Typ && p.From.Typ <= D_INDIR+D_DI {
			p.From.Typ = D_INDIR + D_TLS
			p.From.Scale = 0
			p.From.Index = D_NONE
		}
		if p.To.Index == D_TLS && D_INDIR+D_AX <= p.To.Typ && p.To.Typ <= D_INDIR+D_DI {
			p.To.Typ = D_INDIR + D_TLS
			p.To.Scale = 0
			p.To.Index = D_NONE
		}
	} else {
		// As a courtesy to the C compilers, rewrite TLS local exec load as TLS initial exec load.
		// The instruction
		//	MOVL off(TLS), BX
		// becomes the sequence
		//	MOVL TLS, BX
		//	MOVL off(BX)(TLS*1), BX
		// This allows the C compilers to emit references to m and g using the direct off(TLS) form.
		if p.As == AMOVL && p.From.Typ == D_INDIR+D_TLS && D_AX <= p.To.Typ && p.To.Typ <= D_DI {
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
	if ctxt.Headtype == liblink.Hplan9 {
		if p.From.Scale == 1 && p.From.Index == D_TLS {
			p.From.Scale = 2
		}
		if p.To.Scale == 1 && p.To.Index == D_TLS {
			p.To.Scale = 2
		}
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

func prg() *liblink.Prog {
	var p *liblink.Prog
	p = new(liblink.Prog)
	*p = zprg
	return p
}

func addstacksplit(ctxt *liblink.Link, cursym *liblink.LSym) {
	var p *liblink.Prog
	var q *liblink.Prog
	var autoffset int64
	var deltasp int64
	var a int
	if ctxt.Symmorestack[0] == nil {
		ctxt.Symmorestack[0] = liblink.Linklookup(ctxt, "runtime.morestack", 0)
		ctxt.Symmorestack[1] = liblink.Linklookup(ctxt, "runtime.morestack_noctxt", 0)
	}
	if ctxt.Headtype == liblink.Hplan9 && ctxt.Plan9privates == nil {
		ctxt.Plan9privates = liblink.Linklookup(ctxt, "_privates", 0)
	}
	ctxt.Cursym = cursym
	if cursym.Text == nil || cursym.Text.Link == nil {
		return
	}
	p = cursym.Text
	autoffset = p.To.Offset
	if autoffset < 0 {
		autoffset = 0
	}
	cursym.Locals = autoffset
	cursym.Args = p.To.Offset2
	q = nil
	if p.From.Scale&liblink.NOSPLIT == 0 || (p.From.Scale&liblink.WRAPPER != 0) {
		p = liblink.Appendp(ctxt, p)
		p = load_g_cx(ctxt, p) // load g into CX
	}
	if cursym.Text.From.Scale&liblink.NOSPLIT == 0 {
		p = stacksplit(ctxt, p, autoffset, bool2int(cursym.Text.From.Scale&liblink.NEEDCTXT == 0), &q) // emit split check
	}
	if autoffset != 0 {
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
		// g->panicwrap += autoffset + ctxt->arch->ptrsize;
		p = liblink.Appendp(ctxt, p)
		p.As = AADDL
		p.From.Typ = D_CONST
		p.From.Offset = autoffset + ctxt.Arch.Ptrsize
		p.To.Typ = D_INDIR + D_CX
		p.To.Offset = 2 * ctxt.Arch.Ptrsize
	}
	if ctxt.Debugzerostack != 0 && autoffset != 0 && cursym.Text.From.Scale&liblink.NOSPLIT == 0 {
		// 8l -Z means zero the stack frame on entry.
		// This slows down function calls but can help avoid
		// false positives in garbage collection.
		p = liblink.Appendp(ctxt, p)
		p.As = AMOVL
		p.From.Typ = D_SP
		p.To.Typ = D_DI
		p = liblink.Appendp(ctxt, p)
		p.As = AMOVL
		p.From.Typ = D_CONST
		p.From.Offset = autoffset / 4
		p.To.Typ = D_CX
		p = liblink.Appendp(ctxt, p)
		p.As = AMOVL
		p.From.Typ = D_CONST
		p.From.Offset = 0
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = AREP
		p = liblink.Appendp(ctxt, p)
		p.As = ASTOSL
	}
	for ; p != nil; p = p.Link {
		a = p.From.Typ
		if a == D_AUTO {
			p.From.Offset += deltasp
		}
		if a == D_PARAM {
			p.From.Offset += deltasp + 4
		}
		a = p.To.Typ
		if a == D_AUTO {
			p.To.Offset += deltasp
		}
		if a == D_PARAM {
			p.To.Offset += deltasp + 4
		}
		switch p.As {
		default:
			continue
		case APUSHL,
			APUSHFL:
			deltasp += 4
			p.Spadj = 4
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
			// g->panicwrap -= autoffset + ctxt->arch->ptrsize;
			p.As = ASUBL
			p.From.Typ = D_CONST
			p.From.Offset = autoffset + ctxt.Arch.Ptrsize
			p.To.Typ = D_INDIR + D_CX
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

// Append code to p to load g into cx.
// Overwrites p with the first instruction (no first appendp).
// Overwriting p is unusual but it lets use this in both the
// prologue (caller must call appendp first) and in the epilogue.
// Returns last new instruction.
func load_g_cx(ctxt *liblink.Link, p *liblink.Prog) *liblink.Prog {
	var next *liblink.Prog
	p.As = AMOVL
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
func stacksplit(ctxt *liblink.Link, p *liblink.Prog, framesize int64, noctxt int, jmpok **liblink.Prog) *liblink.Prog {
	var q *liblink.Prog
	var q1 *liblink.Prog
	var arg int
	if ctxt.Debugstack != 0 {
		// 8l -K means check not only for stack
		// overflow but stack underflow.
		// On underflow, INT 3 (breakpoint).
		// Underflow itself is rare but this also
		// catches out-of-sync stack guard info.
		p = liblink.Appendp(ctxt, p)
		p.As = ACMPL
		p.From.Typ = D_INDIR + D_CX
		p.From.Offset = 4
		p.To.Typ = D_SP
		p = liblink.Appendp(ctxt, p)
		p.As = AJCC
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
		//	CMPL SP, stackguard
		p = liblink.Appendp(ctxt, p)
		p.As = ACMPL
		p.From.Typ = D_SP
		p.To.Typ = D_INDIR + D_CX
	} else if framesize <= liblink.StackBig {
		// large stack: SP-framesize <= stackguard-StackSmall
		//	LEAL -(framesize-StackSmall)(SP), AX
		//	CMPL AX, stackguard
		p = liblink.Appendp(ctxt, p)
		p.As = ALEAL
		p.From.Typ = D_INDIR + D_SP
		p.From.Offset = -(framesize - liblink.StackSmall)
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = ACMPL
		p.From.Typ = D_AX
		p.To.Typ = D_INDIR + D_CX
	} else {
		// Such a large stack we need to protect against wraparound
		// if SP is close to zero.
		//	SP-stackguard+StackGuard <= framesize + (StackGuard-StackSmall)
		// The +StackGuard on both sides is required to keep the left side positive:
		// SP is allowed to be slightly below stackguard. See stack.h.
		//
		// Preemption sets stackguard to StackPreempt, a very large value.
		// That breaks the math above, so we have to check for that explicitly.
		//	MOVL	stackguard, CX
		//	CMPL	CX, $StackPreempt
		//	JEQ	label-of-call-to-morestack
		//	LEAL	StackGuard(SP), AX
		//	SUBL	stackguard, AX
		//	CMPL	AX, $(framesize+(StackGuard-StackSmall))
		p = liblink.Appendp(ctxt, p)
		p.As = AMOVL
		p.From.Typ = D_INDIR + D_CX
		p.From.Offset = 0
		p.To.Typ = D_SI
		p = liblink.Appendp(ctxt, p)
		p.As = ACMPL
		p.From.Typ = D_SI
		p.To.Typ = D_CONST
		p.To.Offset = int64(uint32(liblink.StackPreempt & 0xFFFFFFFF))
		p = liblink.Appendp(ctxt, p)
		p.As = AJEQ
		p.To.Typ = D_BRANCH
		q1 = p
		p = liblink.Appendp(ctxt, p)
		p.As = ALEAL
		p.From.Typ = D_INDIR + D_SP
		p.From.Offset = liblink.StackGuard
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = ASUBL
		p.From.Typ = D_SI
		p.From.Offset = 0
		p.To.Typ = D_AX
		p = liblink.Appendp(ctxt, p)
		p.As = ACMPL
		p.From.Typ = D_AX
		p.To.Typ = D_CONST
		p.To.Offset = framesize + (liblink.StackGuard - liblink.StackSmall)
	}
	// common
	p = liblink.Appendp(ctxt, p)
	p.As = AJHI
	p.To.Typ = D_BRANCH
	p.To.Offset = 4
	q = p
	p = liblink.Appendp(ctxt, p) // save frame size in DI
	p.As = AMOVL
	p.To.Typ = D_DI
	p.From.Typ = D_CONST
	// If we ask for more stack, we'll get a minimum of StackMin bytes.
	// We need a stack frame large enough to hold the top-of-stack data,
	// the function arguments+results, our caller's PC, our frame,
	// a word for the return PC of the next call, and then the StackLimit bytes
	// that must be available on entry to any function called from a function
	// that did a stack check.  If StackMin is enough, don't ask for a specific
	// amount: then we can use the custom functions and save a few
	// instructions.
	if int64(liblink.StackTop+ctxt.Cursym.Text.To.Offset2)+ctxt.Arch.Ptrsize+framesize+ctxt.Arch.Ptrsize+liblink.StackLimit >= liblink.StackMin {
		p.From.Offset = (framesize + 7) &^ 7
	}
	arg = ctxt.Cursym.Text.To.Offset2
	if arg == 1 { // special marker for known 0
		arg = 0
	}
	if arg&3 != 0 {
		ctxt.Diag("misaligned argument size in stack split")
	}
	p = liblink.Appendp(ctxt, p) // save arg size in AX
	p.As = AMOVL
	p.To.Typ = D_AX
	p.From.Typ = D_CONST
	p.From.Offset = int64(arg)
	p = liblink.Appendp(ctxt, p)
	p.As = ACALL
	p.To.Typ = D_BRANCH
	p.To.Sym = ctxt.Symmorestack[noctxt]
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
		AIRETW,
		AUNDEF:
		return true
	}
	return false
}

func pushpop(a int) int {
	switch a {
	case APUSHL,
		APUSHFL,
		APUSHW,
		APUSHFW,
		APOPL,
		APOPFL,
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
	log.Fatalf("unknown relation: %s", Anames8[a])
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

var Link386 = liblink.LinkArch{
	Name:          "386",
	Thechar:       '8',
	ByteOrder:     binary.LittleEndian,
	Pconv:         Pconv,
	Addstacksplit: addstacksplit,
	Assemble:      span8,
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
	Regsize:       4,
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
