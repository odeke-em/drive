package amd64

import (
	"fmt"
	"log"

	"code.google.com/p/rsc/c2go/liblink"
)

/*
 * this is the ranlib header
 */
const (
	MaxAlign   = 32
	LoopAlign  = 16
	MaxLoopPad = 0
	FuncAlign  = 16
)

type Optab struct {
	as     int
	ytab   []uint8
	prefix int
	op     [23]uint8
}

type Movtab struct {
	as   int
	ft   uint8
	tt   uint8
	code uint8
	op   [4]uint8
}

const (
	Yxxx = 0 + iota
	Ynone
	Yi0
	Yi1
	Yi8
	Ys32
	Yi32
	Yi64
	Yiauto
	Yal
	Ycl
	Yax
	Ycx
	Yrb
	Yrl
	Yrf
	Yf0
	Yrx
	Ymb
	Yml
	Ym
	Ybr
	Ycol
	Ycs
	Yss
	Yds
	Yes
	Yfs
	Ygs
	Ygdtr
	Yidtr
	Yldtr
	Ymsw
	Ytask
	Ycr0
	Ycr1
	Ycr2
	Ycr3
	Ycr4
	Ycr5
	Ycr6
	Ycr7
	Ycr8
	Ydr0
	Ydr1
	Ydr2
	Ydr3
	Ydr4
	Ydr5
	Ydr6
	Ydr7
	Ytr0
	Ytr1
	Ytr2
	Ytr3
	Ytr4
	Ytr5
	Ytr6
	Ytr7
	Yrl32
	Yrl64
	Ymr
	Ymm
	Yxr
	Yxm
	Ytls
	Ymax
	Zxxx = 0 + iota - 67
	Zlit
	Zlitm_r
	Z_rp
	Zbr
	Zcall
	Zcallindreg
	Zib_
	Zib_rp
	Zibo_m
	Zibo_m_xm
	Zil_
	Zil_rp
	Ziq_rp
	Zilo_m
	Ziqo_m
	Zjmp
	Zloop
	Zo_iw
	Zm_o
	Zm_r
	Zm2_r
	Zm_r_xm
	Zm_r_i_xm
	Zm_r_3d
	Zm_r_xm_nr
	Zr_m_xm_nr
	Zibm_r
	Zmb_r
	Zaut_r
	Zo_m
	Zo_m64
	Zpseudo
	Zr_m
	Zr_m_xm
	Zr_m_i_xm
	Zrp_
	Z_ib
	Z_il
	Zm_ibo
	Zm_ilo
	Zib_rr
	Zil_rr
	Zclr
	Zbyte
	Zmax
	Px     = 0
	P32    = 0x32
	Pe     = 0x66
	Pm     = 0x0f
	Pq     = 0xff
	Pb     = 0xfe
	Pf2    = 0xf2
	Pf3    = 0xf3
	Pq3    = 0x67
	Pw     = 0x48
	Py     = 0x80
	Rxf    = 1 << 9
	Rxt    = 1 << 8
	Rxw    = 1 << 3
	Rxr    = 1 << 2
	Rxx    = 1 << 1
	Rxb    = 1 << 0
	Maxand = 10
)

var ycover [Ymax * Ymax]uint8

var reg [D_NONE]int

var regrex [D_NONE + 1]int

var ynone = []uint8{
	Ynone,
	Ynone,
	Zlit,
	1,
	0,
}

var ytext = []uint8{
	Ymb,
	Yi64,
	Zpseudo,
	1,
	0,
}

var ynop = []uint8{
	Ynone,
	Ynone,
	Zpseudo,
	0,
	Ynone,
	Yiauto,
	Zpseudo,
	0,
	Ynone,
	Yml,
	Zpseudo,
	0,
	Ynone,
	Yrf,
	Zpseudo,
	0,
	Ynone,
	Yxr,
	Zpseudo,
	0,
	Yiauto,
	Ynone,
	Zpseudo,
	0,
	Yml,
	Ynone,
	Zpseudo,
	0,
	Yrf,
	Ynone,
	Zpseudo,
	0,
	Yxr,
	Ynone,
	Zpseudo,
	1,
	0,
}

var yfuncdata = []uint8{
	Yi32,
	Ym,
	Zpseudo,
	0,
	0,
}

var ypcdata = []uint8{
	Yi32,
	Yi32,
	Zpseudo,
	0,
	0,
}

var yxorb = []uint8{
	Yi32,
	Yal,
	Zib_,
	1,
	Yi32,
	Ymb,
	Zibo_m,
	2,
	Yrb,
	Ymb,
	Zr_m,
	1,
	Ymb,
	Yrb,
	Zm_r,
	1,
	0,
}

var yxorl = []uint8{
	Yi8,
	Yml,
	Zibo_m,
	2,
	Yi32,
	Yax,
	Zil_,
	1,
	Yi32,
	Yml,
	Zilo_m,
	2,
	Yrl,
	Yml,
	Zr_m,
	1,
	Yml,
	Yrl,
	Zm_r,
	1,
	0,
}

var yaddl = []uint8{
	Yi8,
	Yml,
	Zibo_m,
	2,
	Yi32,
	Yax,
	Zil_,
	1,
	Yi32,
	Yml,
	Zilo_m,
	2,
	Yrl,
	Yml,
	Zr_m,
	1,
	Yml,
	Yrl,
	Zm_r,
	1,
	0,
}

var yincb = []uint8{
	Ynone,
	Ymb,
	Zo_m,
	2,
	0,
}

var yincw = []uint8{
	Ynone,
	Yml,
	Zo_m,
	2,
	0,
}

var yincl = []uint8{
	Ynone,
	Yml,
	Zo_m,
	2,
	0,
}

var ycmpb = []uint8{
	Yal,
	Yi32,
	Z_ib,
	1,
	Ymb,
	Yi32,
	Zm_ibo,
	2,
	Ymb,
	Yrb,
	Zm_r,
	1,
	Yrb,
	Ymb,
	Zr_m,
	1,
	0,
}

var ycmpl = []uint8{
	Yml,
	Yi8,
	Zm_ibo,
	2,
	Yax,
	Yi32,
	Z_il,
	1,
	Yml,
	Yi32,
	Zm_ilo,
	2,
	Yml,
	Yrl,
	Zm_r,
	1,
	Yrl,
	Yml,
	Zr_m,
	1,
	0,
}

var yshb = []uint8{
	Yi1,
	Ymb,
	Zo_m,
	2,
	Yi32,
	Ymb,
	Zibo_m,
	2,
	Ycx,
	Ymb,
	Zo_m,
	2,
	0,
}

var yshl = []uint8{
	Yi1,
	Yml,
	Zo_m,
	2,
	Yi32,
	Yml,
	Zibo_m,
	2,
	Ycl,
	Yml,
	Zo_m,
	2,
	Ycx,
	Yml,
	Zo_m,
	2,
	0,
}

var ytestb = []uint8{
	Yi32,
	Yal,
	Zib_,
	1,
	Yi32,
	Ymb,
	Zibo_m,
	2,
	Yrb,
	Ymb,
	Zr_m,
	1,
	Ymb,
	Yrb,
	Zm_r,
	1,
	0,
}

var ytestl = []uint8{
	Yi32,
	Yax,
	Zil_,
	1,
	Yi32,
	Yml,
	Zilo_m,
	2,
	Yrl,
	Yml,
	Zr_m,
	1,
	Yml,
	Yrl,
	Zm_r,
	1,
	0,
}

var ymovb = []uint8{
	Yrb,
	Ymb,
	Zr_m,
	1,
	Ymb,
	Yrb,
	Zm_r,
	1,
	Yi32,
	Yrb,
	Zib_rp,
	1,
	Yi32,
	Ymb,
	Zibo_m,
	2,
	0,
}

var ymbs = []uint8{
	Ymb,
	Ynone,
	Zm_o,
	2,
	0,
}

var ybtl = []uint8{
	Yi8,
	Yml,
	Zibo_m,
	2,
	Yrl,
	Yml,
	Zr_m,
	1,
	0,
}

var ymovw = []uint8{
	Yrl,
	Yml,
	Zr_m,
	1,
	Yml,
	Yrl,
	Zm_r,
	1,
	Yi0,
	Yrl,
	Zclr,
	1,
	Yi32,
	Yrl,
	Zil_rp,
	1,
	Yi32,
	Yml,
	Zilo_m,
	2,
	Yiauto,
	Yrl,
	Zaut_r,
	2,
	0,
}

var ymovl = []uint8{
	Yrl,
	Yml,
	Zr_m,
	1,
	Yml,
	Yrl,
	Zm_r,
	1,
	Yi0,
	Yrl,
	Zclr,
	1,
	Yi32,
	Yrl,
	Zil_rp,
	1,
	Yi32,
	Yml,
	Zilo_m,
	2,
	Yml,
	Ymr,
	Zm_r_xm,
	1, // MMX MOVD
	Ymr,
	Yml,
	Zr_m_xm,
	1, // MMX MOVD
	Yml,
	Yxr,
	Zm_r_xm,
	2, // XMM MOVD (32 bit)
	Yxr,
	Yml,
	Zr_m_xm,
	2, // XMM MOVD (32 bit)
	Yiauto,
	Yrl,
	Zaut_r,
	2,
	0,
}

var yret = []uint8{
	Ynone,
	Ynone,
	Zo_iw,
	1,
	Yi32,
	Ynone,
	Zo_iw,
	1,
	0,
}

var ymovq = []uint8{
	Yrl,
	Yml,
	Zr_m,
	1, // 0x89
	Yml,
	Yrl,
	Zm_r,
	1, // 0x8b
	Yi0,
	Yrl,
	Zclr,
	1, // 0x31
	Ys32,
	Yrl,
	Zilo_m,
	2, // 32 bit signed 0xc7,(0)
	Yi64,
	Yrl,
	Ziq_rp,
	1, // 0xb8 -- 32/64 bit immediate
	Yi32,
	Yml,
	Zilo_m,
	2, // 0xc7,(0)
	Ym,
	Ymr,
	Zm_r_xm_nr,
	1, // MMX MOVQ (shorter encoding)
	Ymr,
	Ym,
	Zr_m_xm_nr,
	1, // MMX MOVQ
	Ymm,
	Ymr,
	Zm_r_xm,
	1, // MMX MOVD
	Ymr,
	Ymm,
	Zr_m_xm,
	1, // MMX MOVD
	Yxr,
	Ymr,
	Zm_r_xm_nr,
	2, // MOVDQ2Q
	Yxm,
	Yxr,
	Zm_r_xm_nr,
	2, // MOVQ xmm1/m64 -> xmm2
	Yxr,
	Yxm,
	Zr_m_xm_nr,
	2, // MOVQ xmm1 -> xmm2/m64
	Yml,
	Yxr,
	Zm_r_xm,
	2, // MOVD xmm load
	Yxr,
	Yml,
	Zr_m_xm,
	2, // MOVD xmm store
	Yiauto,
	Yrl,
	Zaut_r,
	2, // built-in LEAQ
	0,
}

var ym_rl = []uint8{
	Ym,
	Yrl,
	Zm_r,
	1,
	0,
}

var yrl_m = []uint8{
	Yrl,
	Ym,
	Zr_m,
	1,
	0,
}

var ymb_rl = []uint8{
	Ymb,
	Yrl,
	Zmb_r,
	1,
	0,
}

var yml_rl = []uint8{
	Yml,
	Yrl,
	Zm_r,
	1,
	0,
}

var yrl_ml = []uint8{
	Yrl,
	Yml,
	Zr_m,
	1,
	0,
}

var yml_mb = []uint8{
	Yrb,
	Ymb,
	Zr_m,
	1,
	Ymb,
	Yrb,
	Zm_r,
	1,
	0,
}

var yrb_mb = []uint8{
	Yrb,
	Ymb,
	Zr_m,
	1,
	0,
}

var yxchg = []uint8{
	Yax,
	Yrl,
	Z_rp,
	1,
	Yrl,
	Yax,
	Zrp_,
	1,
	Yrl,
	Yml,
	Zr_m,
	1,
	Yml,
	Yrl,
	Zm_r,
	1,
	0,
}

var ydivl = []uint8{
	Yml,
	Ynone,
	Zm_o,
	2,
	0,
}

var ydivb = []uint8{
	Ymb,
	Ynone,
	Zm_o,
	2,
	0,
}

var yimul = []uint8{
	Yml,
	Ynone,
	Zm_o,
	2,
	Yi8,
	Yrl,
	Zib_rr,
	1,
	Yi32,
	Yrl,
	Zil_rr,
	1,
	Yml,
	Yrl,
	Zm_r,
	2,
	0,
}

var yimul3 = []uint8{
	Yml,
	Yrl,
	Zibm_r,
	2,
	0,
}

var ybyte = []uint8{
	Yi64,
	Ynone,
	Zbyte,
	1,
	0,
}

var yin = []uint8{
	Yi32,
	Ynone,
	Zib_,
	1,
	Ynone,
	Ynone,
	Zlit,
	1,
	0,
}

var yint = []uint8{
	Yi32,
	Ynone,
	Zib_,
	1,
	0,
}

var ypushl = []uint8{
	Yrl,
	Ynone,
	Zrp_,
	1,
	Ym,
	Ynone,
	Zm_o,
	2,
	Yi8,
	Ynone,
	Zib_,
	1,
	Yi32,
	Ynone,
	Zil_,
	1,
	0,
}

var ypopl = []uint8{
	Ynone,
	Yrl,
	Z_rp,
	1,
	Ynone,
	Ym,
	Zo_m,
	2,
	0,
}

var ybswap = []uint8{
	Ynone,
	Yrl,
	Z_rp,
	2,
	0,
}

var yscond = []uint8{
	Ynone,
	Ymb,
	Zo_m,
	2,
	0,
}

var yjcond = []uint8{
	Ynone,
	Ybr,
	Zbr,
	0,
	Yi0,
	Ybr,
	Zbr,
	0,
	Yi1,
	Ybr,
	Zbr,
	1,
	0,
}

var yloop = []uint8{
	Ynone,
	Ybr,
	Zloop,
	1,
	0,
}

var ycall = []uint8{
	Ynone,
	Yml,
	Zcallindreg,
	0,
	Yrx,
	Yrx,
	Zcallindreg,
	2,
	Ynone,
	Ybr,
	Zcall,
	1,
	0,
}

var yduff = []uint8{
	Ynone,
	Yi32,
	Zcall,
	1,
	0,
}

var yjmp = []uint8{
	Ynone,
	Yml,
	Zo_m64,
	2,
	Ynone,
	Ybr,
	Zjmp,
	1,
	0,
}

var yfmvd = []uint8{
	Ym,
	Yf0,
	Zm_o,
	2,
	Yf0,
	Ym,
	Zo_m,
	2,
	Yrf,
	Yf0,
	Zm_o,
	2,
	Yf0,
	Yrf,
	Zo_m,
	2,
	0,
}

var yfmvdp = []uint8{
	Yf0,
	Ym,
	Zo_m,
	2,
	Yf0,
	Yrf,
	Zo_m,
	2,
	0,
}

var yfmvf = []uint8{
	Ym,
	Yf0,
	Zm_o,
	2,
	Yf0,
	Ym,
	Zo_m,
	2,
	0,
}

var yfmvx = []uint8{
	Ym,
	Yf0,
	Zm_o,
	2,
	0,
}

var yfmvp = []uint8{
	Yf0,
	Ym,
	Zo_m,
	2,
	0,
}

var yfadd = []uint8{
	Ym,
	Yf0,
	Zm_o,
	2,
	Yrf,
	Yf0,
	Zm_o,
	2,
	Yf0,
	Yrf,
	Zo_m,
	2,
	0,
}

var yfaddp = []uint8{
	Yf0,
	Yrf,
	Zo_m,
	2,
	0,
}

var yfxch = []uint8{
	Yf0,
	Yrf,
	Zo_m,
	2,
	Yrf,
	Yf0,
	Zm_o,
	2,
	0,
}

var ycompp = []uint8{
	Yf0,
	Yrf,
	Zo_m,
	2, /* botch is really f0,f1 */
	0,
}

var ystsw = []uint8{
	Ynone,
	Ym,
	Zo_m,
	2,
	Ynone,
	Yax,
	Zlit,
	1,
	0,
}

var ystcw = []uint8{
	Ynone,
	Ym,
	Zo_m,
	2,
	Ym,
	Ynone,
	Zm_o,
	2,
	0,
}

var ysvrs = []uint8{
	Ynone,
	Ym,
	Zo_m,
	2,
	Ym,
	Ynone,
	Zm_o,
	2,
	0,
}

var ymm = []uint8{
	Ymm,
	Ymr,
	Zm_r_xm,
	1,
	Yxm,
	Yxr,
	Zm_r_xm,
	2,
	0,
}

var yxm = []uint8{
	Yxm,
	Yxr,
	Zm_r_xm,
	1,
	0,
}

var yxcvm1 = []uint8{
	Yxm,
	Yxr,
	Zm_r_xm,
	2,
	Yxm,
	Ymr,
	Zm_r_xm,
	2,
	0,
}

var yxcvm2 = []uint8{
	Yxm,
	Yxr,
	Zm_r_xm,
	2,
	Ymm,
	Yxr,
	Zm_r_xm,
	2,
	0,
}

/*
static uchar	yxmq[] =
{
	Yxm,	Yxr,	Zm_r_xm,	2,
	0
};
*/
var yxr = []uint8{
	Yxr,
	Yxr,
	Zm_r_xm,
	1,
	0,
}

var yxr_ml = []uint8{
	Yxr,
	Yml,
	Zr_m_xm,
	1,
	0,
}

var ymr = []uint8{
	Ymr,
	Ymr,
	Zm_r,
	1,
	0,
}

var ymr_ml = []uint8{
	Ymr,
	Yml,
	Zr_m_xm,
	1,
	0,
}

var yxcmp = []uint8{
	Yxm,
	Yxr,
	Zm_r_xm,
	1,
	0,
}

var yxcmpi = []uint8{
	Yxm,
	Yxr,
	Zm_r_i_xm,
	2,
	0,
}

var yxmov = []uint8{
	Yxm,
	Yxr,
	Zm_r_xm,
	1,
	Yxr,
	Yxm,
	Zr_m_xm,
	1,
	0,
}

var yxcvfl = []uint8{
	Yxm,
	Yrl,
	Zm_r_xm,
	1,
	0,
}

var yxcvlf = []uint8{
	Yml,
	Yxr,
	Zm_r_xm,
	1,
	0,
}

var yxcvfq = []uint8{
	Yxm,
	Yrl,
	Zm_r_xm,
	2,
	0,
}

var yxcvqf = []uint8{
	Yml,
	Yxr,
	Zm_r_xm,
	2,
	0,
}

var yps = []uint8{
	Ymm,
	Ymr,
	Zm_r_xm,
	1,
	Yi8,
	Ymr,
	Zibo_m_xm,
	2,
	Yxm,
	Yxr,
	Zm_r_xm,
	2,
	Yi8,
	Yxr,
	Zibo_m_xm,
	3,
	0,
}

var yxrrl = []uint8{
	Yxr,
	Yrl,
	Zm_r,
	1,
	0,
}

var ymfp = []uint8{
	Ymm,
	Ymr,
	Zm_r_3d,
	1,
	0,
}

var ymrxr = []uint8{
	Ymr,
	Yxr,
	Zm_r,
	1,
	Yxm,
	Yxr,
	Zm_r_xm,
	1,
	0,
}

var ymshuf = []uint8{
	Ymm,
	Ymr,
	Zibm_r,
	2,
	0,
}

var ymshufb = []uint8{
	Yxm,
	Yxr,
	Zm2_r,
	2,
	0,
}

var yxshuf = []uint8{
	Yxm,
	Yxr,
	Zibm_r,
	2,
	0,
}

var yextrw = []uint8{
	Yxr,
	Yrl,
	Zibm_r,
	2,
	0,
}

var yinsrw = []uint8{
	Yml,
	Yxr,
	Zibm_r,
	2,
	0,
}

var yinsr = []uint8{
	Ymm,
	Yxr,
	Zibm_r,
	3,
	0,
}

