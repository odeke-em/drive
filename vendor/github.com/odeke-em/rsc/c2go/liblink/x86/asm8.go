package x86

import (
	"fmt"
	"log"

	"code.google.com/p/rsc/c2go/liblink"
)

/*
 * this is the ranlib header
 */
const (
	MaxAlign  = 32
	FuncAlign = 16
)

type Optab struct {
	as     int
	ytab   []uint8
	prefix int
	op     [13]uint8
}

const (
	Yxxx = 0 + iota
	Ynone
	Yi0
	Yi1
	Yi8
	Yi32
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
	Ytls
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
	Ymr
	Ymm
	Yxr
	Yxm
	Ymax
	Zxxx = 0 + iota - 62
	Zlit
	Zlitm_r
	Z_rp
	Zbr
	Zcall
	Zcallcon
	Zcallind
	Zcallindreg
	Zib_
	Zib_rp
	Zibo_m
	Zil_
	Zil_rp
	Zilo_m
	Zjmp
	Zjmpcon
	Zloop
	Zm_o
	Zm_r
	Zm2_r
	Zm_r_xm
	Zm_r_i_xm
	Zaut_r
	Zo_m
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
	Zibm_r
	Zbyte
	Zmov
	Zmax
	Px  = 0
	Pe  = 0x66
	Pm  = 0x0f
	Pq  = 0xff
	Pb  = 0xfe
	Pf2 = 0xf2
	Pf3 = 0xf3
)

var ycover [Ymax * Ymax]uint8

var reg [D_NONE]int

var ynone = []uint8{
	Ynone,
	Ynone,
	Zlit,
	1,
	0,
}