var ypsdq = []uint8{
	Yi8,
	Yxr,
	Zibo_m,
	2,
	0,
}

var ymskb = []uint8{
	Yxr,
	Yrl,
	Zm_r_xm,
	2,
	Ymr,
	Yrl,
	Zm_r_xm,
	1,
	0,
}

var ycrc32l = []uint8{Yml, Yrl, Zlitm_r, 0}

var yprefetch = []uint8{
	Ym,
	Ynone,
	Zm_o,
	2,
	0,
}

var yaes = []uint8{
	Yxm,
	Yxr,
	Zlitm_r,
	2,
	0,
}

var yaes2 = []uint8{
	Yxm,
	Yxr,
	Zibm_r,
	2,
	0,
}

/*
 * You are doasm, holding in your hand a Prog* with p->as set to, say, ACRC32,
 * and p->from and p->to as operands (Addr*).  The linker scans optab to find
 * the entry with the given p->as and then looks through the ytable for that
 * instruction (the second field in the optab struct) for a line whose first
 * two values match the Ytypes of the p->from and p->to operands.  The function
 * oclass in span.c computes the specific Ytype of an operand and then the set
 * of more general Ytypes that it satisfies is implied by the ycover table, set
 * up in instinit.  For example, oclass distinguishes the constants 0 and 1
 * from the more general 8-bit constants, but instinit says
 *
 *        ycover[Yi0*Ymax + Ys32] = 1;
 *        ycover[Yi1*Ymax + Ys32] = 1;
 *        ycover[Yi8*Ymax + Ys32] = 1;
 *
 * which means that Yi0, Yi1, and Yi8 all count as Ys32 (signed 32)
 * if that's what an instruction can handle.
 *
 * In parallel with the scan through the ytable for the appropriate line, there
 * is a z pointer that starts out pointing at the strange magic byte list in
 * the Optab struct.  With each step past a non-matching ytable line, z
 * advances by the 4th entry in the line.  When a matching line is found, that
 * z pointer has the extra data to use in laying down the instruction bytes.
 * The actual bytes laid down are a function of the 3rd entry in the line (that
 * is, the Ztype) and the z bytes.
 *
 * For example, let's look at AADDL.  The optab line says:
 *        { AADDL,        yaddl,  Px, 0x83,(00),0x05,0x81,(00),0x01,0x03 },
 *
 * and yaddl says
 *        uchar   yaddl[] =
 *        {
 *                Yi8,    Yml,    Zibo_m, 2,
 *                Yi32,   Yax,    Zil_,   1,
 *                Yi32,   Yml,    Zilo_m, 2,
 *                Yrl,    Yml,    Zr_m,   1,
 *                Yml,    Yrl,    Zm_r,   1,
 *                0
 *        };
 *
 * so there are 5 possible types of ADDL instruction that can be laid down, and
 * possible states used to lay them down (Ztype and z pointer, assuming z
 * points at {0x83,(00),0x05,0x81,(00),0x01,0x03}) are:
 *
 *        Yi8, Yml -> Zibo_m, z (0x83, 00)
 *        Yi32, Yax -> Zil_, z+2 (0x05)
 *        Yi32, Yml -> Zilo_m, z+2+1 (0x81, 0x00)
 *        Yrl, Yml -> Zr_m, z+2+1+2 (0x01)
 *        Yml, Yrl -> Zm_r, z+2+1+2+1 (0x03)
 *
 * The Pconstant in the optab line controls the prefix bytes to emit.  That's
 * relatively straightforward as this program goes.
 *
 * The switch on t[2] in doasm implements the various Z cases.  Zibo_m, for
 * example, is an opcode byte (z[0]) then an asmando (which is some kind of
 * encoded addressing mode for the Yml arg), and then a single immediate byte.
 * Zilo_m is the same but a long (32-bit) immediate.
 */
var optab = /*	as, ytab, andproto, opcode */
[]Optab{
	{AXXX, nil, 0, [23]uint8{}},
	{AAAA, ynone, P32, [23]uint8{0x37}},
	{AAAD, ynone, P32, [23]uint8{0xd5, 0x0a}},
	{AAAM, ynone, P32, [23]uint8{0xd4, 0x0a}},
	{AAAS, ynone, P32, [23]uint8{0x3f}},
	{AADCB, yxorb, Pb, [23]uint8{0x14, 0x80, 02, 0x10, 0x10}},
	{AADCL, yxorl, Px, [23]uint8{0x83, 02, 0x15, 0x81, 02, 0x11, 0x13}},
	{AADCQ, yxorl, Pw, [23]uint8{0x83, 02, 0x15, 0x81, 02, 0x11, 0x13}},
	{AADCW, yxorl, Pe, [23]uint8{0x83, 02, 0x15, 0x81, 02, 0x11, 0x13}},
	{AADDB, yxorb, Pb, [23]uint8{0x04, 0x80, 00, 0x00, 0x02}},
	{AADDL, yaddl, Px, [23]uint8{0x83, 00, 0x05, 0x81, 00, 0x01, 0x03}},
	{AADDPD, yxm, Pq, [23]uint8{0x58}},
	{AADDPS, yxm, Pm, [23]uint8{0x58}},
	{AADDQ, yaddl, Pw, [23]uint8{0x83, 00, 0x05, 0x81, 00, 0x01, 0x03}},
	{AADDSD, yxm, Pf2, [23]uint8{0x58}},
	{AADDSS, yxm, Pf3, [23]uint8{0x58}},
	{AADDW, yaddl, Pe, [23]uint8{0x83, 00, 0x05, 0x81, 00, 0x01, 0x03}},
	{AADJSP, nil, 0, [23]uint8{}},
	{AANDB, yxorb, Pb, [23]uint8{0x24, 0x80, 04, 0x20, 0x22}},
	{AANDL, yxorl, Px, [23]uint8{0x83, 04, 0x25, 0x81, 04, 0x21, 0x23}},
	{AANDNPD, yxm, Pq, [23]uint8{0x55}},
	{AANDNPS, yxm, Pm, [23]uint8{0x55}},
	{AANDPD, yxm, Pq, [23]uint8{0x54}},
	{AANDPS, yxm, Pq, [23]uint8{0x54}},
	{AANDQ, yxorl, Pw, [23]uint8{0x83, 04, 0x25, 0x81, 04, 0x21, 0x23}},
	{AANDW, yxorl, Pe, [23]uint8{0x83, 04, 0x25, 0x81, 04, 0x21, 0x23}},
	{AARPL, yrl_ml, P32, [23]uint8{0x63}},
	{ABOUNDL, yrl_m, P32, [23]uint8{0x62}},
	{ABOUNDW, yrl_m, Pe, [23]uint8{0x62}},
	{ABSFL, yml_rl, Pm, [23]uint8{0xbc}},
	{ABSFQ, yml_rl, Pw, [23]uint8{0x0f, 0xbc}},
	{ABSFW, yml_rl, Pq, [23]uint8{0xbc}},
	{ABSRL, yml_rl, Pm, [23]uint8{0xbd}},
	{ABSRQ, yml_rl, Pw, [23]uint8{0x0f, 0xbd}},
	{ABSRW, yml_rl, Pq, [23]uint8{0xbd}},
	{ABSWAPL, ybswap, Px, [23]uint8{0x0f, 0xc8}},
	{ABSWAPQ, ybswap, Pw, [23]uint8{0x0f, 0xc8}},
	{ABTCL, ybtl, Pm, [23]uint8{0xba, 07, 0xbb}},
	{ABTCQ, ybtl, Pw, [23]uint8{0x0f, 0xba, 07, 0x0f, 0xbb}},
	{ABTCW, ybtl, Pq, [23]uint8{0xba, 07, 0xbb}},
	{ABTL, ybtl, Pm, [23]uint8{0xba, 04, 0xa3}},
	{ABTQ, ybtl, Pw, [23]uint8{0x0f, 0xba, 04, 0x0f, 0xa3}},
	{ABTRL, ybtl, Pm, [23]uint8{0xba, 06, 0xb3}},
	{ABTRQ, ybtl, Pw, [23]uint8{0x0f, 0xba, 06, 0x0f, 0xb3}},
	{ABTRW, ybtl, Pq, [23]uint8{0xba, 06, 0xb3}},
	{ABTSL, ybtl, Pm, [23]uint8{0xba, 05, 0xab}},
	{ABTSQ, ybtl, Pw, [23]uint8{0x0f, 0xba, 05, 0x0f, 0xab}},
	{ABTSW, ybtl, Pq, [23]uint8{0xba, 05, 0xab}},
	{ABTW, ybtl, Pq, [23]uint8{0xba, 04, 0xa3}},
	{ABYTE, ybyte, Px, [23]uint8{1}},
	{ACALL, ycall, Px, [23]uint8{0xff, 02, 0xe8}},
	{ACDQ, ynone, Px, [23]uint8{0x99}},
	{ACLC, ynone, Px, [23]uint8{0xf8}},
	{ACLD, ynone, Px, [23]uint8{0xfc}},
	{ACLI, ynone, Px, [23]uint8{0xfa}},
	{ACLTS, ynone, Pm, [23]uint8{0x06}},
	{ACMC, ynone, Px, [23]uint8{0xf5}},
	{ACMOVLCC, yml_rl, Pm, [23]uint8{0x43}},
	{ACMOVLCS, yml_rl, Pm, [23]uint8{0x42}},
	{ACMOVLEQ, yml_rl, Pm, [23]uint8{0x44}},
	{ACMOVLGE, yml_rl, Pm, [23]uint8{0x4d}},
	{ACMOVLGT, yml_rl, Pm, [23]uint8{0x4f}},
	{ACMOVLHI, yml_rl, Pm, [23]uint8{0x47}},
	{ACMOVLLE, yml_rl, Pm, [23]uint8{0x4e}},
	{ACMOVLLS, yml_rl, Pm, [23]uint8{0x46}},
	{ACMOVLLT, yml_rl, Pm, [23]uint8{0x4c}},
	{ACMOVLMI, yml_rl, Pm, [23]uint8{0x48}},
	{ACMOVLNE, yml_rl, Pm, [23]uint8{0x45}},
	{ACMOVLOC, yml_rl, Pm, [23]uint8{0x41}},
	{ACMOVLOS, yml_rl, Pm, [23]uint8{0x40}},
	{ACMOVLPC, yml_rl, Pm, [23]uint8{0x4b}},
	{ACMOVLPL, yml_rl, Pm, [23]uint8{0x49}},
	{ACMOVLPS, yml_rl, Pm, [23]uint8{0x4a}},
	{ACMOVQCC, yml_rl, Pw, [23]uint8{0x0f, 0x43}},
	{ACMOVQCS, yml_rl, Pw, [23]uint8{0x0f, 0x42}},
	{ACMOVQEQ, yml_rl, Pw, [23]uint8{0x0f, 0x44}},
	{ACMOVQGE, yml_rl, Pw, [23]uint8{0x0f, 0x4d}},
	{ACMOVQGT, yml_rl, Pw, [23]uint8{0x0f, 0x4f}},
	{ACMOVQHI, yml_rl, Pw, [23]uint8{0x0f, 0x47}},
	{ACMOVQLE, yml_rl, Pw, [23]uint8{0x0f, 0x4e}},
	{ACMOVQLS, yml_rl, Pw, [23]uint8{0x0f, 0x46}},
	{ACMOVQLT, yml_rl, Pw, [23]uint8{0x0f, 0x4c}},
	{ACMOVQMI, yml_rl, Pw, [23]uint8{0x0f, 0x48}},
	{ACMOVQNE, yml_rl, Pw, [23]uint8{0x0f, 0x45}},
	{ACMOVQOC, yml_rl, Pw, [23]uint8{0x0f, 0x41}},
	{ACMOVQOS, yml_rl, Pw, [23]uint8{0x0f, 0x40}},
	{ACMOVQPC, yml_rl, Pw, [23]uint8{0x0f, 0x4b}},
	{ACMOVQPL, yml_rl, Pw, [23]uint8{0x0f, 0x49}},
	{ACMOVQPS, yml_rl, Pw, [23]uint8{0x0f, 0x4a}},
	{ACMOVWCC, yml_rl, Pq, [23]uint8{0x43}},
	{ACMOVWCS, yml_rl, Pq, [23]uint8{0x42}},
	{ACMOVWEQ, yml_rl, Pq, [23]uint8{0x44}},
	{ACMOVWGE, yml_rl, Pq, [23]uint8{0x4d}},
	{ACMOVWGT, yml_rl, Pq, [23]uint8{0x4f}},
	{ACMOVWHI, yml_rl, Pq, [23]uint8{0x47}},
	{ACMOVWLE, yml_rl, Pq, [23]uint8{0x4e}},
	{ACMOVWLS, yml_rl, Pq, [23]uint8{0x46}},
	{ACMOVWLT, yml_rl, Pq, [23]uint8{0x4c}},
	{ACMOVWMI, yml_rl, Pq, [23]uint8{0x48}},
	{ACMOVWNE, yml_rl, Pq, [23]uint8{0x45}},
	{ACMOVWOC, yml_rl, Pq, [23]uint8{0x41}},
	{ACMOVWOS, yml_rl, Pq, [23]uint8{0x40}},
	{ACMOVWPC, yml_rl, Pq, [23]uint8{0x4b}},
	{ACMOVWPL, yml_rl, Pq, [23]uint8{0x49}},
	{ACMOVWPS, yml_rl, Pq, [23]uint8{0x4a}},
	{ACMPB, ycmpb, Pb, [23]uint8{0x3c, 0x80, 07, 0x38, 0x3a}},
	{ACMPL, ycmpl, Px, [23]uint8{0x83, 07, 0x3d, 0x81, 07, 0x39, 0x3b}},
	{ACMPPD, yxcmpi, Px, [23]uint8{Pe, 0xc2}},
	{ACMPPS, yxcmpi, Pm, [23]uint8{0xc2, 0}},
	{ACMPQ, ycmpl, Pw, [23]uint8{0x83, 07, 0x3d, 0x81, 07, 0x39, 0x3b}},
	{ACMPSB, ynone, Pb, [23]uint8{0xa6}},
	{ACMPSD, yxcmpi, Px, [23]uint8{Pf2, 0xc2}},
	{ACMPSL, ynone, Px, [23]uint8{0xa7}},
	{ACMPSQ, ynone, Pw, [23]uint8{0xa7}},
	{ACMPSS, yxcmpi, Px, [23]uint8{Pf3, 0xc2}},
	{ACMPSW, ynone, Pe, [23]uint8{0xa7}},
	{ACMPW, ycmpl, Pe, [23]uint8{0x83, 07, 0x3d, 0x81, 07, 0x39, 0x3b}},
	{ACOMISD, yxcmp, Pe, [23]uint8{0x2f}},
	{ACOMISS, yxcmp, Pm, [23]uint8{0x2f}},
	{ACPUID, ynone, Pm, [23]uint8{0xa2}},
	{ACVTPL2PD, yxcvm2, Px, [23]uint8{Pf3, 0xe6, Pe, 0x2a}},
	{ACVTPL2PS, yxcvm2, Pm, [23]uint8{0x5b, 0, 0x2a, 0}},
	{ACVTPD2PL, yxcvm1, Px, [23]uint8{Pf2, 0xe6, Pe, 0x2d}},
	{ACVTPD2PS, yxm, Pe, [23]uint8{0x5a}},
	{ACVTPS2PL, yxcvm1, Px, [23]uint8{Pe, 0x5b, Pm, 0x2d}},
	{ACVTPS2PD, yxm, Pm, [23]uint8{0x5a}},
	{API2FW, ymfp, Px, [23]uint8{0x0c}},
	{ACVTSD2SL, yxcvfl, Pf2, [23]uint8{0x2d}},
	{ACVTSD2SQ, yxcvfq, Pw, [23]uint8{Pf2, 0x2d}},
	{ACVTSD2SS, yxm, Pf2, [23]uint8{0x5a}},
	{ACVTSL2SD, yxcvlf, Pf2, [23]uint8{0x2a}},
	{ACVTSQ2SD, yxcvqf, Pw, [23]uint8{Pf2, 0x2a}},
	{ACVTSL2SS, yxcvlf, Pf3, [23]uint8{0x2a}},
	{ACVTSQ2SS, yxcvqf, Pw, [23]uint8{Pf3, 0x2a}},
	{ACVTSS2SD, yxm, Pf3, [23]uint8{0x5a}},
	{ACVTSS2SL, yxcvfl, Pf3, [23]uint8{0x2d}},
	{ACVTSS2SQ, yxcvfq, Pw, [23]uint8{Pf3, 0x2d}},
	{ACVTTPD2PL, yxcvm1, Px, [23]uint8{Pe, 0xe6, Pe, 0x2c}},
	{ACVTTPS2PL, yxcvm1, Px, [23]uint8{Pf3, 0x5b, Pm, 0x2c}},
	{ACVTTSD2SL, yxcvfl, Pf2, [23]uint8{0x2c}},
	{ACVTTSD2SQ, yxcvfq, Pw, [23]uint8{Pf2, 0x2c}},
	{ACVTTSS2SL, yxcvfl, Pf3, [23]uint8{0x2c}},
	{ACVTTSS2SQ, yxcvfq, Pw, [23]uint8{Pf3, 0x2c}},
	{ACWD, ynone, Pe, [23]uint8{0x99}},
	{ACQO, ynone, Pw, [23]uint8{0x99}},
	{ADAA, ynone, P32, [23]uint8{0x27}},
	{ADAS, ynone, P32, [23]uint8{0x2f}},
	{ADATA, nil, 0, [23]uint8{}},
	{ADECB, yincb, Pb, [23]uint8{0xfe, 01}},
	{ADECL, yincl, Px, [23]uint8{0xff, 01}},
	{ADECQ, yincl, Pw, [23]uint8{0xff, 01}},
	{ADECW, yincw, Pe, [23]uint8{0xff, 01}},
	{ADIVB, ydivb, Pb, [23]uint8{0xf6, 06}},
	{ADIVL, ydivl, Px, [23]uint8{0xf7, 06}},
	{ADIVPD, yxm, Pe, [23]uint8{0x5e}},
	{ADIVPS, yxm, Pm, [23]uint8{0x5e}},
	{ADIVQ, ydivl, Pw, [23]uint8{0xf7, 06}},
	{ADIVSD, yxm, Pf2, [23]uint8{0x5e}},
	{ADIVSS, yxm, Pf3, [23]uint8{0x5e}},
	{ADIVW, ydivl, Pe, [23]uint8{0xf7, 06}},
	{AEMMS, ynone, Pm, [23]uint8{0x77}},
	{AENTER, nil, 0, [23]uint8{}}, /* botch */
	{AFXRSTOR, ysvrs, Pm, [23]uint8{0xae, 01, 0xae, 01}},
	{AFXSAVE, ysvrs, Pm, [23]uint8{0xae, 00, 0xae, 00}},
	{AFXRSTOR64, ysvrs, Pw, [23]uint8{0x0f, 0xae, 01, 0x0f, 0xae, 01}},
	{AFXSAVE64, ysvrs, Pw, [23]uint8{0x0f, 0xae, 00, 0x0f, 0xae, 00}},
	{AGLOBL, nil, 0, [23]uint8{}},
	{AGOK, nil, 0, [23]uint8{}},
	{AHISTORY, nil, 0, [23]uint8{}},
	{AHLT, ynone, Px, [23]uint8{0xf4}},
	{AIDIVB, ydivb, Pb, [23]uint8{0xf6, 07}},
	{AIDIVL, ydivl, Px, [23]uint8{0xf7, 07}},
	{AIDIVQ, ydivl, Pw, [23]uint8{0xf7, 07}},
	{AIDIVW, ydivl, Pe, [23]uint8{0xf7, 07}},
	{AIMULB, ydivb, Pb, [23]uint8{0xf6, 05}},
	{AIMULL, yimul, Px, [23]uint8{0xf7, 05, 0x6b, 0x69, Pm, 0xaf}},
	{AIMULQ, yimul, Pw, [23]uint8{0xf7, 05, 0x6b, 0x69, Pm, 0xaf}},
	{AIMULW, yimul, Pe, [23]uint8{0xf7, 05, 0x6b, 0x69, Pm, 0xaf}},
	{AIMUL3Q, yimul3, Pw, [23]uint8{0x6b, 00}},
	{AINB, yin, Pb, [23]uint8{0xe4, 0xec}},
	{AINCB, yincb, Pb, [23]uint8{0xfe, 00}},
	{AINCL, yincl, Px, [23]uint8{0xff, 00}},
	{AINCQ, yincl, Pw, [23]uint8{0xff, 00}},
	{AINCW, yincw, Pe, [23]uint8{0xff, 00}},
	{AINL, yin, Px, [23]uint8{0xe5, 0xed}},
	{AINSB, ynone, Pb, [23]uint8{0x6c}},
	{AINSL, ynone, Px, [23]uint8{0x6d}},
	{AINSW, ynone, Pe, [23]uint8{0x6d}},
	{AINT, yint, Px, [23]uint8{0xcd}},
	{AINTO, ynone, P32, [23]uint8{0xce}},
	{AINW, yin, Pe, [23]uint8{0xe5, 0xed}},
	{AIRETL, ynone, Px, [23]uint8{0xcf}},
	{AIRETQ, ynone, Pw, [23]uint8{0xcf}},
	{AIRETW, ynone, Pe, [23]uint8{0xcf}},
	{AJCC, yjcond, Px, [23]uint8{0x73, 0x83, 00}},
	{AJCS, yjcond, Px, [23]uint8{0x72, 0x82}},
	{AJCXZL, yloop, Px, [23]uint8{0xe3}},
	{AJCXZQ, yloop, Px, [23]uint8{0xe3}},
	{AJEQ, yjcond, Px, [23]uint8{0x74, 0x84}},
	{AJGE, yjcond, Px, [23]uint8{0x7d, 0x8d}},
	{AJGT, yjcond, Px, [23]uint8{0x7f, 0x8f}},
	{AJHI, yjcond, Px, [23]uint8{0x77, 0x87}},
	{AJLE, yjcond, Px, [23]uint8{0x7e, 0x8e}},
	{AJLS, yjcond, Px, [23]uint8{0x76, 0x86}},
	{AJLT, yjcond, Px, [23]uint8{0x7c, 0x8c}},
	{AJMI, yjcond, Px, [23]uint8{0x78, 0x88}},
	{AJMP, yjmp, Px, [23]uint8{0xff, 04, 0xeb, 0xe9}},
	{AJNE, yjcond, Px, [23]uint8{0x75, 0x85}},
	{AJOC, yjcond, Px, [23]uint8{0x71, 0x81, 00}},
	{AJOS, yjcond, Px, [23]uint8{0x70, 0x80, 00}},
	{AJPC, yjcond, Px, [23]uint8{0x7b, 0x8b}},
	{AJPL, yjcond, Px, [23]uint8{0x79, 0x89}},
	{AJPS, yjcond, Px, [23]uint8{0x7a, 0x8a}},
	{ALAHF, ynone, Px, [23]uint8{0x9f}},
	{ALARL, yml_rl, Pm, [23]uint8{0x02}},
	{ALARW, yml_rl, Pq, [23]uint8{0x02}},
	{ALDMXCSR, ysvrs, Pm, [23]uint8{0xae, 02, 0xae, 02}},
	{ALEAL, ym_rl, Px, [23]uint8{0x8d}},
	{ALEAQ, ym_rl, Pw, [23]uint8{0x8d}},
	{ALEAVEL, ynone, P32, [23]uint8{0xc9}},
	{ALEAVEQ, ynone, Py, [23]uint8{0xc9}},
	{ALEAVEW, ynone, Pe, [23]uint8{0xc9}},
	{ALEAW, ym_rl, Pe, [23]uint8{0x8d}},
	{ALOCK, ynone, Px, [23]uint8{0xf0}},
	{ALODSB, ynone, Pb, [23]uint8{0xac}},
	{ALODSL, ynone, Px, [23]uint8{0xad}},
	{ALODSQ, ynone, Pw, [23]uint8{0xad}},
	{ALODSW, ynone, Pe, [23]uint8{0xad}},
	{ALONG, ybyte, Px, [23]uint8{4}},
	{ALOOP, yloop, Px, [23]uint8{0xe2}},
	{ALOOPEQ, yloop, Px, [23]uint8{0xe1}},
	{ALOOPNE, yloop, Px, [23]uint8{0xe0}},
	{ALSLL, yml_rl, Pm, [23]uint8{0x03}},
	{ALSLW, yml_rl, Pq, [23]uint8{0x03}},
	{AMASKMOVOU, yxr, Pe, [23]uint8{0xf7}},
	{AMASKMOVQ, ymr, Pm, [23]uint8{0xf7}},
	{AMAXPD, yxm, Pe, [23]uint8{0x5f}},
	{AMAXPS, yxm, Pm, [23]uint8{0x5f}},
	{AMAXSD, yxm, Pf2, [23]uint8{0x5f}},
	{AMAXSS, yxm, Pf3, [23]uint8{0x5f}},
	{AMINPD, yxm, Pe, [23]uint8{0x5d}},
	{AMINPS, yxm, Pm, [23]uint8{0x5d}},
	{AMINSD, yxm, Pf2, [23]uint8{0x5d}},
	{AMINSS, yxm, Pf3, [23]uint8{0x5d}},
	{AMOVAPD, yxmov, Pe, [23]uint8{0x28, 0x29}},
	{AMOVAPS, yxmov, Pm, [23]uint8{0x28, 0x29}},
	{AMOVB, ymovb, Pb, [23]uint8{0x88, 0x8a, 0xb0, 0xc6, 00}},
	{AMOVBLSX, ymb_rl, Pm, [23]uint8{0xbe}},
	{AMOVBLZX, ymb_rl, Pm, [23]uint8{0xb6}},
	{AMOVBQSX, ymb_rl, Pw, [23]uint8{0x0f, 0xbe}},
	{AMOVBQZX, ymb_rl, Pw, [23]uint8{0x0f, 0xb6}},
	{AMOVBWSX, ymb_rl, Pq, [23]uint8{0xbe}},
	{AMOVBWZX, ymb_rl, Pq, [23]uint8{0xb6}},
	{AMOVO, yxmov, Pe, [23]uint8{0x6f, 0x7f}},
	{AMOVOU, yxmov, Pf3, [23]uint8{0x6f, 0x7f}},
	{AMOVHLPS, yxr, Pm, [23]uint8{0x12}},
	{AMOVHPD, yxmov, Pe, [23]uint8{0x16, 0x17}},
	{AMOVHPS, yxmov, Pm, [23]uint8{0x16, 0x17}},
	{AMOVL, ymovl, Px, [23]uint8{0x89, 0x8b, 0x31, 0xb8, 0xc7, 00, 0x6e, 0x7e, Pe, 0x6e, Pe, 0x7e, 0}},
	{AMOVLHPS, yxr, Pm, [23]uint8{0x16}},
	{AMOVLPD, yxmov, Pe, [23]uint8{0x12, 0x13}},
	{AMOVLPS, yxmov, Pm, [23]uint8{0x12, 0x13}},
	{AMOVLQSX, yml_rl, Pw, [23]uint8{0x63}},
	{AMOVLQZX, yml_rl, Px, [23]uint8{0x8b}},
	{AMOVMSKPD, yxrrl, Pq, [23]uint8{0x50}},
	{AMOVMSKPS, yxrrl, Pm, [23]uint8{0x50}},
	{AMOVNTO, yxr_ml, Pe, [23]uint8{0xe7}},
	{AMOVNTPD, yxr_ml, Pe, [23]uint8{0x2b}},
	{AMOVNTPS, yxr_ml, Pm, [23]uint8{0x2b}},
	{AMOVNTQ, ymr_ml, Pm, [23]uint8{0xe7}},
	{AMOVQ, ymovq, Pw, [23]uint8{0x89, 0x8b, 0x31, 0xc7, 00, 0xb8, 0xc7, 00, 0x6f, 0x7f, 0x6e, 0x7e, Pf2, 0xd6, Pf3, 0x7e, Pe, 0xd6, Pe, 0x6e, Pe, 0x7e, 0}},
	{AMOVQOZX, ymrxr, Pf3, [23]uint8{0xd6, 0x7e}},
	{AMOVSB, ynone, Pb, [23]uint8{0xa4}},
	{AMOVSD, yxmov, Pf2, [23]uint8{0x10, 0x11}},
	{AMOVSL, ynone, Px, [23]uint8{0xa5}},
	{AMOVSQ, ynone, Pw, [23]uint8{0xa5}},
	{AMOVSS, yxmov, Pf3, [23]uint8{0x10, 0x11}},
	{AMOVSW, ynone, Pe, [23]uint8{0xa5}},
	{AMOVUPD, yxmov, Pe, [23]uint8{0x10, 0x11}},
	{AMOVUPS, yxmov, Pm, [23]uint8{0x10, 0x11}},
	{AMOVW, ymovw, Pe, [23]uint8{0x89, 0x8b, 0x31, 0xb8, 0xc7, 00, 0}},
	{AMOVWLSX, yml_rl, Pm, [23]uint8{0xbf}},
	{AMOVWLZX, yml_rl, Pm, [23]uint8{0xb7}},
	{AMOVWQSX, yml_rl, Pw, [23]uint8{0x0f, 0xbf}},
	{AMOVWQZX, yml_rl, Pw, [23]uint8{0x0f, 0xb7}},
	{AMULB, ydivb, Pb, [23]uint8{0xf6, 04}},
	{AMULL, ydivl, Px, [23]uint8{0xf7, 04}},
	{AMULPD, yxm, Pe, [23]uint8{0x59}},
	{AMULPS, yxm, Ym, [23]uint8{0x59}},
	{AMULQ, ydivl, Pw, [23]uint8{0xf7, 04}},
	{AMULSD, yxm, Pf2, [23]uint8{0x59}},
	{AMULSS, yxm, Pf3, [23]uint8{0x59}},
	{AMULW, ydivl, Pe, [23]uint8{0xf7, 04}},
	{ANAME, nil, 0, [23]uint8{}},
	{ANEGB, yscond, Pb, [23]uint8{0xf6, 03}},
	{ANEGL, yscond, Px, [23]uint8{0xf7, 03}},
	{ANEGQ, yscond, Pw, [23]uint8{0xf7, 03}},
	{ANEGW, yscond, Pe, [23]uint8{0xf7, 03}},
	{ANOP, ynop, Px, [23]uint8{0, 0}},
	{ANOTB, yscond, Pb, [23]uint8{0xf6, 02}},
	{ANOTL, yscond, Px, [23]uint8{0xf7, 02}},
	{ANOTQ, yscond, Pw, [23]uint8{0xf7, 02}},
	{ANOTW, yscond, Pe, [23]uint8{0xf7, 02}},
	{AORB, yxorb, Pb, [23]uint8{0x0c, 0x80, 01, 0x08, 0x0a}},
	{AORL, yxorl, Px, [23]uint8{0x83, 01, 0x0d, 0x81, 01, 0x09, 0x0b}},
	{AORPD, yxm, Pq, [23]uint8{0x56}},
	{AORPS, yxm, Pm, [23]uint8{0x56}},
	{AORQ, yxorl, Pw, [23]uint8{0x83, 01, 0x0d, 0x81, 01, 0x09, 0x0b}},
	{AORW, yxorl, Pe, [23]uint8{0x83, 01, 0x0d, 0x81, 01, 0x09, 0x0b}},
	{AOUTB, yin, Pb, [23]uint8{0xe6, 0xee}},
	{AOUTL, yin, Px, [23]uint8{0xe7, 0xef}},
	{AOUTSB, ynone, Pb, [23]uint8{0x6e}},
	{AOUTSL, ynone, Px, [23]uint8{0x6f}},
	{AOUTSW, ynone, Pe, [23]uint8{0x6f}},
	{AOUTW, yin, Pe, [23]uint8{0xe7, 0xef}},
	{APACKSSLW, ymm, Py, [23]uint8{0x6b, Pe, 0x6b}},
	{APACKSSWB, ymm, Py, [23]uint8{0x63, Pe, 0x63}},
	{APACKUSWB, ymm, Py, [23]uint8{0x67, Pe, 0x67}},
	{APADDB, ymm, Py, [23]uint8{0xfc, Pe, 0xfc}},
	{APADDL, ymm, Py, [23]uint8{0xfe, Pe, 0xfe}},
	{APADDQ, yxm, Pe, [23]uint8{0xd4}},
	{APADDSB, ymm, Py, [23]uint8{0xec, Pe, 0xec}},
	{APADDSW, ymm, Py, [23]uint8{0xed, Pe, 0xed}},
	{APADDUSB, ymm, Py, [23]uint8{0xdc, Pe, 0xdc}},
	{APADDUSW, ymm, Py, [23]uint8{0xdd, Pe, 0xdd}},
	{APADDW, ymm, Py, [23]uint8{0xfd, Pe, 0xfd}},
	{APAND, ymm, Py, [23]uint8{0xdb, Pe, 0xdb}},
	{APANDN, ymm, Py, [23]uint8{0xdf, Pe, 0xdf}},
	{APAUSE, ynone, Px, [23]uint8{0xf3, 0x90}},
	{APAVGB, ymm, Py, [23]uint8{0xe0, Pe, 0xe0}},
	{APAVGW, ymm, Py, [23]uint8{0xe3, Pe, 0xe3}},
	{APCMPEQB, ymm, Py, [23]uint8{0x74, Pe, 0x74}},
	{APCMPEQL, ymm, Py, [23]uint8{0x76, Pe, 0x76}},
	{APCMPEQW, ymm, Py, [23]uint8{0x75, Pe, 0x75}},
	{APCMPGTB, ymm, Py, [23]uint8{0x64, Pe, 0x64}},
	{APCMPGTL, ymm, Py, [23]uint8{0x66, Pe, 0x66}},
	{APCMPGTW, ymm, Py, [23]uint8{0x65, Pe, 0x65}},
	{APEXTRW, yextrw, Pq, [23]uint8{0xc5, 00}},
	{APF2IL, ymfp, Px, [23]uint8{0x1d}},
	{APF2IW, ymfp, Px, [23]uint8{0x1c}},
	{API2FL, ymfp, Px, [23]uint8{0x0d}},
	{APFACC, ymfp, Px, [23]uint8{0xae}},
	{APFADD, ymfp, Px, [23]uint8{0x9e}},
	{APFCMPEQ, ymfp, Px, [23]uint8{0xb0}},
	{APFCMPGE, ymfp, Px, [23]uint8{0x90}},
	{APFCMPGT, ymfp, Px, [23]uint8{0xa0}},
	{APFMAX, ymfp, Px, [23]uint8{0xa4}},
	{APFMIN, ymfp, Px, [23]uint8{0x94}},
	{APFMUL, ymfp, Px, [23]uint8{0xb4}},
	{APFNACC, ymfp, Px, [23]uint8{0x8a}},
	{APFPNACC, ymfp, Px, [23]uint8{0x8e}},
	{APFRCP, ymfp, Px, [23]uint8{0x96}},
	{APFRCPIT1, ymfp, Px, [23]uint8{0xa6}},
	{APFRCPI2T, ymfp, Px, [23]uint8{0xb6}},
	{APFRSQIT1, ymfp, Px, [23]uint8{0xa7}},
	{APFRSQRT, ymfp, Px, [23]uint8{0x97}},
	{APFSUB, ymfp, Px, [23]uint8{0x9a}},
	{APFSUBR, ymfp, Px, [23]uint8{0xaa}},
	{APINSRW, yinsrw, Pq, [23]uint8{0xc4, 00}},
	{APINSRD, yinsr, Pq, [23]uint8{0x3a, 0x22, 00}},
	{APINSRQ, yinsr, Pq3, [23]uint8{0x3a, 0x22, 00}},
	{APMADDWL, ymm, Py, [23]uint8{0xf5, Pe, 0xf5}},
	{APMAXSW, yxm, Pe, [23]uint8{0xee}},
	{APMAXUB, yxm, Pe, [23]uint8{0xde}},
	{APMINSW, yxm, Pe, [23]uint8{0xea}},
	{APMINUB, yxm, Pe, [23]uint8{0xda}},
	{APMOVMSKB, ymskb, Px, [23]uint8{Pe, 0xd7, 0xd7}},
	{APMULHRW, ymfp, Px, [23]uint8{0xb7}},
	{APMULHUW, ymm, Py, [23]uint8{0xe4, Pe, 0xe4}},
	{APMULHW, ymm, Py, [23]uint8{0xe5, Pe, 0xe5}},
	{APMULLW, ymm, Py, [23]uint8{0xd5, Pe, 0xd5}},
	{APMULULQ, ymm, Py, [23]uint8{0xf4, Pe, 0xf4}},
	{APOPAL, ynone, P32, [23]uint8{0x61}},
	{APOPAW, ynone, Pe, [23]uint8{0x61}},
	{APOPFL, ynone, P32, [23]uint8{0x9d}},
	{APOPFQ, ynone, Py, [23]uint8{0x9d}},
	{APOPFW, ynone, Pe, [23]uint8{0x9d}},
	{APOPL, ypopl, P32, [23]uint8{0x58, 0x8f, 00}},
	{APOPQ, ypopl, Py, [23]uint8{0x58, 0x8f, 00}},
	{APOPW, ypopl, Pe, [23]uint8{0x58, 0x8f, 00}},
	{APOR, ymm, Py, [23]uint8{0xeb, Pe, 0xeb}},
	{APSADBW, yxm, Pq, [23]uint8{0xf6}},
	{APSHUFHW, yxshuf, Pf3, [23]uint8{0x70, 00}},
	{APSHUFL, yxshuf, Pq, [23]uint8{0x70, 00}},
	{APSHUFLW, yxshuf, Pf2, [23]uint8{0x70, 00}},
	{APSHUFW, ymshuf, Pm, [23]uint8{0x70, 00}},
	{APSHUFB, ymshufb, Pq, [23]uint8{0x38, 0x00}},
	{APSLLO, ypsdq, Pq, [23]uint8{0x73, 07}},
	{APSLLL, yps, Py, [23]uint8{0xf2, 0x72, 06, Pe, 0xf2, Pe, 0x72, 06}},
	{APSLLQ, yps, Py, [23]uint8{0xf3, 0x73, 06, Pe, 0xf3, Pe, 0x73, 06}},
	{APSLLW, yps, Py, [23]uint8{0xf1, 0x71, 06, Pe, 0xf1, Pe, 0x71, 06}},
	{APSRAL, yps, Py, [23]uint8{0xe2, 0x72, 04, Pe, 0xe2, Pe, 0x72, 04}},
	{APSRAW, yps, Py, [23]uint8{0xe1, 0x71, 04, Pe, 0xe1, Pe, 0x71, 04}},
	{APSRLO, ypsdq, Pq, [23]uint8{0x73, 03}},
	{APSRLL, yps, Py, [23]uint8{0xd2, 0x72, 02, Pe, 0xd2, Pe, 0x72, 02}},
	{APSRLQ, yps, Py, [23]uint8{0xd3, 0x73, 02, Pe, 0xd3, Pe, 0x73, 02}},
	{APSRLW, yps, Py, [23]uint8{0xd1, 0x71, 02, Pe, 0xe1, Pe, 0x71, 02}},
	{APSUBB, yxm, Pe, [23]uint8{0xf8}},
	{APSUBL, yxm, Pe, [23]uint8{0xfa}},
	{APSUBQ, yxm, Pe, [23]uint8{0xfb}},
	{APSUBSB, yxm, Pe, [23]uint8{0xe8}},
	{APSUBSW, yxm, Pe, [23]uint8{0xe9}},
	{APSUBUSB, yxm, Pe, [23]uint8{0xd8}},
	{APSUBUSW, yxm, Pe, [23]uint8{0xd9}},
	{APSUBW, yxm, Pe, [23]uint8{0xf9}},
	{APSWAPL, ymfp, Px, [23]uint8{0xbb}},
	{APUNPCKHBW, ymm, Py, [23]uint8{0x68, Pe, 0x68}},
	{APUNPCKHLQ, ymm, Py, [23]uint8{0x6a, Pe, 0x6a}},
	{APUNPCKHQDQ, yxm, Pe, [23]uint8{0x6d}},
	{APUNPCKHWL, ymm, Py, [23]uint8{0x69, Pe, 0x69}},
	{APUNPCKLBW, ymm, Py, [23]uint8{0x60, Pe, 0x60}},
	{APUNPCKLLQ, ymm, Py, [23]uint8{0x62, Pe, 0x62}},
	{APUNPCKLQDQ, yxm, Pe, [23]uint8{0x6c}},
	{APUNPCKLWL, ymm, Py, [23]uint8{0x61, Pe, 0x61}},
	{APUSHAL, ynone, P32, [23]uint8{0x60}},
	{APUSHAW, ynone, Pe, [23]uint8{0x60}},
	{APUSHFL, ynone, P32, [23]uint8{0x9c}},
	{APUSHFQ, ynone, Py, [23]uint8{0x9c}},
	{APUSHFW, ynone, Pe, [23]uint8{0x9c}},
	{APUSHL, ypushl, P32, [23]uint8{0x50, 0xff, 06, 0x6a, 0x68}},
	{APUSHQ, ypushl, Py, [23]uint8{0x50, 0xff, 06, 0x6a, 0x68}},
	{APUSHW, ypushl, Pe, [23]uint8{0x50, 0xff, 06, 0x6a, 0x68}},
	{APXOR, ymm, Py, [23]uint8{0xef, Pe, 0xef}},
	{AQUAD, ybyte, Px, [23]uint8{8}},
	{ARCLB, yshb, Pb, [23]uint8{0xd0, 02, 0xc0, 02, 0xd2, 02}},
	{ARCLL, yshl, Px, [23]uint8{0xd1, 02, 0xc1, 02, 0xd3, 02, 0xd3, 02}},
	{ARCLQ, yshl, Pw, [23]uint8{0xd1, 02, 0xc1, 02, 0xd3, 02, 0xd3, 02}},
	{ARCLW, yshl, Pe, [23]uint8{0xd1, 02, 0xc1, 02, 0xd3, 02, 0xd3, 02}},
	{ARCPPS, yxm, Pm, [23]uint8{0x53}},
	{ARCPSS, yxm, Pf3, [23]uint8{0x53}},
	{ARCRB, yshb, Pb, [23]uint8{0xd0, 03, 0xc0, 03, 0xd2, 03}},
	{ARCRL, yshl, Px, [23]uint8{0xd1, 03, 0xc1, 03, 0xd3, 03, 0xd3, 03}},
	{ARCRQ, yshl, Pw, [23]uint8{0xd1, 03, 0xc1, 03, 0xd3, 03, 0xd3, 03}},
	{ARCRW, yshl, Pe, [23]uint8{0xd1, 03, 0xc1, 03, 0xd3, 03, 0xd3, 03}},
	{AREP, ynone, Px, [23]uint8{0xf3}},
	{AREPN, ynone, Px, [23]uint8{0xf2}},
	{ARET, ynone, Px, [23]uint8{0xc3}},
	{ARETFW, yret, Pe, [23]uint8{0xcb, 0xca}},
	{ARETFL, yret, Px, [23]uint8{0xcb, 0xca}},
	{ARETFQ, yret, Pw, [23]uint8{0xcb, 0xca}},
	{AROLB, yshb, Pb, [23]uint8{0xd0, 00, 0xc0, 00, 0xd2, 00}},
	{AROLL, yshl, Px, [23]uint8{0xd1, 00, 0xc1, 00, 0xd3, 00, 0xd3, 00}},
	{AROLQ, yshl, Pw, [23]uint8{0xd1, 00, 0xc1, 00, 0xd3, 00, 0xd3, 00}},
	{AROLW, yshl, Pe, [23]uint8{0xd1, 00, 0xc1, 00, 0xd3, 00, 0xd3, 00}},
	{ARORB, yshb, Pb, [23]uint8{0xd0, 01, 0xc0, 01, 0xd2, 01}},
	{ARORL, yshl, Px, [23]uint8{0xd1, 01, 0xc1, 01, 0xd3, 01, 0xd3, 01}},
	{ARORQ, yshl, Pw, [23]uint8{0xd1, 01, 0xc1, 01, 0xd3, 01, 0xd3, 01}},
	{ARORW, yshl, Pe, [23]uint8{0xd1, 01, 0xc1, 01, 0xd3, 01, 0xd3, 01}},
	{ARSQRTPS, yxm, Pm, [23]uint8{0x52}},
	{ARSQRTSS, yxm, Pf3, [23]uint8{0x52}},
	{ASAHF, ynone, Px, [23]uint8{0x86, 0xe0, 0x50, 0x9d}}, /* XCHGB AH,AL; PUSH AX; POPFL */
	{ASALB, yshb, Pb, [23]uint8{0xd0, 04, 0xc0, 04, 0xd2, 04}},
	{ASALL, yshl, Px, [23]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASALQ, yshl, Pw, [23]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASALW, yshl, Pe, [23]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASARB, yshb, Pb, [23]uint8{0xd0, 07, 0xc0, 07, 0xd2, 07}},
	{ASARL, yshl, Px, [23]uint8{0xd1, 07, 0xc1, 07, 0xd3, 07, 0xd3, 07}},
	{ASARQ, yshl, Pw, [23]uint8{0xd1, 07, 0xc1, 07, 0xd3, 07, 0xd3, 07}},
	{ASARW, yshl, Pe, [23]uint8{0xd1, 07, 0xc1, 07, 0xd3, 07, 0xd3, 07}},
	{ASBBB, yxorb, Pb, [23]uint8{0x1c, 0x80, 03, 0x18, 0x1a}},
	{ASBBL, yxorl, Px, [23]uint8{0x83, 03, 0x1d, 0x81, 03, 0x19, 0x1b}},
	{ASBBQ, yxorl, Pw, [23]uint8{0x83, 03, 0x1d, 0x81, 03, 0x19, 0x1b}},
	{ASBBW, yxorl, Pe, [23]uint8{0x83, 03, 0x1d, 0x81, 03, 0x19, 0x1b}},
	{ASCASB, ynone, Pb, [23]uint8{0xae}},
	{ASCASL, ynone, Px, [23]uint8{0xaf}},
	{ASCASQ, ynone, Pw, [23]uint8{0xaf}},
	{ASCASW, ynone, Pe, [23]uint8{0xaf}},
	{ASETCC, yscond, Pm, [23]uint8{0x93, 00}},
	{ASETCS, yscond, Pm, [23]uint8{0x92, 00}},
	{ASETEQ, yscond, Pm, [23]uint8{0x94, 00}},
	{ASETGE, yscond, Pm, [23]uint8{0x9d, 00}},
	{ASETGT, yscond, Pm, [23]uint8{0x9f, 00}},
	{ASETHI, yscond, Pm, [23]uint8{0x97, 00}},
	{ASETLE, yscond, Pm, [23]uint8{0x9e, 00}},
	{ASETLS, yscond, Pm, [23]uint8{0x96, 00}},
	{ASETLT, yscond, Pm, [23]uint8{0x9c, 00}},
	{ASETMI, yscond, Pm, [23]uint8{0x98, 00}},
	{ASETNE, yscond, Pm, [23]uint8{0x95, 00}},
	{ASETOC, yscond, Pm, [23]uint8{0x91, 00}},
	{ASETOS, yscond, Pm, [23]uint8{0x90, 00}},
	{ASETPC, yscond, Pm, [23]uint8{0x96, 00}},
	{ASETPL, yscond, Pm, [23]uint8{0x99, 00}},
	{ASETPS, yscond, Pm, [23]uint8{0x9a, 00}},
	{ASHLB, yshb, Pb, [23]uint8{0xd0, 04, 0xc0, 04, 0xd2, 04}},
	{ASHLL, yshl, Px, [23]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASHLQ, yshl, Pw, [23]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASHLW, yshl, Pe, [23]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASHRB, yshb, Pb, [23]uint8{0xd0, 05, 0xc0, 05, 0xd2, 05}},
	{ASHRL, yshl, Px, [23]uint8{0xd1, 05, 0xc1, 05, 0xd3, 05, 0xd3, 05}},
	{ASHRQ, yshl, Pw, [23]uint8{0xd1, 05, 0xc1, 05, 0xd3, 05, 0xd3, 05}},
	{ASHRW, yshl, Pe, [23]uint8{0xd1, 05, 0xc1, 05, 0xd3, 05, 0xd3, 05}},
	{ASHUFPD, yxshuf, Pq, [23]uint8{0xc6, 00}},
	{ASHUFPS, yxshuf, Pm, [23]uint8{0xc6, 00}},
	{ASQRTPD, yxm, Pe, [23]uint8{0x51}},
	{ASQRTPS, yxm, Pm, [23]uint8{0x51}},
	{ASQRTSD, yxm, Pf2, [23]uint8{0x51}},
	{ASQRTSS, yxm, Pf3, [23]uint8{0x51}},
	{ASTC, ynone, Px, [23]uint8{0xf9}},
	{ASTD, ynone, Px, [23]uint8{0xfd}},
	{ASTI, ynone, Px, [23]uint8{0xfb}},
	{ASTMXCSR, ysvrs, Pm, [23]uint8{0xae, 03, 0xae, 03}},
	{ASTOSB, ynone, Pb, [23]uint8{0xaa}},
	{ASTOSL, ynone, Px, [23]uint8{0xab}},
	{ASTOSQ, ynone, Pw, [23]uint8{0xab}},
	{ASTOSW, ynone, Pe, [23]uint8{0xab}},
	{ASUBB, yxorb, Pb, [23]uint8{0x2c, 0x80, 05, 0x28, 0x2a}},
	{ASUBL, yaddl, Px, [23]uint8{0x83, 05, 0x2d, 0x81, 05, 0x29, 0x2b}},
	{ASUBPD, yxm, Pe, [23]uint8{0x5c}},
	{ASUBPS, yxm, Pm, [23]uint8{0x5c}},
	{ASUBQ, yaddl, Pw, [23]uint8{0x83, 05, 0x2d, 0x81, 05, 0x29, 0x2b}},
	{ASUBSD, yxm, Pf2, [23]uint8{0x5c}},
	{ASUBSS, yxm, Pf3, [23]uint8{0x5c}},
	{ASUBW, yaddl, Pe, [23]uint8{0x83, 05, 0x2d, 0x81, 05, 0x29, 0x2b}},
	{ASWAPGS, ynone, Pm, [23]uint8{0x01, 0xf8}},
	{ASYSCALL, ynone, Px, [23]uint8{0x0f, 0x05}}, /* fast syscall */
	{ATESTB, ytestb, Pb, [23]uint8{0xa8, 0xf6, 00, 0x84, 0x84}},
	{ATESTL, ytestl, Px, [23]uint8{0xa9, 0xf7, 00, 0x85, 0x85}},
	{ATESTQ, ytestl, Pw, [23]uint8{0xa9, 0xf7, 00, 0x85, 0x85}},
	{ATESTW, ytestl, Pe, [23]uint8{0xa9, 0xf7, 00, 0x85, 0x85}},
	{ATEXT, ytext, Px, [23]uint8{}},
	{AUCOMISD, yxcmp, Pe, [23]uint8{0x2e}},
	{AUCOMISS, yxcmp, Pm, [23]uint8{0x2e}},
	{AUNPCKHPD, yxm, Pe, [23]uint8{0x15}},
	{AUNPCKHPS, yxm, Pm, [23]uint8{0x15}},
	{AUNPCKLPD, yxm, Pe, [23]uint8{0x14}},
	{AUNPCKLPS, yxm, Pm, [23]uint8{0x14}},
	{AVERR, ydivl, Pm, [23]uint8{0x00, 04}},
	{AVERW, ydivl, Pm, [23]uint8{0x00, 05}},
	{AWAIT, ynone, Px, [23]uint8{0x9b}},
	{AWORD, ybyte, Px, [23]uint8{2}},
	{AXCHGB, yml_mb, Pb, [23]uint8{0x86, 0x86}},
	{AXCHGL, yxchg, Px, [23]uint8{0x90, 0x90, 0x87, 0x87}},
	{AXCHGQ, yxchg, Pw, [23]uint8{0x90, 0x90, 0x87, 0x87}},
	{AXCHGW, yxchg, Pe, [23]uint8{0x90, 0x90, 0x87, 0x87}},
	{AXLAT, ynone, Px, [23]uint8{0xd7}},
	{AXORB, yxorb, Pb, [23]uint8{0x34, 0x80, 06, 0x30, 0x32}},
	{AXORL, yxorl, Px, [23]uint8{0x83, 06, 0x35, 0x81, 06, 0x31, 0x33}},
	{AXORPD, yxm, Pe, [23]uint8{0x57}},
	{AXORPS, yxm, Pm, [23]uint8{0x57}},
	{AXORQ, yxorl, Pw, [23]uint8{0x83, 06, 0x35, 0x81, 06, 0x31, 0x33}},
	{AXORW, yxorl, Pe, [23]uint8{0x83, 06, 0x35, 0x81, 06, 0x31, 0x33}},
	{AFMOVB, yfmvx, Px, [23]uint8{0xdf, 04}},
	{AFMOVBP, yfmvp, Px, [23]uint8{0xdf, 06}},
	{AFMOVD, yfmvd, Px, [23]uint8{0xdd, 00, 0xdd, 02, 0xd9, 00, 0xdd, 02}},
	{AFMOVDP, yfmvdp, Px, [23]uint8{0xdd, 03, 0xdd, 03}},
	{AFMOVF, yfmvf, Px, [23]uint8{0xd9, 00, 0xd9, 02}},
	{AFMOVFP, yfmvp, Px, [23]uint8{0xd9, 03}},
	{AFMOVL, yfmvf, Px, [23]uint8{0xdb, 00, 0xdb, 02}},
	{AFMOVLP, yfmvp, Px, [23]uint8{0xdb, 03}},
	{AFMOVV, yfmvx, Px, [23]uint8{0xdf, 05}},
	{AFMOVVP, yfmvp, Px, [23]uint8{0xdf, 07}},
	{AFMOVW, yfmvf, Px, [23]uint8{0xdf, 00, 0xdf, 02}},
	{AFMOVWP, yfmvp, Px, [23]uint8{0xdf, 03}},
	{AFMOVX, yfmvx, Px, [23]uint8{0xdb, 05}},
	{AFMOVXP, yfmvp, Px, [23]uint8{0xdb, 07}},
	{AFCOMB, nil, 0, [23]uint8{}},
	{AFCOMBP, nil, 0, [23]uint8{}},
	{AFCOMD, yfadd, Px, [23]uint8{0xdc, 02, 0xd8, 02, 0xdc, 02}},  /* botch */
	{AFCOMDP, yfadd, Px, [23]uint8{0xdc, 03, 0xd8, 03, 0xdc, 03}}, /* botch */
	{AFCOMDPP, ycompp, Px, [23]uint8{0xde, 03}},
	{AFCOMF, yfmvx, Px, [23]uint8{0xd8, 02}},
	{AFCOMFP, yfmvx, Px, [23]uint8{0xd8, 03}},
	{AFCOML, yfmvx, Px, [23]uint8{0xda, 02}},
	{AFCOMLP, yfmvx, Px, [23]uint8{0xda, 03}},
	{AFCOMW, yfmvx, Px, [23]uint8{0xde, 02}},
	{AFCOMWP, yfmvx, Px, [23]uint8{0xde, 03}},
	{AFUCOM, ycompp, Px, [23]uint8{0xdd, 04}},
	{AFUCOMP, ycompp, Px, [23]uint8{0xdd, 05}},
	{AFUCOMPP, ycompp, Px, [23]uint8{0xda, 13}},
	{AFADDDP, yfaddp, Px, [23]uint8{0xde, 00}},
	{AFADDW, yfmvx, Px, [23]uint8{0xde, 00}},
	{AFADDL, yfmvx, Px, [23]uint8{0xda, 00}},
	{AFADDF, yfmvx, Px, [23]uint8{0xd8, 00}},
	{AFADDD, yfadd, Px, [23]uint8{0xdc, 00, 0xd8, 00, 0xdc, 00}},
	{AFMULDP, yfaddp, Px, [23]uint8{0xde, 01}},
	{AFMULW, yfmvx, Px, [23]uint8{0xde, 01}},
	{AFMULL, yfmvx, Px, [23]uint8{0xda, 01}},
	{AFMULF, yfmvx, Px, [23]uint8{0xd8, 01}},
	{AFMULD, yfadd, Px, [23]uint8{0xdc, 01, 0xd8, 01, 0xdc, 01}},
	{AFSUBDP, yfaddp, Px, [23]uint8{0xde, 05}},
	{AFSUBW, yfmvx, Px, [23]uint8{0xde, 04}},
	{AFSUBL, yfmvx, Px, [23]uint8{0xda, 04}},
	{AFSUBF, yfmvx, Px, [23]uint8{0xd8, 04}},
	{AFSUBD, yfadd, Px, [23]uint8{0xdc, 04, 0xd8, 04, 0xdc, 05}},
	{AFSUBRDP, yfaddp, Px, [23]uint8{0xde, 04}},
	{AFSUBRW, yfmvx, Px, [23]uint8{0xde, 05}},
	{AFSUBRL, yfmvx, Px, [23]uint8{0xda, 05}},
	{AFSUBRF, yfmvx, Px, [23]uint8{0xd8, 05}},
	{AFSUBRD, yfadd, Px, [23]uint8{0xdc, 05, 0xd8, 05, 0xdc, 04}},
	{AFDIVDP, yfaddp, Px, [23]uint8{0xde, 07}},
	{AFDIVW, yfmvx, Px, [23]uint8{0xde, 06}},
	{AFDIVL, yfmvx, Px, [23]uint8{0xda, 06}},
	{AFDIVF, yfmvx, Px, [23]uint8{0xd8, 06}},
	{AFDIVD, yfadd, Px, [23]uint8{0xdc, 06, 0xd8, 06, 0xdc, 07}},
	{AFDIVRDP, yfaddp, Px, [23]uint8{0xde, 06}},
	{AFDIVRW, yfmvx, Px, [23]uint8{0xde, 07}},
	{AFDIVRL, yfmvx, Px, [23]uint8{0xda, 07}},
	{AFDIVRF, yfmvx, Px, [23]uint8{0xd8, 07}},
	{AFDIVRD, yfadd, Px, [23]uint8{0xdc, 07, 0xd8, 07, 0xdc, 06}},
	{AFXCHD, yfxch, Px, [23]uint8{0xd9, 01, 0xd9, 01}},
	{AFFREE, nil, 0, [23]uint8{}},
	{AFLDCW, ystcw, Px, [23]uint8{0xd9, 05, 0xd9, 05}},
	{AFLDENV, ystcw, Px, [23]uint8{0xd9, 04, 0xd9, 04}},
	{AFRSTOR, ysvrs, Px, [23]uint8{0xdd, 04, 0xdd, 04}},
	{AFSAVE, ysvrs, Px, [23]uint8{0xdd, 06, 0xdd, 06}},
	{AFSTCW, ystcw, Px, [23]uint8{0xd9, 07, 0xd9, 07}},
	{AFSTENV, ystcw, Px, [23]uint8{0xd9, 06, 0xd9, 06}},
	{AFSTSW, ystsw, Px, [23]uint8{0xdd, 07, 0xdf, 0xe0}},
	{AF2XM1, ynone, Px, [23]uint8{0xd9, 0xf0}},
	{AFABS, ynone, Px, [23]uint8{0xd9, 0xe1}},
	{AFCHS, ynone, Px, [23]uint8{0xd9, 0xe0}},
	{AFCLEX, ynone, Px, [23]uint8{0xdb, 0xe2}},
	{AFCOS, ynone, Px, [23]uint8{0xd9, 0xff}},
	{AFDECSTP, ynone, Px, [23]uint8{0xd9, 0xf6}},
	{AFINCSTP, ynone, Px, [23]uint8{0xd9, 0xf7}},
	{AFINIT, ynone, Px, [23]uint8{0xdb, 0xe3}},
	{AFLD1, ynone, Px, [23]uint8{0xd9, 0xe8}},
	{AFLDL2E, ynone, Px, [23]uint8{0xd9, 0xea}},
	{AFLDL2T, ynone, Px, [23]uint8{0xd9, 0xe9}},
	{AFLDLG2, ynone, Px, [23]uint8{0xd9, 0xec}},
	{AFLDLN2, ynone, Px, [23]uint8{0xd9, 0xed}},
	{AFLDPI, ynone, Px, [23]uint8{0xd9, 0xeb}},
	{AFLDZ, ynone, Px, [23]uint8{0xd9, 0xee}},
	{AFNOP, ynone, Px, [23]uint8{0xd9, 0xd0}},
	{AFPATAN, ynone, Px, [23]uint8{0xd9, 0xf3}},
	{AFPREM, ynone, Px, [23]uint8{0xd9, 0xf8}},
	{AFPREM1, ynone, Px, [23]uint8{0xd9, 0xf5}},
	{AFPTAN, ynone, Px, [23]uint8{0xd9, 0xf2}},
	{AFRNDINT, ynone, Px, [23]uint8{0xd9, 0xfc}},
	{AFSCALE, ynone, Px, [23]uint8{0xd9, 0xfd}},
	{AFSIN, ynone, Px, [23]uint8{0xd9, 0xfe}},
	{AFSINCOS, ynone, Px, [23]uint8{0xd9, 0xfb}},
	{AFSQRT, ynone, Px, [23]uint8{0xd9, 0xfa}},
	{AFTST, ynone, Px, [23]uint8{0xd9, 0xe4}},
	{AFXAM, ynone, Px, [23]uint8{0xd9, 0xe5}},
	{AFXTRACT, ynone, Px, [23]uint8{0xd9, 0xf4}},
	{AFYL2X, ynone, Px, [23]uint8{0xd9, 0xf1}},
	{AFYL2XP1, ynone, Px, [23]uint8{0xd9, 0xf9}},
	{ACMPXCHGB, yrb_mb, Pb, [23]uint8{0x0f, 0xb0}},
	{ACMPXCHGL, yrl_ml, Px, [23]uint8{0x0f, 0xb1}},
	{ACMPXCHGW, yrl_ml, Pe, [23]uint8{0x0f, 0xb1}},
	{ACMPXCHGQ, yrl_ml, Pw, [23]uint8{0x0f, 0xb1}},
	{ACMPXCHG8B, yscond, Pm, [23]uint8{0xc7, 01}},
	{AINVD, ynone, Pm, [23]uint8{0x08}},
	{AINVLPG, ymbs, Pm, [23]uint8{0x01, 07}},
	{ALFENCE, ynone, Pm, [23]uint8{0xae, 0xe8}},
	{AMFENCE, ynone, Pm, [23]uint8{0xae, 0xf0}},
	{AMOVNTIL, yrl_ml, Pm, [23]uint8{0xc3}},
	{AMOVNTIQ, yrl_ml, Pw, [23]uint8{0x0f, 0xc3}},
	{ARDMSR, ynone, Pm, [23]uint8{0x32}},
	{ARDPMC, ynone, Pm, [23]uint8{0x33}},
	{ARDTSC, ynone, Pm, [23]uint8{0x31}},
	{ARSM, ynone, Pm, [23]uint8{0xaa}},
	{ASFENCE, ynone, Pm, [23]uint8{0xae, 0xf8}},
	{ASYSRET, ynone, Pm, [23]uint8{0x07}},
	{AWBINVD, ynone, Pm, [23]uint8{0x09}},
	{AWRMSR, ynone, Pm, [23]uint8{0x30}},
	{AXADDB, yrb_mb, Pb, [23]uint8{0x0f, 0xc0}},
	{AXADDL, yrl_ml, Px, [23]uint8{0x0f, 0xc1}},
	{AXADDQ, yrl_ml, Pw, [23]uint8{0x0f, 0xc1}},
	{AXADDW, yrl_ml, Pe, [23]uint8{0x0f, 0xc1}},
	{ACRC32B, ycrc32l, Px, [23]uint8{0xf2, 0x0f, 0x38, 0xf0, 0}},
	{ACRC32Q, ycrc32l, Pw, [23]uint8{0xf2, 0x0f, 0x38, 0xf1, 0}},
	{APREFETCHT0, yprefetch, Pm, [23]uint8{0x18, 01}},
	{APREFETCHT1, yprefetch, Pm, [23]uint8{0x18, 02}},
	{APREFETCHT2, yprefetch, Pm, [23]uint8{0x18, 03}},
	{APREFETCHNTA, yprefetch, Pm, [23]uint8{0x18, 00}},
	{AMOVQL, yrl_ml, Px, [23]uint8{0x89}},
	{AUNDEF, ynone, Px, [23]uint8{0x0f, 0x0b}},
	{AAESENC, yaes, Pq, [23]uint8{0x38, 0xdc, 0}},
	{AAESENCLAST, yaes, Pq, [23]uint8{0x38, 0xdd, 0}},
	{AAESDEC, yaes, Pq, [23]uint8{0x38, 0xde, 0}},
	{AAESDECLAST, yaes, Pq, [23]uint8{0x38, 0xdf, 0}},
	{AAESIMC, yaes, Pq, [23]uint8{0x38, 0xdb, 0}},
	{AAESKEYGENASSIST, yaes2, Pq, [23]uint8{0x3a, 0xdf, 0}},
	{APSHUFD, yaes2, Pq, [23]uint8{0x70, 0}},
	{APCLMULQDQ, yxshuf, Pq, [23]uint8{0x3a, 0x44, 0}},
	{AUSEFIELD, ynop, Px, [23]uint8{0, 0}},
	{ATYPE, nil, 0, [23]uint8{}},
	{AFUNCDATA, yfuncdata, Px, [23]uint8{0, 0}},
	{APCDATA, ypcdata, Px, [23]uint8{0, 0}},
	{ACHECKNIL, nil, 0, [23]uint8{}},
	{AVARDEF, nil, 0, [23]uint8{}},
	{AVARKILL, nil, 0, [23]uint8{}},
	{ADUFFCOPY, yduff, Px, [23]uint8{0xe8}},
	{ADUFFZERO, yduff, Px, [23]uint8{0xe8}},
	{AEND, nil, 0, [23]uint8{}},
	{0, nil, 0, [23]uint8{}},
}