var ytext = []uint8{
	Ymb,
	Yi32,
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
	Yiauto,
	Ynone,
	Zpseudo,
	0,
	Ynone,
	Yxr,
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

var yincl = []uint8{
	Ynone,
	Yrl,
	Z_rp,
	1,
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
	1 + 2,
	//	Yi0,	Yml,	Zibo_m,	2,	// shorter but slower AND $0,dst
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
	1,
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
	1 + 2,
	//	Yi0,	Yml,	Zibo_m,	2,	// shorter but slower AND $0,dst
	Yi32,
	Yrl,
	Zil_rp,
	1,
	Yi32,
	Yml,
	Zilo_m,
	2,
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
	1,
	0,
}

var ymovq = []uint8{
	Yml,
	Yxr,
	Zm_r_xm,
	2,
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
	Zm_r,
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

var yrb_mb = []uint8{
	Yrb,
	Ymb,
	Zr_m,
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
	0,
}

var ybyte = []uint8{
	Yi32,
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
	1,
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
	Ycol,
	Zcallind,
	2,
	Ynone,
	Ybr,
	Zcall,
	0,
	Ynone,
	Yi32,
	Zcallcon,
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
	Zo_m,
	2,
	Ynone,
	Ybr,
	Zjmp,
	0,
	Ynone,
	Yi32,
	Zjmpcon,
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

var yfcmv = []uint8{
	Yrf,
	Yf0,
	Zm_o,
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

var yxmq = []uint8{
	Yxm,
	Yxr,
	Zm_r_xm,
	2,
	0,
}

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

/*
static uchar	yxcvfq[] =
{
	Yxm,	Yrl,	Zm_r_xm,	2,
	0
};
static uchar	yxcvqf[] =
{
	Yml,	Yxr,	Zm_r_xm,	2,
	0
};
*/
var yxrrl = []uint8{
	Yxr,
	Yrl,
	Zm_r,
	1,
	0,
}

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

var yinsrd = []uint8{
	Yml,
	Yxr,
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

var optab = /*	as, ytab, andproto, opcode */
[]Optab{
	{AXXX, nil, 0, [13]uint8{}},
	{AAAA, ynone, Px, [13]uint8{0x37}},
	{AAAD, ynone, Px, [13]uint8{0xd5, 0x0a}},
	{AAAM, ynone, Px, [13]uint8{0xd4, 0x0a}},
	{AAAS, ynone, Px, [13]uint8{0x3f}},
	{AADCB, yxorb, Pb, [13]uint8{0x14, 0x80, 02, 0x10, 0x10}},
	{AADCL, yxorl, Px, [13]uint8{0x83, 02, 0x15, 0x81, 02, 0x11, 0x13}},
	{AADCW, yxorl, Pe, [13]uint8{0x83, 02, 0x15, 0x81, 02, 0x11, 0x13}},
	{AADDB, yxorb, Px, [13]uint8{0x04, 0x80, 00, 0x00, 0x02}},
	{AADDL, yaddl, Px, [13]uint8{0x83, 00, 0x05, 0x81, 00, 0x01, 0x03}},
	{AADDW, yaddl, Pe, [13]uint8{0x83, 00, 0x05, 0x81, 00, 0x01, 0x03}},
	{AADJSP, nil, 0, [13]uint8{}},
	{AANDB, yxorb, Pb, [13]uint8{0x24, 0x80, 04, 0x20, 0x22}},
	{AANDL, yxorl, Px, [13]uint8{0x83, 04, 0x25, 0x81, 04, 0x21, 0x23}},
	{AANDW, yxorl, Pe, [13]uint8{0x83, 04, 0x25, 0x81, 04, 0x21, 0x23}},
	{AARPL, yrl_ml, Px, [13]uint8{0x63}},
	{ABOUNDL, yrl_m, Px, [13]uint8{0x62}},
	{ABOUNDW, yrl_m, Pe, [13]uint8{0x62}},
	{ABSFL, yml_rl, Pm, [13]uint8{0xbc}},
	{ABSFW, yml_rl, Pq, [13]uint8{0xbc}},
	{ABSRL, yml_rl, Pm, [13]uint8{0xbd}},
	{ABSRW, yml_rl, Pq, [13]uint8{0xbd}},
	{ABTL, yml_rl, Pm, [13]uint8{0xa3}},
	{ABTW, yml_rl, Pq, [13]uint8{0xa3}},
	{ABTCL, yml_rl, Pm, [13]uint8{0xbb}},
	{ABTCW, yml_rl, Pq, [13]uint8{0xbb}},
	{ABTRL, yml_rl, Pm, [13]uint8{0xb3}},
	{ABTRW, yml_rl, Pq, [13]uint8{0xb3}},
	{ABTSL, yml_rl, Pm, [13]uint8{0xab}},
	{ABTSW, yml_rl, Pq, [13]uint8{0xab}},
	{ABYTE, ybyte, Px, [13]uint8{1}},
	{ACALL, ycall, Px, [13]uint8{0xff, 02, 0xff, 0x15, 0xe8}},
	{ACLC, ynone, Px, [13]uint8{0xf8}},
	{ACLD, ynone, Px, [13]uint8{0xfc}},
	{ACLI, ynone, Px, [13]uint8{0xfa}},
	{ACLTS, ynone, Pm, [13]uint8{0x06}},
	{ACMC, ynone, Px, [13]uint8{0xf5}},
	{ACMPB, ycmpb, Pb, [13]uint8{0x3c, 0x80, 07, 0x38, 0x3a}},
	{ACMPL, ycmpl, Px, [13]uint8{0x83, 07, 0x3d, 0x81, 07, 0x39, 0x3b}},
	{ACMPW, ycmpl, Pe, [13]uint8{0x83, 07, 0x3d, 0x81, 07, 0x39, 0x3b}},
	{ACMPSB, ynone, Pb, [13]uint8{0xa6}},
	{ACMPSL, ynone, Px, [13]uint8{0xa7}},
	{ACMPSW, ynone, Pe, [13]uint8{0xa7}},
	{ADAA, ynone, Px, [13]uint8{0x27}},
	{ADAS, ynone, Px, [13]uint8{0x2f}},
	{ADATA, nil, 0, [13]uint8{}},
	{ADECB, yincb, Pb, [13]uint8{0xfe, 01}},
	{ADECL, yincl, Px, [13]uint8{0x48, 0xff, 01}},
	{ADECW, yincl, Pe, [13]uint8{0x48, 0xff, 01}},
	{ADIVB, ydivb, Pb, [13]uint8{0xf6, 06}},
	{ADIVL, ydivl, Px, [13]uint8{0xf7, 06}},
	{ADIVW, ydivl, Pe, [13]uint8{0xf7, 06}},
	{AENTER, nil, 0, [13]uint8{}}, /* botch */
	{AGLOBL, nil, 0, [13]uint8{}},
	{AGOK, nil, 0, [13]uint8{}},
	{AHISTORY, nil, 0, [13]uint8{}},
	{AHLT, ynone, Px, [13]uint8{0xf4}},
	{AIDIVB, ydivb, Pb, [13]uint8{0xf6, 07}},
	{AIDIVL, ydivl, Px, [13]uint8{0xf7, 07}},
	{AIDIVW, ydivl, Pe, [13]uint8{0xf7, 07}},
	{AIMULB, ydivb, Pb, [13]uint8{0xf6, 05}},
	{AIMULL, yimul, Px, [13]uint8{0xf7, 05, 0x6b, 0x69}},
	{AIMULW, yimul, Pe, [13]uint8{0xf7, 05, 0x6b, 0x69}},
	{AINB, yin, Pb, [13]uint8{0xe4, 0xec}},
	{AINL, yin, Px, [13]uint8{0xe5, 0xed}},
	{AINW, yin, Pe, [13]uint8{0xe5, 0xed}},
	{AINCB, yincb, Pb, [13]uint8{0xfe, 00}},
	{AINCL, yincl, Px, [13]uint8{0x40, 0xff, 00}},
	{AINCW, yincl, Pe, [13]uint8{0x40, 0xff, 00}},
	{AINSB, ynone, Pb, [13]uint8{0x6c}},
	{AINSL, ynone, Px, [13]uint8{0x6d}},
	{AINSW, ynone, Pe, [13]uint8{0x6d}},
	{AINT, yint, Px, [13]uint8{0xcd}},
	{AINTO, ynone, Px, [13]uint8{0xce}},
	{AIRETL, ynone, Px, [13]uint8{0xcf}},
	{AIRETW, ynone, Pe, [13]uint8{0xcf}},
	{AJCC, yjcond, Px, [13]uint8{0x73, 0x83, 00}},
	{AJCS, yjcond, Px, [13]uint8{0x72, 0x82}},
	{AJCXZL, yloop, Px, [13]uint8{0xe3}},
	{AJCXZW, yloop, Px, [13]uint8{0xe3}},
	{AJEQ, yjcond, Px, [13]uint8{0x74, 0x84}},
	{AJGE, yjcond, Px, [13]uint8{0x7d, 0x8d}},
	{AJGT, yjcond, Px, [13]uint8{0x7f, 0x8f}},
	{AJHI, yjcond, Px, [13]uint8{0x77, 0x87}},
	{AJLE, yjcond, Px, [13]uint8{0x7e, 0x8e}},
	{AJLS, yjcond, Px, [13]uint8{0x76, 0x86}},
	{AJLT, yjcond, Px, [13]uint8{0x7c, 0x8c}},
	{AJMI, yjcond, Px, [13]uint8{0x78, 0x88}},
	{AJMP, yjmp, Px, [13]uint8{0xff, 04, 0xeb, 0xe9}},
	{AJNE, yjcond, Px, [13]uint8{0x75, 0x85}},
	{AJOC, yjcond, Px, [13]uint8{0x71, 0x81, 00}},
	{AJOS, yjcond, Px, [13]uint8{0x70, 0x80, 00}},
	{AJPC, yjcond, Px, [13]uint8{0x7b, 0x8b}},
	{AJPL, yjcond, Px, [13]uint8{0x79, 0x89}},
	{AJPS, yjcond, Px, [13]uint8{0x7a, 0x8a}},
	{ALAHF, ynone, Px, [13]uint8{0x9f}},
	{ALARL, yml_rl, Pm, [13]uint8{0x02}},
	{ALARW, yml_rl, Pq, [13]uint8{0x02}},
	{ALEAL, ym_rl, Px, [13]uint8{0x8d}},
	{ALEAW, ym_rl, Pe, [13]uint8{0x8d}},
	{ALEAVEL, ynone, Px, [13]uint8{0xc9}},
	{ALEAVEW, ynone, Pe, [13]uint8{0xc9}},
	{ALOCK, ynone, Px, [13]uint8{0xf0}},
	{ALODSB, ynone, Pb, [13]uint8{0xac}},
	{ALODSL, ynone, Px, [13]uint8{0xad}},
	{ALODSW, ynone, Pe, [13]uint8{0xad}},
	{ALONG, ybyte, Px, [13]uint8{4}},
	{ALOOP, yloop, Px, [13]uint8{0xe2}},
	{ALOOPEQ, yloop, Px, [13]uint8{0xe1}},
	{ALOOPNE, yloop, Px, [13]uint8{0xe0}},
	{ALSLL, yml_rl, Pm, [13]uint8{0x03}},
	{ALSLW, yml_rl, Pq, [13]uint8{0x03}},
	{AMOVB, ymovb, Pb, [13]uint8{0x88, 0x8a, 0xb0, 0xc6, 00}},
	{AMOVL, ymovl, Px, [13]uint8{0x89, 0x8b, 0x31, 0x83, 04, 0xb8, 0xc7, 00, Pe, 0x6e, Pe, 0x7e, 0}},
	{AMOVW, ymovw, Pe, [13]uint8{0x89, 0x8b, 0x31, 0x83, 04, 0xb8, 0xc7, 00, 0}},
	{AMOVQ, ymovq, Pf3, [13]uint8{0x7e}},
	{AMOVBLSX, ymb_rl, Pm, [13]uint8{0xbe}},
	{AMOVBLZX, ymb_rl, Pm, [13]uint8{0xb6}},
	{AMOVBWSX, ymb_rl, Pq, [13]uint8{0xbe}},
	{AMOVBWZX, ymb_rl, Pq, [13]uint8{0xb6}},
	{AMOVWLSX, yml_rl, Pm, [13]uint8{0xbf}},
	{AMOVWLZX, yml_rl, Pm, [13]uint8{0xb7}},
	{AMOVSB, ynone, Pb, [13]uint8{0xa4}},
	{AMOVSL, ynone, Px, [13]uint8{0xa5}},
	{AMOVSW, ynone, Pe, [13]uint8{0xa5}},
	{AMULB, ydivb, Pb, [13]uint8{0xf6, 04}},
	{AMULL, ydivl, Px, [13]uint8{0xf7, 04}},
	{AMULW, ydivl, Pe, [13]uint8{0xf7, 04}},
	{ANAME, nil, 0, [13]uint8{}},
	{ANEGB, yscond, Px, [13]uint8{0xf6, 03}},
	{ANEGL, yscond, Px, [13]uint8{0xf7, 03}},
	{ANEGW, yscond, Pe, [13]uint8{0xf7, 03}},
	{ANOP, ynop, Px, [13]uint8{0, 0}},
	{ANOTB, yscond, Px, [13]uint8{0xf6, 02}},
	{ANOTL, yscond, Px, [13]uint8{0xf7, 02}},
	{ANOTW, yscond, Pe, [13]uint8{0xf7, 02}},
	{AORB, yxorb, Pb, [13]uint8{0x0c, 0x80, 01, 0x08, 0x0a}},
	{AORL, yxorl, Px, [13]uint8{0x83, 01, 0x0d, 0x81, 01, 0x09, 0x0b}},
	{AORW, yxorl, Pe, [13]uint8{0x83, 01, 0x0d, 0x81, 01, 0x09, 0x0b}},
	{AOUTB, yin, Pb, [13]uint8{0xe6, 0xee}},
	{AOUTL, yin, Px, [13]uint8{0xe7, 0xef}},
	{AOUTW, yin, Pe, [13]uint8{0xe7, 0xef}},
	{AOUTSB, ynone, Pb, [13]uint8{0x6e}},
	{AOUTSL, ynone, Px, [13]uint8{0x6f}},
	{AOUTSW, ynone, Pe, [13]uint8{0x6f}},
	{APAUSE, ynone, Px, [13]uint8{0xf3, 0x90}},
	{APOPAL, ynone, Px, [13]uint8{0x61}},
	{APOPAW, ynone, Pe, [13]uint8{0x61}},
	{APOPFL, ynone, Px, [13]uint8{0x9d}},
	{APOPFW, ynone, Pe, [13]uint8{0x9d}},
	{APOPL, ypopl, Px, [13]uint8{0x58, 0x8f, 00}},
	{APOPW, ypopl, Pe, [13]uint8{0x58, 0x8f, 00}},
	{APUSHAL, ynone, Px, [13]uint8{0x60}},
	{APUSHAW, ynone, Pe, [13]uint8{0x60}},
	{APUSHFL, ynone, Px, [13]uint8{0x9c}},
	{APUSHFW, ynone, Pe, [13]uint8{0x9c}},
	{APUSHL, ypushl, Px, [13]uint8{0x50, 0xff, 06, 0x6a, 0x68}},
	{APUSHW, ypushl, Pe, [13]uint8{0x50, 0xff, 06, 0x6a, 0x68}},
	{ARCLB, yshb, Pb, [13]uint8{0xd0, 02, 0xc0, 02, 0xd2, 02}},
	{ARCLL, yshl, Px, [13]uint8{0xd1, 02, 0xc1, 02, 0xd3, 02, 0xd3, 02}},
	{ARCLW, yshl, Pe, [13]uint8{0xd1, 02, 0xc1, 02, 0xd3, 02, 0xd3, 02}},
	{ARCRB, yshb, Pb, [13]uint8{0xd0, 03, 0xc0, 03, 0xd2, 03}},
	{ARCRL, yshl, Px, [13]uint8{0xd1, 03, 0xc1, 03, 0xd3, 03, 0xd3, 03}},
	{ARCRW, yshl, Pe, [13]uint8{0xd1, 03, 0xc1, 03, 0xd3, 03, 0xd3, 03}},
	{AREP, ynone, Px, [13]uint8{0xf3}},
	{AREPN, ynone, Px, [13]uint8{0xf2}},
	{ARET, ynone, Px, [13]uint8{0xc3}},
	{AROLB, yshb, Pb, [13]uint8{0xd0, 00, 0xc0, 00, 0xd2, 00}},
	{AROLL, yshl, Px, [13]uint8{0xd1, 00, 0xc1, 00, 0xd3, 00, 0xd3, 00}},
	{AROLW, yshl, Pe, [13]uint8{0xd1, 00, 0xc1, 00, 0xd3, 00, 0xd3, 00}},
	{ARORB, yshb, Pb, [13]uint8{0xd0, 01, 0xc0, 01, 0xd2, 01}},
	{ARORL, yshl, Px, [13]uint8{0xd1, 01, 0xc1, 01, 0xd3, 01, 0xd3, 01}},
	{ARORW, yshl, Pe, [13]uint8{0xd1, 01, 0xc1, 01, 0xd3, 01, 0xd3, 01}},
	{ASAHF, ynone, Px, [13]uint8{0x9e}},
	{ASALB, yshb, Pb, [13]uint8{0xd0, 04, 0xc0, 04, 0xd2, 04}},
	{ASALL, yshl, Px, [13]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASALW, yshl, Pe, [13]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASARB, yshb, Pb, [13]uint8{0xd0, 07, 0xc0, 07, 0xd2, 07}},
	{ASARL, yshl, Px, [13]uint8{0xd1, 07, 0xc1, 07, 0xd3, 07, 0xd3, 07}},
	{ASARW, yshl, Pe, [13]uint8{0xd1, 07, 0xc1, 07, 0xd3, 07, 0xd3, 07}},
	{ASBBB, yxorb, Pb, [13]uint8{0x1c, 0x80, 03, 0x18, 0x1a}},
	{ASBBL, yxorl, Px, [13]uint8{0x83, 03, 0x1d, 0x81, 03, 0x19, 0x1b}},
	{ASBBW, yxorl, Pe, [13]uint8{0x83, 03, 0x1d, 0x81, 03, 0x19, 0x1b}},
	{ASCASB, ynone, Pb, [13]uint8{0xae}},
	{ASCASL, ynone, Px, [13]uint8{0xaf}},
	{ASCASW, ynone, Pe, [13]uint8{0xaf}},
	{ASETCC, yscond, Pm, [13]uint8{0x93, 00}},
	{ASETCS, yscond, Pm, [13]uint8{0x92, 00}},
	{ASETEQ, yscond, Pm, [13]uint8{0x94, 00}},
	{ASETGE, yscond, Pm, [13]uint8{0x9d, 00}},
	{ASETGT, yscond, Pm, [13]uint8{0x9f, 00}},
	{ASETHI, yscond, Pm, [13]uint8{0x97, 00}},
	{ASETLE, yscond, Pm, [13]uint8{0x9e, 00}},
	{ASETLS, yscond, Pm, [13]uint8{0x96, 00}},
	{ASETLT, yscond, Pm, [13]uint8{0x9c, 00}},
	{ASETMI, yscond, Pm, [13]uint8{0x98, 00}},
	{ASETNE, yscond, Pm, [13]uint8{0x95, 00}},
	{ASETOC, yscond, Pm, [13]uint8{0x91, 00}},
	{ASETOS, yscond, Pm, [13]uint8{0x90, 00}},
	{ASETPC, yscond, Pm, [13]uint8{0x96, 00}},
	{ASETPL, yscond, Pm, [13]uint8{0x99, 00}},
	{ASETPS, yscond, Pm, [13]uint8{0x9a, 00}},
	{ACDQ, ynone, Px, [13]uint8{0x99}},
	{ACWD, ynone, Pe, [13]uint8{0x99}},
	{ASHLB, yshb, Pb, [13]uint8{0xd0, 04, 0xc0, 04, 0xd2, 04}},
	{ASHLL, yshl, Px, [13]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASHLW, yshl, Pe, [13]uint8{0xd1, 04, 0xc1, 04, 0xd3, 04, 0xd3, 04}},
	{ASHRB, yshb, Pb, [13]uint8{0xd0, 05, 0xc0, 05, 0xd2, 05}},
	{ASHRL, yshl, Px, [13]uint8{0xd1, 05, 0xc1, 05, 0xd3, 05, 0xd3, 05}},
	{ASHRW, yshl, Pe, [13]uint8{0xd1, 05, 0xc1, 05, 0xd3, 05, 0xd3, 05}},
	{ASTC, ynone, Px, [13]uint8{0xf9}},
	{ASTD, ynone, Px, [13]uint8{0xfd}},
	{ASTI, ynone, Px, [13]uint8{0xfb}},
	{ASTOSB, ynone, Pb, [13]uint8{0xaa}},
	{ASTOSL, ynone, Px, [13]uint8{0xab}},
	{ASTOSW, ynone, Pe, [13]uint8{0xab}},
	{ASUBB, yxorb, Pb, [13]uint8{0x2c, 0x80, 05, 0x28, 0x2a}},
	{ASUBL, yaddl, Px, [13]uint8{0x83, 05, 0x2d, 0x81, 05, 0x29, 0x2b}},
	{ASUBW, yaddl, Pe, [13]uint8{0x83, 05, 0x2d, 0x81, 05, 0x29, 0x2b}},
	{ASYSCALL, ynone, Px, [13]uint8{0xcd, 100}},
	{ATESTB, ytestb, Pb, [13]uint8{0xa8, 0xf6, 00, 0x84, 0x84}},
	{ATESTL, ytestl, Px, [13]uint8{0xa9, 0xf7, 00, 0x85, 0x85}},
	{ATESTW, ytestl, Pe, [13]uint8{0xa9, 0xf7, 00, 0x85, 0x85}},
	{ATEXT, ytext, Px, [13]uint8{}},
	{AVERR, ydivl, Pm, [13]uint8{0x00, 04}},
	{AVERW, ydivl, Pm, [13]uint8{0x00, 05}},
	{AWAIT, ynone, Px, [13]uint8{0x9b}},
	{AWORD, ybyte, Px, [13]uint8{2}},
	{AXCHGB, yml_mb, Pb, [13]uint8{0x86, 0x86}},
	{AXCHGL, yxchg, Px, [13]uint8{0x90, 0x90, 0x87, 0x87}},
	{AXCHGW, yxchg, Pe, [13]uint8{0x90, 0x90, 0x87, 0x87}},
	{AXLAT, ynone, Px, [13]uint8{0xd7}},
	{AXORB, yxorb, Pb, [13]uint8{0x34, 0x80, 06, 0x30, 0x32}},
	{AXORL, yxorl, Px, [13]uint8{0x83, 06, 0x35, 0x81, 06, 0x31, 0x33}},
	{AXORW, yxorl, Pe, [13]uint8{0x83, 06, 0x35, 0x81, 06, 0x31, 0x33}},
	{AFMOVB, yfmvx, Px, [13]uint8{0xdf, 04}},
	{AFMOVBP, yfmvp, Px, [13]uint8{0xdf, 06}},
	{AFMOVD, yfmvd, Px, [13]uint8{0xdd, 00, 0xdd, 02, 0xd9, 00, 0xdd, 02}},
	{AFMOVDP, yfmvdp, Px, [13]uint8{0xdd, 03, 0xdd, 03}},
	{AFMOVF, yfmvf, Px, [13]uint8{0xd9, 00, 0xd9, 02}},
	{AFMOVFP, yfmvp, Px, [13]uint8{0xd9, 03}},
	{AFMOVL, yfmvf, Px, [13]uint8{0xdb, 00, 0xdb, 02}},
	{AFMOVLP, yfmvp, Px, [13]uint8{0xdb, 03}},
	{AFMOVV, yfmvx, Px, [13]uint8{0xdf, 05}},
	{AFMOVVP, yfmvp, Px, [13]uint8{0xdf, 07}},
	{AFMOVW, yfmvf, Px, [13]uint8{0xdf, 00, 0xdf, 02}},
	{AFMOVWP, yfmvp, Px, [13]uint8{0xdf, 03}},
	{AFMOVX, yfmvx, Px, [13]uint8{0xdb, 05}},
	{AFMOVXP, yfmvp, Px, [13]uint8{0xdb, 07}},
	{AFCOMB, nil, 0, [13]uint8{}},
	{AFCOMBP, nil, 0, [13]uint8{}},
	{AFCOMD, yfadd, Px, [13]uint8{0xdc, 02, 0xd8, 02, 0xdc, 02}},  /* botch */
	{AFCOMDP, yfadd, Px, [13]uint8{0xdc, 03, 0xd8, 03, 0xdc, 03}}, /* botch */
	{AFCOMDPP, ycompp, Px, [13]uint8{0xde, 03}},
	{AFCOMF, yfmvx, Px, [13]uint8{0xd8, 02}},
	{AFCOMFP, yfmvx, Px, [13]uint8{0xd8, 03}},
	{AFCOMI, yfmvx, Px, [13]uint8{0xdb, 06}},
	{AFCOMIP, yfmvx, Px, [13]uint8{0xdf, 06}},
	{AFCOML, yfmvx, Px, [13]uint8{0xda, 02}},
	{AFCOMLP, yfmvx, Px, [13]uint8{0xda, 03}},
	{AFCOMW, yfmvx, Px, [13]uint8{0xde, 02}},
	{AFCOMWP, yfmvx, Px, [13]uint8{0xde, 03}},
	{AFUCOM, ycompp, Px, [13]uint8{0xdd, 04}},
	{AFUCOMI, ycompp, Px, [13]uint8{0xdb, 05}},
	{AFUCOMIP, ycompp, Px, [13]uint8{0xdf, 05}},
	{AFUCOMP, ycompp, Px, [13]uint8{0xdd, 05}},
	{AFUCOMPP, ycompp, Px, [13]uint8{0xda, 13}},
	{AFADDDP, yfaddp, Px, [13]uint8{0xde, 00}},
	{AFADDW, yfmvx, Px, [13]uint8{0xde, 00}},
	{AFADDL, yfmvx, Px, [13]uint8{0xda, 00}},
	{AFADDF, yfmvx, Px, [13]uint8{0xd8, 00}},
	{AFADDD, yfadd, Px, [13]uint8{0xdc, 00, 0xd8, 00, 0xdc, 00}},
	{AFMULDP, yfaddp, Px, [13]uint8{0xde, 01}},
	{AFMULW, yfmvx, Px, [13]uint8{0xde, 01}},
	{AFMULL, yfmvx, Px, [13]uint8{0xda, 01}},
	{AFMULF, yfmvx, Px, [13]uint8{0xd8, 01}},
	{AFMULD, yfadd, Px, [13]uint8{0xdc, 01, 0xd8, 01, 0xdc, 01}},
	{AFSUBDP, yfaddp, Px, [13]uint8{0xde, 05}},
	{AFSUBW, yfmvx, Px, [13]uint8{0xde, 04}},
	{AFSUBL, yfmvx, Px, [13]uint8{0xda, 04}},
	{AFSUBF, yfmvx, Px, [13]uint8{0xd8, 04}},
	{AFSUBD, yfadd, Px, [13]uint8{0xdc, 04, 0xd8, 04, 0xdc, 05}},
	{AFSUBRDP, yfaddp, Px, [13]uint8{0xde, 04}},
	{AFSUBRW, yfmvx, Px, [13]uint8{0xde, 05}},
	{AFSUBRL, yfmvx, Px, [13]uint8{0xda, 05}},
	{AFSUBRF, yfmvx, Px, [13]uint8{0xd8, 05}},
	{AFSUBRD, yfadd, Px, [13]uint8{0xdc, 05, 0xd8, 05, 0xdc, 04}},
	{AFDIVDP, yfaddp, Px, [13]uint8{0xde, 07}},
	{AFDIVW, yfmvx, Px, [13]uint8{0xde, 06}},
	{AFDIVL, yfmvx, Px, [13]uint8{0xda, 06}},
	{AFDIVF, yfmvx, Px, [13]uint8{0xd8, 06}},
	{AFDIVD, yfadd, Px, [13]uint8{0xdc, 06, 0xd8, 06, 0xdc, 07}},
	{AFDIVRDP, yfaddp, Px, [13]uint8{0xde, 06}},
	{AFDIVRW, yfmvx, Px, [13]uint8{0xde, 07}},
	{AFDIVRL, yfmvx, Px, [13]uint8{0xda, 07}},
	{AFDIVRF, yfmvx, Px, [13]uint8{0xd8, 07}},
	{AFDIVRD, yfadd, Px, [13]uint8{0xdc, 07, 0xd8, 07, 0xdc, 06}},
	{AFXCHD, yfxch, Px, [13]uint8{0xd9, 01, 0xd9, 01}},
	{AFFREE, nil, 0, [13]uint8{}},
	{AFLDCW, ystcw, Px, [13]uint8{0xd9, 05, 0xd9, 05}},
	{AFLDENV, ystcw, Px, [13]uint8{0xd9, 04, 0xd9, 04}},
	{AFRSTOR, ysvrs, Px, [13]uint8{0xdd, 04, 0xdd, 04}},
	{AFSAVE, ysvrs, Px, [13]uint8{0xdd, 06, 0xdd, 06}},
	{AFSTCW, ystcw, Px, [13]uint8{0xd9, 07, 0xd9, 07}},
	{AFSTENV, ystcw, Px, [13]uint8{0xd9, 06, 0xd9, 06}},
	{AFSTSW, ystsw, Px, [13]uint8{0xdd, 07, 0xdf, 0xe0}},
	{AF2XM1, ynone, Px, [13]uint8{0xd9, 0xf0}},
	{AFABS, ynone, Px, [13]uint8{0xd9, 0xe1}},
	{AFCHS, ynone, Px, [13]uint8{0xd9, 0xe0}},
	{AFCLEX, ynone, Px, [13]uint8{0xdb, 0xe2}},
	{AFCOS, ynone, Px, [13]uint8{0xd9, 0xff}},
	{AFDECSTP, ynone, Px, [13]uint8{0xd9, 0xf6}},
	{AFINCSTP, ynone, Px, [13]uint8{0xd9, 0xf7}},
	{AFINIT, ynone, Px, [13]uint8{0xdb, 0xe3}},
	{AFLD1, ynone, Px, [13]uint8{0xd9, 0xe8}},
	{AFLDL2E, ynone, Px, [13]uint8{0xd9, 0xea}},
	{AFLDL2T, ynone, Px, [13]uint8{0xd9, 0xe9}},
	{AFLDLG2, ynone, Px, [13]uint8{0xd9, 0xec}},
	{AFLDLN2, ynone, Px, [13]uint8{0xd9, 0xed}},
	{AFLDPI, ynone, Px, [13]uint8{0xd9, 0xeb}},
	{AFLDZ, ynone, Px, [13]uint8{0xd9, 0xee}},
	{AFNOP, ynone, Px, [13]uint8{0xd9, 0xd0}},
	{AFPATAN, ynone, Px, [13]uint8{0xd9, 0xf3}},
	{AFPREM, ynone, Px, [13]uint8{0xd9, 0xf8}},
	{AFPREM1, ynone, Px, [13]uint8{0xd9, 0xf5}},
	{AFPTAN, ynone, Px, [13]uint8{0xd9, 0xf2}},
	{AFRNDINT, ynone, Px, [13]uint8{0xd9, 0xfc}},
	{AFSCALE, ynone, Px, [13]uint8{0xd9, 0xfd}},
	{AFSIN, ynone, Px, [13]uint8{0xd9, 0xfe}},
	{AFSINCOS, ynone, Px, [13]uint8{0xd9, 0xfb}},
	{AFSQRT, ynone, Px, [13]uint8{0xd9, 0xfa}},
	{AFTST, ynone, Px, [13]uint8{0xd9, 0xe4}},
	{AFXAM, ynone, Px, [13]uint8{0xd9, 0xe5}},
	{AFXTRACT, ynone, Px, [13]uint8{0xd9, 0xf4}},
	{AFYL2X, ynone, Px, [13]uint8{0xd9, 0xf1}},
	{AFYL2XP1, ynone, Px, [13]uint8{0xd9, 0xf9}},
	{AEND, nil, 0, [13]uint8{}},
	{ADYNT_, nil, 0, [13]uint8{}},
	{AINIT_, nil, 0, [13]uint8{}},
	{ASIGNAME, nil, 0, [13]uint8{}},
	{ACMPXCHGB, yrb_mb, Pm, [13]uint8{0xb0}},
	{ACMPXCHGL, yrl_ml, Pm, [13]uint8{0xb1}},
	{ACMPXCHGW, yrl_ml, Pm, [13]uint8{0xb1}},
	{ACMPXCHG8B, yscond, Pm, [13]uint8{0xc7, 01}},
	{ACPUID, ynone, Pm, [13]uint8{0xa2}},
	{ARDTSC, ynone, Pm, [13]uint8{0x31}},
	{AXADDB, yrb_mb, Pb, [13]uint8{0x0f, 0xc0}},
	{AXADDL, yrl_ml, Pm, [13]uint8{0xc1}},
	{AXADDW, yrl_ml, Pe, [13]uint8{0x0f, 0xc1}},
	{ACMOVLCC, yml_rl, Pm, [13]uint8{0x43}},
	{ACMOVLCS, yml_rl, Pm, [13]uint8{0x42}},
	{ACMOVLEQ, yml_rl, Pm, [13]uint8{0x44}},
	{ACMOVLGE, yml_rl, Pm, [13]uint8{0x4d}},
	{ACMOVLGT, yml_rl, Pm, [13]uint8{0x4f}},
	{ACMOVLHI, yml_rl, Pm, [13]uint8{0x47}},
	{ACMOVLLE, yml_rl, Pm, [13]uint8{0x4e}},
	{ACMOVLLS, yml_rl, Pm, [13]uint8{0x46}},
	{ACMOVLLT, yml_rl, Pm, [13]uint8{0x4c}},
	{ACMOVLMI, yml_rl, Pm, [13]uint8{0x48}},
	{ACMOVLNE, yml_rl, Pm, [13]uint8{0x45}},
	{ACMOVLOC, yml_rl, Pm, [13]uint8{0x41}},
	{ACMOVLOS, yml_rl, Pm, [13]uint8{0x40}},
	{ACMOVLPC, yml_rl, Pm, [13]uint8{0x4b}},
	{ACMOVLPL, yml_rl, Pm, [13]uint8{0x49}},
	{ACMOVLPS, yml_rl, Pm, [13]uint8{0x4a}},
	{ACMOVWCC, yml_rl, Pq, [13]uint8{0x43}},
	{ACMOVWCS, yml_rl, Pq, [13]uint8{0x42}},
	{ACMOVWEQ, yml_rl, Pq, [13]uint8{0x44}},
	{ACMOVWGE, yml_rl, Pq, [13]uint8{0x4d}},
	{ACMOVWGT, yml_rl, Pq, [13]uint8{0x4f}},
	{ACMOVWHI, yml_rl, Pq, [13]uint8{0x47}},
	{ACMOVWLE, yml_rl, Pq, [13]uint8{0x4e}},
	{ACMOVWLS, yml_rl, Pq, [13]uint8{0x46}},
	{ACMOVWLT, yml_rl, Pq, [13]uint8{0x4c}},
	{ACMOVWMI, yml_rl, Pq, [13]uint8{0x48}},
	{ACMOVWNE, yml_rl, Pq, [13]uint8{0x45}},
	{ACMOVWOC, yml_rl, Pq, [13]uint8{0x41}},
	{ACMOVWOS, yml_rl, Pq, [13]uint8{0x40}},
	{ACMOVWPC, yml_rl, Pq, [13]uint8{0x4b}},
	{ACMOVWPL, yml_rl, Pq, [13]uint8{0x49}},
	{ACMOVWPS, yml_rl, Pq, [13]uint8{0x4a}},
	{AFCMOVCC, yfcmv, Px, [13]uint8{0xdb, 00}},
	{AFCMOVCS, yfcmv, Px, [13]uint8{0xda, 00}},
	{AFCMOVEQ, yfcmv, Px, [13]uint8{0xda, 01}},
	{AFCMOVHI, yfcmv, Px, [13]uint8{0xdb, 02}},
	{AFCMOVLS, yfcmv, Px, [13]uint8{0xda, 02}},
	{AFCMOVNE, yfcmv, Px, [13]uint8{0xdb, 01}},
	{AFCMOVNU, yfcmv, Px, [13]uint8{0xdb, 03}},
	{AFCMOVUN, yfcmv, Px, [13]uint8{0xda, 03}},
	{ALFENCE, ynone, Pm, [13]uint8{0xae, 0xe8}},
	{AMFENCE, ynone, Pm, [13]uint8{0xae, 0xf0}},
	{ASFENCE, ynone, Pm, [13]uint8{0xae, 0xf8}},
	{AEMMS, ynone, Pm, [13]uint8{0x77}},
	{APREFETCHT0, yprefetch, Pm, [13]uint8{0x18, 01}},
	{APREFETCHT1, yprefetch, Pm, [13]uint8{0x18, 02}},
	{APREFETCHT2, yprefetch, Pm, [13]uint8{0x18, 03}},
	{APREFETCHNTA, yprefetch, Pm, [13]uint8{0x18, 00}},
	{ABSWAPL, ybswap, Pm, [13]uint8{0xc8}},
	{AUNDEF, ynone, Px, [13]uint8{0x0f, 0x0b}},
	{AADDPD, yxm, Pq, [13]uint8{0x58}},
	{AADDPS, yxm, Pm, [13]uint8{0x58}},
	{AADDSD, yxm, Pf2, [13]uint8{0x58}},
	{AADDSS, yxm, Pf3, [13]uint8{0x58}},
	{AANDNPD, yxm, Pq, [13]uint8{0x55}},
	{AANDNPS, yxm, Pm, [13]uint8{0x55}},
	{AANDPD, yxm, Pq, [13]uint8{0x54}},
	{AANDPS, yxm, Pq, [13]uint8{0x54}},
	{ACMPPD, yxcmpi, Px, [13]uint8{Pe, 0xc2}},
	{ACMPPS, yxcmpi, Pm, [13]uint8{0xc2, 0}},
	{ACMPSD, yxcmpi, Px, [13]uint8{Pf2, 0xc2}},
	{ACMPSS, yxcmpi, Px, [13]uint8{Pf3, 0xc2}},
	{ACOMISD, yxcmp, Pe, [13]uint8{0x2f}},
	{ACOMISS, yxcmp, Pm, [13]uint8{0x2f}},
	{ACVTPL2PD, yxcvm2, Px, [13]uint8{Pf3, 0xe6, Pe, 0x2a}},
	{ACVTPL2PS, yxcvm2, Pm, [13]uint8{0x5b, 0, 0x2a, 0}},
	{ACVTPD2PL, yxcvm1, Px, [13]uint8{Pf2, 0xe6, Pe, 0x2d}},
	{ACVTPD2PS, yxm, Pe, [13]uint8{0x5a}},
	{ACVTPS2PL, yxcvm1, Px, [13]uint8{Pe, 0x5b, Pm, 0x2d}},
	{ACVTPS2PD, yxm, Pm, [13]uint8{0x5a}},
	{ACVTSD2SL, yxcvfl, Pf2, [13]uint8{0x2d}},
	{ACVTSD2SS, yxm, Pf2, [13]uint8{0x5a}},
	{ACVTSL2SD, yxcvlf, Pf2, [13]uint8{0x2a}},
	{ACVTSL2SS, yxcvlf, Pf3, [13]uint8{0x2a}},
	{ACVTSS2SD, yxm, Pf3, [13]uint8{0x5a}},
	{ACVTSS2SL, yxcvfl, Pf3, [13]uint8{0x2d}},
	{ACVTTPD2PL, yxcvm1, Px, [13]uint8{Pe, 0xe6, Pe, 0x2c}},
	{ACVTTPS2PL, yxcvm1, Px, [13]uint8{Pf3, 0x5b, Pm, 0x2c}},
	{ACVTTSD2SL, yxcvfl, Pf2, [13]uint8{0x2c}},
	{ACVTTSS2SL, yxcvfl, Pf3, [13]uint8{0x2c}},
	{ADIVPD, yxm, Pe, [13]uint8{0x5e}},
	{ADIVPS, yxm, Pm, [13]uint8{0x5e}},
	{ADIVSD, yxm, Pf2, [13]uint8{0x5e}},
	{ADIVSS, yxm, Pf3, [13]uint8{0x5e}},
	{AMASKMOVOU, yxr, Pe, [13]uint8{0xf7}},
	{AMAXPD, yxm, Pe, [13]uint8{0x5f}},
	{AMAXPS, yxm, Pm, [13]uint8{0x5f}},
	{AMAXSD, yxm, Pf2, [13]uint8{0x5f}},
	{AMAXSS, yxm, Pf3, [13]uint8{0x5f}},
	{AMINPD, yxm, Pe, [13]uint8{0x5d}},
	{AMINPS, yxm, Pm, [13]uint8{0x5d}},
	{AMINSD, yxm, Pf2, [13]uint8{0x5d}},
	{AMINSS, yxm, Pf3, [13]uint8{0x5d}},
	{AMOVAPD, yxmov, Pe, [13]uint8{0x28, 0x29}},
	{AMOVAPS, yxmov, Pm, [13]uint8{0x28, 0x29}},
	{AMOVO, yxmov, Pe, [13]uint8{0x6f, 0x7f}},
	{AMOVOU, yxmov, Pf3, [13]uint8{0x6f, 0x7f}},
	{AMOVHLPS, yxr, Pm, [13]uint8{0x12}},
	{AMOVHPD, yxmov, Pe, [13]uint8{0x16, 0x17}},
	{AMOVHPS, yxmov, Pm, [13]uint8{0x16, 0x17}},
	{AMOVLHPS, yxr, Pm, [13]uint8{0x16}},
	{AMOVLPD, yxmov, Pe, [13]uint8{0x12, 0x13}},
	{AMOVLPS, yxmov, Pm, [13]uint8{0x12, 0x13}},
	{AMOVMSKPD, yxrrl, Pq, [13]uint8{0x50}},
	{AMOVMSKPS, yxrrl, Pm, [13]uint8{0x50}},
	{AMOVNTO, yxr_ml, Pe, [13]uint8{0xe7}},
	{AMOVNTPD, yxr_ml, Pe, [13]uint8{0x2b}},
	{AMOVNTPS, yxr_ml, Pm, [13]uint8{0x2b}},
	{AMOVSD, yxmov, Pf2, [13]uint8{0x10, 0x11}},
	{AMOVSS, yxmov, Pf3, [13]uint8{0x10, 0x11}},
	{AMOVUPD, yxmov, Pe, [13]uint8{0x10, 0x11}},
	{AMOVUPS, yxmov, Pm, [13]uint8{0x10, 0x11}},
	{AMULPD, yxm, Pe, [13]uint8{0x59}},
	{AMULPS, yxm, Ym, [13]uint8{0x59}},
	{AMULSD, yxm, Pf2, [13]uint8{0x59}},
	{AMULSS, yxm, Pf3, [13]uint8{0x59}},
	{AORPD, yxm, Pq, [13]uint8{0x56}},
	{AORPS, yxm, Pm, [13]uint8{0x56}},
	{APADDQ, yxm, Pe, [13]uint8{0xd4}},
	{APAND, yxm, Pe, [13]uint8{0xdb}},
	{APCMPEQB, yxmq, Pe, [13]uint8{0x74}},
	{APMAXSW, yxm, Pe, [13]uint8{0xee}},
	{APMAXUB, yxm, Pe, [13]uint8{0xde}},
	{APMINSW, yxm, Pe, [13]uint8{0xea}},
	{APMINUB, yxm, Pe, [13]uint8{0xda}},
	{APMOVMSKB, ymskb, Px, [13]uint8{Pe, 0xd7, 0xd7}},
	{APSADBW, yxm, Pq, [13]uint8{0xf6}},
	{APSUBB, yxm, Pe, [13]uint8{0xf8}},
	{APSUBL, yxm, Pe, [13]uint8{0xfa}},
	{APSUBQ, yxm, Pe, [13]uint8{0xfb}},
	{APSUBSB, yxm, Pe, [13]uint8{0xe8}},
	{APSUBSW, yxm, Pe, [13]uint8{0xe9}},
	{APSUBUSB, yxm, Pe, [13]uint8{0xd8}},
	{APSUBUSW, yxm, Pe, [13]uint8{0xd9}},
	{APSUBW, yxm, Pe, [13]uint8{0xf9}},
	{APUNPCKHQDQ, yxm, Pe, [13]uint8{0x6d}},
	{APUNPCKLQDQ, yxm, Pe, [13]uint8{0x6c}},
	{APXOR, yxm, Pe, [13]uint8{0xef}},
	{ARCPPS, yxm, Pm, [13]uint8{0x53}},
	{ARCPSS, yxm, Pf3, [13]uint8{0x53}},
	{ARSQRTPS, yxm, Pm, [13]uint8{0x52}},
	{ARSQRTSS, yxm, Pf3, [13]uint8{0x52}},
	{ASQRTPD, yxm, Pe, [13]uint8{0x51}},
	{ASQRTPS, yxm, Pm, [13]uint8{0x51}},
	{ASQRTSD, yxm, Pf2, [13]uint8{0x51}},
	{ASQRTSS, yxm, Pf3, [13]uint8{0x51}},
	{ASUBPD, yxm, Pe, [13]uint8{0x5c}},
	{ASUBPS, yxm, Pm, [13]uint8{0x5c}},
	{ASUBSD, yxm, Pf2, [13]uint8{0x5c}},
	{ASUBSS, yxm, Pf3, [13]uint8{0x5c}},
	{AUCOMISD, yxcmp, Pe, [13]uint8{0x2e}},
	{AUCOMISS, yxcmp, Pm, [13]uint8{0x2e}},
	{AUNPCKHPD, yxm, Pe, [13]uint8{0x15}},
	{AUNPCKHPS, yxm, Pm, [13]uint8{0x15}},
	{AUNPCKLPD, yxm, Pe, [13]uint8{0x14}},
	{AUNPCKLPS, yxm, Pm, [13]uint8{0x14}},
	{AXORPD, yxm, Pe, [13]uint8{0x57}},
	{AXORPS, yxm, Pm, [13]uint8{0x57}},
	{AAESENC, yaes, Pq, [13]uint8{0x38, 0xdc, 0}},
	{APINSRD, yinsrd, Pq, [13]uint8{0x3a, 0x22, 00}},
	{APSHUFB, ymshufb, Pq, [13]uint8{0x38, 0x00}},
	{AUSEFIELD, ynop, Px, [13]uint8{0, 0}},
	{ATYPE, nil, 0, [13]uint8{}},
	{AFUNCDATA, yfuncdata, Px, [13]uint8{0, 0}},
	{APCDATA, ypcdata, Px, [13]uint8{0, 0}},
	{ACHECKNIL, nil, 0, [13]uint8{}},
	{AVARDEF, nil, 0, [13]uint8{}},
	{AVARKILL, nil, 0, [13]uint8{}},
	{ADUFFCOPY, yduff, Px, [13]uint8{0xe8}},
	{ADUFFZERO, yduff, Px, [13]uint8{0xe8}},
	{0, nil, 0, [13]uint8{}},
}

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

func span8(ctxt *liblink.Link, s *liblink.LSym) {
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
	if s.Text == nil || s.Text.Link == nil {
		return
	}
	if ycover[0] == 0 {
		instinit()
	}
	for p = s.Text; p != nil; p = p.Link {
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
			p.As = AADDL
			if v < 0 {
				p.As = ASUBL
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
		}
		if p.As == AADJSP {
			p.To.Typ = D_SP
			v = int(-p.From.Offset)
			p.From.Offset = int64(v)
			p.As = AADDL
			if v < 0 {
				p.As = ASUBL
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
				if p.As == AREP && c>>5 != (c+3-1)>>5 {
					c = naclpad(ctxt, s, c, int(-c&31))
				}
				// same for LOCK.
				// various instructions follow; the longest is 4 bytes.
				// give ourselves 8 bytes so as to avoid surprises.
				if p.As == ALOCK && c>>5 != (c+8-1)>>5 {
					c = naclpad(ctxt, s, c, int(-c&31))
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
					if q.As == AJCXZW {
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
			log.Fatalf("bad code")
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
	var i int
	for i = 1; optab[i].as != 0; i++ {
		if i != optab[i].as {
			log.Fatalf("phase error in optab: at %v found %v", Aconv(i), Aconv(optab[i].as))
		}
	}
	for i = 0; i < Ymax; i++ {
		ycover[i*Ymax+i] = 1
	}
	ycover[Yi0*Ymax+Yi8] = 1
	ycover[Yi1*Ymax+Yi8] = 1
	ycover[Yi0*Ymax+Yi32] = 1
	ycover[Yi1*Ymax+Yi32] = 1
	ycover[Yi8*Ymax+Yi32] = 1
	ycover[Yal*Ymax+Yrb] = 1
	ycover[Ycl*Ymax+Yrb] = 1
	ycover[Yax*Ymax+Yrb] = 1
	ycover[Ycx*Ymax+Yrb] = 1
	ycover[Yrx*Ymax+Yrb] = 1
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
		if i >= D_AL && i <= D_BH {
			reg[i] = (i - D_AL) & 7
		}
		if i >= D_AX && i <= D_DI {
			reg[i] = (i - D_AX) & 7
		}
		if i >= D_F0 && i <= D_F0+7 {
			reg[i] = (i - D_F0) & 7
		}
		if i >= D_X0 && i <= D_X0+7 {
			reg[i] = (i - D_X0) & 7
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
		case liblink.Hdarwin,
			liblink.Hdragonfly,
			liblink.Hfreebsd,
			liblink.Hnetbsd,
			liblink.Hopenbsd:
			return 0x65 // GS
		}
	}
	return 0
}

func oclass(a *liblink.Addr) int {
	var v int
	if (a.Typ >= D_INDIR && a.Typ < 2*D_INDIR) || a.Index != D_NONE {
		if a.Index != D_NONE && a.Scale == 0 {
			if a.Typ == D_ADDR {
				switch a.Index {
				case D_EXTERN,
					D_STATIC:
					return Yi32
				case D_AUTO,
					D_PARAM:
					return Yiauto
				}
				return Yxxx
			}
			//if(a->type == D_INDIR+D_ADDR)
			//	print("*Ycol\n");
			return Ycol
		}
		return Ym
	}
	switch a.Typ {
	case D_AL:
		return Yal
	case D_AX:
		return Yax
	case D_CL,
		D_DL,
		D_BL,
		D_AH,
		D_CH,
		D_DH,
		D_BH:
		return Yrb
	case D_CX:
		return Ycx
	case D_DX,
		D_BX:
		return Yrx
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
	case D_X0 + 0,
		D_X0 + 1,
		D_X0 + 2,
		D_X0 + 3,
		D_X0 + 4,
		D_X0 + 5,
		D_X0 + 6,
		D_X0 + 7:
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
		D_CONST2,
		D_ADDR:
		if a.Sym == nil {
			v = int(int32(a.Offset))
			if v == 0 {
				return Yi0
			}
			if v == 1 {
				return Yi1
			}
			if v >= -128 && v <= 127 {
				return Yi8
			}
		}
		return Yi32
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
	ctxt.Diag("asmidx: bad address %d,%d,%d", scale, index, base)
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
		if s != nil {
			if r == nil {
				ctxt.Diag("need reloc for %D", a)
				log.Fatalf("bad code")
			}
			r.Typ = liblink.R_ADDR
			r.Siz = 4
			r.Off = -1
			r.Sym = s
			r.Add = v
			v = 0
		}
	case D_INDIR + D_TLS:
		if r == nil {
			ctxt.Diag("need reloc for %D", a)
			log.Fatalf("bad code")
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

func asmand(ctxt *liblink.Link, a *liblink.Addr, r int) {
	var v int64
	var t int
	var scale int
	var rel liblink.Reloc
	v = a.Offset
	t = a.Typ
	rel.Siz = 0
	if a.Index != D_NONE && a.Index != D_TLS {
		if t < D_INDIR || t >= 2*D_INDIR {
			switch t {
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
		} else {
			t -= D_INDIR
		}
		if t == D_NONE {
			ctxt.Andptr[0] = uint8(0<<6 | 4<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			asmidx(ctxt, int(a.Scale), a.Index, t)
			goto putrelv
		}
		if v == 0 && rel.Siz == 0 && t != D_BP {
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
	if t >= D_AL && t <= D_F7 || t >= D_X0 && t <= D_X7 {
		if v != 0 {
			goto bad
		}
		ctxt.Andptr[0] = uint8(3<<6 | reg[t]<<0 | r<<3)
		ctxt.Andptr = ctxt.Andptr[1:]
		return
	}
	scale = int(a.Scale)
	if t < D_INDIR || t >= 2*D_INDIR {
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
	if t == D_NONE || (D_CS <= t && t <= D_GS) || t == D_TLS {
		ctxt.Andptr[0] = uint8(0<<6 | 5<<0 | r<<3)
		ctxt.Andptr = ctxt.Andptr[1:]
		goto putrelv
	}
	if t == D_SP {
		if v == 0 && rel.Siz == 0 {
			ctxt.Andptr[0] = uint8(0<<6 | 4<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			asmidx(ctxt, scale, D_NONE, t)
			return
		}
		if v >= -128 && v < 128 && rel.Siz == 0 {
			ctxt.Andptr[0] = uint8(1<<6 | 4<<0 | r<<3)
			ctxt.Andptr = ctxt.Andptr[1:]
			asmidx(ctxt, scale, D_NONE, t)
			ctxt.Andptr[0] = uint8(v)
			ctxt.Andptr = ctxt.Andptr[1:]
			return
		}
		ctxt.Andptr[0] = uint8(2<<6 | 4<<0 | r<<3)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmidx(ctxt, scale, D_NONE, t)
		goto putrelv
	}
	if t >= D_AX && t <= D_DI {
		if a.Index == D_TLS {
			rel = liblink.Reloc{}
			rel.Typ = liblink.R_TLS_IE
			rel.Siz = 4
			rel.Sym = nil
			rel.Add = v
			v = 0
		}
		if v == 0 && rel.Siz == 0 && t != D_BP {
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

const (
	E = 0xff
)

var ymovtab = []uint8{
	/* push */
	APUSHL,
	Ycs,
	Ynone,
	0,
	0x0e,
	E,
	0,
	0,
	APUSHL,
	Yss,
	Ynone,
	0,
	0x16,
	E,
	0,
	0,
	APUSHL,
	Yds,
	Ynone,
	0,
	0x1e,
	E,
	0,
	0,
	APUSHL,
	Yes,
	Ynone,
	0,
	0x06,
	E,
	0,
	0,
	APUSHL,
	Yfs,
	Ynone,
	0,
	0x0f,
	0xa0,
	E,
	0,
	APUSHL,
	Ygs,
	Ynone,
	0,
	0x0f,
	0xa8,
	E,
	0,
	APUSHW,
	Ycs,
	Ynone,
	0,
	Pe,
	0x0e,
	E,
	0,
	APUSHW,
	Yss,
	Ynone,
	0,
	Pe,
	0x16,
	E,
	0,
	APUSHW,
	Yds,
	Ynone,
	0,
	Pe,
	0x1e,
	E,
	0,
	APUSHW,
	Yes,
	Ynone,
	0,
	Pe,
	0x06,
	E,
	0,
	APUSHW,
	Yfs,
	Ynone,
	0,
	Pe,
	0x0f,
	0xa0,
	E,
	APUSHW,
	Ygs,
	Ynone,
	0,
	Pe,
	0x0f,
	0xa8,
	E,
	/* pop */
	APOPL,
	Ynone,
	Yds,
	0,
	0x1f,
	E,
	0,
	0,
	APOPL,
	Ynone,
	Yes,
	0,
	0x07,
	E,
	0,
	0,
	APOPL,
	Ynone,
	Yss,
	0,
	0x17,
	E,
	0,
	0,
	APOPL,
	Ynone,
	Yfs,
	0,
	0x0f,
	0xa1,
	E,
	0,
	APOPL,
	Ynone,
	Ygs,
	0,
	0x0f,
	0xa9,
	E,
	0,
	APOPW,
	Ynone,
	Yds,
	0,
	Pe,
	0x1f,
	E,
	0,
	APOPW,
	Ynone,
	Yes,
	0,
	Pe,
	0x07,
	E,
	0,
	APOPW,
	Ynone,
	Yss,
	0,
	Pe,
	0x17,
	E,
	0,
	APOPW,
	Ynone,
	Yfs,
	0,
	Pe,
	0x0f,
	0xa1,
	E,
	APOPW,
	Ynone,
	Ygs,
	0,
	Pe,
	0x0f,
	0xa9,
	E,
	/* mov seg */
	AMOVW,
	Yes,
	Yml,
	1,
	0x8c,
	0,
	0,
	0,
	AMOVW,
	Ycs,
	Yml,
	1,
	0x8c,
	1,
	0,
	0,
	AMOVW,
	Yss,
	Yml,
	1,
	0x8c,
	2,
	0,
	0,
	AMOVW,
	Yds,
	Yml,
	1,
	0x8c,
	3,
	0,
	0,
	AMOVW,
	Yfs,
	Yml,
	1,
	0x8c,
	4,
	0,
	0,
	AMOVW,
	Ygs,
	Yml,
	1,
	0x8c,
	5,
	0,
	0,
	AMOVW,
	Yml,
	Yes,
	2,
	0x8e,
	0,
	0,
	0,
	AMOVW,
	Yml,
	Ycs,
	2,
	0x8e,
	1,
	0,
	0,
	AMOVW,
	Yml,
	Yss,
	2,
	0x8e,
	2,
	0,
	0,
	AMOVW,
	Yml,
	Yds,
	2,
	0x8e,
	3,
	0,
	0,
	AMOVW,
	Yml,
	Yfs,
	2,
	0x8e,
	4,
	0,
	0,
	AMOVW,
	Yml,
	Ygs,
	2,
	0x8e,
	5,
	0,
	0,
	/* mov cr */
	AMOVL,
	Ycr0,
	Yml,
	3,
	0x0f,
	0x20,
	0,
	0,
	AMOVL,
	Ycr2,
	Yml,
	3,
	0x0f,
	0x20,
	2,
	0,
	AMOVL,
	Ycr3,
	Yml,
	3,
	0x0f,
	0x20,
	3,
	0,
	AMOVL,
	Ycr4,
	Yml,
	3,
	0x0f,
	0x20,
	4,
	0,
	AMOVL,
	Yml,
	Ycr0,
	4,
	0x0f,
	0x22,
	0,
	0,
	AMOVL,
	Yml,
	Ycr2,
	4,
	0x0f,
	0x22,
	2,
	0,
	AMOVL,
	Yml,
	Ycr3,
	4,
	0x0f,
	0x22,
	3,
	0,
	AMOVL,
	Yml,
	Ycr4,
	4,
	0x0f,
	0x22,
	4,
	0,
	/* mov dr */
	AMOVL,
	Ydr0,
	Yml,
	3,
	0x0f,
	0x21,
	0,
	0,
	AMOVL,
	Ydr6,
	Yml,
	3,
	0x0f,
	0x21,
	6,
	0,
	AMOVL,
	Ydr7,
	Yml,
	3,
	0x0f,
	0x21,
	7,
	0,
	AMOVL,
	Yml,
	Ydr0,
	4,
	0x0f,
	0x23,
	0,
	0,
	AMOVL,
	Yml,
	Ydr6,
	4,
	0x0f,
	0x23,
	6,
	0,
	AMOVL,
	Yml,
	Ydr7,
	4,
	0x0f,
	0x23,
	7,
	0,
	/* mov tr */
	AMOVL,
	Ytr6,
	Yml,
	3,
	0x0f,
	0x24,
	6,
	0,
	AMOVL,
	Ytr7,
	Yml,
	3,
	0x0f,
	0x24,
	7,
	0,
	AMOVL,
	Yml,
	Ytr6,
	4,
	0x0f,
	0x26,
	6,
	E,
	AMOVL,
	Yml,
	Ytr7,
	4,
	0x0f,
	0x26,
	7,
	E,
	/* lgdt, sgdt, lidt, sidt */
	AMOVL,
	Ym,
	Ygdtr,
	4,
	0x0f,
	0x01,
	2,
	0,
	AMOVL,
	Ygdtr,
	Ym,
	3,
	0x0f,
	0x01,
	0,
	0,
	AMOVL,
	Ym,
	Yidtr,
	4,
	0x0f,
	0x01,
	3,
	0,
	AMOVL,
	Yidtr,
	Ym,
	3,
	0x0f,
	0x01,
	1,
	0,
	/* lldt, sldt */
	AMOVW,
	Yml,
	Yldtr,
	4,
	0x0f,
	0x00,
	2,
	0,
	AMOVW,
	Yldtr,
	Yml,
	3,
	0x0f,
	0x00,
	0,
	0,
	/* lmsw, smsw */
	AMOVW,
	Yml,
	Ymsw,
	4,
	0x0f,
	0x01,
	6,
	0,
	AMOVW,
	Ymsw,
	Yml,
	3,
	0x0f,
	0x01,
	4,
	0,
	/* ltr, str */
	AMOVW,
	Yml,
	Ytask,
	4,
	0x0f,
	0x00,
	3,
	0,
	AMOVW,
	Ytask,
	Yml,
	3,
	0x0f,
	0x00,
	1,
	0,
	/* load full pointer */
	AMOVL,
	Yml,
	Ycol,
	5,
	0,
	0,
	0,
	0,
	AMOVW,
	Yml,
	Ycol,
	5,
	Pe,
	0,
	0,
	0,
	/* double shift */
	ASHLL,
	Ycol,
	Yml,
	6,
	0xa4,
	0xa5,
	0,
	0,
	ASHRL,
	Ycol,
	Yml,
	6,
	0xac,
	0xad,
	0,
	0,
	/* extra imul */
	AIMULW,
	Yml,
	Yrl,
	7,
	Pq,
	0xaf,
	0,
	0,
	AIMULL,
	Yml,
	Yrl,
	7,
	Pm,
	0xaf,
	0,
	0,
	/* load TLS base pointer */
	AMOVL,
	Ytls,
	Yrl,
	8,
	0,
	0,
	0,
	0,
	0,
}

// byteswapreg returns a byte-addressable register (AX, BX, CX, DX)
// which is not referenced in a->type.
// If a is empty, it returns BX to account for MULB-like instructions
// that might use DX and AX.
func byteswapreg(ctxt *liblink.Link, a *liblink.Addr) int {
	var cana int
	var canb int
	var canc int
	var cand int
	cand = 1
	canc = cand
	canb = canc
	cana = canb
	switch a.Typ {
	case D_NONE:
		cand = 0
		cana = cand
	case D_AX,
		D_AL,
		D_AH,
		D_INDIR + D_AX:
		cana = 0
	case D_BX,
		D_BL,
		D_BH,
		D_INDIR + D_BX:
		canb = 0
	case D_CX,
		D_CL,
		D_CH,
		D_INDIR + D_CX:
		canc = 0
	case D_DX,
		D_DL,
		D_DH,
		D_INDIR + D_DX:
		cand = 0
		break
	}
	switch a.Index {
	case D_AX:
		cana = 0
	case D_BX:
		canb = 0
	case D_CX:
		canc = 0
	case D_DX:
		cand = 0
		break
	}
	if cana != 0 {
		return D_AX
	}
	if canb != 0 {
		return D_BX
	}
	if canc != 0 {
		return D_CX
	}
	if cand != 0 {
		return D_DX
	}
	ctxt.Diag("impossible byte register")
	log.Fatalf("bad code")
	return 0
}

func subreg(p *liblink.Prog, from int, to int) {
	if false { /* debug['Q'] */
		fmt.Printf("\n%v\ts/%v/%v/\n", p, Rconv(from), Rconv(to))
	}
	if p.From.Typ == from {
		p.From.Typ = to
		p.Ft = 0
	}
	if p.To.Typ == from {
		p.To.Typ = to
		p.Tt = 0
	}
	if p.From.Index == from {
		p.From.Index = to
		p.Ft = 0
	}
	if p.To.Index == from {
		p.To.Index = to
		p.Tt = 0
	}
	from += D_INDIR
	if p.From.Typ == from {
		p.From.Typ = to + D_INDIR
		p.Ft = 0
	}
	if p.To.Typ == from {
		p.To.Typ = to + D_INDIR
		p.Tt = 0
	}
	if false { /* debug['Q'] */
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
	var z int
	var op int
	var ft int
	var tt int
	var breg int
	var v int64
	var pre int
	var rel liblink.Reloc
	var r *liblink.Reloc
	var a *liblink.Addr
	ctxt.Curp = p // TODO
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
		p.Ft = uint8(oclass(&p.From))
	}
	if p.Tt == 0 {
		p.Tt = uint8(oclass(&p.To))
	}
	ft = int(p.Ft) * Ymax
	tt = int(p.Tt) * Ymax
	o = &optab[p.As]
	t = o.ytab
	if t == nil {
		ctxt.Diag("asmins: noproto %P", p)
		return
	}
	for z = 0; t[0] != 0; (func() { z += int(t[3]); t = t[4:] })() {
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
	case Pb: /* botch */
		break
	}
	op = int(o.op[z])
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
		asmand(ctxt, &p.From, reg[p.To.Typ])
	case Zm_r:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, reg[p.To.Typ])
	case Zm2_r:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = o.op[z+1]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, reg[p.To.Typ])
	case Zm_r_xm:
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.From, reg[p.To.Typ])
	case Zm_r_i_xm:
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.From, reg[p.To.Typ])
		ctxt.Andptr[0] = uint8(p.To.Offset)
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zibm_r:
		for {
			tmp2 := z
			z++
			op = int(o.op[tmp2])
			if op == 0 {
				break
			}
			ctxt.Andptr[0] = uint8(op)
			ctxt.Andptr = ctxt.Andptr[1:]
		}
		asmand(ctxt, &p.From, reg[p.To.Typ])
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
		p.Ft = 0
		asmand(ctxt, &p.From, reg[p.To.Typ])
		p.From.Index = p.From.Typ
		p.From.Typ = D_ADDR
		p.Ft = 0
	case Zm_o:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, int(o.op[z+1]))
	case Zr_m:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, reg[p.From.Typ])
	case Zr_m_xm:
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.To, reg[p.From.Typ])
	case Zr_m_i_xm:
		mediaop(ctxt, o, op, int(t[3]), z)
		asmand(ctxt, &p.To, reg[p.From.Typ])
		ctxt.Andptr[0] = uint8(p.From.Offset)
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zcallindreg:
		r = liblink.Addrel(ctxt.Cursym)
		r.Off = p.Pc
		r.Typ = liblink.R_CALLIND
		r.Siz = 0
		fallthrough
	// fallthrough
	case Zo_m:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, int(o.op[z+1]))
	case Zm_ibo:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, int(o.op[z+1]))
		ctxt.Andptr[0] = uint8(vaddr(ctxt, &p.To, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zibo_m:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, int(o.op[z+1]))
		ctxt.Andptr[0] = uint8(vaddr(ctxt, &p.From, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Z_ib,
		Zib_:
		if t[2] == Zib_ {
			a = &p.From
		} else {
			a = &p.To
		}
		v = vaddr(ctxt, a, nil)
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = uint8(v)
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zib_rp:
		ctxt.Andptr[0] = uint8(op + reg[p.To.Typ])
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = uint8(vaddr(ctxt, &p.From, nil))
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zil_rp:
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
	case Zib_rr:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, reg[p.To.Typ])
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
			asmand(ctxt, &p.To, int(o.op[z+1]))
		} else {
			a = &p.To
			asmand(ctxt, &p.From, int(o.op[z+1]))
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
		asmand(ctxt, &p.To, reg[p.To.Typ])
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
		ctxt.Andptr[0] = uint8(op + reg[p.To.Typ])
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zrp_:
		ctxt.Andptr[0] = uint8(op + reg[p.From.Typ])
		ctxt.Andptr = ctxt.Andptr[1:]
	case Zclr:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, reg[p.To.Typ])
	case Zcall:
		if p.To.Sym == nil {
			ctxt.Diag("call without target")
			log.Fatalf("bad code")
		}
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		r = liblink.Addrel(ctxt.Cursym)
		r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
		r.Typ = liblink.R_CALL
		r.Siz = 4
		r.Sym = p.To.Sym
		r.Add = p.To.Offset
		put4(ctxt, 0)
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
		// Fill in backward jump now.
		q = p.Pcond
		if q == nil {
			ctxt.Diag("jmp/branch/loop without target")
			log.Fatalf("bad code")
		}
		if p.Back&1 != 0 {
			v = q.Pc - (p.Pc + 2)
			if v >= -128 {
				if p.As == AJCXZW {
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
			if p.As == AJCXZW {
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
	case Zcallcon,
		Zjmpcon:
		if t[2] == Zcallcon {
			ctxt.Andptr[0] = uint8(op)
			ctxt.Andptr = ctxt.Andptr[1:]
		} else {
			ctxt.Andptr[0] = o.op[z+1]
			ctxt.Andptr = ctxt.Andptr[1:]
		}
		r = liblink.Addrel(ctxt.Cursym)
		r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
		r.Typ = liblink.R_PCREL
		r.Siz = 4
		r.Add = p.To.Offset
		put4(ctxt, 0)
	case Zcallind:
		ctxt.Andptr[0] = uint8(op)
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = o.op[z+1]
		ctxt.Andptr = ctxt.Andptr[1:]
		r = liblink.Addrel(ctxt.Cursym)
		r.Off = p.Pc + int64(-cap(ctxt.Andptr)+cap(ctxt.And[:]))
		r.Typ = liblink.R_ADDR
		r.Siz = 4
		r.Add = p.To.Offset
		r.Sym = p.To.Sym
		put4(ctxt, 0)
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
			}
		}
	case Zmov:
		goto domov
	}
	return
domov:
	for t = ymovtab; t[0] != 0; t = t[8:] {
		if p.As == int(t[0]) {
			if ycover[ft+int(t[1])] != 0 {
				if ycover[tt+int(t[2])] != 0 {
					goto mfound
				}
			}
		}
	}
	/*
	 * here, the assembly has failed.
	 * if its a byte instruction that has
	 * unaddressable registers, try to
	 * exchange registers and reissue the
	 * instruction with the operands renamed.
	 */
bad:
	pp = *p
	z = p.From.Typ
	if z >= D_BP && z <= D_DI {
		breg = byteswapreg(ctxt, &p.To)
		if breg != D_AX {
			ctxt.Andptr[0] = 0x87
			ctxt.Andptr = ctxt.Andptr[1:] /* xchg lhs,bx */
			asmand(ctxt, &p.From, reg[breg])
			subreg(&pp, z, breg)
			doasm(ctxt, &pp)
			ctxt.Andptr[0] = 0x87
			ctxt.Andptr = ctxt.Andptr[1:] /* xchg lhs,bx */
			asmand(ctxt, &p.From, reg[breg])
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
		breg = byteswapreg(ctxt, &p.From)
		if breg != D_AX {
			ctxt.Andptr[0] = 0x87
			ctxt.Andptr = ctxt.Andptr[1:] /* xchg rhs,bx */
			asmand(ctxt, &p.To, reg[breg])
			subreg(&pp, z, breg)
			doasm(ctxt, &pp)
			ctxt.Andptr[0] = 0x87
			ctxt.Andptr = ctxt.Andptr[1:] /* xchg rhs,bx */
			asmand(ctxt, &p.To, reg[breg])
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
	ctxt.Diag("doasm: notfound t2=%ux from=%ux to=%ux %P", t[2], p.From.Typ, p.To.Typ, p)
	return
mfound:
	switch t[3] {
	default:
		ctxt.Diag("asmins: unknown mov %d %P", t[3], p)
	case 0: /* lit */
		for z = 4; t[z] != E; z++ {
			ctxt.Andptr[0] = t[z]
			ctxt.Andptr = ctxt.Andptr[1:]
		}
	case 1: /* r,m */
		ctxt.Andptr[0] = t[4]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, int(t[5]))
	case 2: /* m,r */
		ctxt.Andptr[0] = t[4]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, int(t[5]))
	case 3: /* r,m - 2op */
		ctxt.Andptr[0] = t[4]
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = t[5]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.To, int(t[6]))
	case 4: /* m,r - 2op */
		ctxt.Andptr[0] = t[4]
		ctxt.Andptr = ctxt.Andptr[1:]
		ctxt.Andptr[0] = t[5]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, int(t[6]))
	case 5: /* load full pointer, trash heap */
		if t[4] != 0 {
			ctxt.Andptr[0] = t[4]
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
		asmand(ctxt, &p.From, reg[p.To.Typ])
	case 6: /* double shift */
		z = p.From.Typ
		switch z {
		default:
			goto bad
		case D_CONST:
			ctxt.Andptr[0] = 0x0f
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = t[4]
			ctxt.Andptr = ctxt.Andptr[1:]
			asmand(ctxt, &p.To, reg[p.From.Index])
			ctxt.Andptr[0] = uint8(p.From.Offset)
			ctxt.Andptr = ctxt.Andptr[1:]
		case D_CL,
			D_CX:
			ctxt.Andptr[0] = 0x0f
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = t[5]
			ctxt.Andptr = ctxt.Andptr[1:]
			asmand(ctxt, &p.To, reg[p.From.Index])
			break
		}
	case 7: /* imul rm,r */
		if t[4] == Pq {
			ctxt.Andptr[0] = Pe
			ctxt.Andptr = ctxt.Andptr[1:]
			ctxt.Andptr[0] = Pm
			ctxt.Andptr = ctxt.Andptr[1:]
		} else {
			ctxt.Andptr[0] = t[4]
			ctxt.Andptr = ctxt.Andptr[1:]
		}
		ctxt.Andptr[0] = t[5]
		ctxt.Andptr = ctxt.Andptr[1:]
		asmand(ctxt, &p.From, reg[p.To.Typ])
	// NOTE: The systems listed here are the ones that use the "TLS initial exec" model,
	// where you load the TLS base register into a register and then index off that
	// register to access the actual TLS variables. Systems that allow direct TLS access
	// are handled in prefixof above and should not be listed here.
	case 8: /* mov tls, r */
		switch ctxt.Headtype {
		default:
			log.Fatalf("unknown TLS base location for %s", liblink.Headstr(ctxt.Headtype))
		// ELF TLS base is 0(GS).
		case liblink.Hlinux,
			liblink.Hnacl:
			pp.From = p.From
			pp.From.Typ = D_INDIR + D_GS
			pp.From.Offset = 0
			pp.From.Index = D_NONE
			pp.From.Scale = 0
			ctxt.Andptr[0] = 0x65
			ctxt.Andptr = ctxt.Andptr[1:] // GS
			ctxt.Andptr[0] = 0x8B
			ctxt.Andptr = ctxt.Andptr[1:]
			asmand(ctxt, &pp.From, reg[p.To.Typ])
		case liblink.Hplan9:
			if ctxt.Plan9privates == nil {
				ctxt.Plan9privates = liblink.Linklookup(ctxt, "_privates", 0)
			}
			pp.From = liblink.Addr{}
			pp.From.Typ = D_EXTERN
			pp.From.Sym = ctxt.Plan9privates
			pp.From.Offset = 0
			pp.From.Index = D_NONE
			ctxt.Andptr[0] = 0x8B
			ctxt.Andptr = ctxt.Andptr[1:]
			asmand(ctxt, &pp.From, reg[p.To.Typ])
		// Windows TLS base is always 0x14(FS).
		case liblink.Hwindows:
			pp.From = p.From
			pp.From.Typ = D_INDIR + D_FS
			pp.From.Offset = 0x14
			pp.From.Index = D_NONE
			pp.From.Scale = 0
			ctxt.Andptr[0] = 0x64
			ctxt.Andptr = ctxt.Andptr[1:] // FS
			ctxt.Andptr[0] = 0x8B
			ctxt.Andptr = ctxt.Andptr[1:]
			asmand(ctxt, &pp.From, reg[p.To.Typ])
			break
		}
		break
	}
}

var naclret = []uint8{
	0x5d, // POPL BP
	// 0x8b, 0x7d, 0x00, // MOVL (BP), DI - catch return to invalid address, for debugging
	0x83,
	0xe5,
	0xe0, // ANDL $~31, BP
	0xff,
	0xe5, // JMP BP
}

func asmins(ctxt *liblink.Link, p *liblink.Prog) {
	var r *liblink.Reloc
	ctxt.Andptr = ctxt.And[:]
	if p.As == AUSEFIELD {
		r = liblink.Addrel(ctxt.Cursym)
		r.Off = 0
		r.Sym = p.From.Sym
		r.Typ = liblink.R_USEFIELD
		r.Siz = 0
		return
	}
	if ctxt.Headtype == liblink.Hnacl {
		switch p.As {
		case ARET:
			copy(ctxt.Andptr, naclret)
			ctxt.Andptr = ctxt.Andptr[len(naclret):]
			return
		case ACALL,
			AJMP:
			if D_AX <= p.To.Typ && p.To.Typ <= D_DI {
				ctxt.Andptr[0] = 0x83
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = uint8(0xe0 | (p.To.Typ - D_AX))
				ctxt.Andptr = ctxt.Andptr[1:]
				ctxt.Andptr[0] = 0xe0
				ctxt.Andptr = ctxt.Andptr[1:]
			}
		case AINT:
			ctxt.Andptr[0] = 0xf4
			ctxt.Andptr = ctxt.Andptr[1:]
			return
		}
	}
	doasm(ctxt, p)
	if -cap(ctxt.Andptr) > -cap(ctxt.And[len(ctxt.And):]) {
		fmt.Printf("and[] is too short - %d byte instruction\n", -cap(ctxt.Andptr)+cap(ctxt.And[:]))
		log.Fatalf("bad code")
	}
}