var opindex [ALAST + 1]*Optab

// single-instruction no-ops of various lengths.
// constructed by hand and disassembled with gdb to verify.
// see http://www.agner.org/optimize/optimizing_assembly.pdf for discussion.
var nop = [][16]uint8{
	{0x90},
	{0x66, 0x90},
	{0x0F, 0x1F, 0x00},
	{0x0F, 0x1F, 0x40, 0x00},
	{0x0F, 0x1F, 0x44, 0x00, 0x00},
	{0x66, 0x0F, 0x1F, 0x44, 0x00, 0x00},
	{0x0F, 0x1F, 0x80, 0x00, 0x00, 0x00, 0x00},
	{0x0F, 0x1F, 0x84, 0x00, 0x00, 0x00, 0x00, 0x00},
	{0x66, 0x0F, 0x1F, 0x84, 0x00, 0x00, 0x00, 0x00, 0x00},
}

// Native Client rejects the repeated 0x66 prefix.
// {0x66, 0x66, 0x0F, 0x1F, 0x84, 0x00, 0x00, 0x00, 0x00, 0x00},
func fillnop(p []uint8, n int) {
	var m int
	for n > 0 {
		m = n
		if m > len(nop) {
			m = len(nop)
		}
		copy(p, nop[m-1][:m])
		p = p[m:]
		n -= m
	}
}

func naclpad(ctxt *liblink.Link, s *liblink.LSym, c int64, pad int) int64 {
	liblink.Symgrow(ctxt, s, c+int64(pad))
	fillnop(s.P[c:], pad)
	return c + int64(pad)
}

func spadjop(ctxt *liblink.Link, p *liblink.Prog, l int, q int) int {
	if p.Mode != 64 || ctxt.Arch.Ptrsize == 4 {
		return l
	}
	return q
}

func span6(ctxt *liblink.Link, s *liblink.LSym) {
	var p *liblink.Prog
	var q *liblink.Prog
	var c int64
	var v int
	var loop int
	var bp []uint8
	var n int
	var m int
	var i int
	ctxt.Cursym = s
	if s.P != nil {
		return
	}
	if ycover[0] == 0 {
		instinit()
	}
	for p = ctxt.Cursym.Text; p != nil; p = p.Link {
		n = 0
		if p.To.Typ == D_BRANCH {
			if p.Pcond == nil {
				p.Pcond = p
			}
		}
		q = p.Pcond
		if q != nil {
			if q.Back != 2 {
				n = 1
			}
		}
		p.Back = n
		if p.As == AADJSP {
			p.To.Typ = D_SP
			v = int(-p.From.Offset)
			p.From.Offset = int64(v)
			p.As = spadjop(ctxt, p, AADDL, AADDQ)
			if v < 0 {
				p.As = spadjop(ctxt, p, ASUBL, ASUBQ)
				v = -v
				p.From.Offset = int64(v)
			}
			if v == 0 {
				p.As = ANOP
			}
		}
	}
	for p = s.Text; p != nil; p = p.Link {
		p.Back = 2 // use short branches first time through
		q = p.Pcond
		if q != nil && (q.Back&2 != 0) {
			p.Back |= 1 // backward jump
			q.Back |= 4 // loop head
		}
		if p.As == AADJSP {
			p.To.Typ = D_SP
			v = int(-p.From.Offset)
			p.From.Offset = int64(v)
			p.As = spadjop(ctxt, p, AADDL, AADDQ)
			if v < 0 {
				p.As = spadjop(ctxt, p, ASUBL, ASUBQ)
				v = -v
				p.From.Offset = int64(v)
			}
			if v == 0 {
				p.As = ANOP
			}
		}
	}
	n = 0
	for {
		loop = 0
		for i = 0; i < len(s.R); i++ {
			s.R[i] = liblink.Reloc{}
		}
		s.R = s.R[:0]
		s.P = s.P[:0]
		c = 0
		for p = s.Text; p != nil; p = p.Link {
			if ctxt.Headtype == liblink.Hnacl && p.Isize > 0 {
				var deferreturn *liblink.LSym
				if deferreturn == nil {
					deferreturn = liblink.Linklookup(ctxt, "runtime.deferreturn", 0)
				}
				// pad everything to avoid crossing 32-byte boundary
				if c>>5 != (c+int64(p.Isize)-1)>>5 {
					c = naclpad(ctxt, s, c, int(-c&31))
				}
				// pad call deferreturn to start at 32-byte boundary
				// so that subtracting 5 in jmpdefer will jump back
				// to that boundary and rerun the call.
				if p.As == ACALL && p.To.Sym == deferreturn {
					c = naclpad(ctxt, s, c, int(-c&31))
				}
				// pad call to end at 32-byte boundary
				if p.As == ACALL {
					c = naclpad(ctxt, s, c, int(-(c+int64(p.Isize))&31))
				}
				// the linker treats REP and STOSQ as different instructions
				// but in fact the REP is a prefix on the STOSQ.
				// make sure REP has room for 2 more bytes, so that
				// padding will not be inserted before the next instruction.
				if (p.As == AREP || p.As == AREPN) && c>>5 != (c+3-1)>>5 {
					c = naclpad(ctxt, s, c, int(-c&31))
				}
				// same for LOCK.
				// various instructions follow; the longest is 4 bytes.
				// give ourselves 8 bytes so as to avoid surprises.
				if p.As == ALOCK && c>>5 != (c+8-1)>>5 {
					c = naclpad(ctxt, s, c, int(-c&31))
				}
			}
			if (p.Back&4 != 0) && c&(LoopAlign-1) != 0 {
				// pad with NOPs
				v = int(-c & int64(LoopAlign-1))
				if v <= MaxLoopPad {
					liblink.Symgrow(ctxt, s, c+int64(v))
					fillnop(s.P[c:], v)
					c += int64(v)
				}
			}
			p.Pc = c
			// process forward jumps to p
			for q = p.Comefrom; q != nil; q = q.Forwd {
				v = int(p.Pc - (q.Pc + int64(q.Mark)))
				if q.Back&2 != 0 { // short
					if v > 127 {
						loop++
						q.Back ^= 2
					}
					if q.As == AJCXZL {
						s.P[q.Pc+2] = uint8(v)
					} else {
						s.P[q.Pc+1] = uint8(v)
					}
				} else {
					bp = s.P[q.Pc+int64(q.Mark)-4:]
					bp[0] = uint8(v)
					bp = bp[1:]
					bp[0] = uint8(v >> 8)
					bp = bp[1:]
					bp[0] = uint8(v >> 16)
					bp = bp[1:]
					bp[0] = uint8(v >> 24)
				}
			}
			p.Comefrom = nil
			p.Pc = c
			asmins(ctxt, p)
			m = -cap(ctxt.Andptr) + cap(ctxt.And[:])
			if p.Isize != m {
				p.Isize = m
				loop++
			}
			liblink.Symgrow(ctxt, s, p.Pc+int64(m))
			copy(s.P[p.Pc:], ctxt.And[:m])
			p.Mark = m
			c += int64(m)
		}
		n++
		if n > 20 {
			ctxt.Diag("span must be looping")
			log.Fatalf("loop")
		}
		if loop == 0 {
			break
		}
	}
	if ctxt.Headtype == liblink.Hnacl {
		c = naclpad(ctxt, s, c, int(-c&31))
	}
	c += -c & (FuncAlign - 1)
	s.Size = c
	if false { /* debug['a'] > 1 */
		fmt.Printf("span1 %s %d (%d tries)\n %.6x", s.Name, s.Size, n, 0)
		for i = 0; i < len(s.P); i++ {
			fmt.Printf(" %.2x", s.P[i])
			if i%16 == 15 {
				fmt.Printf("\n  %.6x", uint(i+1))
			}
		}
		if i%16 != 0 {
			fmt.Printf("\n")
		}
		for i = 0; i < len(s.R); i++ {
			var r *liblink.Reloc
			r = &s.R[i]
			fmt.Printf(" rel %#.4x/%d %s%+d\n", uint64(r.Off), r.Siz, r.Sym.Name, r.Add)
		}
	}
}

func instinit() {
	var c int
	var i int
	for i = 1; optab[i].as != 0; i++ {
		c = optab[i].as
		if opindex[c] != nil {
			log.Fatalf("phase error in optab: %d (%v)", i, Aconv(c))
		}
		opindex[c] = &optab[i]
	}
	for i = 0; i < Ymax; i++ {
		ycover[i*Ymax+i] = 1
	}
	ycover[Yi0*Ymax+Yi8] = 1
	ycover[Yi1*Ymax+Yi8] = 1
	ycover[Yi0*Ymax+Ys32] = 1
	ycover[Yi1*Ymax+Ys32] = 1
	ycover[Yi8*Ymax+Ys32] = 1
	ycover[Yi0*Ymax+Yi32] = 1
	ycover[Yi1*Ymax+Yi32] = 1
	ycover[Yi8*Ymax+Yi32] = 1
	ycover[Ys32*Ymax+Yi32] = 1
	ycover[Yi0*Ymax+Yi64] = 1
	ycover[Yi1*Ymax+Yi64] = 1
	ycover[Yi8*Ymax+Yi64] = 1
	ycover[Ys32*Ymax+Yi64] = 1
	ycover[Yi32*Ymax+Yi64] = 1
	ycover[Yal*Ymax+Yrb] = 1
	ycover[Ycl*Ymax+Yrb] = 1
	ycover[Yax*Ymax+Yrb] = 1
	ycover[Ycx*Ymax+Yrb] = 1
	ycover[Yrx*Ymax+Yrb] = 1
	ycover[Yrl*Ymax+Yrb] = 1
	ycover[Ycl*Ymax+Ycx] = 1
	ycover[Yax*Ymax+Yrx] = 1
	ycover[Ycx*Ymax+Yrx] = 1
	ycover[Yax*Ymax+Yrl] = 1
	ycover[Ycx*Ymax+Yrl] = 1
	ycover[Yrx*Ymax+Yrl] = 1
	ycover[Yf0*Ymax+Yrf] = 1
	ycover[Yal*Ymax+Ymb] = 1
	ycover[Ycl*Ymax+Ymb] = 1
	ycover[Yax*Ymax+Ymb] = 1
	ycover[Ycx*Ymax+Ymb] = 1
	ycover[Yrx*Ymax+Ymb] = 1
	ycover[Yrb*Ymax+Ymb] = 1
	ycover[Yrl*Ymax+Ymb] = 1
	ycover[Ym*Ymax+Ymb] = 1
	ycover[Yax*Ymax+Yml] = 1
	ycover[Ycx*Ymax+Yml] = 1
	ycover[Yrx*Ymax+Yml] = 1
	ycover[Yrl*Ymax+Yml] = 1
	ycover[Ym*Ymax+Yml] = 1
	ycover[Yax*Ymax+Ymm] = 1
	ycover[Ycx*Ymax+Ymm] = 1
	ycover[Yrx*Ymax+Ymm] = 1
	ycover[Yrl*Ymax+Ymm] = 1
	ycover[Ym*Ymax+Ymm] = 1
	ycover[Ymr*Ymax+Ymm] = 1
	ycover[Ym*Ymax+Yxm] = 1
	ycover[Yxr*Ymax+Yxm] = 1
	for i = 0; i < D_NONE; i++ {
		reg[i] = -1
		if i >= D_AL && i <= D_R15B {
			reg[i] = (i - D_AL) & 7
			if i >= D_SPB && i <= D_DIB {
				regrex[i] = 0x40
			}
			if i >= D_R8B && i <= D_R15B {
				regrex[i] = Rxr | Rxx | Rxb
			}
		}
		if i >= D_AH && i <= D_BH {
			reg[i] = 4 + ((i - D_AH) & 7)
		}
		if i >= D_AX && i <= D_R15 {
			reg[i] = (i - D_AX) & 7
			if i >= D_R8 {
				regrex[i] = Rxr | Rxx | Rxb
			}
		}
		if i >= D_F0 && i <= D_F0+7 {
			reg[i] = (i - D_F0) & 7
		}
		if i >= D_M0 && i <= D_M0+7 {
			reg[i] = (i - D_M0) & 7
		}
		if i >= D_X0 && i <= D_X0+15 {
			reg[i] = (i - D_X0) & 7
			if i >= D_X0+8 {
				regrex[i] = Rxr | Rxx | Rxb
			}
		}
		if i >= D_CR+8 && i <= D_CR+15 {
			regrex[i] = Rxr
		}
	}
}

func prefixof(ctxt *liblink.Link, a *liblink.Addr) int {
	switch a.Typ {
	case D_INDIR + D_CS:
		return 0x2e
	case D_INDIR + D_DS:
		return 0x3e
	case D_INDIR + D_ES:
		return 0x26
	case D_INDIR + D_FS:
		return 0x64
	case D_INDIR + D_GS:
		return 0x65
	// NOTE: Systems listed here should be only systems that
	// support direct TLS references like 8(TLS) implemented as
	// direct references from FS or GS. Systems that require
	// the initial-exec model, where you load the TLS base into
	// a register and then index from that register, do not reach
	// this code and should not be listed.
	case D_INDIR + D_TLS:
		switch ctxt.Headtype {
		default:
			log.Fatalf("unknown TLS base register for %s", liblink.Headstr(ctxt.Headtype))
		case liblink.Hdragonfly,
			liblink.Hfreebsd,
			liblink.Hlinux,
			liblink.Hnetbsd,
			liblink.Hopenbsd,
			liblink.Hsolaris:
			return 0x64 // FS
		case liblink.Hdarwin:
			return 0x65 // GS
		}
	}
	switch a.Index {
	case D_CS:
		return 0x2e
	case D_DS:
		return 0x3e
	case D_ES:
		return 0x26
	case D_FS:
		return 0x64
	case D_GS:
		return 0x65
	}
	return 0
}

func oclass(ctxt *liblink.Link, a *liblink.Addr) int {
	var v int64
	var l int32
	if a.Typ >= D_INDIR || a.Index != D_NONE {
		if a.Index != D_NONE && a.Scale == 0 {
			if a.Typ == D_ADDR {
				switch a.Index {
				case D_EXTERN,
					D_STATIC:
					if ctxt.Flag_shared != 0 || ctxt.Headtype == liblink.Hnacl {
						return Yiauto
					} else {
						return Yi32 /* TO DO: Yi64 */
					}
					fallthrough
				case D_AUTO,
					D_PARAM:
					return Yiauto
				}
				return Yxxx
			}
			return Ycol
		}
		return Ym
	}
	switch a.Typ {
	case D_AL:
		return Yal
	case D_AX:
		return Yax
	/*
		case D_SPB:
	*/
	case D_BPB,
		D_SIB,
		D_DIB,
		D_R8B,
		D_R9B,
		D_R10B,
		D_R11B,
		D_R12B,
		D_R13B,
		D_R14B,
		D_R15B:
		if ctxt.Asmode != 64 {
			return Yxxx
		}
		fallthrough
	case D_DL,
		D_BL,
		D_AH,
		D_CH,
		D_DH,
		D_BH:
		return Yrb
	case D_CL:
		return Ycl
	case D_CX:
		return Ycx
	case D_DX,
		D_BX:
		return Yrx
	case D_R8, /* not really Yrl */
		D_R9,
		D_R10,
		D_R11,
		D_R12,
		D_R13,
		D_R14,
		D_R15:
		if ctxt.Asmode != 64 {
			return Yxxx
		}
		fallthrough
	case D_SP,
		D_BP,
		D_SI,
		D_DI:
		return Yrl
	case D_F0 + 0:
		return Yf0
	case D_F0 + 1,
		D_F0 + 2,
		D_F0 + 3,
		D_F0 + 4,
		D_F0 + 5,
		D_F0 + 6,
		D_F0 + 7:
		return Yrf
	case D_M0 + 0,
		D_M0 + 1,
		D_M0 + 2,
		D_M0 + 3,
		D_M0 + 4,
		D_M0 + 5,
		D_M0 + 6,
		D_M0 + 7:
		return Ymr
	case D_X0 + 0,
		D_X0 + 1,
		D_X0 + 2,
		D_X0 + 3,
		D_X0 + 4,
		D_X0 + 5,
		D_X0 + 6,
		D_X0 + 7,
		D_X0 + 8,
		D_X0 + 9,
		D_X0 + 10,
		D_X0 + 11,
		D_X0 + 12,
		D_X0 + 13,
		D_X0 + 14,
		D_X0 + 15:
		return Yxr
	case D_NONE:
		return Ynone
	case D_CS:
		return Ycs
	case D_SS:
		return Yss
	case D_DS:
		return Yds
	case D_ES:
		return Yes
	case D_FS:
		return Yfs
	case D_GS:
		return Ygs
	case D_TLS:
		return Ytls
	case D_GDTR:
		return Ygdtr
	case D_IDTR:
		return Yidtr
	case D_LDTR:
		return Yldtr
	case D_MSW:
		return Ymsw
	case D_TASK:
		return Ytask
	case D_CR + 0:
		return Ycr0
	case D_CR + 1:
		return Ycr1
	case D_CR + 2:
		return Ycr2
	case D_CR + 3:
		return Ycr3
	case D_CR + 4:
		return Ycr4
	case D_CR + 5:
		return Ycr5
	case D_CR + 6:
		return Ycr6
	case D_CR + 7:
		return Ycr7
	case D_CR + 8:
		return Ycr8
	case D_DR + 0:
		return Ydr0
	case D_DR + 1:
		return Ydr1
	case D_DR + 2:
		return Ydr2
	case D_DR + 3:
		return Ydr3
	case D_DR + 4:
		return Ydr4
	case D_DR + 5:
		return Ydr5
	case D_DR + 6:
		return Ydr6
	case D_DR + 7:
		return Ydr7
	case D_TR + 0:
		return Ytr0
	case D_TR + 1:
		return Ytr1
	case D_TR + 2:
		return Ytr2
	case D_TR + 3:
		return Ytr3
	case D_TR + 4:
		return Ytr4
	case D_TR + 5:
		return Ytr5
	case D_TR + 6:
		return Ytr6
	case D_TR + 7:
		return Ytr7
	case D_EXTERN,
		D_STATIC,
		D_AUTO,
		D_PARAM:
		return Ym
	case D_CONST,
		D_ADDR:
		if a.Sym == nil {
			v = a.Offset
			if v == 0 {
				return Yi0
			}
			if v == 1 {
				return Yi1
			}
			if v >= -128 && v <= 127 {
				return Yi8
			}
			l = int32(v)
			if int64(l) == v {
				return Ys32 /* can sign extend */
			}
			if v>>32 == 0 {
				return Yi32 /* unsigned */
			}
			return Yi64
		}
		return Yi32 /* TO DO: D_ADDR as Yi64 */
	case D_BRANCH:
		return Ybr
	}
	return Yxxx
}

func asmidx(ctxt *liblink.Link, scale int, index int, base int) {
	var i int
	switch index {
	default:
		goto bad
	case D_NONE:
		i = 4 << 3
		goto bas
	case D_R8,
		D_R9,
		D_R10,
		D_R11,
		D_R12,
		D_R13,
		D_R14,
		D_R15:
		if ctxt.Asmode != 64 {
			goto bad
		}
		fallthrough
	case D_AX,
		D_CX,
		D_DX,
		D_BX,
		D_BP,
		D_SI,
		D_DI:
		i = reg[index] << 3
		break
	}
	switch scale {
	default:
		goto bad
	case 1:
		break
	case 2:
		i |= 1 << 6
	case 4:
		i |= 2 << 6
	case 8:
		i |= 3 << 6
		break
	}
bas:
	switch base {
	default:
		goto bad
	case D_NONE: /* must be mod=00 */
		i |= 5
	case D_R8,
		D_R9,
		D_R10,
		D_R11,
		D_R12,
		D_R13,
		D_R14,
		D_R15:
		if ctxt.Asmode != 64 {
			goto bad
		}
		fallthrough
	case D_AX,
		D_CX,
		D_DX,
		D_BX,
		D_SP,
		D_BP,
		D_SI,
		D_DI:
		i |= reg[base]
		break
	}
	ctxt.Andptr[0] = uint8(i)
	ctxt.Andptr = ctxt.Andptr[1:]
	return
bad:
	ctxt.Diag("asmidx: bad address %d/%d/%d", scale, index, base)
	ctxt.Andptr[0] = 0
	ctxt.Andptr = ctxt.Andptr[1:]
	return
}

func put4(ctxt *liblink.Link, v int64) {
	ctxt.Andptr[0] = uint8(v)
	ctxt.Andptr[1] = uint8(v >> 8)
	ctxt.Andptr[2] = uint8(v >> 16)
	ctxt.Andptr[3] = uint8(v >> 24)
	ctxt.Andptr = ctxt.Andptr[4:]
}

func relput4(ctxt *liblink.Link, p *liblink.Prog, a *liblink.Addr) {
	var v int64
	var rel liblink.Reloc
	var r *liblink.Reloc
	v = vaddr(ctxt, a, &rel)
	if rel.Siz != 0 {
		if rel.Siz != 4 {
			ctxt.Diag("bad reloc")
		}
		r = liblink.Addrel(ctxt.Cursym)
		*r = rel
		r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
	}
	put4(ctxt, v)
}

func put8(ctxt *liblink.Link, v int64) {
	ctxt.Andptr[0] = uint8(v)
	ctxt.Andptr[1] = uint8(v >> 8)
	ctxt.Andptr[2] = uint8(v >> 16)
	ctxt.Andptr[3] = uint8(v >> 24)
	ctxt.Andptr[4] = uint8(v >> 32)
	ctxt.Andptr[5] = uint8(v >> 40)
	ctxt.Andptr[6] = uint8(v >> 48)
	ctxt.Andptr[7] = uint8(v >> 56)
	ctxt.Andptr = ctxt.Andptr[8:]
}

/*
static void
relput8(Prog *p, Addr *a)
{
	vlong v;
	Reloc rel, *r;

	v = vaddr(ctxt, a, &rel);
	if(rel.siz != 0) {
		r = addrel(ctxt->cursym);
		*r = rel;
		r->siz = 8;
		r->off = p->pc + ctxt->andptr - ctxt->and;
	}
	put8(ctxt, v);
}
*/
func vaddr(ctxt *liblink.Link, a *liblink.Addr, r *liblink.Reloc) int64 {
	var t int
	var v int64
	var s *liblink.LSym
	if r != nil {
		*r = liblink.Reloc{}
	}
	t = a.Typ
	v = a.Offset
	if t == D_ADDR {
		t = a.Index
	}
	switch t {
	case D_STATIC,
		D_EXTERN:
		s = a.Sym
		if r == nil {
			ctxt.Diag("need reloc for %D", a)
			log.Fatalf("reloc")
		}
		r.Siz = 4  // TODO: 8 for external symbols
		r.Off = -1 // caller must fill in
		r.Sym = s
		r.Add = v
		v = 0
		if ctxt.Flag_shared != 0 || ctxt.Headtype == liblink.Hnacl {
			if s.Typ == liblink.STLSBSS {
				r.Xadd = r.Add - int64(r.Siz)
				r.Typ = liblink.R_TLS
				r.Xsym = s
			} else {
				r.Typ = liblink.R_PCREL
			}
		} else {
			r.Typ = liblink.R_ADDR
		}
	case D_INDIR + D_TLS:
		if r == nil {
			ctxt.Diag("need reloc for %D", a)
			log.Fatalf("reloc")
		}
		r.Typ = liblink.R_TLS_LE
		r.Siz = 4
		r.Off = -1 // caller must fill in
		r.Add = v
		v = 0
		break
	}
	return v
}

func asmandsz(ctxt *liblink.Link, a *liblink.Addr, r int, rex int, m64 int) {
	var v int64
	var t int
	var scale int
	var rel liblink.Reloc
	rex &= 0x40 | Rxr
	v = a.Offset
	t = a.Typ
	rel.Siz = 0
	if a.Index != D_NONE && a.Index != D_TLS {
		if t < D_INDIR {
			switch t {
			default:
				goto bad
			case D_STATIC,
				D_EXTERN:
				if ctxt.Flag_shared != 0 || ctxt.Headtype == liblink.Hnacl {
					goto bad
				}
				t = D_NONE
				v = vaddr(ctxt, a, &rel)
			case D_AUTO,
				D_PARAM:
				t = D_SP
				break
			}
		} else {
			t -= D_INDIR
		}
		ctxt.Rexflag |= regrex[int(a.Index)]&Rxx | regrex[t]&Rxb | rex
		if t == D_NONE {
			ctxt.Andptr[0] = uint8(0<<6 | 4<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			asmidx(ctxt, int(a.Scale), a.Index, t)
			goto putrelv
		}
		if v == 0 && rel.Siz == 0 && t != D_BP && t != D_R13 {
			ctxt.Andptr[0] = uint8(0<<6 | 4<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			asmidx(ctxt, int(a.Scale), a.Index, t)
			return
		}
		if v >= -128 && v < 128 && rel.Siz == 0 {
			ctxt.Andptr[0] = uint8(1<<6 | 4<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			asmidx(ctxt, int(a.Scale), a.Index, t)
			ctxt.Andptr[0] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[1:]
			return
		}
		ctxt.Andptr[0] = uint8(2<<6 | 4<<0 | r<<3)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmidx(ctxt, int(a.Scale), a.Index, t)
		goto putrelv
	}
	if t >= D_AL && t <= D_X0+15 {
		if v != 0 {
			goto bad
		}
		ctxt.Andptr[0] = uint8(3<<6 | reg[t]<<0 | r<<3)
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Rexflag |= regrex[t]&(0x40|Rxb) | rex
		return
	}
	scale = int(a.Scale)
	if t < D_INDIR {
		switch a.Typ {
		default:
			goto bad
		case D_STATIC,
			D_EXTERN:
			t = D_NONE
			v = vaddr(ctxt, a, &rel)
		case D_AUTO,
			D_PARAM:
			t = D_SP
			break
		}
		scale = 1
	} else {
		t -= D_INDIR
	}
	if t == D_TLS {
		v = vaddr(ctxt, a, &rel)
	}
	ctxt.Rexflag |= regrex[t]&Rxb | rex
	if t == D_NONE || (D_CS <= t && t <= D_GS) || t == D_TLS {
		if (ctxt.Flag_shared != 0 || ctxt.Headtype == liblink.Hnacl) && t == D_NONE && (a.Typ == D_STATIC || a.Typ == D_EXTERN) || ctxt.Asmode != 64 {
			ctxt.Andptr[0] = uint8(0<<6 | 5<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			goto putrelv
		}
		/* temporary */
		ctxt.Andptr[0] = uint8(0<<6 | 4<<0 | r<<3)
		ctxt.Andptr = ctxt.Andptr[1:] /* sib present */
		ctxt.Andptr[0] = 0<<6 | 4<<3 | 5<<0
		ctxt.Andptr = ctxt.Andptr[1:] /* DS:d32 */
		goto putrelv
	}
	if t == D_SP || t == D_R12 {
		if v == 0 {
			ctxt.Andptr[0] = uint8(0<<6 | reg[t]<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			asmidx(ctxt, scale, D_NONE, t)
			return
		}
		if v >= -128 && v < 128 {
			ctxt.Andptr[0] = uint8(1<<6 | reg[t]<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			asmidx(ctxt, scale, D_NONE, t)
			ctxt.Andptr[0] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[1:]
			return
		}
		ctxt.Andptr[0] = uint8(2<<6 | reg[t]<<0 | r<<3)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmidx(ctxt, scale, D_NONE, t)
		goto putrelv
	}
	if t >= D_AX && t <= D_R15 {
		if a.Index == D_TLS {
			rel = liblink.Reloc{}
			rel.Typ = liblink.R_TLS_IE
			rel.Siz = 4
			rel.Sym = nil
			rel.Add = v
			v = 0
		}
		if v == 0 && rel.Siz == 0 && t != D_BP && t != D_R13 {
			ctxt.Andptr[0] = uint8(0<<6 | reg[t]<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			return
		}
		if v >= -128 && v < 128 && rel.Siz == 0 {
			ctxt.Andptr[0] = uint8(1<<6 | reg[t]<<0 | r<<3)
			ctxt.Andptr[1] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[2:]
			return
		}
		ctxt.Andptr[0] = uint8(2<<6 | reg[t]<<0 | r<<3)
		ctxt.Andptr = ctxt.Andptr[1:]
		goto putrelv
	}
	goto bad
putrelv:
	if rel.Siz != 0 {
		var r *liblink.Reloc
		if rel.Siz != 4 {
			ctxt.Diag("bad rel")
			goto bad
		}
		r = liblink.Addrel(ctxt.Cursym)
		*r = rel
		r.Off = ctxt.Curp.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
	}
	put4(ctxt, v)
	return
bad:
	ctxt.Diag("asmand: bad address %D", a)
	return
}

func asmand(ctxt *liblink.Link, a *liblink.Addr, ra *liblink.Addr) {
	asmandsz(ctxt, a, reg[ra.Typ], regrex[ra.Typ], 0)
}

func asmando(ctxt *liblink.Link, a *liblink.Addr, o int) {
	asmandsz(ctxt, a, o, 0, 0)
}

func bytereg(a *liblink.Addr, t *uint8) {
	if a.Index == D_NONE && (a.Typ >= D_AX && a.Typ <= D_R15) {
		a.Typ = D_AL + (a.Typ - D_AX)
		*t = 0
	}
}

const (
	E = 0xff
)

var ymovtab = []Movtab{
	/* push */
	{APUSHL, Ycs, Ynone, 0, [4]uint8{0x0e, E, 0, 0}},
	{APUSHL, Yss, Ynone, 0, [4]uint8{0x16, E, 0, 0}},
	{APUSHL, Yds, Ynone, 0, [4]uint8{0x1e, E, 0, 0}},
	{APUSHL, Yes, Ynone, 0, [4]uint8{0x06, E, 0, 0}},
	{APUSHL, Yfs, Ynone, 0, [4]uint8{0x0f, 0xa0, E, 0}},
	{APUSHL, Ygs, Ynone, 0, [4]uint8{0x0f, 0xa8, E, 0}},
	{APUSHQ, Yfs, Ynone, 0, [4]uint8{0x0f, 0xa0, E, 0}},
	{APUSHQ, Ygs, Ynone, 0, [4]uint8{0x0f, 0xa8, E, 0}},
	{APUSHW, Ycs, Ynone, 0, [4]uint8{Pe, 0x0e, E, 0}},
	{APUSHW, Yss, Ynone, 0, [4]uint8{Pe, 0x16, E, 0}},
	{APUSHW, Yds, Ynone, 0, [4]uint8{Pe, 0x1e, E, 0}},
	{APUSHW, Yes, Ynone, 0, [4]uint8{Pe, 0x06, E, 0}},
	{APUSHW, Yfs, Ynone, 0, [4]uint8{Pe, 0x0f, 0xa0, E}},
	{APUSHW, Ygs, Ynone, 0, [4]uint8{Pe, 0x0f, 0xa8, E}},
	/* pop */
	{APOPL, Ynone, Yds, 0, [4]uint8{0x1f, E, 0, 0}},
	{APOPL, Ynone, Yes, 0, [4]uint8{0x07, E, 0, 0}},
	{APOPL, Ynone, Yss, 0, [4]uint8{0x17, E, 0, 0}},
	{APOPL, Ynone, Yfs, 0, [4]uint8{0x0f, 0xa1, E, 0}},
	{APOPL, Ynone, Ygs, 0, [4]uint8{0x0f, 0xa9, E, 0}},
	{APOPQ, Ynone, Yfs, 0, [4]uint8{0x0f, 0xa1, E, 0}},
	{APOPQ, Ynone, Ygs, 0, [4]uint8{0x0f, 0xa9, E, 0}},
	{APOPW, Ynone, Yds, 0, [4]uint8{Pe, 0x1f, E, 0}},
	{APOPW, Ynone, Yes, 0, [4]uint8{Pe, 0x07, E, 0}},
	{APOPW, Ynone, Yss, 0, [4]uint8{Pe, 0x17, E, 0}},
	{APOPW, Ynone, Yfs, 0, [4]uint8{Pe, 0x0f, 0xa1, E}},
	{APOPW, Ynone, Ygs, 0, [4]uint8{Pe, 0x0f, 0xa9, E}},
	/* mov seg */
	{AMOVW, Yes, Yml, 1, [4]uint8{0x8c, 0, 0, 0}},
	{AMOVW, Ycs, Yml, 1, [4]uint8{0x8c, 1, 0, 0}},
	{AMOVW, Yss, Yml, 1, [4]uint8{0x8c, 2, 0, 0}},
	{AMOVW, Yds, Yml, 1, [4]uint8{0x8c, 3, 0, 0}},
	{AMOVW, Yfs, Yml, 1, [4]uint8{0x8c, 4, 0, 0}},
	{AMOVW, Ygs, Yml, 1, [4]uint8{0x8c, 5, 0, 0}},
	{AMOVW, Yml, Yes, 2, [4]uint8{0x8e, 0, 0, 0}},
	{AMOVW, Yml, Ycs, 2, [4]uint8{0x8e, 1, 0, 0}},
	{AMOVW, Yml, Yss, 2, [4]uint8{0x8e, 2, 0, 0}},
	{AMOVW, Yml, Yds, 2, [4]uint8{0x8e, 3, 0, 0}},
	{AMOVW, Yml, Yfs, 2, [4]uint8{0x8e, 4, 0, 0}},
	{AMOVW, Yml, Ygs, 2, [4]uint8{0x8e, 5, 0, 0}},
	/* mov cr */
	{AMOVL, Ycr0, Yml, 3, [4]uint8{0x0f, 0x20, 0, 0}},
	{AMOVL, Ycr2, Yml, 3, [4]uint8{0x0f, 0x20, 2, 0}},
	{AMOVL, Ycr3, Yml, 3, [4]uint8{0x0f, 0x20, 3, 0}},
	{AMOVL, Ycr4, Yml, 3, [4]uint8{0x0f, 0x20, 4, 0}},
	{AMOVL, Ycr8, Yml, 3, [4]uint8{0x0f, 0x20, 8, 0}},
	{AMOVQ, Ycr0, Yml, 3, [4]uint8{0x0f, 0x20, 0, 0}},
	{AMOVQ, Ycr2, Yml, 3, [4]uint8{0x0f, 0x20, 2, 0}},
	{AMOVQ, Ycr3, Yml, 3, [4]uint8{0x0f, 0x20, 3, 0}},
	{AMOVQ, Ycr4, Yml, 3, [4]uint8{0x0f, 0x20, 4, 0}},
	{AMOVQ, Ycr8, Yml, 3, [4]uint8{0x0f, 0x20, 8, 0}},
	{AMOVL, Yml, Ycr0, 4, [4]uint8{0x0f, 0x22, 0, 0}},
	{AMOVL, Yml, Ycr2, 4, [4]uint8{0x0f, 0x22, 2, 0}},
	{AMOVL, Yml, Ycr3, 4, [4]uint8{0x0f, 0x22, 3, 0}},
	{AMOVL, Yml, Ycr4, 4, [4]uint8{0x0f, 0x22, 4, 0}},
	{AMOVL, Yml, Ycr8, 4, [4]uint8{0x0f, 0x22, 8, 0}},
	{AMOVQ, Yml, Ycr0, 4, [4]uint8{0x0f, 0x22, 0, 0}},
	{AMOVQ, Yml, Ycr2, 4, [4]uint8{0x0f, 0x22, 2, 0}},
	{AMOVQ, Yml, Ycr3, 4, [4]uint8{0x0f, 0x22, 3, 0}},
	{AMOVQ, Yml, Ycr4, 4, [4]uint8{0x0f, 0x22, 4, 0}},
	{AMOVQ, Yml, Ycr8, 4, [4]uint8{0x0f, 0x22, 8, 0}},
	/* mov dr */
	{AMOVL, Ydr0, Yml, 3, [4]uint8{0x0f, 0x21, 0, 0}},
	{AMOVL, Ydr6, Yml, 3, [4]uint8{0x0f, 0x21, 6, 0}},
	{AMOVL, Ydr7, Yml, 3, [4]uint8{0x0f, 0x21, 7, 0}},
	{AMOVQ, Ydr0, Yml, 3, [4]uint8{0x0f, 0x21, 0, 0}},
	{AMOVQ, Ydr6, Yml, 3, [4]uint8{0x0f, 0x21, 6, 0}},
	{AMOVQ, Ydr7, Yml, 3, [4]uint8{0x0f, 0x21, 7, 0}},
	{AMOVL, Yml, Ydr0, 4, [4]uint8{0x0f, 0x23, 0, 0}},
	{AMOVL, Yml, Ydr6, 4, [4]uint8{0x0f, 0x23, 6, 0}},
	{AMOVL, Yml, Ydr7, 4, [4]uint8{0x0f, 0x23, 7, 0}},
	{AMOVQ, Yml, Ydr0, 4, [4]uint8{0x0f, 0x23, 0, 0}},
	{AMOVQ, Yml, Ydr6, 4, [4]uint8{0x0f, 0x23, 6, 0}},
	{AMOVQ, Yml, Ydr7, 4, [4]uint8{0x0f, 0x23, 7, 0}},
	/* mov tr */
	{AMOVL, Ytr6, Yml, 3, [4]uint8{0x0f, 0x24, 6, 0}},
	{AMOVL, Ytr7, Yml, 3, [4]uint8{0x0f, 0x24, 7, 0}},
	{AMOVL, Yml, Ytr6, 4, [4]uint8{0x0f, 0x26, 6, E}},
	{AMOVL, Yml, Ytr7, 4, [4]uint8{0x0f, 0x26, 7, E}},
	/* lgdt, sgdt, lidt, sidt */
	{AMOVL, Ym, Ygdtr, 4, [4]uint8{0x0f, 0x01, 2, 0}},
	{AMOVL, Ygdtr, Ym, 3, [4]uint8{0x0f, 0x01, 0, 0}},
	{AMOVL, Ym, Yidtr, 4, [4]uint8{0x0f, 0x01, 3, 0}},
	{AMOVL, Yidtr, Ym, 3, [4]uint8{0x0f, 0x01, 1, 0}},
	{AMOVQ, Ym, Ygdtr, 4, [4]uint8{0x0f, 0x01, 2, 0}},
	{AMOVQ, Ygdtr, Ym, 3, [4]uint8{0x0f, 0x01, 0, 0}},
	{AMOVQ, Ym, Yidtr, 4, [4]uint8{0x0f, 0x01, 3, 0}},
	{AMOVQ, Yidtr, Ym, 3, [4]uint8{0x0f, 0x01, 1, 0}},
	/* lldt, sldt */
	{AMOVW, Yml, Yldtr, 4, [4]uint8{0x0f, 0x00, 2, 0}},
	{AMOVW, Yldtr, Yml, 3, [4]uint8{0x0f, 0x00, 0, 0}},
	/* lmsw, smsw */
	{AMOVW, Yml, Ymsw, 4, [4]uint8{0x0f, 0x01, 6, 0}},
	{AMOVW, Ymsw, Yml, 3, [4]uint8{0x0f, 0x01, 4, 0}},
	/* ltr, str */
	{AMOVW, Yml, Ytask, 4, [4]uint8{0x0f, 0x00, 3, 0}},
	{AMOVW, Ytask, Yml, 3, [4]uint8{0x0f, 0x00, 1, 0}},
	/* load full pointer */
	{AMOVL, Yml, Ycol, 5, [4]uint8{0, 0, 0, 0}},
	{AMOVW, Yml, Ycol, 5, [4]uint8{Pe, 0, 0, 0}},
	/* double shift */
	{ASHLL, Ycol, Yml, 6, [4]uint8{0xa4, 0xa5, 0, 0}},
	{ASHRL, Ycol, Yml, 6, [4]uint8{0xac, 0xad, 0, 0}},
	{ASHLQ, Ycol, Yml, 6, [4]uint8{Pw, 0xa4, 0xa5, 0}},
	{ASHRQ, Ycol, Yml, 6, [4]uint8{Pw, 0xac, 0xad, 0}},
	{ASHLW, Ycol, Yml, 6, [4]uint8{Pe, 0xa4, 0xa5, 0}},
	{ASHRW, Ycol, Yml, 6, [4]uint8{Pe, 0xac, 0xad, 0}},
	/* load TLS base */
	{AMOVQ, Ytls, Yrl, 7, [4]uint8{0, 0, 0, 0}},
	{0, 0, 0, 0, [4]uint8{}},
}

func isax(a *liblink.Addr) bool {
	switch a.Typ {
	case D_AX,
		D_AL,
		D_AH,
		D_INDIR + D_AX:
		return true
	}
	if a.Index == D_AX {
		return true
	}
	return false
}

func subreg(p *liblink.Prog, from int, to int) {
	if false { /*debug['Q']*/
		fmt.Printf("\n%v\ts/%v/%v/\n", p, Rconv(from), Rconv(to))
	}
	if p.From.Typ == from {
		p.From.Typ = to
	}
	if p.To.Typ == from {
		p.To.Typ = to
	}
	if p.From.Index == from {
		p.From.Index = to
	}
	if p.To.Index == from {
		p.To.Index = to
	}
	from += D_INDIR
	if p.From.Typ == from {
		p.From.Typ = to + D_INDIR
	}
	if p.To.Typ == from {
		p.To.Typ = to + D_INDIR
	}
	if false { /*debug['Q']*/
		fmt.Printf("%v\n", p)
	}
}

func mediaop(ctxt *liblink.Link, o *Optab, op int, osize int, z int) int {
	switch op {
	case Pm,
		Pe,
		Pf2,
		Pf3:
		if osize != 1 {
			if op != Pm {
				ctxt.Andptr[0] = uint8(op)
				ctxt.Andptr = ctxt.Andptr[1:]
			}
			ctxt.Andptr[0] = Pm
			ctxt.Andptr = ctxt.Andptr[1:]
			z++
			op = int(o.op[z])
			break
		}
		fallthrough
	default:
		if -cap(ctxt.Andptr) == -cap(ctxt.And) || ctxt.And[-cap(ctxt.Andptr)+cap(ctxt.And[:])-1] != Pm {
			ctxt.Andptr[0] = Pm
			ctxt.Andptr = ctxt.Andptr[1:]
		}
		break
	}
	ctxt.Andptr[0] = uint8(op)
	ctxt.Andptr = ctxt.Andptr[1:]
	return z
}

func doasm(ctxt *liblink.Link, p *liblink.Prog) {
	var o *Optab
	var q *liblink.Prog
	var pp liblink.Prog
	var t []uint8
	var mo []Movtab
	var z int
	var op int
	var ft int
	var tt int
	var xo int
	var l int
	var pre int
	var v int64
	var rel liblink.Reloc
	var r *liblink.Reloc
	var a *liblink.Addr
	ctxt.Curp = p // TODO
	o = opindex[p.As]
	if o == nil {
		ctxt.Diag("asmins: missing op %P", p)
		return
	}
	pre = prefixof(ctxt, &p.From)
	if pre != 0 {
		ctxt.Andptr[0] = uint8(pre)
		ctxt.Andptr = ctxt.Andptr[1:]
	}
	pre = prefixof(ctxt, &p.To)
	if pre != 0 {
		ctxt.Andptr[0] = uint8(pre)
		ctxt.Andptr = ctxt.Andptr[1:]
	}
	if p.Ft == 0 {
		p.Ft = uint8(oclass(ctxt, &p.From))
	}
	if p.Tt == 0 {
		p.Tt = uint8(oclass(ctxt, &p.To))
	}
	ft = int(p.Ft) * Ymax
	tt = int(p.Tt) * Ymax
	t = o.ytab
	if t == nil {
		ctxt.Diag("asmins: noproto %P", p)
		return
	}
	xo = bool2int(o.op[0] == 0x0f)
	for z = 0; t[0] != 0; (func() { z += int(t[3]) + xo; t = t[4:] })() {
		if ycover[ft+int(t[0])] != 0 {
			if ycover[tt+int(t[1])] != 0 {
				goto found
			}
		}
	}
	goto domov
found:
	switch o.prefix {
	case Pq: /* 16 bit escape and opcode escape */
		ctxt.Andptr[0] = Pe
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = Pm
		ctxt.Andptr = ctxt.Andptr[1:]
	case Pq3: /* 16 bit escape, Rex.w, and opcode escape */
		ctxt.Andptr[0] = Pe
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = Pw
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = Pm
		ctxt.Andptr = ctxt.Andptr[1:]
	case Pf2, /* xmm opcode escape */
		Pf3:
		ctxt.Andptr[0] = uint8(o.prefix)
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = Pm
		ctxt.Andptr = ctxt.Andptr[1:]
	case Pm: /* opcode escape */
		ctxt.Andptr[0] = Pm
		ctxt.Andptr = ctxt.Andptr[1:]
	case Pe: /* 16 bit escape */
		ctxt.Andptr[0] = Pe
		ctxt.Andptr = ctxt.Andptr[1:]
	case Pw: /* 64-bit escape */
		if p.Mode != 64 {
			ctxt.Diag("asmins: illegal 64: %P", p)
		}
		ctxt.Rexflag |= Pw
	case Pb: /* botch */
		bytereg(&p.From, &p.Ft)
		bytereg(&p.To, &p.Tt)
	case P32: /* 32 bit but illegal if 64-bit mode */
		if p.Mode == 64 {
			ctxt.Diag("asmins: illegal in 64-bit mode: %P", p)
		}
	case Py: /* 64-bit only, no prefix */
		if p.Mode != 64 {
			ctxt.Diag("asmins: illegal in %d-bit mode: %P", p.Mode, p)
		}
		break
	}
	if z >= len(o.op) {
		log.Fatalf("asmins bad table %v", p)
	}
	op = int(o.op[z])
	if op == 0x0f {
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		z++
		op = int(o.op[z])
	}
	switch t[2] {
	default:
		ctxt.Diag("asmins: unknown z %d %P", t[2], p)
		return
	case Zpseudo:
		break
	case Zlit:
		for ; ; z++ {
			op = int(o.op[z])
			if op == 0 {
				break
			}
			ctxt.Andptr[0] = uint8(op)
			ctxt.Andptr = ctxt.Andptr[1:]
		}
	case Zlitm_r:
		for ; ; z++ {
			op = int(o.op[z])
			if op == 0 {
				break
			}
			ctxt.Andptr[0] = uint8(op)
			ctxt.Andptr = ctxt.Andptr[1:]
		}
		asmand(ctxt, &p.From, &p.To)
	case Zmb_r:
		bytereg(&p.From, &p.Ft)
		fallthrough
	/* fall through */
	case Zm_r:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, &p.To)
	case Zm2_r:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = o.op[z+1]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, &p.To)
	case Zm_r_xm:
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.From, &p.To)
	case Zm_r_xm_nr:
		ctxt.Rexflag = 0
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.From, &p.To)
	case Zm_r_i_xm:
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.From, &p.To)
		ctxt.Andptr[0] = uint8(p.To.Offset)
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zm_r_3d:
		ctxt.Andptr[0] = 0x0f
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = 0x0f
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, &p.To)
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zibm_r:
		for {
			tmp1 := z
			z++
			op = int(o.op[tmp1])
			if op == 0 {
				break
			}
			ctxt.Andptr[0] = uint8(op)
			ctxt.Andptr = ctxt.Andptr[1:]
		}
		asmand(ctxt, &p.From, &p.To)
		ctxt.Andptr[0] = uint8(p.To.Offset)
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zaut_r:
		ctxt.Andptr[0] = 0x8d
		ctxt.Andptr = ctxt.Andptr[1:] /* leal */
		if p.From.Typ != D_ADDR {
			ctxt.Diag("asmins: Zaut sb type ADDR")
		}
		p.From.Typ = p.From.Index
		p.From.Index = D_NONE
		asmand(ctxt, &p.From, &p.To)
		p.From.Index = p.From.Typ
		p.From.Typ = D_ADDR
	case Zm_o:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmando(ctxt, &p.From, int(o.op[z+1]))
	case Zr_m:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, &p.From)
	case Zr_m_xm:
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.To, &p.From)
	case Zr_m_xm_nr:
		ctxt.Rexflag = 0
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.To, &p.From)
	case Zr_m_i_xm:
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.To, &p.From)
		ctxt.Andptr[0] = uint8(p.From.Offset)
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zo_m:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmando(ctxt, &p.To, int(o.op[z+1]))
	case Zcallindreg:
		r = liblink.Addrel(ctxt.Cursym)
		r.Off = p.Pc
		r.Typ = liblink.R_CALLIND
		r.Siz = 0
		fallthrough
	// fallthrough
	case Zo_m64:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmandsz(ctxt, &p.To, int(o.op[z+1]), 0, 1)
	case Zm_ibo:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmando(ctxt, &p.From, int(o.op[z+1]))
		ctxt.Andptr[0] = uint8(vaddr(ctxt, &p.To, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zibo_m:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmando(ctxt, &p.To, int(o.op[z+1]))
		ctxt.Andptr[0] = uint8(vaddr(ctxt, &p.From, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zibo_m_xm:
		z = mediaop(ctxt, o, op, int(t[3]), z)
		asmando(ctxt, &p.To, int(o.op[z+1]))
		ctxt.Andptr[0] = uint8(vaddr(ctxt, &p.From, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Z_ib,
		Zib_:
		if t[2] == Zib_ {
			a = &p.From
		} else {
			a = &p.To
		}
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = uint8(vaddr(ctxt, a, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zib_rp:
		ctxt.Rexflag |= regrex[p.To.Typ] & (Rxb | 0x40)
		ctxt.Andptr[0] = uint8(op + reg[p.To.Typ])
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = uint8(vaddr(ctxt, &p.From, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zil_rp:
		ctxt.Rexflag |= regrex[p.To.Typ] & Rxb
		ctxt.Andptr[0] = uint8(op + reg[p.To.Typ])
		ctxt.Andptr = ctxt.Andptr[1:]
		if o.prefix == Pe {
			v = vaddr(ctxt, &p.From, nil)
			ctxt.Andptr[0] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = uint8(v >> 8)
			ctxt.Andptr = ctxt.Andptr[1:]
		} else {
			relput4(ctxt, p, &p.From)
		}
	case Zo_iw:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		if p.From.Typ != D_NONE {
			v = vaddr(ctxt, &p.From, nil)
			ctxt.Andptr[0] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = uint8(v >> 8)
			ctxt.Andptr = ctxt.Andptr[1:]
		}
	case Ziq_rp:
		v = vaddr(ctxt, &p.From, &rel)
		l = int(v >> 32)
		if l == 0 && rel.Siz != 8 {
			//p->mark |= 0100;
			//print("zero: %llux %P\n", v, p);
			ctxt.Rexflag &^= (0x40 | Rxw)
			ctxt.Rexflag |= regrex[p.To.Typ] & Rxb
			ctxt.Andptr[0] = uint8(0xb8 + reg[p.To.Typ])
			ctxt.Andptr = ctxt.Andptr[1:]
			if rel.Typ != 0 {
				r = liblink.Addrel(ctxt.Cursym)
				*r = rel
				r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
			}
			put4(ctxt, v)
		} else if l == -1 && uint64(v)&(uint64(1)<<31) != 0 { /* sign extend */
			//p->mark |= 0100;
			//print("sign: %llux %P\n", v, p);
			ctxt.Andptr[0] = 0xc7
			ctxt.Andptr = ctxt.Andptr[1:]
			asmando(ctxt, &p.To, 0)
			put4(ctxt, v) /* need all 8 */
		} else {
			//print("all: %llux %P\n", v, p);
			ctxt.Rexflag |= regrex[p.To.Typ] & Rxb
			ctxt.Andptr[0] = uint8(op + reg[p.To.Typ])
			ctxt.Andptr = ctxt.Andptr[1:]
			if rel.Typ != 0 {
				r = liblink.Addrel(ctxt.Cursym)
				*r = rel
				r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
			}
			put8(ctxt, v)
		}
	case Zib_rr:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, &p.To)
		ctxt.Andptr[0] = uint8(vaddr(ctxt, &p.From, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Z_il,
		Zil_:
		if t[2] == Zil_ {
			a = &p.From
		} else {
			a = &p.To
		}
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		if o.prefix == Pe {
			v = vaddr(ctxt, a, nil)
			ctxt.Andptr[0] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = uint8(v >> 8)
			ctxt.Andptr = ctxt.Andptr[1:]
		} else {
			relput4(ctxt, p, a)
		}
	case Zm_ilo,
		Zilo_m:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		if t[2] == Zilo_m {
			a = &p.From
			asmando(ctxt, &p.To, int(o.op[z+1]))
		} else {
			a = &p.To
			asmando(ctxt, &p.From, int(o.op[z+1]))
		}
		if o.prefix == Pe {
			v = vaddr(ctxt, a, nil)
			ctxt.Andptr[0] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = uint8(v >> 8)
			ctxt.Andptr = ctxt.Andptr[1:]
		} else {
			relput4(ctxt, p, a)
		}
	case Zil_rr:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, &p.To)
		if o.prefix == Pe {
			v = vaddr(ctxt, &p.From, nil)
			ctxt.Andptr[0] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = uint8(v >> 8)
			ctxt.Andptr = ctxt.Andptr[1:]
		} else {
			relput4(ctxt, p, &p.From)
		}
	case Z_rp:
		ctxt.Rexflag |= regrex[p.To.Typ] & (Rxb | 0x40)
		ctxt.Andptr[0] = uint8(op + reg[p.To.Typ])
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zrp_:
		ctxt.Rexflag |= regrex[p.From.Typ] & (Rxb | 0x40)
		ctxt.Andptr[0] = uint8(op + reg[p.From.Typ])
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zclr:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, &p.To)
	case Zcall:
		if p.To.Sym == nil {
			ctxt.Diag("call without target")
			log.Fatalf("bad code")
		}
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		r = liblink.Addrel(ctxt.Cursym)
		r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
		r.Sym = p.To.Sym
		r.Add = p.To.Offset
		r.Typ = liblink.R_CALL
		r.Siz = 4
		put4(ctxt, 0)
	// TODO: jump across functions needs reloc
	case Zbr,
		Zjmp,
		Zloop:
		if p.To.Sym != nil {
			if t[2] != Zjmp {
				ctxt.Diag("branch to ATEXT")
				log.Fatalf("bad code")
			}
			ctxt.Andptr[0] = o.op[z+1]
			ctxt.Andptr = ctxt.Andptr[1:]
			r = liblink.Addrel(ctxt.Cursym)
			r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
			r.Sym = p.To.Sym
			r.Typ = liblink.R_PCREL
			r.Siz = 4
			put4(ctxt, 0)
			break
		}
		// Assumes q is in this function.
		// TODO: Check in input, preserve in brchain.
		// Fill in backward jump now.
		q = p.Pcond
		if q == nil {
			ctxt.Diag("jmp/branch/loop without target")
			log.Fatalf("bad code")
		}
		if p.Back&1 != 0 {
			v = q.Pc - (p.Pc + 2)
			if v >= -128 {
				if p.As == AJCXZL {
					ctxt.Andptr[0] = 0x67
					ctxt.Andptr = ctxt.Andptr[1:]
				}
				ctxt.Andptr[0] = uint8(op)
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(v)
				ctxt.Andptr = ctxt.Andptr[1:]
			} else if t[2] == Zloop {
				ctxt.Diag("loop too far: %P", p)
			} else {
				v -= 5 - 2
				if t[2] == Zbr {
					ctxt.Andptr[0] = 0x0f
					ctxt.Andptr = ctxt.Andptr[1:]
					v--
				}
				ctxt.Andptr[0] = o.op[z+1]
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(v)
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(v >> 8)
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(v >> 16)
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(v >> 24)
				ctxt.Andptr = ctxt.Andptr[1:]
			}
			break
		}
		// Annotate target; will fill in later.
		p.Forwd = q.Comefrom
		q.Comefrom = p
		if p.Back&2 != 0 { // short
			if p.As == AJCXZL {
				ctxt.Andptr[0] = 0x67
				ctxt.Andptr = ctxt.Andptr[1:]
			}
			ctxt.Andptr[0] = uint8(op)
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = 0
			ctxt.Andptr = ctxt.Andptr[1:]
		} else if t[2] == Zloop {
			ctxt.Diag("loop too far: %P", p)
		} else {
			if t[2] == Zbr {
				ctxt.Andptr[0] = 0x0f
				ctxt.Andptr = ctxt.Andptr[1:]
			}
			ctxt.Andptr[0] = o.op[z+1]
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = 0
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = 0
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = 0
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = 0
			ctxt.Andptr = ctxt.Andptr[1:]
		}
		break
		/*
			v = q->pc - p->pc - 2;
			if((v >= -128 && v <= 127) || p->pc == -1 || q->pc == -1) {
				*ctxt->andptr++ = op;
				*ctxt->andptr++ = v;
			} else {
				v -= 5-2;
				if(t[2] == Zbr) {
					*ctxt->andptr++ = 0x0f;
					v--;
				}
				*ctxt->andptr++ = o->op[z+1];
				*ctxt->andptr++ = v;
				*ctxt->andptr++ = v>>8;
				*ctxt->andptr++ = v>>16;
				*ctxt->andptr++ = v>>24;
			}
		*/

	case Zbyte:
		v = vaddr(ctxt, &p.From, &rel)
		if rel.Siz != 0 {
			rel.Siz = uint8(op)
			r = liblink.Addrel(ctxt.Cursym)
			*r = rel
			r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
		}
		ctxt.Andptr[0] = uint8(v)
		ctxt.Andptr = ctxt.Andptr[1:]
		if op > 1 {
			ctxt.Andptr[0] = uint8(v >> 8)
			ctxt.Andptr = ctxt.Andptr[1:]
			if op > 2 {
				ctxt.Andptr[0] = uint8(v >> 16)
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(v >> 24)
				ctxt.Andptr = ctxt.Andptr[1:]
				if op > 4 {
					ctxt.Andptr[0] = uint8(v >> 32)
					ctxt.Andptr = ctxt.Andptr[1:]
					ctxt.Andptr[0] = uint8(v >> 40)
					ctxt.Andptr = ctxt.Andptr[1:]
					ctxt.Andptr[0] = uint8(v >> 48)
					ctxt.Andptr = ctxt.Andptr[1:]
					ctxt.Andptr[0] = uint8(v >> 56)
					ctxt.Andptr = ctxt.Andptr[1:]
				}
			}
		}
		break
	}
	return
domov:
	for mo = ymovtab; mo[0].as != 0; mo = mo[1:] {
		if p.As == mo[0].as {
			if ycover[ft+int(mo[0].ft)] != 0 {
				if ycover[tt+int(mo[0].tt)] != 0 {
					t = mo[0].op[:]
					goto mfound
				}
			}
		}
	}
bad:
	if p.Mode != 64 {
		/*
		 * here, the assembly has failed.
		 * if its a byte instruction that has
		 * unaddressable registers, try to
		 * exchange registers and reissue the
		 * instruction with the operands renamed.
		 */
		pp = *p
		z = p.From.Typ
		if z >= D_BP && z <= D_DI {
			if isax(&p.To) || p.To.Typ == D_NONE {
				// We certainly don't want to exchange
				// with AX if the op is MUL or DIV.
				ctxt.Andptr[0] = 0x87
				ctxt.Andptr = ctxt.Andptr[1:] /* xchg lhs,bx */
				asmando(ctxt, &p.From, reg[D_BX])
				subreg(&pp, z, D_BX)
				doasm(ctxt, &pp)
				ctxt.Andptr[0] = 0x87
				ctxt.Andptr = ctxt.Andptr[1:] /* xchg lhs,bx */
				asmando(ctxt, &p.From, reg[D_BX])
			} else {
				ctxt.Andptr[0] = uint8(0x90 + reg[z])
				ctxt.Andptr = ctxt.Andptr[1:] /* xchg lsh,ax */
				subreg(&pp, z, D_AX)
				doasm(ctxt, &pp)
				ctxt.Andptr[0] = uint8(0x90 + reg[z])
				ctxt.Andptr = ctxt.Andptr[1:] /* xchg lsh,ax */
			}
			return
		}
		z = p.To.Typ
		if z >= D_BP && z <= D_DI {
			if isax(&p.From) {
				ctxt.Andptr[0] = 0x87
				ctxt.Andptr = ctxt.Andptr[1:] /* xchg rhs,bx */
				asmando(ctxt, &p.To, reg[D_BX])
				subreg(&pp, z, D_BX)
				doasm(ctxt, &pp)
				ctxt.Andptr[0] = 0x87
				ctxt.Andptr = ctxt.Andptr[1:] /* xchg rhs,bx */
				asmando(ctxt, &p.To, reg[D_BX])
			} else {
				ctxt.Andptr[0] = uint8(0x90 + reg[z])
				ctxt.Andptr = ctxt.Andptr[1:] /* xchg rsh,ax */
				subreg(&pp, z, D_AX)
				doasm(ctxt, &pp)
				ctxt.Andptr[0] = uint8(0x90 + reg[z])
				ctxt.Andptr = ctxt.Andptr[1:] /* xchg rsh,ax */
			}
			return
		}
	}
	ctxt.Diag("doasm: notfound from=%ux to=%ux %P", p.From.Typ, p.To.Typ, p)
	return
mfound:
	switch mo[0].code {
	default:
		ctxt.Diag("asmins: unknown mov %d %P", mo[0].code, p)
	case 0: /* lit */
		for z = 0; t[z] != E; z++ {
			ctxt.Andptr[0] = t[z]
			ctxt.Andptr = ctxt.Andptr[1:]
		}
	case 1: /* r,m */
		ctxt.Andptr[0] = t[0]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmando(ctxt, &p.To, int(t[1]))
	case 2: /* m,r */
		ctxt.Andptr[0] = t[0]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmando(ctxt, &p.From, int(t[1]))
	case 3: /* r,m - 2op */
		ctxt.Andptr[0] = t[0]
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = t[1]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmando(ctxt, &p.To, int(t[2]))
		ctxt.Rexflag |= regrex[p.From.Typ] & (Rxr | 0x40)
	case 4: /* m,r - 2op */
		ctxt.Andptr[0] = t[0]
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = t[1]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmando(ctxt, &p.From, int(t[2]))
		ctxt.Rexflag |= regrex[p.To.Typ] & (Rxr | 0x40)
	case 5: /* load full pointer, trash heap */
		if t[0] != 0 {
			ctxt.Andptr[0] = t[0]
			ctxt.Andptr = ctxt.Andptr[1:]
		}
		switch p.To.Index {
		default:
			goto bad
		case D_DS:
			ctxt.Andptr[0] = 0xc5
			ctxt.Andptr = ctxt.Andptr[1:]
		case D_SS:
			ctxt.Andptr[0] = 0x0f
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = 0xb2
			ctxt.Andptr = ctxt.Andptr[1:]
		case D_ES:
			ctxt.Andptr[0] = 0xc4
			ctxt.Andptr = ctxt.Andptr[1:]
		case D_FS:
			ctxt.Andptr[0] = 0x0f
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = 0xb4
			ctxt.Andptr = ctxt.Andptr[1:]
		case D_GS:
			ctxt.Andptr[0] = 0x0f
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = 0xb5
			ctxt.Andptr = ctxt.Andptr[1:]
			break
		}
		asmand(ctxt, &p.From, &p.To)
	case 6: /* double shift */
		if t[0] == Pw {
			if p.Mode != 64 {
				ctxt.Diag("asmins: illegal 64: %P", p)
			}
			ctxt.Rexflag |= Pw
			t = t[1:]
		} else if t[0] == Pe {
			ctxt.Andptr[0] = Pe
			ctxt.Andptr = ctxt.Andptr[1:]
			t = t[1:]
		}
		z = p.From.Typ
		switch z {
		default:
			goto bad
		case D_CONST:
			ctxt.Andptr[0] = 0x0f
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = t[0]
			ctxt.Andptr = ctxt.Andptr[1:]
			asmandsz(ctxt, &p.To, reg[int(p.From.Index)], regrex[int(p.From.Index)], 0)
			ctxt.Andptr[0] = uint8(p.From.Offset)
			ctxt.Andptr = ctxt.Andptr[1:]
		case D_CL,
			D_CX:
			ctxt.Andptr[0] = 0x0f
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = t[1]
			ctxt.Andptr = ctxt.Andptr[1:]
			asmandsz(ctxt, &p.To, reg[int(p.From.Index)], regrex[int(p.From.Index)], 0)
			break
		}
	// NOTE: The systems listed here are the ones that use the "TLS initial exec" model,
	// where you load the TLS base register into a register and then index off that
	// register to access the actual TLS variables. Systems that allow direct TLS access
	// are handled in prefixof above and should not be listed here.
	case 7: /* mov tls, r */
		switch ctxt.Headtype {
		default:
			log.Fatalf("unknown TLS base location for %s", liblink.Headstr(ctxt.Headtype))
		case liblink.Hplan9:
			if ctxt.Plan9privates == nil {
				ctxt.Plan9privates = liblink.Linklookup(ctxt, "_privates", 0)
			}
			pp.From = liblink.Addr{}
			pp.From.Typ = D_EXTERN
			pp.From.Sym = ctxt.Plan9privates
			pp.From.Offset = 0
			pp.From.Index = D_NONE
			ctxt.Rexflag |= Pw
			ctxt.Andptr[0] = 0x8B
			ctxt.Andptr = ctxt.Andptr[1:]
			asmand(ctxt, &pp.From, &p.To)
		// TLS base is 0(FS).
		case liblink.Hsolaris: // TODO(rsc): Delete Hsolaris from list. Should not use this code. See progedit in obj6.c.
			pp.From = p.From
			pp.From.Typ = D_INDIR + D_NONE
			pp.From.Offset = 0
			pp.From.Index = D_NONE
			pp.From.Scale = 0
			ctxt.Rexflag |= Pw
			ctxt.Andptr[0] = 0x64
			ctxt.Andptr = ctxt.Andptr[1:] // FS
			ctxt.Andptr[0] = 0x8B
			ctxt.Andptr = ctxt.Andptr[1:]
			asmand(ctxt, &pp.From, &p.To)
		// Windows TLS base is always 0x28(GS).
		case liblink.Hwindows:
			pp.From = p.From
			pp.From.Typ = D_INDIR + D_GS
			pp.From.Offset = 0x28
			pp.From.Index = D_NONE
			pp.From.Scale = 0
			ctxt.Rexflag |= Pw
			ctxt.Andptr[0] = 0x65
			ctxt.Andptr = ctxt.Andptr[1:] // GS
			ctxt.Andptr[0] = 0x8B
			ctxt.Andptr = ctxt.Andptr[1:]
			asmand(ctxt, &pp.From, &p.To)
			break
		}
		break
	}
}

var naclret = []uint8{
	0x5e, // POPL SI
	// 0x8b, 0x7d, 0x00, // MOVL (BP), DI - catch return to invalid address, for debugging
	0x83,
	0xe6,
	0xe0, // ANDL $~31, SI
	0x4c,
	0x01,
	0xfe, // ADDQ R15, SI
	0xff,
	0xe6, // JMP SI
}

var naclspfix = []uint8{0x4c, 0x01, 0xfc} // ADDQ R15, SP

var naclbpfix = []uint8{0x4c, 0x01, 0xfd} // ADDQ R15, BP

var naclmovs = []uint8{
	0x89,
	0xf6, // MOVL SI, SI
	0x49,
	0x8d,
	0x34,
	0x37, // LEAQ (R15)(SI*1), SI
	0x89,
	0xff, // MOVL DI, DI
	0x49,
	0x8d,
	0x3c,
	0x3f, // LEAQ (R15)(DI*1), DI
}

var naclstos = []uint8{
	0x89,
	0xff, // MOVL DI, DI
	0x49,
	0x8d,
	0x3c,
	0x3f, // LEAQ (R15)(DI*1), DI
}

func nacltrunc(ctxt *liblink.Link, reg int) {
	if reg >= D_R8 {
		ctxt.Andptr[0] = 0x45
		ctxt.Andptr = ctxt.Andptr[1:]
	}
	reg = (reg - D_AX) & 7
	ctxt.Andptr[0] = 0x89
	ctxt.Andptr = ctxt.Andptr[1:]
	ctxt.Andptr[0] = uint8(3<<6 | reg<<3 | reg)
	ctxt.Andptr = ctxt.Andptr[1:]
}

func asmins(ctxt *liblink.Link, p *liblink.Prog) {
	var i int
	var n int
	var np int
	var c int
	var and0 []uint8
	var r *liblink.Reloc
	ctxt.Andptr = ctxt.And[:]
	ctxt.Asmode = p.Mode
	if p.As == AUSEFIELD {
		r = liblink.Addrel(ctxt.Cursym)
		r.Off = 0
		r.Siz = 0
		r.Sym = p.From.Sym
		r.Typ = liblink.R_USEFIELD
		return
	}
	if ctxt.Headtype == liblink.Hnacl {
		if p.As == AREP {
			ctxt.Rep++
			return
		}
		if p.As == AREPN {
			ctxt.Repn++
			return
		}
		if p.As == ALOCK {
			ctxt.Lock++
			return
		}
		if p.As != ALEAQ && p.As != ALEAL {
			if p.From.Index != D_NONE && p.From.Scale > 0 {
				nacltrunc(ctxt, p.From.Index)
			}
			if p.To.Index != D_NONE && p.To.Scale > 0 {
				nacltrunc(ctxt, p.To.Index)
			}
		}
		switch p.As {
		case ARET:
			copy(ctxt.Andptr, naclret)
			ctxt.Andptr = ctxt.Andptr[len(naclret):]
			return
		case ACALL,
			AJMP:
			if D_AX <= p.To.Typ && p.To.Typ <= D_DI {
				// ANDL $~31, reg
				ctxt.Andptr[0] = 0x83
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(0xe0 | (p.To.Typ - D_AX))
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = 0xe0
				ctxt.Andptr = ctxt.Andptr[1:]
				// ADDQ R15, reg
				ctxt.Andptr[0] = 0x4c
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = 0x01
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(0xf8 | (p.To.Typ - D_AX))
				ctxt.Andptr = ctxt.Andptr[1:]
			}
			if D_R8 <= p.To.Typ && p.To.Typ <= D_R15 {
				// ANDL $~31, reg
				ctxt.Andptr[0] = 0x41
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = 0x83
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(0xe0 | (p.To.Typ - D_R8))
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = 0xe0
				ctxt.Andptr = ctxt.Andptr[1:]
				// ADDQ R15, reg
				ctxt.Andptr[0] = 0x4d
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = 0x01
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(0xf8 | (p.To.Typ - D_R8))
				ctxt.Andptr = ctxt.Andptr[1:]
			}
		case AINT:
			ctxt.Andptr[0] = 0xf4
			ctxt.Andptr = ctxt.Andptr[1:]
			return
		case ASCASB,
			ASCASW,
			ASCASL,
			ASCASQ,
			ASTOSB,
			ASTOSW,
			ASTOSL,
			ASTOSQ:
			copy(ctxt.Andptr, naclstos)
			ctxt.Andptr = ctxt.Andptr[len(naclstos):]
		case AMOVSB,
			AMOVSW,
			AMOVSL,
			AMOVSQ:
			copy(ctxt.Andptr, naclmovs)
			ctxt.Andptr = ctxt.Andptr[len(naclmovs):]
			break
		}
		if ctxt.Rep != 0 {
			ctxt.Andptr[0] = 0xf3
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Rep = 0
		}
		if ctxt.Repn != 0 {
			ctxt.Andptr[0] = 0xf2
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Repn = 0
		}
		if ctxt.Lock != 0 {
			ctxt.Andptr[0] = 0xf0
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Lock = 0
		}
	}
	ctxt.Rexflag = 0
	and0 = ctxt.Andptr
	ctxt.Asmode = p.Mode
	doasm(ctxt, p)
	if ctxt.Rexflag != 0 {
		/*
		 * as befits the whole approach of the architecture,
		 * the rex prefix must appear before the first opcode byte
		 * (and thus after any 66/67/f2/f3/26/2e/3e prefix bytes, but
		 * before the 0f opcode escape!), or it might be ignored.
		 * note that the handbook often misleadingly shows 66/f2/f3 in `opcode'.
		 */
		if p.Mode != 64 {
			ctxt.Diag("asmins: illegal in mode %d: %P", p.Mode, p)
		}
		n = -cap(ctxt.Andptr) + cap(and0)
		for np = 0; np < n; np++ {
			c = int(and0[np])
			if c != 0xf2 && c != 0xf3 && (c < 0x64 || c > 0x67) && c != 0x2e && c != 0x3e && c != 0x26 {
				break
			}
		}
		copy(and0[np+1:], and0[np:][:n-np])
		and0[np] = uint8(0x40 | ctxt.Rexflag)
		ctxt.Andptr = ctxt.Andptr[1:]
	}
	n = -cap(ctxt.Andptr) + cap(ctxt.And[:])
	for i = len(ctxt.Cursym.R) - 1; i >= 0; i-- {
		r = &ctxt.Cursym.R[i:][0]
		if r.Off < p.Pc {
			break
		}
		if ctxt.Rexflag != 0 {
			r.Off++
		}
		if r.Typ == liblink.R_PCREL || r.Typ == liblink.R_CALL {
			r.Add -= p.Pc + int64(n) - (r.Off + int64(r.Siz))
		}
	}
	if ctxt.Headtype == liblink.Hnacl && p.As != ACMPL && p.As != ACMPQ {
		switch p.To.Typ {
		case D_SP:
			copy(ctxt.Andptr, naclspfix)
			ctxt.Andptr = ctxt.Andptr[len(naclspfix):]
		case D_BP:
			copy(ctxt.Andptr, naclbpfix)
			ctxt.Andptr = ctxt.Andptr[len(naclbpfix):]
			break
		}
	}
}
