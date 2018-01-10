package arm

import (
	"fmt"
	"log"
	"math"
	"sort"

	"code.google.com/p/rsc/c2go/liblink"
)

type Optab struct {
	as       int
	a1       uint8
	a2       int
	a3       uint8
	typ      uint8
	size     int
	param    int
	flag     int8
	pcrelsiz uint8
}

type Oprang struct {
	start []Optab
	stop  []Optab
}

type Opcross [32][2][32]uint8

const (
	LFROM  = 1 << 0
	LTO    = 1 << 1
	LPOOL  = 1 << 2
	LPCREL = 1 << 3
	C_NONE = 0 + iota - 4
	C_REG
	C_REGREG
	C_REGREG2
	C_SHIFT
	C_FREG
	C_PSR
	C_FCR
	C_RCON
	C_NCON
	C_SCON
	C_LCON
	C_LCONADDR
	C_ZFCON
	C_SFCON
	C_LFCON
	C_RACON
	C_LACON
	C_SBRA
	C_LBRA
	C_HAUTO
	C_FAUTO
	C_HFAUTO
	C_SAUTO
	C_LAUTO
	C_HOREG
	C_FOREG
	C_HFOREG
	C_SOREG
	C_ROREG
	C_SROREG
	C_LOREG
	C_PC
	C_SP
	C_HREG
	C_ADDR
	C_GOK
)

var optab = []Optab{
	/* struct Optab:
	OPCODE,	from, prog->reg, to,		 type,size,param,flag */
	{ATEXT, C_ADDR, C_NONE, C_LCON, 0, 0, 0, 0, 0},
	{ATEXT, C_ADDR, C_REG, C_LCON, 0, 0, 0, 0, 0},
	{AADD, C_REG, C_REG, C_REG, 1, 4, 0, 0, 0},
	{AADD, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0},
	{AMVN, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0},
	{ACMP, C_REG, C_REG, C_NONE, 1, 4, 0, 0, 0},
	{AADD, C_RCON, C_REG, C_REG, 2, 4, 0, 0, 0},
	{AADD, C_RCON, C_NONE, C_REG, 2, 4, 0, 0, 0},
	{AMOVW, C_RCON, C_NONE, C_REG, 2, 4, 0, 0, 0},
	{AMVN, C_RCON, C_NONE, C_REG, 2, 4, 0, 0, 0},
	{ACMP, C_RCON, C_REG, C_NONE, 2, 4, 0, 0, 0},
	{AADD, C_SHIFT, C_REG, C_REG, 3, 4, 0, 0, 0},
	{AADD, C_SHIFT, C_NONE, C_REG, 3, 4, 0, 0, 0},
	{AMVN, C_SHIFT, C_NONE, C_REG, 3, 4, 0, 0, 0},
	{ACMP, C_SHIFT, C_REG, C_NONE, 3, 4, 0, 0, 0},
	{AMOVW, C_RACON, C_NONE, C_REG, 4, 4, REGSP, 0, 0},
	{AB, C_NONE, C_NONE, C_SBRA, 5, 4, 0, LPOOL, 0},
	{ABL, C_NONE, C_NONE, C_SBRA, 5, 4, 0, 0, 0},
	{ABX, C_NONE, C_NONE, C_SBRA, 74, 20, 0, 0, 0},
	{ABEQ, C_NONE, C_NONE, C_SBRA, 5, 4, 0, 0, 0},
	{AB, C_NONE, C_NONE, C_ROREG, 6, 4, 0, LPOOL, 0},
	{ABL, C_NONE, C_NONE, C_ROREG, 7, 4, 0, 0, 0},
	{ABL, C_REG, C_NONE, C_ROREG, 7, 4, 0, 0, 0},
	{ABX, C_NONE, C_NONE, C_ROREG, 75, 12, 0, 0, 0},
	{ABXRET, C_NONE, C_NONE, C_ROREG, 76, 4, 0, 0, 0},
	{ASLL, C_RCON, C_REG, C_REG, 8, 4, 0, 0, 0},
	{ASLL, C_RCON, C_NONE, C_REG, 8, 4, 0, 0, 0},
	{ASLL, C_REG, C_NONE, C_REG, 9, 4, 0, 0, 0},
	{ASLL, C_REG, C_REG, C_REG, 9, 4, 0, 0, 0},
	{ASWI, C_NONE, C_NONE, C_NONE, 10, 4, 0, 0, 0},
	{ASWI, C_NONE, C_NONE, C_LOREG, 10, 4, 0, 0, 0},
	{ASWI, C_NONE, C_NONE, C_LCON, 10, 4, 0, 0, 0},
	{AWORD, C_NONE, C_NONE, C_LCON, 11, 4, 0, 0, 0},
	{AWORD, C_NONE, C_NONE, C_LCONADDR, 11, 4, 0, 0, 0},
	{AWORD, C_NONE, C_NONE, C_ADDR, 11, 4, 0, 0, 0},
	{AMOVW, C_NCON, C_NONE, C_REG, 12, 4, 0, 0, 0},
	{AMOVW, C_LCON, C_NONE, C_REG, 12, 4, 0, LFROM, 0},
	{AMOVW, C_LCONADDR, C_NONE, C_REG, 12, 4, 0, LFROM | LPCREL, 4},
	{AADD, C_NCON, C_REG, C_REG, 13, 8, 0, 0, 0},
	{AADD, C_NCON, C_NONE, C_REG, 13, 8, 0, 0, 0},
	{AMVN, C_NCON, C_NONE, C_REG, 13, 8, 0, 0, 0},
	{ACMP, C_NCON, C_REG, C_NONE, 13, 8, 0, 0, 0},
	{AADD, C_LCON, C_REG, C_REG, 13, 8, 0, LFROM, 0},
	{AADD, C_LCON, C_NONE, C_REG, 13, 8, 0, LFROM, 0},
	{AMVN, C_LCON, C_NONE, C_REG, 13, 8, 0, LFROM, 0},
	{ACMP, C_LCON, C_REG, C_NONE, 13, 8, 0, LFROM, 0},
	{AMOVB, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0},
	{AMOVBS, C_REG, C_NONE, C_REG, 14, 8, 0, 0, 0},
	{AMOVBU, C_REG, C_NONE, C_REG, 58, 4, 0, 0, 0},
	{AMOVH, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0},
	{AMOVHS, C_REG, C_NONE, C_REG, 14, 8, 0, 0, 0},
	{AMOVHU, C_REG, C_NONE, C_REG, 14, 8, 0, 0, 0},
	{AMUL, C_REG, C_REG, C_REG, 15, 4, 0, 0, 0},
	{AMUL, C_REG, C_NONE, C_REG, 15, 4, 0, 0, 0},
	{ADIV, C_REG, C_REG, C_REG, 16, 4, 0, 0, 0},
	{ADIV, C_REG, C_NONE, C_REG, 16, 4, 0, 0, 0},
	{AMULL, C_REG, C_REG, C_REGREG, 17, 4, 0, 0, 0},
	{AMULA, C_REG, C_REG, C_REGREG2, 17, 4, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_SAUTO, 20, 4, REGSP, 0, 0},
	{AMOVW, C_REG, C_NONE, C_SOREG, 20, 4, 0, 0, 0},
	{AMOVB, C_REG, C_NONE, C_SAUTO, 20, 4, REGSP, 0, 0},
	{AMOVB, C_REG, C_NONE, C_SOREG, 20, 4, 0, 0, 0},
	{AMOVBS, C_REG, C_NONE, C_SAUTO, 20, 4, REGSP, 0, 0},
	{AMOVBS, C_REG, C_NONE, C_SOREG, 20, 4, 0, 0, 0},
	{AMOVBU, C_REG, C_NONE, C_SAUTO, 20, 4, REGSP, 0, 0},
	{AMOVBU, C_REG, C_NONE, C_SOREG, 20, 4, 0, 0, 0},
	{AMOVW, C_SAUTO, C_NONE, C_REG, 21, 4, REGSP, 0, 0},
	{AMOVW, C_SOREG, C_NONE, C_REG, 21, 4, 0, 0, 0},
	{AMOVBU, C_SAUTO, C_NONE, C_REG, 21, 4, REGSP, 0, 0},
	{AMOVBU, C_SOREG, C_NONE, C_REG, 21, 4, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_LAUTO, 30, 8, REGSP, LTO, 0},
	{AMOVW, C_REG, C_NONE, C_LOREG, 30, 8, 0, LTO, 0},
	{AMOVW, C_REG, C_NONE, C_ADDR, 64, 8, 0, LTO | LPCREL, 4},
	{AMOVB, C_REG, C_NONE, C_LAUTO, 30, 8, REGSP, LTO, 0},
	{AMOVB, C_REG, C_NONE, C_LOREG, 30, 8, 0, LTO, 0},
	{AMOVB, C_REG, C_NONE, C_ADDR, 64, 8, 0, LTO | LPCREL, 4},
	{AMOVBS, C_REG, C_NONE, C_LAUTO, 30, 8, REGSP, LTO, 0},
	{AMOVBS, C_REG, C_NONE, C_LOREG, 30, 8, 0, LTO, 0},
	{AMOVBS, C_REG, C_NONE, C_ADDR, 64, 8, 0, LTO | LPCREL, 4},
	{AMOVBU, C_REG, C_NONE, C_LAUTO, 30, 8, REGSP, LTO, 0},
	{AMOVBU, C_REG, C_NONE, C_LOREG, 30, 8, 0, LTO, 0},
	{AMOVBU, C_REG, C_NONE, C_ADDR, 64, 8, 0, LTO | LPCREL, 4},
	{AMOVW, C_LAUTO, C_NONE, C_REG, 31, 8, REGSP, LFROM, 0},
	{AMOVW, C_LOREG, C_NONE, C_REG, 31, 8, 0, LFROM, 0},
	{AMOVW, C_ADDR, C_NONE, C_REG, 65, 8, 0, LFROM | LPCREL, 4},
	{AMOVBU, C_LAUTO, C_NONE, C_REG, 31, 8, REGSP, LFROM, 0},
	{AMOVBU, C_LOREG, C_NONE, C_REG, 31, 8, 0, LFROM, 0},
	{AMOVBU, C_ADDR, C_NONE, C_REG, 65, 8, 0, LFROM | LPCREL, 4},
	{AMOVW, C_LACON, C_NONE, C_REG, 34, 8, REGSP, LFROM, 0},
	{AMOVW, C_PSR, C_NONE, C_REG, 35, 4, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_PSR, 36, 4, 0, 0, 0},
	{AMOVW, C_RCON, C_NONE, C_PSR, 37, 4, 0, 0, 0},
	{AMOVM, C_LCON, C_NONE, C_SOREG, 38, 4, 0, 0, 0},
	{AMOVM, C_SOREG, C_NONE, C_LCON, 39, 4, 0, 0, 0},
	{ASWPW, C_SOREG, C_REG, C_REG, 40, 4, 0, 0, 0},
	{ARFE, C_NONE, C_NONE, C_NONE, 41, 4, 0, 0, 0},
	{AMOVF, C_FREG, C_NONE, C_FAUTO, 50, 4, REGSP, 0, 0},
	{AMOVF, C_FREG, C_NONE, C_FOREG, 50, 4, 0, 0, 0},
	{AMOVF, C_FAUTO, C_NONE, C_FREG, 51, 4, REGSP, 0, 0},
	{AMOVF, C_FOREG, C_NONE, C_FREG, 51, 4, 0, 0, 0},
	{AMOVF, C_FREG, C_NONE, C_LAUTO, 52, 12, REGSP, LTO, 0},
	{AMOVF, C_FREG, C_NONE, C_LOREG, 52, 12, 0, LTO, 0},
	{AMOVF, C_LAUTO, C_NONE, C_FREG, 53, 12, REGSP, LFROM, 0},
	{AMOVF, C_LOREG, C_NONE, C_FREG, 53, 12, 0, LFROM, 0},
	{AMOVF, C_FREG, C_NONE, C_ADDR, 68, 8, 0, LTO | LPCREL, 4},
	{AMOVF, C_ADDR, C_NONE, C_FREG, 69, 8, 0, LFROM | LPCREL, 4},
	{AADDF, C_FREG, C_NONE, C_FREG, 54, 4, 0, 0, 0},
	{AADDF, C_FREG, C_REG, C_FREG, 54, 4, 0, 0, 0},
	{AMOVF, C_FREG, C_NONE, C_FREG, 54, 4, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_FCR, 56, 4, 0, 0, 0},
	{AMOVW, C_FCR, C_NONE, C_REG, 57, 4, 0, 0, 0},
	{AMOVW, C_SHIFT, C_NONE, C_REG, 59, 4, 0, 0, 0},
	{AMOVBU, C_SHIFT, C_NONE, C_REG, 59, 4, 0, 0, 0},
	{AMOVB, C_SHIFT, C_NONE, C_REG, 60, 4, 0, 0, 0},
	{AMOVBS, C_SHIFT, C_NONE, C_REG, 60, 4, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_SHIFT, 61, 4, 0, 0, 0},
	{AMOVB, C_REG, C_NONE, C_SHIFT, 61, 4, 0, 0, 0},
	{AMOVBS, C_REG, C_NONE, C_SHIFT, 61, 4, 0, 0, 0},
	{AMOVBU, C_REG, C_NONE, C_SHIFT, 61, 4, 0, 0, 0},
	{ACASE, C_REG, C_NONE, C_NONE, 62, 4, 0, LPCREL, 8},
	{ABCASE, C_NONE, C_NONE, C_SBRA, 63, 4, 0, LPCREL, 0},
	{AMOVH, C_REG, C_NONE, C_HAUTO, 70, 4, REGSP, 0, 0},
	{AMOVH, C_REG, C_NONE, C_HOREG, 70, 4, 0, 0, 0},
	{AMOVHS, C_REG, C_NONE, C_HAUTO, 70, 4, REGSP, 0, 0},
	{AMOVHS, C_REG, C_NONE, C_HOREG, 70, 4, 0, 0, 0},
	{AMOVHU, C_REG, C_NONE, C_HAUTO, 70, 4, REGSP, 0, 0},
	{AMOVHU, C_REG, C_NONE, C_HOREG, 70, 4, 0, 0, 0},
	{AMOVB, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0},
	{AMOVB, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0},
	{AMOVBS, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0},
	{AMOVBS, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0},
	{AMOVH, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0},
	{AMOVH, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0},
	{AMOVHS, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0},
	{AMOVHS, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0},
	{AMOVHU, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0},
	{AMOVHU, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0},
	{AMOVH, C_REG, C_NONE, C_LAUTO, 72, 8, REGSP, LTO, 0},
	{AMOVH, C_REG, C_NONE, C_LOREG, 72, 8, 0, LTO, 0},
	{AMOVH, C_REG, C_NONE, C_ADDR, 94, 8, 0, LTO | LPCREL, 4},
	{AMOVHS, C_REG, C_NONE, C_LAUTO, 72, 8, REGSP, LTO, 0},
	{AMOVHS, C_REG, C_NONE, C_LOREG, 72, 8, 0, LTO, 0},
	{AMOVHS, C_REG, C_NONE, C_ADDR, 94, 8, 0, LTO | LPCREL, 4},
	{AMOVHU, C_REG, C_NONE, C_LAUTO, 72, 8, REGSP, LTO, 0},
	{AMOVHU, C_REG, C_NONE, C_LOREG, 72, 8, 0, LTO, 0},
	{AMOVHU, C_REG, C_NONE, C_ADDR, 94, 8, 0, LTO | LPCREL, 4},
	{AMOVB, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0},
	{AMOVB, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0},
	{AMOVB, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4},
	{AMOVBS, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0},
	{AMOVBS, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0},
	{AMOVBS, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4},
	{AMOVH, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0},
	{AMOVH, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0},
	{AMOVH, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4},
	{AMOVHS, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0},
	{AMOVHS, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0},
	{AMOVHS, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4},
	{AMOVHU, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0},
	{AMOVHU, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0},
	{AMOVHU, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4},
	{ALDREX, C_SOREG, C_NONE, C_REG, 77, 4, 0, 0, 0},
	{ASTREX, C_SOREG, C_REG, C_REG, 78, 4, 0, 0, 0},
	{AMOVF, C_ZFCON, C_NONE, C_FREG, 80, 8, 0, 0, 0},
	{AMOVF, C_SFCON, C_NONE, C_FREG, 81, 4, 0, 0, 0},
	{ACMPF, C_FREG, C_REG, C_NONE, 82, 8, 0, 0, 0},
	{ACMPF, C_FREG, C_NONE, C_NONE, 83, 8, 0, 0, 0},
	{AMOVFW, C_FREG, C_NONE, C_FREG, 84, 4, 0, 0, 0},
	{AMOVWF, C_FREG, C_NONE, C_FREG, 85, 4, 0, 0, 0},
	{AMOVFW, C_FREG, C_NONE, C_REG, 86, 8, 0, 0, 0},
	{AMOVWF, C_REG, C_NONE, C_FREG, 87, 8, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_FREG, 88, 4, 0, 0, 0},
	{AMOVW, C_FREG, C_NONE, C_REG, 89, 4, 0, 0, 0},
	{ATST, C_REG, C_NONE, C_NONE, 90, 4, 0, 0, 0},
	{ALDREXD, C_SOREG, C_NONE, C_REG, 91, 4, 0, 0, 0},
	{ASTREXD, C_SOREG, C_REG, C_REG, 92, 4, 0, 0, 0},
	{APLD, C_SOREG, C_NONE, C_NONE, 95, 4, 0, 0, 0},
	{AUNDEF, C_NONE, C_NONE, C_NONE, 96, 4, 0, 0, 0},
	{ACLZ, C_REG, C_NONE, C_REG, 97, 4, 0, 0, 0},
	{AMULWT, C_REG, C_REG, C_REG, 98, 4, 0, 0, 0},
	{AMULAWT, C_REG, C_REG, C_REGREG2, 99, 4, 0, 0, 0},
	{AUSEFIELD, C_ADDR, C_NONE, C_NONE, 0, 0, 0, 0, 0},
	{APCDATA, C_LCON, C_NONE, C_LCON, 0, 0, 0, 0, 0},
	{AFUNCDATA, C_LCON, C_NONE, C_ADDR, 0, 0, 0, 0, 0},
	{ADUFFZERO, C_NONE, C_NONE, C_SBRA, 5, 4, 0, 0, 0}, // same as ABL
	{ADUFFCOPY, C_NONE, C_NONE, C_SBRA, 5, 4, 0, 0, 0}, // same as ABL
	{ADATABUNDLE, C_NONE, C_NONE, C_NONE, 100, 4, 0, 0, 0},
	{ADATABUNDLEEND, C_NONE, C_NONE, C_NONE, 100, 0, 0, 0, 0},
	{AXXX, C_NONE, C_NONE, C_NONE, 0, 4, 0, 0, 0},
}

var pool struct {
	start int64
	size  int64
	extra uint32
}

var oprange [ALAST]Oprang

var xcmp [C_GOK + 1][C_GOK + 1]uint8

var zprg_asm5 = liblink.Prog{
	As:    AGOK,
	Scond: C_SCOND_NONE,
	Reg:   NREG,
	From: liblink.Addr{
		Name: D_NONE,
		Typ:  D_NONE,
		Reg:  NREG,
	},
	To: liblink.Addr{
		Name: D_NONE,
		Typ:  D_NONE,
		Reg:  NREG,
	},
}

var deferreturn *liblink.LSym

func nocache_asm5(p *liblink.Prog) {
	p.Optab = 0
	p.From.Class = 0
	p.To.Class = 0
}

/* size of a case statement including jump table */
func casesz(ctxt *liblink.Link, p *liblink.Prog) int {
	var jt int = 0
	var n int = 0
	var o []Optab
	for ; p != nil; p = p.Link {
		if p.As == ABCASE {
			jt = 1
		} else if jt != 0 {
			break
		}
		o = oplook(ctxt, p)
		n += o[0].size
	}
	return n
}

// asmoutnacl assembles the instruction p. It replaces asmout for NaCl.
// It returns the total number of bytes put in out, and it can change
// p->pc if extra padding is necessary.
// In rare cases, asmoutnacl might split p into two instructions.
// origPC is the PC for this Prog (no padding is taken into account).
func asmoutnacl(ctxt *liblink.Link, origPC int64, p *liblink.Prog, o []Optab, out []uint32) int {
	var size int
	var reg int
	var q *liblink.Prog
	var a *liblink.Addr
	var a2 *liblink.Addr
	size = o[0].size
	// instruction specific
	switch p.As {
	default:
		if out != nil {
			asmout(ctxt, p, o, out)
		}
	case ADATABUNDLE, // align to 16-byte boundary
		ADATABUNDLEEND: // zero width instruction, just to align next instruction to 16-byte boundary
		p.Pc = (p.Pc + 15) &^ 15
		if out != nil {
			asmout(ctxt, p, o, out)
		}
	case AUNDEF,
		APLD:
		size = 4
		if out != nil {
			switch p.As {
			case AUNDEF:
				out[0] = 0xe7fedef0 // NACL_INSTR_ARM_ABORT_NOW (UDF #0xEDE0)
			case APLD:
				out[0] = 0xe1a01001 // (MOVW R1, R1)
				break
			}
		}
	case AB,
		ABL:
		if p.To.Typ != D_OREG {
			if out != nil {
				asmout(ctxt, p, o, out)
			}
		} else {
			if p.To.Offset != 0 || size != 4 || p.To.Reg >= 16 || p.To.Reg < 0 {
				ctxt.Diag("unsupported instruction: %P", p)
			}
			if p.Pc&15 == 12 {
				p.Pc += 4
			}
			if out != nil {
				out[0] = (uint32(p.Scond)&C_SCOND)<<28 | 0x03c0013f | uint32(p.To.Reg)<<12 | uint32(p.To.Reg)<<16 // BIC $0xc000000f, Rx
				if p.As == AB {
					out[1] = (uint32(p.Scond)&C_SCOND)<<28 | 0x012fff10 | uint32(p.To.Reg) // BX Rx // ABL
				} else {
					out[1] = (uint32(p.Scond)&C_SCOND)<<28 | 0x012fff30 | uint32(p.To.Reg) // BLX Rx
				}
			}
			size = 8
		}
		// align the last instruction (the actual BL) to the last instruction in a bundle
		if p.As == ABL {
			if deferreturn == nil {
				deferreturn = liblink.Linklookup(ctxt, "runtime.deferreturn", 0)
			}
			if p.To.Sym == deferreturn {
				p.Pc = ((origPC + 15) &^ 15) + 16 - int64(size)
			} else {
				p.Pc += (16 - ((p.Pc + int64(size)) & 15)) & 15
			}
		}
	case ALDREX,
		ALDREXD,
		AMOVB,
		AMOVBS,
		AMOVBU,
		AMOVD,
		AMOVF,
		AMOVH,
		AMOVHS,
		AMOVHU,
		AMOVM,
		AMOVW,
		ASTREX,
		ASTREXD:
		if p.To.Typ == D_REG && p.To.Reg == 15 && p.From.Reg == 13 { // MOVW.W x(R13), PC
			if out != nil {
				asmout(ctxt, p, o, out)
			}
			if size == 4 {
				if out != nil {
					// Note: 5c and 5g reg.c know that DIV/MOD smashes R12
					// so that this return instruction expansion is valid.
					out[0] = out[0] &^ 0x3000                           // change PC to R12
					out[1] = (uint32(p.Scond)&C_SCOND)<<28 | 0x03ccc13f // BIC $0xc000000f, R12
					out[2] = (uint32(p.Scond)&C_SCOND)<<28 | 0x012fff1c // BX R12
				}
				size += 8
				if (p.Pc+int64(size))&15 == 4 {
					p.Pc += 4
				}
				break
			} else {
				// if the instruction used more than 4 bytes, then it must have used a very large
				// offset to update R13, so we need to additionally mask R13.
				if out != nil {
					out[size/4-1] &^= 0x3000                                   // change PC to R12
					out[size/4] = (uint32(p.Scond)&C_SCOND)<<28 | 0x03cdd103   // BIC $0xc0000000, R13
					out[size/4+1] = (uint32(p.Scond)&C_SCOND)<<28 | 0x03ccc13f // BIC $0xc000000f, R12
					out[size/4+2] = (uint32(p.Scond)&C_SCOND)<<28 | 0x012fff1c // BX R12
				}
				// p->pc+size is only ok at 4 or 12 mod 16.
				if (p.Pc+int64(size))%8 == 0 {
					p.Pc += 4
				}
				size += 12
				break
			}
		}
		if p.To.Typ == D_REG && p.To.Reg == 15 {
			ctxt.Diag("unsupported instruction (move to another register and use indirect jump instead): %P", p)
		}
		if p.To.Typ == D_OREG && p.To.Reg == 13 && (p.Scond&C_WBIT != 0) && size > 4 {
			// function prolog with very large frame size: MOVW.W R14,-100004(R13)
			// split it into two instructions:
			// 	ADD $-100004, R13
			// 	MOVW R14, 0(R13)
			q = ctxt.Prg()
			p.Scond &^= C_WBIT
			*q = *p
			a = &p.To
			if p.To.Typ == D_OREG {
				a2 = &q.To
			} else {
				a2 = &q.From
			}
			nocache_asm5(q)
			nocache_asm5(p)
			// insert q after p
			q.Link = p.Link
			p.Link = q
			q.Pcond = nil
			// make p into ADD $X, R13
			p.As = AADD
			p.From = *a
			p.From.Reg = NREG
			p.From.Typ = D_CONST
			p.To = zprg_asm5.To
			p.To.Typ = D_REG
			p.To.Reg = 13
			// make q into p but load/store from 0(R13)
			q.Spadj = 0
			*a2 = zprg_asm5.From
			a2.Typ = D_OREG
			a2.Reg = 13
			a2.Sym = nil
			a2.Offset = 0
			size = oplook(ctxt, p)[0].size
			break
		}
		if (p.To.Typ == D_OREG && p.To.Reg != 13 && p.To.Reg != 9) || (p.From.Typ == D_OREG && p.From.Reg != 13 && p.From.Reg != 9) { // MOVW Rx, X(Ry), y != 13 && y != 9 // MOVW X(Rx), Ry, x != 13 && x != 9
			if p.To.Typ == D_OREG {
				a = &p.To
			} else {
				a = &p.From
			}
			reg = a.Reg
			if size == 4 {
				// if addr.reg == NREG, then it is probably load from x(FP) with small x, no need to modify.
				if reg == NREG {
					if out != nil {
						asmout(ctxt, p, o, out)
					}
				} else {
					if out != nil {
						out[0] = (uint32(p.Scond)&C_SCOND)<<28 | 0x03c00103 | uint32(reg)<<16 | uint32(reg)<<12 // BIC $0xc0000000, Rx
					}
					if p.Pc&15 == 12 {
						p.Pc += 4
					}
					size += 4
					if out != nil {
						asmout(ctxt, p, o, out[1:])
					}
				}
				break
			} else {
				// if a load/store instruction takes more than 1 word to implement, then
				// we need to seperate the instruction into two:
				// 1. explicitly load the address into R11.
				// 2. load/store from R11.
				// This won't handle .W/.P, so we should reject such code.
				if p.Scond&(C_PBIT|C_WBIT) != 0 {
					ctxt.Diag("unsupported instruction (.P/.W): %P", p)
				}
				q = ctxt.Prg()
				*q = *p
				if p.To.Typ == D_OREG {
					a2 = &q.To
				} else {
					a2 = &q.From
				}
				nocache_asm5(q)
				nocache_asm5(p)
				// insert q after p
				q.Link = p.Link
				p.Link = q
				q.Pcond = nil
				// make p into MOVW $X(R), R11
				p.As = AMOVW
				p.From = *a
				p.From.Typ = D_CONST
				p.To = zprg_asm5.To
				p.To.Typ = D_REG
				p.To.Reg = 11
				// make q into p but load/store from 0(R11)
				*a2 = zprg_asm5.From
				a2.Typ = D_OREG
				a2.Reg = 11
				a2.Sym = nil
				a2.Offset = 0
				size = oplook(ctxt, p)[0].size
				break
			}
		} else if out != nil {
			asmout(ctxt, p, o, out)
		}
		break
	}
	// destination register specific
	if p.To.Typ == D_REG {
		switch p.To.Reg {
		case 9:
			ctxt.Diag("invalid instruction, cannot write to R9: %P", p)
		case 13:
			if out != nil {
				out[size/4] = 0xe3cdd103 // BIC $0xc0000000, R13
			}
			if (p.Pc+int64(size))&15 == 0 {
				p.Pc += 4
			}
			size += 4
			break
		}
	}
	return size
}

func span5(ctxt *liblink.Link, cursym *liblink.LSym) {
	var p *liblink.Prog
	var op *liblink.Prog
	var o []Optab
	var m int
	var bflag int
	var i int
	var v int
	var times int
	var c int64
	var opc int64
	var out [9]uint32
	var bp []uint8
	p = cursym.Text
	if p == nil || p.Link == nil { // handle external functions and ELF section symbols
		return
	}
	if oprange[AAND].start == nil {
		buildop(ctxt)
	}
	ctxt.Cursym = cursym
	ctxt.Autosize = int(p.To.Offset + 4)
	c = 0
	op = p
	p = p.Link
	for ; p != nil || ctxt.Blitrl != nil; (func() { op = p; p = p.Link })() {
		if p == nil {
			if checkpool(ctxt, op, 0) {
				p = op
				continue
			}
			// can't happen: blitrl is not nil, but checkpool didn't flushpool
			ctxt.Diag("internal inconsistency")
			break
		}
		ctxt.Curp = p
		p.Pc = c
		o = oplook(ctxt, p)
		if ctxt.Headtype != liblink.Hnacl {
			m = o[0].size
		} else {
			m = asmoutnacl(ctxt, c, p, o, nil)
			c = p.Pc            // asmoutnacl might change pc for alignment
			o = oplook(ctxt, p) // asmoutnacl might change p in rare cases
		}
		if m%4 != 0 || p.Pc%4 != 0 {
			ctxt.Diag("!pc invalid: %P size=%d", p, m)
		}
		// must check literal pool here in case p generates many instructions
		if ctxt.Blitrl != nil {
			var tmp int
			if p.As == ACASE {
				tmp = casesz(ctxt, p)
			} else {
				tmp = m
			}
			if checkpool(ctxt, op, tmp) {
				p = op
				continue
			}
		}
		if m == 0 && (p.As != AFUNCDATA && p.As != APCDATA && p.As != ADATABUNDLEEND) {
			ctxt.Diag("zero-width instruction\n%P", p)
			continue
		}
		switch o[0].flag & (LFROM | LTO | LPOOL) {
		case LFROM:
			addpool(ctxt, p, &p.From)
		case LTO:
			addpool(ctxt, p, &p.To)
		case LPOOL:
			if p.Scond&C_SCOND == C_SCOND_NONE {
				flushpool(ctxt, p, 0, 0)
			}
			break
		}
		if p.As == AMOVW && p.To.Typ == D_REG && p.To.Reg == REGPC && p.Scond&C_SCOND == C_SCOND_NONE {
			flushpool(ctxt, p, 0, 0)
		}
		c += int64(m)
	}
	cursym.Size = c
	/*
	 * if any procedure is large enough to
	 * generate a large SBRA branch, then
	 * generate extra passes putting branches
	 * around jmps to fix. this is rare.
	 */
	times = 0
	for {
		if ctxt.Debugvlog != 0 {
			fmt.Fprintf(ctxt.Bso, "%5.2f span1\n", liblink.Cputime())
		}
		bflag = 0
		c = 0
		times++
		cursym.Text.Pc = 0 // force re-layout the code.
		for p = cursym.Text; p != nil; p = p.Link {
			ctxt.Curp = p
			o = oplook(ctxt, p)
			if c > p.Pc {
				p.Pc = c
			}
			/* very large branches
			if(o->type == 6 && p->pcond) {
				otxt = p->pcond->pc - c;
				if(otxt < 0)
					otxt = -otxt;
				if(otxt >= (1L<<17) - 10) {
					q = ctxt->arch->prg();
					q->link = p->link;
					p->link = q;
					q->as = AB;
					q->to.type = D_BRANCH;
					q->pcond = p->pcond;
					p->pcond = q;
					q = ctxt->arch->prg();
					q->link = p->link;
					p->link = q;
					q->as = AB;
					q->to.type = D_BRANCH;
					q->pcond = q->link->link;
					bflag = 1;
				}
			}
			*/
			opc = p.Pc
			if ctxt.Headtype != liblink.Hnacl {
				m = o[0].size
			} else {
				m = asmoutnacl(ctxt, c, p, o, nil)
			}
			if p.Pc != opc {
				bflag = 1
			}
			//print("%P pc changed %d to %d in iter. %d\n", p, opc, (int32)p->pc, times);
			c = p.Pc + int64(m)
			if m%4 != 0 || p.Pc%4 != 0 {
				ctxt.Diag("pc invalid: %P size=%d", p, m)
			}
			if m/4 > len(out) {
				ctxt.Diag("instruction size too large: %d > %d", m/4, len(out))
			}
			if m == 0 && (p.As != AFUNCDATA && p.As != APCDATA && p.As != ADATABUNDLEEND) {
				if p.As == ATEXT {
					ctxt.Autosize = int(p.To.Offset + 4)
					continue
				}
				ctxt.Diag("zero-width instruction\n%P", p)
				continue
			}
		}
		cursym.Size = c
		if bflag == 0 {
			break
		}
	}
	if c%4 != 0 {
		ctxt.Diag("sym->size=%d, invalid", c)
	}
	/*
	 * lay out the code.  all the pc-relative code references,
	 * even cross-function, are resolved now;
	 * only data references need to be relocated.
	 * with more work we could leave cross-function
	 * code references to be relocated too, and then
	 * perhaps we'd be able to parallelize the span loop above.
	 */
	if ctxt.Tlsg == nil {
		ctxt.Tlsg = liblink.Linklookup(ctxt, "runtime.tlsg", 0)
	}
	p = cursym.Text
	ctxt.Autosize = int(p.To.Offset + 4)
	liblink.Symgrow(ctxt, cursym, cursym.Size)
	bp = cursym.P
	c = p.Pc // even p->link might need extra padding
	for p = p.Link; p != nil; p = p.Link {
		ctxt.Pc = p.Pc
		ctxt.Curp = p
		o = oplook(ctxt, p)
		opc = p.Pc
		if ctxt.Headtype != liblink.Hnacl {
			asmout(ctxt, p, o, out[:])
			m = o[0].size
		} else {
			m = asmoutnacl(ctxt, c, p, o, out[:])
			if opc != p.Pc {
				ctxt.Diag("asmoutnacl broken: pc changed (%d->%d) in last stage: %P", opc, int(p.Pc), p)
			}
		}
		if m%4 != 0 || p.Pc%4 != 0 {
			ctxt.Diag("final stage: pc invalid: %P size=%d", p, m)
		}
		if c > p.Pc {
			ctxt.Diag("PC padding invalid: want %#lld, has %#d: %P", p.Pc, c, p)
		}
		for c != p.Pc {
			// emit 0xe1a00000 (MOVW R0, R0)
			bp[0] = 0x00
			bp = bp[1:]
			bp[0] = 0x00
			bp = bp[1:]
			bp[0] = 0xa0
			bp = bp[1:]
			bp[0] = 0xe1
			bp = bp[1:]
			c += 4
		}
		for i = 0; i < m/4; i++ {
			v = int(out[i])
			bp[0] = uint8(v)
			bp = bp[1:]
			bp[0] = uint8(v >> 8)
			bp = bp[1:]
			bp[0] = uint8(v >> 16)
			bp = bp[1:]
			bp[0] = uint8(v >> 24)
			bp = bp[1:]
		}
		c += int64(m)
	}
}

/*
 * when the first reference to the literal pool threatens
 * to go out of range of a 12-bit PC-relative offset,
 * drop the pool now, and branch round it.
 * this happens only in extended basic blocks that exceed 4k.
 */
func checkpool(ctxt *liblink.Link, p *liblink.Prog, sz int) bool {
	if pool.size >= 0xff0 || immaddr(int((p.Pc+int64(sz)+4)+4+(12+pool.size)-(pool.start+8))) == 0 {
		return flushpool(ctxt, p, 1, 0)
	} else if p.Link == nil {
		return flushpool(ctxt, p, 2, 0)
	}
	return false
}

func flushpool(ctxt *liblink.Link, p *liblink.Prog, skip int, force int) bool {
	var q *liblink.Prog
	if ctxt.Blitrl != nil {
		if skip != 0 {
			if false && skip == 1 {
				fmt.Printf("note: flush literal pool at %x: len=%d ref=%x\n", uint64(p.Pc+4), uint64(pool.size), uint64(pool.start))
			}
			q = ctxt.Prg()
			q.As = AB
			q.To.Typ = D_BRANCH
			q.Pcond = p.Link
			q.Link = ctxt.Blitrl
			q.Lineno = p.Lineno
			ctxt.Blitrl = q
		} else if force == 0 && (p.Pc+(12+pool.size)-pool.start < 2048) { // 12 take into account the maximum nacl literal pool alignment padding size
			return false
		}
		if ctxt.Headtype == liblink.Hnacl && pool.size%16 != 0 {
			// if pool is not multiple of 16 bytes, add an alignment marker
			q = ctxt.Prg()
			q.As = ADATABUNDLEEND
			ctxt.Elitrl.Link = q
			ctxt.Elitrl = q
		}
		ctxt.Elitrl.Link = p.Link
		p.Link = ctxt.Blitrl
		// BUG(minux): how to correctly handle line number for constant pool entries?
		// for now, we set line number to the last instruction preceding them at least
		// this won't bloat the .debug_line tables
		for ctxt.Blitrl != nil {
			ctxt.Blitrl.Lineno = p.Lineno
			ctxt.Blitrl = ctxt.Blitrl.Link
		}
		ctxt.Blitrl = nil /* BUG: should refer back to values until out-of-range */
		ctxt.Elitrl = nil
		pool.size = 0
		pool.start = 0
		pool.extra = 0
		return true
	}
	return false
}

func addpool(ctxt *liblink.Link, p *liblink.Prog, a *liblink.Addr) {
	var q *liblink.Prog
	var t liblink.Prog
	var c int
	c = aclass(ctxt, a)
	t = zprg_asm5
	t.Ctxt = ctxt
	t.As = AWORD
	switch c {
	default:
		t.To.Offset = a.Offset
		t.To.Sym = a.Sym
		t.To.Typ = a.Typ
		t.To.Name = a.Name
		if ctxt.Flag_shared != 0 && t.To.Sym != nil {
			t.Pcrel = p
		}
	case C_SROREG,
		C_LOREG,
		C_ROREG,
		C_FOREG,
		C_SOREG,
		C_HOREG,
		C_FAUTO,
		C_SAUTO,
		C_LAUTO,
		C_LACON:
		t.To.Typ = D_CONST
		t.To.Offset = int64(ctxt.Instoffset)
		break
	}
	if t.Pcrel == nil {
		for q = ctxt.Blitrl; q != nil; q = q.Link { /* could hash on t.t0.offset */
			if q.Pcrel == nil && q.To == t.To {
				p.Pcond = q
				return
			}
		}
	}
	if ctxt.Headtype == liblink.Hnacl && pool.size%16 == 0 {
		// start a new data bundle
		q = ctxt.Prg()
		*q = zprg_asm5
		q.As = ADATABUNDLE
		q.Pc = pool.size
		pool.size += 4
		if ctxt.Blitrl == nil {
			ctxt.Blitrl = q
			pool.start = p.Pc
		} else {
			ctxt.Elitrl.Link = q
		}
		ctxt.Elitrl = q
	}
	q = ctxt.Prg()
	*q = t
	q.Pc = pool.size
	if ctxt.Blitrl == nil {
		ctxt.Blitrl = q
		pool.start = p.Pc
	} else {
		ctxt.Elitrl.Link = q
	}
	ctxt.Elitrl = q
	pool.size += 4
	p.Pcond = q
}

func regoff(ctxt *liblink.Link, a *liblink.Addr) int {
	ctxt.Instoffset = 0
	aclass(ctxt, a)
	return ctxt.Instoffset
}

func immrot(v uint32) int {
	var i int
	for i = 0; i < 16; i++ {
		if v&^0xff == 0 {
			return int(uint32(i<<8) | v | 1<<25)
		}
		v = v<<2 | v>>30
	}
	return 0
}

func immaddr(v int) int {
	if v >= 0 && v <= 0xfff {
		return v&0xfff | 1<<24 | 1<<23 /* pre indexing */ /* pre indexing, up */
	}
	if v >= -0xfff && v < 0 {
		return -v&0xfff | 1<<24 /* pre indexing */
	}
	return 0
}

func immfloat(v int) bool {
	return v&0xC03 == 0 /* offset will fit in floating-point load/store */
}

func immhalf(v int) int {
	if v >= 0 && v <= 0xff {
		return v | 1<<24 | 1<<23 /* pre indexing */ /* pre indexing, up */
	}
	if v >= -0xff && v < 0 {
		return -v&0xff | 1<<24 /* pre indexing */
	}
	return 0
}

func aclass(ctxt *liblink.Link, a *liblink.Addr) int {
	var s *liblink.LSym
	var t int
	switch a.Typ {
	case D_NONE:
		return C_NONE
	case D_REG:
		return C_REG
	case D_REGREG:
		return C_REGREG
	case D_REGREG2:
		return C_REGREG2
	case D_SHIFT:
		return C_SHIFT
	case D_FREG:
		return C_FREG
	case D_FPCR:
		return C_FCR
	case D_OREG:
		switch a.Name {
		case D_EXTERN,
			D_STATIC:
			if a.Sym == nil || a.Sym.Name == "" {
				fmt.Printf("null sym external\n")
				return C_GOK
			}
			ctxt.Instoffset = 0 // s.b. unused but just in case
			return C_ADDR
		case D_AUTO:
			ctxt.Instoffset = int(int64(ctxt.Autosize) + a.Offset)
			t = immaddr(ctxt.Instoffset)
			if t != 0 {
				if immhalf(ctxt.Instoffset) != 0 {
					var tmp int
					if immfloat(t) {
						tmp = C_HFAUTO
					} else {
						tmp = C_HAUTO
					}
					return tmp
				}
				if immfloat(t) {
					return C_FAUTO
				}
				return C_SAUTO
			}
			return C_LAUTO
		case D_PARAM:
			ctxt.Instoffset = int(int64(ctxt.Autosize) + a.Offset + 4)
			t = immaddr(ctxt.Instoffset)
			if t != 0 {
				if immhalf(ctxt.Instoffset) != 0 {
					var tmp int
					if immfloat(t) {
						tmp = C_HFAUTO
					} else {
						tmp = C_HAUTO
					}
					return tmp
				}
				if immfloat(t) {
					return C_FAUTO
				}
				return C_SAUTO
			}
			return C_LAUTO
		case D_NONE:
			ctxt.Instoffset = int(a.Offset)
			t = immaddr(ctxt.Instoffset)
			if t != 0 {
				if immhalf(ctxt.Instoffset) != 0 { /* n.b. that it will also satisfy immrot */
					var tmp int
					if immfloat(t) {
						tmp = C_HFOREG
					} else {
						tmp = C_HOREG
					}
					return tmp
				}
				if immfloat(t) {
					return C_FOREG /* n.b. that it will also satisfy immrot */
				}
				t = immrot(uint32(ctxt.Instoffset))
				if t != 0 {
					return C_SROREG
				}
				if immhalf(ctxt.Instoffset) != 0 {
					return C_HOREG
				}
				return C_SOREG
			}
			t = immrot(uint32(ctxt.Instoffset))
			if t != 0 {
				return C_ROREG
			}
			return C_LOREG
		}
		return C_GOK
	case D_PSR:
		return C_PSR
	case D_OCONST:
		switch a.Name {
		case D_EXTERN,
			D_STATIC:
			ctxt.Instoffset = 0 // s.b. unused but just in case
			return C_ADDR
		}
		return C_GOK
	case D_FCONST:
		if chipzero5(ctxt, a.U.Dval) >= 0 {
			return C_ZFCON
		}
		if chipfloat5(ctxt, a.U.Dval) >= 0 {
			return C_SFCON
		}
		return C_LFCON
	case D_CONST,
		D_CONST2:
		switch a.Name {
		case D_NONE:
			ctxt.Instoffset = int(a.Offset)
			if a.Reg != NREG {
				return aconsize(ctxt)
			}
			t = immrot(uint32(ctxt.Instoffset))
			if t != 0 {
				return C_RCON
			}
			t = immrot(uint32(^ctxt.Instoffset))
			if t != 0 {
				return C_NCON
			}
			return C_LCON
		case D_EXTERN,
			D_STATIC:
			s = a.Sym
			if s == nil {
				break
			}
			ctxt.Instoffset = 0 // s.b. unused but just in case
			return C_LCONADDR
		case D_AUTO:
			ctxt.Instoffset = int(int64(ctxt.Autosize) + a.Offset)
			return aconsize(ctxt)
		case D_PARAM:
			ctxt.Instoffset = int(int64(ctxt.Autosize) + a.Offset + 4)
			return aconsize(ctxt)
		}
		return C_GOK
	case D_BRANCH:
		return C_SBRA
	}
	return C_GOK
}

func aconsize(ctxt *liblink.Link) int {
	var t int
	t = immrot(uint32(ctxt.Instoffset))
	if t != 0 {
		return C_RACON
	}
	return C_LACON
}

func prasm(p *liblink.Prog) {
	fmt.Printf("%v\n", p)
}

func oplook(ctxt *liblink.Link, p *liblink.Prog) []Optab {
	var a1 int
	var a2 int
	var a3 int
	var r int
	var c1 []uint8
	var c3 []uint8
	var o []Optab
	var e []Optab
	a1 = p.Optab
	if a1 != 0 {
		return optab[a1-1:]
	}
	a1 = p.From.Class
	if a1 == 0 {
		a1 = aclass(ctxt, &p.From) + 1
		p.From.Class = a1
	}
	a1--
	a3 = p.To.Class
	if a3 == 0 {
		a3 = aclass(ctxt, &p.To) + 1
		p.To.Class = a3
	}
	a3--
	a2 = C_NONE
	if p.Reg != NREG {
		a2 = C_REG
	}
	r = p.As
	o = oprange[r].start
	if o == nil {
		o = oprange[r].stop /* just generate an error */
	}
	if false { /*debug['O']*/
		fmt.Printf("oplook %v %d %d %d\n", Aconv(int(p.As)), a1, a2, a3)
		fmt.Printf("\t\t%d %d\n", p.From.Typ, p.To.Typ)
	}
	e = oprange[r].stop
	c1 = xcmp[a1][:]
	c3 = xcmp[a3][:]
	for ; -cap(o) < -cap(e); o = o[1:] {
		if o[0].a2 == a2 {
			if c1[o[0].a1] != 0 {
				if c3[o[0].a3] != 0 {
					p.Optab = (-cap(o) + cap(optab)) + 1
					return o
				}
			}
		}
	}
	ctxt.Diag("illegal combination %P; %d %d %d, %d %d", p, a1, a2, a3, p.From.Typ, p.To.Typ)
	ctxt.Diag("from %d %d to %d %d\n", p.From.Typ, p.From.Name, p.To.Typ, p.To.Name)
	prasm(p)
	if o == nil {
		o = optab
	}
	return o
}

func cmp(a int, b int) int {
	if a == b {
		return 1
	}
	switch a {
	case C_LCON:
		if b == C_RCON || b == C_NCON {
			return 1
		}
	case C_LACON:
		if b == C_RACON {
			return 1
		}
	case C_LFCON:
		if b == C_ZFCON || b == C_SFCON {
			return 1
		}
	case C_HFAUTO:
		return bool2int(b == C_HAUTO || b == C_FAUTO)
	case C_FAUTO,
		C_HAUTO:
		return bool2int(b == C_HFAUTO)
	case C_SAUTO:
		return cmp(C_HFAUTO, b)
	case C_LAUTO:
		return cmp(C_SAUTO, b)
	case C_HFOREG:
		return bool2int(b == C_HOREG || b == C_FOREG)
	case C_FOREG,
		C_HOREG:
		return bool2int(b == C_HFOREG)
	case C_SROREG:
		return bool2int(cmp(C_SOREG, b) != 0 || cmp(C_ROREG, b) != 0)
	case C_SOREG,
		C_ROREG:
		return bool2int(b == C_SROREG || cmp(C_HFOREG, b) != 0)
	case C_LOREG:
		return cmp(C_SROREG, b)
	case C_LBRA:
		if b == C_SBRA {
			return 1
		}
	case C_HREG:
		return bool2int(cmp(C_SP, b) != 0 || cmp(C_PC, b) != 0)
	}
	return 0
}

type ocmp []Optab

func (x ocmp) Len() int {
	return len(x)
}

func (x ocmp) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x ocmp) Less(i, j int) bool {
	var p1 *Optab
	var p2 *Optab
	var n int
	p1 = &x[i]
	p2 = &x[j]
	n = p1.as - p2.as
	if n != 0 {
		return n < 0
	}
	n = int(p1.a1) - int(p2.a1)
	if n != 0 {
		return n < 0
	}
	n = p1.a2 - p2.a2
	if n != 0 {
		return n < 0
	}
	n = int(p1.a3) - int(p2.a3)
	if n != 0 {
		return n < 0
	}
	return false
}

func buildop(ctxt *liblink.Link) {
	var i int
	var n int
	var r int
	for i = 0; i < C_GOK; i++ {
		for n = 0; n < C_GOK; n++ {
			xcmp[i][n] = uint8(cmp(n, i))
		}
	}
	for n = 0; optab[n].as != AXXX; n++ {
		if optab[n].flag&LPCREL != 0 {
			if ctxt.Flag_shared != 0 {
				optab[n].size += int(optab[n].pcrelsiz)
			} else {
				optab[n].flag &^= LPCREL
			}
		}
	}
	sort.Sort(ocmp(optab[:n]))
	for i = 0; i < n; i++ {
		r = optab[i].as
		oprange[r].start = optab[i:]
		for optab[i].as == r {
			i++
		}
		oprange[r].stop = optab[i:]
		i--
		switch r {
		default:
			ctxt.Diag("unknown op in build: %A", r)
			log.Fatalf("bad code")
		case AADD:
			oprange[AAND] = oprange[r]
			oprange[AEOR] = oprange[r]
			oprange[ASUB] = oprange[r]
			oprange[ARSB] = oprange[r]
			oprange[AADC] = oprange[r]
			oprange[ASBC] = oprange[r]
			oprange[ARSC] = oprange[r]
			oprange[AORR] = oprange[r]
			oprange[ABIC] = oprange[r]
		case ACMP:
			oprange[ATEQ] = oprange[r]
			oprange[ACMN] = oprange[r]
		case AMVN:
			break
		case ABEQ:
			oprange[ABNE] = oprange[r]
			oprange[ABCS] = oprange[r]
			oprange[ABHS] = oprange[r]
			oprange[ABCC] = oprange[r]
			oprange[ABLO] = oprange[r]
			oprange[ABMI] = oprange[r]
			oprange[ABPL] = oprange[r]
			oprange[ABVS] = oprange[r]
			oprange[ABVC] = oprange[r]
			oprange[ABHI] = oprange[r]
			oprange[ABLS] = oprange[r]
			oprange[ABGE] = oprange[r]
			oprange[ABLT] = oprange[r]
			oprange[ABGT] = oprange[r]
			oprange[ABLE] = oprange[r]
		case ASLL:
			oprange[ASRL] = oprange[r]
			oprange[ASRA] = oprange[r]
		case AMUL:
			oprange[AMULU] = oprange[r]
		case ADIV:
			oprange[AMOD] = oprange[r]
			oprange[AMODU] = oprange[r]
			oprange[ADIVU] = oprange[r]
		case AMOVW,
			AMOVB,
			AMOVBS,
			AMOVBU,
			AMOVH,
			AMOVHS,
			AMOVHU:
			break
		case ASWPW:
			oprange[ASWPBU] = oprange[r]
		case AB,
			ABL,
			ABX,
			ABXRET,
			ADUFFZERO,
			ADUFFCOPY,
			ASWI,
			AWORD,
			AMOVM,
			ARFE,
			ATEXT,
			AUSEFIELD,
			ACASE,
			ABCASE,
			ATYPE:
			break
		case AADDF:
			oprange[AADDD] = oprange[r]
			oprange[ASUBF] = oprange[r]
			oprange[ASUBD] = oprange[r]
			oprange[AMULF] = oprange[r]
			oprange[AMULD] = oprange[r]
			oprange[ADIVF] = oprange[r]
			oprange[ADIVD] = oprange[r]
			oprange[ASQRTF] = oprange[r]
			oprange[ASQRTD] = oprange[r]
			oprange[AMOVFD] = oprange[r]
			oprange[AMOVDF] = oprange[r]
			oprange[AABSF] = oprange[r]
			oprange[AABSD] = oprange[r]
		case ACMPF:
			oprange[ACMPD] = oprange[r]
		case AMOVF:
			oprange[AMOVD] = oprange[r]
		case AMOVFW:
			oprange[AMOVDW] = oprange[r]
		case AMOVWF:
			oprange[AMOVWD] = oprange[r]
		case AMULL:
			oprange[AMULAL] = oprange[r]
			oprange[AMULLU] = oprange[r]
			oprange[AMULALU] = oprange[r]
		case AMULWT:
			oprange[AMULWB] = oprange[r]
		case AMULAWT:
			oprange[AMULAWB] = oprange[r]
		case AMULA,
			ALDREX,
			ASTREX,
			ALDREXD,
			ASTREXD,
			ATST,
			APLD,
			AUNDEF,
			ACLZ,
			AFUNCDATA,
			APCDATA,
			ADATABUNDLE,
			ADATABUNDLEEND:
			break
		}
	}
}

func asmout(ctxt *liblink.Link, p *liblink.Prog, o []Optab, out []uint32) {
	var o1 uint32
	var o2 uint32
	var o3 uint32
	var o4 uint32
	var o5 uint32
	var o6 uint32
	var v int
	var r int
	var rf int
	var rt int
	var rt2 int
	var rel *liblink.Reloc
	ctxt.Printp = p
	o1 = 0
	o2 = 0
	o3 = 0
	o4 = 0
	o5 = 0
	o6 = 0
	ctxt.Armsize += o[0].size
	if false { /*debug['P']*/
		fmt.Printf("%x: %v\ttype %d\n", uint32(p.Pc), p, o[0].typ)
	}
	switch o[0].typ {
	default:
		ctxt.Diag("unknown asm %d", o[0].typ)
		prasm(p)
	case 0: /* pseudo ops */
		if false { /*debug['G']*/
			fmt.Printf("%x: %s: arm %d\n", uint32(p.Pc), p.From.Sym.Name, p.From.Sym.Fnptr)
		}
	case 1: /* op R,[R],R */
		o1 = oprrr(ctxt, p.As, p.Scond)
		rf = p.From.Reg
		rt = p.To.Reg
		r = p.Reg
		if p.To.Typ == D_NONE {
			rt = 0
		}
		if p.As == AMOVB || p.As == AMOVH || p.As == AMOVW || p.As == AMVN {
			r = 0
		} else if r == NREG {
			r = rt
		}
		o1 |= uint32(rf) | uint32(r)<<16 | uint32(rt)<<12
	case 2: /* movbu $I,[R],R */
		aclass(ctxt, &p.From)
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(immrot(uint32(ctxt.Instoffset)))
		rt = p.To.Reg
		r = p.Reg
		if p.To.Typ == D_NONE {
			rt = 0
		}
		if p.As == AMOVW || p.As == AMVN {
			r = 0
		} else if r == NREG {
			r = rt
		}
		o1 |= uint32(r)<<16 | uint32(rt)<<12
	case 3: /* add R<<[IR],[R],R */
		o1 = mov(ctxt, p)
	case 4: /* add $I,[R],R */
		aclass(ctxt, &p.From)
		o1 = oprrr(ctxt, AADD, p.Scond)
		o1 |= uint32(immrot(uint32(ctxt.Instoffset)))
		r = p.From.Reg
		if r == NREG {
			r = o[0].param
		}
		o1 |= uint32(r) << 16
		o1 |= uint32(p.To.Reg) << 12
	case 5: /* bra s */
		o1 = opbra(ctxt, p.As, p.Scond)
		v = -8
		if p.To.Sym != nil {
			rel = liblink.Addrel(ctxt.Cursym)
			rel.Off = ctxt.Pc
			rel.Siz = 4
			rel.Sym = p.To.Sym
			v += int(p.To.Offset)
			rel.Add = int64(int32(o1) | (int32(v) >> 2 & 0xffffff))
			rel.Typ = liblink.R_CALLARM
			break
		}
		if p.Pcond != nil {
			v = int((p.Pcond.Pc - ctxt.Pc) - 8)
		}
		o1 |= (uint32(v) >> 2) & 0xffffff
	case 6: /* b ,O(R) -> add $O,R,PC */
		aclass(ctxt, &p.To)
		o1 = oprrr(ctxt, AADD, p.Scond)
		o1 |= uint32(immrot(uint32(ctxt.Instoffset)))
		o1 |= uint32(p.To.Reg) << 16
		o1 |= REGPC << 12
	case 7: /* bl (R) -> blx R */
		aclass(ctxt, &p.To)
		if ctxt.Instoffset != 0 {
			ctxt.Diag("%P: doesn't support BL offset(REG) where offset != 0", p)
		}
		o1 = oprrr(ctxt, ABL, p.Scond)
		o1 |= uint32(p.To.Reg)
		rel = liblink.Addrel(ctxt.Cursym)
		rel.Off = ctxt.Pc
		rel.Siz = 0
		rel.Typ = liblink.R_CALLIND
	case 8: /* sll $c,[R],R -> mov (R<<$c),R */
		aclass(ctxt, &p.From)
		o1 = oprrr(ctxt, p.As, p.Scond)
		r = p.Reg
		if r == NREG {
			r = p.To.Reg
		}
		o1 |= uint32(r)
		o1 |= (uint32(ctxt.Instoffset) & 31) << 7
		o1 |= uint32(p.To.Reg) << 12
	case 9: /* sll R,[R],R -> mov (R<<R),R */
		o1 = oprrr(ctxt, p.As, p.Scond)
		r = p.Reg
		if r == NREG {
			r = p.To.Reg
		}
		o1 |= uint32(r)
		o1 |= uint32(p.From.Reg)<<8 | 1<<4
		o1 |= uint32(p.To.Reg) << 12
	case 10: /* swi [$con] */
		o1 = oprrr(ctxt, p.As, p.Scond)
		if p.To.Typ != D_NONE {
			aclass(ctxt, &p.To)
			o1 |= uint32(ctxt.Instoffset) & 0xffffff
		}
	case 11: /* word */
		aclass(ctxt, &p.To)
		o1 = uint32(ctxt.Instoffset)
		if p.To.Sym != nil {
			// This case happens with words generated
			// in the PC stream as part of the literal pool.
			rel = liblink.Addrel(ctxt.Cursym)
			rel.Off = ctxt.Pc
			rel.Siz = 4
			rel.Sym = p.To.Sym
			rel.Add = p.To.Offset
			// runtime.tlsg is special.
			// Its "address" is the offset from the TLS thread pointer
			// to the thread-local g and m pointers.
			// Emit a TLS relocation instead of a standard one.
			if rel.Sym == ctxt.Tlsg {
				rel.Typ = liblink.R_TLS
				if ctxt.Flag_shared != 0 {
					rel.Add += ctxt.Pc - p.Pcrel.Pc - 8 - int64(rel.Siz)
				}
				rel.Xadd = rel.Add
				rel.Xsym = rel.Sym
			} else if ctxt.Flag_shared != 0 {
				rel.Typ = liblink.R_PCREL
				rel.Add += ctxt.Pc - p.Pcrel.Pc - 8
			} else {
				rel.Typ = liblink.R_ADDR
			}
			o1 = 0
		}
	case 12: /* movw $lcon, reg */
		o1 = omvl(ctxt, p, &p.From, p.To.Reg)
		if o[0].flag&LPCREL != 0 {
			o2 = oprrr(ctxt, AADD, p.Scond) | uint32(p.To.Reg) | REGPC<<16 | uint32(p.To.Reg)<<12
		}
	case 13: /* op $lcon, [R], R */
		o1 = omvl(ctxt, p, &p.From, REGTMP)
		if o1 == 0 {
			break
		}
		o2 = oprrr(ctxt, p.As, p.Scond)
		o2 |= REGTMP
		r = p.Reg
		if p.As == AMOVW || p.As == AMVN {
			r = 0
		} else if r == NREG {
			r = p.To.Reg
		}
		o2 |= uint32(r) << 16
		if p.To.Typ != D_NONE {
			o2 |= uint32(p.To.Reg) << 12
		}
	case 14: /* movb/movbu/movh/movhu R,R */
		o1 = oprrr(ctxt, ASLL, p.Scond)
		if p.As == AMOVBU || p.As == AMOVHU {
			o2 = oprrr(ctxt, ASRL, p.Scond)
		} else {
			o2 = oprrr(ctxt, ASRA, p.Scond)
		}
		r = p.To.Reg
		o1 |= uint32(p.From.Reg) | uint32(r)<<12
		o2 |= uint32(r) | uint32(r)<<12
		if p.As == AMOVB || p.As == AMOVBS || p.As == AMOVBU {
			o1 |= 24 << 7
			o2 |= 24 << 7
		} else {
			o1 |= 16 << 7
			o2 |= 16 << 7
		}
	case 15: /* mul r,[r,]r */
		o1 = oprrr(ctxt, p.As, p.Scond)
		rf = p.From.Reg
		rt = p.To.Reg
		r = p.Reg
		if r == NREG {
			r = rt
		}
		if rt == r {
			r = rf
			rf = rt
		}
		if false {
			if rt == r || rf == REGPC || r == REGPC || rt == REGPC {
				ctxt.Diag("bad registers in MUL")
				prasm(p)
			}
		}
		o1 |= uint32(rf)<<8 | uint32(r) | uint32(rt)<<16
	case 16: /* div r,[r,]r */
		o1 = 0xf << 28
		o2 = 0
	case 17:
		o1 = oprrr(ctxt, p.As, p.Scond)
		rf = p.From.Reg
		rt = p.To.Reg
		rt2 = int(p.To.Offset)
		r = p.Reg
		o1 |= uint32(rf)<<8 | uint32(r) | uint32(rt)<<16 | uint32(rt2)<<12
	case 20: /* mov/movb/movbu R,O(R) */
		aclass(ctxt, &p.To)
		r = p.To.Reg
		if r == NREG {
			r = o[0].param
		}
		o1 = osr(ctxt, p.As, p.From.Reg, ctxt.Instoffset, r, p.Scond)
	case 21: /* mov/movbu O(R),R -> lr */
		aclass(ctxt, &p.From)
		r = p.From.Reg
		if r == NREG {
			r = o[0].param
		}
		o1 = olr(ctxt, ctxt.Instoffset, r, p.To.Reg, p.Scond)
		if p.As != AMOVW {
			o1 |= 1 << 22
		}
	case 30: /* mov/movb/movbu R,L(R) */
		o1 = omvl(ctxt, p, &p.To, REGTMP)
		if o1 == 0 {
			break
		}
		r = p.To.Reg
		if r == NREG {
			r = o[0].param
		}
		o2 = osrr(ctxt, p.From.Reg, REGTMP, r, p.Scond)
		if p.As != AMOVW {
			o2 |= 1 << 22
		}
	case 31: /* mov/movbu L(R),R -> lr[b] */
		o1 = omvl(ctxt, p, &p.From, REGTMP)
		if o1 == 0 {
			break
		}
		r = p.From.Reg
		if r == NREG {
			r = o[0].param
		}
		o2 = olrr(ctxt, REGTMP, r, p.To.Reg, p.Scond)
		if p.As == AMOVBU || p.As == AMOVBS || p.As == AMOVB {
			o2 |= 1 << 22
		}
	case 34: /* mov $lacon,R */
		o1 = omvl(ctxt, p, &p.From, REGTMP)
		if o1 == 0 {
			break
		}
		o2 = oprrr(ctxt, AADD, p.Scond)
		o2 |= REGTMP
		r = p.From.Reg
		if r == NREG {
			r = o[0].param
		}
		o2 |= uint32(r) << 16
		if p.To.Typ != D_NONE {
			o2 |= uint32(p.To.Reg) << 12
		}
	case 35: /* mov PSR,R */
		o1 = 2<<23 | 0xf<<16 | 0<<0
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
		o1 |= (uint32(p.From.Reg) & 1) << 22
		o1 |= uint32(p.To.Reg) << 12
	case 36: /* mov R,PSR */
		o1 = 2<<23 | 0x29f<<12 | 0<<4
		if p.Scond&C_FBIT != 0 {
			o1 ^= 0x010 << 12
		}
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
		o1 |= (uint32(p.To.Reg) & 1) << 22
		o1 |= uint32(p.From.Reg) << 0
	case 37: /* mov $con,PSR */
		aclass(ctxt, &p.From)
		o1 = 2<<23 | 0x29f<<12 | 0<<4
		if p.Scond&C_FBIT != 0 {
			o1 ^= 0x010 << 12
		}
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
		o1 |= uint32(immrot(uint32(ctxt.Instoffset)))
		o1 |= (uint32(p.To.Reg) & 1) << 22
		o1 |= uint32(p.From.Reg) << 0
	case 38,
		39:
		switch o[0].typ {
		case 38: /* movm $con,oreg -> stm */
			o1 = 0x4 << 25
			o1 |= uint32(p.From.Offset & 0xffff)
			o1 |= uint32(p.To.Reg) << 16
			aclass(ctxt, &p.To)
		case 39: /* movm oreg,$con -> ldm */
			o1 = 0x4<<25 | 1<<20
			o1 |= uint32(p.To.Offset & 0xffff)
			o1 |= uint32(p.From.Reg) << 16
			aclass(ctxt, &p.From)
			break
		}
		if ctxt.Instoffset != 0 {
			ctxt.Diag("offset must be zero in MOVM; %P", p)
		}
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
		if p.Scond&C_PBIT != 0 {
			o1 |= 1 << 24
		}
		if p.Scond&C_UBIT != 0 {
			o1 |= 1 << 23
		}
		if p.Scond&C_SBIT != 0 {
			o1 |= 1 << 22
		}
		if p.Scond&C_WBIT != 0 {
			o1 |= 1 << 21
		}
	case 40: /* swp oreg,reg,reg */
		aclass(ctxt, &p.From)
		if ctxt.Instoffset != 0 {
			ctxt.Diag("offset must be zero in SWP")
		}
		o1 = 0x2<<23 | 0x9<<4
		if p.As != ASWPW {
			o1 |= 1 << 22
		}
		o1 |= uint32(p.From.Reg) << 16
		o1 |= uint32(p.Reg) << 0
		o1 |= uint32(p.To.Reg) << 12
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
	case 41: /* rfe -> movm.s.w.u 0(r13),[r15] */
		o1 = 0xe8fd8000
	case 50: /* floating point store */
		v = regoff(ctxt, &p.To)
		r = p.To.Reg
		if r == NREG {
			r = o[0].param
		}
		o1 = ofsr(ctxt, p.As, p.From.Reg, v, r, p.Scond, p)
	case 51: /* floating point load */
		v = regoff(ctxt, &p.From)
		r = p.From.Reg
		if r == NREG {
			r = o[0].param
		}
		o1 = ofsr(ctxt, p.As, p.To.Reg, v, r, p.Scond, p) | 1<<20
	case 52: /* floating point store, int32 offset UGLY */
		o1 = omvl(ctxt, p, &p.To, REGTMP)
		if o1 == 0 {
			break
		}
		r = p.To.Reg
		if r == NREG {
			r = o[0].param
		}
		o2 = oprrr(ctxt, AADD, p.Scond) | REGTMP<<12 | REGTMP<<16 | uint32(r)
		o3 = ofsr(ctxt, p.As, p.From.Reg, 0, REGTMP, p.Scond, p)
	case 53: /* floating point load, int32 offset UGLY */
		o1 = omvl(ctxt, p, &p.From, REGTMP)
		if o1 == 0 {
			break
		}
		r = p.From.Reg
		if r == NREG {
			r = o[0].param
		}
		o2 = oprrr(ctxt, AADD, p.Scond) | REGTMP<<12 | REGTMP<<16 | uint32(r)
		o3 = ofsr(ctxt, p.As, p.To.Reg, 0, REGTMP, p.Scond, p) | 1<<20
	case 54: /* floating point arith */
		o1 = oprrr(ctxt, p.As, p.Scond)
		rf = p.From.Reg
		rt = p.To.Reg
		r = p.Reg
		if r == NREG {
			r = rt
			if p.As == AMOVF || p.As == AMOVD || p.As == AMOVFD || p.As == AMOVDF || p.As == ASQRTF || p.As == ASQRTD || p.As == AABSF || p.As == AABSD {
				r = 0
			}
		}
		o1 |= uint32(rf) | uint32(r)<<16 | uint32(rt)<<12
	case 56: /* move to FP[CS]R */
		o1 = (uint32(p.Scond)&C_SCOND)<<28 | 0xe<<24 | 1<<8 | 1<<4
		o1 |= (uint32(p.To.Reg)+1)<<21 | uint32(p.From.Reg)<<12
	case 57: /* move from FP[CS]R */
		o1 = (uint32(p.Scond)&C_SCOND)<<28 | 0xe<<24 | 1<<8 | 1<<4
		o1 |= (uint32(p.From.Reg)+1)<<21 | uint32(p.To.Reg)<<12 | 1<<20
	case 58: /* movbu R,R */
		o1 = oprrr(ctxt, AAND, p.Scond)
		o1 |= uint32(immrot(0xff))
		rt = p.To.Reg
		r = p.From.Reg
		if p.To.Typ == D_NONE {
			rt = 0
		}
		if r == NREG {
			r = rt
		}
		o1 |= uint32(r)<<16 | uint32(rt)<<12
	case 59: /* movw/bu R<<I(R),R -> ldr indexed */
		if p.From.Reg == NREG {
			if p.As != AMOVW {
				ctxt.Diag("byte MOV from shifter operand")
			}
			o1 = mov(ctxt, p)
			break
		}
		if p.From.Offset&(1<<4) != 0 {
			ctxt.Diag("bad shift in LDR")
		}
		o1 = olrr(ctxt, int(p.From.Offset), p.From.Reg, p.To.Reg, p.Scond)
		if p.As == AMOVBU {
			o1 |= 1 << 22
		}
	case 60: /* movb R(R),R -> ldrsb indexed */
		if p.From.Reg == NREG {
			ctxt.Diag("byte MOV from shifter operand")
			o1 = mov(ctxt, p)
			break
		}
		if p.From.Offset&(^0xf) != 0 {
			ctxt.Diag("bad shift in LDRSB")
		}
		o1 = olhrr(ctxt, int(p.From.Offset), p.From.Reg, p.To.Reg, p.Scond)
		o1 ^= 1<<5 | 1<<6
	case 61: /* movw/b/bu R,R<<[IR](R) -> str indexed */
		if p.To.Reg == NREG {
			ctxt.Diag("MOV to shifter operand")
		}
		o1 = osrr(ctxt, p.From.Reg, int(p.To.Offset), p.To.Reg, p.Scond)
		if p.As == AMOVB || p.As == AMOVBS || p.As == AMOVBU {
			o1 |= 1 << 22
		}
	case 62: /* case R -> movw	R<<2(PC),PC */
		if o[0].flag&LPCREL != 0 {
			o1 = oprrr(ctxt, AADD, p.Scond) | uint32(immrot(1)) | uint32(p.From.Reg)<<16 | REGTMP<<12
			o2 = olrr(ctxt, REGTMP, REGPC, REGTMP, p.Scond)
			o2 |= 2 << 7
			o3 = oprrr(ctxt, AADD, p.Scond) | REGTMP | REGPC<<16 | REGPC<<12
		} else {
			o1 = olrr(ctxt, p.From.Reg, REGPC, REGPC, p.Scond)
			o1 |= 2 << 7
		}
	case 63: /* bcase */
		if p.Pcond != nil {
			rel = liblink.Addrel(ctxt.Cursym)
			rel.Off = ctxt.Pc
			rel.Siz = 4
			if p.To.Sym != nil && p.To.Sym.Typ != 0 {
				rel.Sym = p.To.Sym
				rel.Add = p.To.Offset
			} else {
				rel.Sym = ctxt.Cursym
				rel.Add = p.Pcond.Pc
			}
			if o[0].flag&LPCREL != 0 {
				rel.Typ = liblink.R_PCREL
				rel.Add += ctxt.Pc - p.Pcrel.Pc - 16 + int64(rel.Siz)
			} else {
				rel.Typ = liblink.R_ADDR
			}
			o1 = 0
		}
	/* reloc ops */
	case 64: /* mov/movb/movbu R,addr */
		o1 = omvl(ctxt, p, &p.To, REGTMP)
		if o1 == 0 {
			break
		}
		o2 = osr(ctxt, p.As, p.From.Reg, 0, REGTMP, p.Scond)
		if o[0].flag&LPCREL != 0 {
			o3 = o2
			o2 = oprrr(ctxt, AADD, p.Scond) | REGTMP | REGPC<<16 | REGTMP<<12
		}
	case 65: /* mov/movbu addr,R */
		o1 = omvl(ctxt, p, &p.From, REGTMP)
		if o1 == 0 {
			break
		}
		o2 = olr(ctxt, 0, REGTMP, p.To.Reg, p.Scond)
		if p.As == AMOVBU || p.As == AMOVBS || p.As == AMOVB {
			o2 |= 1 << 22
		}
		if o[0].flag&LPCREL != 0 {
			o3 = o2
			o2 = oprrr(ctxt, AADD, p.Scond) | REGTMP | REGPC<<16 | REGTMP<<12
		}
	case 68: /* floating point store -> ADDR */
		o1 = omvl(ctxt, p, &p.To, REGTMP)
		if o1 == 0 {
			break
		}
		o2 = ofsr(ctxt, p.As, p.From.Reg, 0, REGTMP, p.Scond, p)
		if o[0].flag&LPCREL != 0 {
			o3 = o2
			o2 = oprrr(ctxt, AADD, p.Scond) | REGTMP | REGPC<<16 | REGTMP<<12
		}
	case 69: /* floating point load <- ADDR */
		o1 = omvl(ctxt, p, &p.From, REGTMP)
		if o1 == 0 {
			break
		}
		o2 = ofsr(ctxt, p.As, p.To.Reg, 0, REGTMP, p.Scond, p) | 1<<20
		if o[0].flag&LPCREL != 0 {
			o3 = o2
			o2 = oprrr(ctxt, AADD, p.Scond) | REGTMP | REGPC<<16 | REGTMP<<12
		}
	/* ArmV4 ops: */
	case 70: /* movh/movhu R,O(R) -> strh */
		aclass(ctxt, &p.To)
		r = p.To.Reg
		if r == NREG {
			r = o[0].param
		}
		o1 = oshr(ctxt, p.From.Reg, ctxt.Instoffset, r, p.Scond)
	case 71: /* movb/movh/movhu O(R),R -> ldrsb/ldrsh/ldrh */
		aclass(ctxt, &p.From)
		r = p.From.Reg
		if r == NREG {
			r = o[0].param
		}
		o1 = olhr(ctxt, ctxt.Instoffset, r, p.To.Reg, p.Scond)
		if p.As == AMOVB || p.As == AMOVBS {
			o1 ^= 1<<5 | 1<<6
		} else if p.As == AMOVH || p.As == AMOVHS {
			o1 ^= (1 << 6)
		}
	case 72: /* movh/movhu R,L(R) -> strh */
		o1 = omvl(ctxt, p, &p.To, REGTMP)
		if o1 == 0 {
			break
		}
		r = p.To.Reg
		if r == NREG {
			r = o[0].param
		}
		o2 = oshrr(ctxt, p.From.Reg, REGTMP, r, p.Scond)
	case 73: /* movb/movh/movhu L(R),R -> ldrsb/ldrsh/ldrh */
		o1 = omvl(ctxt, p, &p.From, REGTMP)
		if o1 == 0 {
			break
		}
		r = p.From.Reg
		if r == NREG {
			r = o[0].param
		}
		o2 = olhrr(ctxt, REGTMP, r, p.To.Reg, p.Scond)
		if p.As == AMOVB || p.As == AMOVBS {
			o2 ^= 1<<5 | 1<<6
		} else if p.As == AMOVH || p.As == AMOVHS {
			o2 ^= (1 << 6)
		}
	case 74: /* bx $I */
		ctxt.Diag("ABX $I")
	case 75: /* bx O(R) */
		aclass(ctxt, &p.To)
		if ctxt.Instoffset != 0 {
			ctxt.Diag("non-zero offset in ABX")
		}
		/*
			o1 = 	oprrr(ctxt, AADD, p->scond) | immrot(0) | (REGPC<<16) | (REGLINK<<12);	// mov PC, LR
			o2 = ((p->scond&C_SCOND)<<28) | (0x12fff<<8) | (1<<4) | p->to.reg;		// BX R
		*/
		// p->to.reg may be REGLINK
		o1 = oprrr(ctxt, AADD, p.Scond)
		o1 |= uint32(immrot(uint32(ctxt.Instoffset)))
		o1 |= uint32(p.To.Reg) << 16
		o1 |= REGTMP << 12
		o2 = oprrr(ctxt, AADD, p.Scond) | uint32(immrot(0)) | REGPC<<16 | REGLINK<<12 // mov PC, LR
		o3 = (uint32(p.Scond)&C_SCOND)<<28 | 0x12fff<<8 | 1<<4 | REGTMP               // BX Rtmp
	case 76: /* bx O(R) when returning from fn*/
		ctxt.Diag("ABXRET")
	case 77: /* ldrex oreg,reg */
		aclass(ctxt, &p.From)
		if ctxt.Instoffset != 0 {
			ctxt.Diag("offset must be zero in LDREX")
		}
		o1 = 0x19<<20 | 0xf9f
		o1 |= uint32(p.From.Reg) << 16
		o1 |= uint32(p.To.Reg) << 12
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
	case 78: /* strex reg,oreg,reg */
		aclass(ctxt, &p.From)
		if ctxt.Instoffset != 0 {
			ctxt.Diag("offset must be zero in STREX")
		}
		o1 = 0x18<<20 | 0xf90
		o1 |= uint32(p.From.Reg) << 16
		o1 |= uint32(p.Reg) << 0
		o1 |= uint32(p.To.Reg) << 12
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
	case 80: /* fmov zfcon,freg */
		if p.As == AMOVD {
			o1 = 0xeeb00b00 // VMOV imm 64
			o2 = oprrr(ctxt, ASUBD, p.Scond)
		} else {
			o1 = 0x0eb00a00 // VMOV imm 32
			o2 = oprrr(ctxt, ASUBF, p.Scond)
		}
		v = 0x70 // 1.0
		r = p.To.Reg
		// movf $1.0, r
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
		o1 |= uint32(r) << 12
		o1 |= (uint32(v) & 0xf) << 0
		o1 |= (uint32(v) & 0xf0) << 12
		// subf r,r,r
		o2 |= uint32(r) | uint32(r)<<16 | uint32(r)<<12
	case 81: /* fmov sfcon,freg */
		o1 = 0x0eb00a00 // VMOV imm 32
		if p.As == AMOVD {
			o1 = 0xeeb00b00 // VMOV imm 64
		}
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
		o1 |= uint32(p.To.Reg) << 12
		v = chipfloat5(ctxt, p.From.U.Dval)
		o1 |= (uint32(v) & 0xf) << 0
		o1 |= (uint32(v) & 0xf0) << 12
	case 82: /* fcmp freg,freg, */
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(p.Reg)<<12 | uint32(p.From.Reg)<<0
		o2 = 0x0ef1fa10 // VMRS R15
		o2 |= (uint32(p.Scond) & C_SCOND) << 28
	case 83: /* fcmp freg,, */
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(p.From.Reg)<<12 | 1<<16
		o2 = 0x0ef1fa10 // VMRS R15
		o2 |= (uint32(p.Scond) & C_SCOND) << 28
	case 84: /* movfw freg,freg - truncate float-to-fix */
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(p.From.Reg) << 0
		o1 |= uint32(p.To.Reg) << 12
	case 85: /* movwf freg,freg - fix-to-float */
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(p.From.Reg) << 0
		o1 |= uint32(p.To.Reg) << 12
	// macro for movfw freg,FTMP; movw FTMP,reg
	case 86: /* movfw freg,reg - truncate float-to-fix */
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(p.From.Reg) << 0
		o1 |= FREGTMP << 12
		o2 = oprrr(ctxt, AMOVFW+AEND, p.Scond)
		o2 |= FREGTMP << 16
		o2 |= uint32(p.To.Reg) << 12
	// macro for movw reg,FTMP; movwf FTMP,freg
	case 87: /* movwf reg,freg - fix-to-float */
		o1 = oprrr(ctxt, AMOVWF+AEND, p.Scond)
		o1 |= uint32(p.From.Reg) << 12
		o1 |= FREGTMP << 16
		o2 = oprrr(ctxt, p.As, p.Scond)
		o2 |= FREGTMP << 0
		o2 |= uint32(p.To.Reg) << 12
	case 88: /* movw reg,freg  */
		o1 = oprrr(ctxt, AMOVWF+AEND, p.Scond)
		o1 |= uint32(p.From.Reg) << 12
		o1 |= uint32(p.To.Reg) << 16
	case 89: /* movw freg,reg  */
		o1 = oprrr(ctxt, AMOVFW+AEND, p.Scond)
		o1 |= uint32(p.From.Reg) << 16
		o1 |= uint32(p.To.Reg) << 12
	case 90: /* tst reg  */
		o1 = oprrr(ctxt, ACMP+AEND, p.Scond)
		o1 |= uint32(p.From.Reg) << 16
	case 91: /* ldrexd oreg,reg */
		aclass(ctxt, &p.From)
		if ctxt.Instoffset != 0 {
			ctxt.Diag("offset must be zero in LDREX")
		}
		o1 = 0x1b<<20 | 0xf9f
		o1 |= uint32(p.From.Reg) << 16
		o1 |= uint32(p.To.Reg) << 12
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
	case 92: /* strexd reg,oreg,reg */
		aclass(ctxt, &p.From)
		if ctxt.Instoffset != 0 {
			ctxt.Diag("offset must be zero in STREX")
		}
		o1 = 0x1a<<20 | 0xf90
		o1 |= uint32(p.From.Reg) << 16
		o1 |= uint32(p.Reg) << 0
		o1 |= uint32(p.To.Reg) << 12
		o1 |= (uint32(p.Scond) & C_SCOND) << 28
	case 93: /* movb/movh/movhu addr,R -> ldrsb/ldrsh/ldrh */
		o1 = omvl(ctxt, p, &p.From, REGTMP)
		if o1 == 0 {
			break
		}
		o2 = olhr(ctxt, 0, REGTMP, p.To.Reg, p.Scond)
		if p.As == AMOVB || p.As == AMOVBS {
			o2 ^= 1<<5 | 1<<6
		} else if p.As == AMOVH || p.As == AMOVHS {
			o2 ^= (1 << 6)
		}
		if o[0].flag&LPCREL != 0 {
			o3 = o2
			o2 = oprrr(ctxt, AADD, p.Scond) | REGTMP | REGPC<<16 | REGTMP<<12
		}
	case 94: /* movh/movhu R,addr -> strh */
		o1 = omvl(ctxt, p, &p.To, REGTMP)
		if o1 == 0 {
			break
		}
		o2 = oshr(ctxt, p.From.Reg, 0, REGTMP, p.Scond)
		if o[0].flag&LPCREL != 0 {
			o3 = o2
			o2 = oprrr(ctxt, AADD, p.Scond) | REGTMP | REGPC<<16 | REGTMP<<12
		}
	case 95: /* PLD off(reg) */
		o1 = 0xf5d0f000
		o1 |= uint32(p.From.Reg) << 16
		if p.From.Offset < 0 {
			o1 &^= (1 << 23)
			o1 |= uint32((-p.From.Offset) & 0xfff)
		} else {
			o1 |= uint32(p.From.Offset & 0xfff)
		}
	// This is supposed to be something that stops execution.
	// It's not supposed to be reached, ever, but if it is, we'd
	// like to be able to tell how we got there.  Assemble as
	// 0xf7fabcfd which is guaranteed to raise undefined instruction
	// exception.
	case 96: /* UNDEF */
		o1 = 0xf7fabcfd
	case 97: /* CLZ Rm, Rd */
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(p.To.Reg) << 12
		o1 |= uint32(p.From.Reg)
	case 98: /* MULW{T,B} Rs, Rm, Rd */
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(p.To.Reg) << 16
		o1 |= uint32(p.From.Reg) << 8
		o1 |= uint32(p.Reg)
	case 99: /* MULAW{T,B} Rs, Rm, Rn, Rd */
		o1 = oprrr(ctxt, p.As, p.Scond)
		o1 |= uint32(p.To.Reg) << 12
		o1 |= uint32(p.From.Reg) << 8
		o1 |= uint32(p.Reg)
		o1 |= uint32(p.To.Offset << 16)
	// DATABUNDLE: BKPT $0x5be0, signify the start of NaCl data bundle;
	// DATABUNDLEEND: zero width alignment marker
	case 100:
		if p.As == ADATABUNDLE {
			o1 = 0xe125be70
		}
		break
	}
	out[0] = o1
	out[1] = o2
	out[2] = o3
	out[3] = o4
	out[4] = o5
	out[5] = o6
	return
}

func mov(ctxt *liblink.Link, p *liblink.Prog) uint32 {
	var o1 uint32
	var rt int
	var r int
	aclass(ctxt, &p.From)
	o1 = oprrr(ctxt, p.As, p.Scond)
	o1 |= uint32(p.From.Offset)
	rt = p.To.Reg
	r = p.Reg
	if p.To.Typ == D_NONE {
		rt = 0
	}
	if p.As == AMOVW || p.As == AMVN {
		r = 0
	} else if r == NREG {
		r = rt
	}
	o1 |= uint32(r)<<16 | uint32(rt)<<12
	return o1
}

func oprrr(ctxt *liblink.Link, a int, sc int) uint32 {
	var o int
	o = (sc & C_SCOND) << 28
	if sc&C_SBIT != 0 {
		o |= 1 << 20
	}
	if sc&(C_PBIT|C_WBIT) != 0 {
		ctxt.Diag(".nil/.W on dp instruction")
	}
	switch a {
	case AMULU,
		AMUL:
		return uint32(o) | 0x0<<21 | 0x9<<4
	case AMULA:
		return uint32(o) | 0x1<<21 | 0x9<<4
	case AMULLU:
		return uint32(o) | 0x4<<21 | 0x9<<4
	case AMULL:
		return uint32(o) | 0x6<<21 | 0x9<<4
	case AMULALU:
		return uint32(o) | 0x5<<21 | 0x9<<4
	case AMULAL:
		return uint32(o) | 0x7<<21 | 0x9<<4
	case AAND:
		return uint32(o) | 0x0<<21
	case AEOR:
		return uint32(o) | 0x1<<21
	case ASUB:
		return uint32(o) | 0x2<<21
	case ARSB:
		return uint32(o) | 0x3<<21
	case AADD:
		return uint32(o) | 0x4<<21
	case AADC:
		return uint32(o) | 0x5<<21
	case ASBC:
		return uint32(o) | 0x6<<21
	case ARSC:
		return uint32(o) | 0x7<<21
	case ATST:
		return uint32(o) | 0x8<<21 | 1<<20
	case ATEQ:
		return uint32(o) | 0x9<<21 | 1<<20
	case ACMP:
		return uint32(o) | 0xa<<21 | 1<<20
	case ACMN:
		return uint32(o) | 0xb<<21 | 1<<20
	case AORR:
		return uint32(o) | 0xc<<21
	case AMOVB,
		AMOVH,
		AMOVW:
		return uint32(o) | 0xd<<21
	case ABIC:
		return uint32(o) | 0xe<<21
	case AMVN:
		return uint32(o) | 0xf<<21
	case ASLL:
		return uint32(o) | 0xd<<21 | 0<<5
	case ASRL:
		return uint32(o) | 0xd<<21 | 1<<5
	case ASRA:
		return uint32(o) | 0xd<<21 | 2<<5
	case ASWI:
		return uint32(o) | 0xf<<24
	case AADDD:
		return uint32(o) | 0xe<<24 | 0x3<<20 | 0xb<<8 | 0<<4
	case AADDF:
		return uint32(o) | 0xe<<24 | 0x3<<20 | 0xa<<8 | 0<<4
	case ASUBD:
		return uint32(o) | 0xe<<24 | 0x3<<20 | 0xb<<8 | 4<<4
	case ASUBF:
		return uint32(o) | 0xe<<24 | 0x3<<20 | 0xa<<8 | 4<<4
	case AMULD:
		return uint32(o) | 0xe<<24 | 0x2<<20 | 0xb<<8 | 0<<4
	case AMULF:
		return uint32(o) | 0xe<<24 | 0x2<<20 | 0xa<<8 | 0<<4
	case ADIVD:
		return uint32(o) | 0xe<<24 | 0x8<<20 | 0xb<<8 | 0<<4
	case ADIVF:
		return uint32(o) | 0xe<<24 | 0x8<<20 | 0xa<<8 | 0<<4
	case ASQRTD:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 1<<16 | 0xb<<8 | 0xc<<4
	case ASQRTF:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 1<<16 | 0xa<<8 | 0xc<<4
	case AABSD:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 0<<16 | 0xb<<8 | 0xc<<4
	case AABSF:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 0<<16 | 0xa<<8 | 0xc<<4
	case ACMPD:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 4<<16 | 0xb<<8 | 0xc<<4
	case ACMPF:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 4<<16 | 0xa<<8 | 0xc<<4
	case AMOVF:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 0<<16 | 0xa<<8 | 4<<4
	case AMOVD:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 0<<16 | 0xb<<8 | 4<<4
	case AMOVDF:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 7<<16 | 0xa<<8 | 0xc<<4 | 1<<8 // dtof
	case AMOVFD:
		return uint32(o) | 0xe<<24 | 0xb<<20 | 7<<16 | 0xa<<8 | 0xc<<4 | 0<<8 // dtof
	case AMOVWF:
		if sc&C_UBIT == 0 {
			o |= 1 << 7 /* signed */
		}
		return uint32(o) | 0xe<<24 | 0xb<<20 | 8<<16 | 0xa<<8 | 4<<4 | 0<<18 | 0<<8 // toint, double
	case AMOVWD:
		if sc&C_UBIT == 0 {
			o |= 1 << 7 /* signed */
		}
		return uint32(o) | 0xe<<24 | 0xb<<20 | 8<<16 | 0xa<<8 | 4<<4 | 0<<18 | 1<<8 // toint, double
	case AMOVFW:
		if sc&C_UBIT == 0 {
			o |= 1 << 16 /* signed */
		}
		return uint32(o) | 0xe<<24 | 0xb<<20 | 8<<16 | 0xa<<8 | 4<<4 | 1<<18 | 0<<8 | 1<<7 // toint, double, trunc
	case AMOVDW:
		if sc&C_UBIT == 0 {
			o |= 1 << 16 /* signed */
		}
		return uint32(o) | 0xe<<24 | 0xb<<20 | 8<<16 | 0xa<<8 | 4<<4 | 1<<18 | 1<<8 | 1<<7 // toint, double, trunc
	case AMOVWF + AEND: // copy WtoF
		return uint32(o) | 0xe<<24 | 0x0<<20 | 0xb<<8 | 1<<4
	case AMOVFW + AEND: // copy FtoW
		return uint32(o) | 0xe<<24 | 0x1<<20 | 0xb<<8 | 1<<4
	case ACMP + AEND: // cmp imm
		return uint32(o) | 0x3<<24 | 0x5<<20
	// CLZ doesn't support .nil
	case ACLZ:
		return uint32(o)&(0xf<<28) | 0x16f<<16 | 0xf1<<4
	case AMULWT:
		return uint32(o)&(0xf<<28) | 0x12<<20 | 0xe<<4
	case AMULWB:
		return uint32(o)&(0xf<<28) | 0x12<<20 | 0xa<<4
	case AMULAWT:
		return uint32(o)&(0xf<<28) | 0x12<<20 | 0xc<<4
	case AMULAWB:
		return uint32(o)&(0xf<<28) | 0x12<<20 | 0x8<<4
	case ABL: // BLX REG
		return uint32(o)&(0xf<<28) | 0x12fff3<<4
	}
	ctxt.Diag("bad rrr %d", a)
	prasm(ctxt.Curp)
	return 0
}

func opbra(ctxt *liblink.Link, a int, sc int) uint32 {
	if sc&(C_SBIT|C_PBIT|C_WBIT) != 0 {
		ctxt.Diag(".nil/.nil/.W on bra instruction")
	}
	sc &= C_SCOND
	if a == ABL || a == ADUFFZERO || a == ADUFFCOPY {
		return uint32(sc)<<28 | 0x5<<25 | 0x1<<24
	}
	if sc != 0xe {
		ctxt.Diag(".COND on bcond instruction")
	}
	switch a {
	case ABEQ:
		return 0x0<<28 | 0x5<<25
	case ABNE:
		return 0x1<<28 | 0x5<<25
	case ABCS:
		return 0x2<<28 | 0x5<<25
	case ABHS:
		return 0x2<<28 | 0x5<<25
	case ABCC:
		return 0x3<<28 | 0x5<<25
	case ABLO:
		return 0x3<<28 | 0x5<<25
	case ABMI:
		return 0x4<<28 | 0x5<<25
	case ABPL:
		return 0x5<<28 | 0x5<<25
	case ABVS:
		return 0x6<<28 | 0x5<<25
	case ABVC:
		return 0x7<<28 | 0x5<<25
	case ABHI:
		return 0x8<<28 | 0x5<<25
	case ABLS:
		return 0x9<<28 | 0x5<<25
	case ABGE:
		return 0xa<<28 | 0x5<<25
	case ABLT:
		return 0xb<<28 | 0x5<<25
	case ABGT:
		return 0xc<<28 | 0x5<<25
	case ABLE:
		return 0xd<<28 | 0x5<<25
	case AB:
		return 0xe<<28 | 0x5<<25
	}
	ctxt.Diag("bad bra %A", a)
	prasm(ctxt.Curp)
	return 0
}

func olr(ctxt *liblink.Link, v int, b int, r int, sc int) uint32 {
	var o uint32
	if sc&C_SBIT != 0 {
		ctxt.Diag(".nil on LDR/STR instruction")
	}
	o = (uint32(sc) & C_SCOND) << 28
	if sc&C_PBIT == 0 {
		o |= 1 << 24
	}
	if sc&C_UBIT == 0 {
		o |= 1 << 23
	}
	if sc&C_WBIT != 0 {
		o |= 1 << 21
	}
	o |= 1<<26 | 1<<20
	if v < 0 {
		if sc&C_UBIT != 0 {
			ctxt.Diag(".U on neg offset")
		}
		v = -v
		o ^= 1 << 23
	}
	if v >= 1<<12 || v < 0 {
		ctxt.Diag("literal span too large: %d (R%d)\n%P", v, b, ctxt.Printp)
	}
	o |= uint32(v)
	o |= uint32(b) << 16
	o |= uint32(r) << 12
	return o
}

func olhr(ctxt *liblink.Link, v int, b int, r int, sc int) uint32 {
	var o uint32
	if sc&C_SBIT != 0 {
		ctxt.Diag(".nil on LDRH/STRH instruction")
	}
	o = (uint32(sc) & C_SCOND) << 28
	if sc&C_PBIT == 0 {
		o |= 1 << 24
	}
	if sc&C_WBIT != 0 {
		o |= 1 << 21
	}
	o |= 1<<23 | 1<<20 | 0xb<<4
	if v < 0 {
		v = -v
		o ^= 1 << 23
	}
	if v >= 1<<8 || v < 0 {
		ctxt.Diag("literal span too large: %d (R%d)\n%P", v, b, ctxt.Printp)
	}
	o |= uint32(v)&0xf | (uint32(v)>>4)<<8 | 1<<22
	o |= uint32(b) << 16
	o |= uint32(r) << 12
	return o
}

func osr(ctxt *liblink.Link, a int, r int, v int, b int, sc int) uint32 {
	var o uint32
	o = olr(ctxt, v, b, r, sc) ^ (1 << 20)
	if a != AMOVW {
		o |= 1 << 22
	}
	return o
}

func oshr(ctxt *liblink.Link, r int, v int, b int, sc int) uint32 {
	var o uint32
	o = olhr(ctxt, v, b, r, sc) ^ (1 << 20)
	return o
}

func osrr(ctxt *liblink.Link, r int, i int, b int, sc int) uint32 {
	return olr(ctxt, i, b, r, sc) ^ (1<<25 | 1<<20)
}

func oshrr(ctxt *liblink.Link, r int, i int, b int, sc int) uint32 {
	return olhr(ctxt, i, b, r, sc) ^ (1<<22 | 1<<20)
}

func olrr(ctxt *liblink.Link, i int, b int, r int, sc int) uint32 {
	return olr(ctxt, i, b, r, sc) ^ (1 << 25)
}

func olhrr(ctxt *liblink.Link, i int, b int, r int, sc int) uint32 {
	return olhr(ctxt, i, b, r, sc) ^ (1 << 22)
}

func ofsr(ctxt *liblink.Link, a int, r int, v int, b int, sc int, p *liblink.Prog) uint32 {
	var o uint32
	if sc&C_SBIT != 0 {
		ctxt.Diag(".nil on FLDR/FSTR instruction")
	}
	o = (uint32(sc) & C_SCOND) << 28
	if sc&C_PBIT == 0 {
		o |= 1 << 24
	}
	if sc&C_WBIT != 0 {
		o |= 1 << 21
	}
	o |= 6<<25 | 1<<24 | 1<<23 | 10<<8
	if v < 0 {
		v = -v
		o ^= 1 << 23
	}
	if v&3 != 0 {
		ctxt.Diag("odd offset for floating point op: %d\n%P", v, p)
	} else if v >= 1<<10 || v < 0 {
		ctxt.Diag("literal span too large: %d\n%P", v, p)
	}
	o |= (uint32(v) >> 2) & 0xFF
	o |= uint32(b) << 16
	o |= uint32(r) << 12
	switch a {
	default:
		ctxt.Diag("bad fst %A", a)
		fallthrough
	case AMOVD:
		o |= 1 << 8
		fallthrough
	case AMOVF:
		break
	}
	return o
}

func omvl(ctxt *liblink.Link, p *liblink.Prog, a *liblink.Addr, dr int) uint32 {
	var v int
	var o1 uint32
	if p.Pcond == nil {
		aclass(ctxt, a)
		v = immrot(uint32(^ctxt.Instoffset))
		if v == 0 {
			ctxt.Diag("missing literal")
			prasm(p)
			return 0
		}
		o1 = oprrr(ctxt, AMVN, p.Scond&C_SCOND)
		o1 |= uint32(v)
		o1 |= uint32(dr) << 12
	} else {
		v = int(p.Pcond.Pc - p.Pc - 8)
		o1 = olr(ctxt, v, REGPC, dr, p.Scond&C_SCOND)
	}
	return o1
}

func chipzero5(ctxt *liblink.Link, e float64) int {
	// We use GOARM=7 to gate the use of VFPv3 vmov (imm) instructions.
	if ctxt.Goarm < 7 || e != 0 {
		return -1
	}
	return 0
}

func chipfloat5(ctxt *liblink.Link, e float64) int {
	var n int
	var h1 uint32
	var l uint32
	var h uint32
	var ei uint64
	// We use GOARM=7 to gate the use of VFPv3 vmov (imm) instructions.
	if ctxt.Goarm < 7 {
		goto no
	}
	ei = math.Float64bits(e)
	l = uint32(ei)
	h = uint32(int(ei >> 32))
	if l != 0 || h&0xffff != 0 {
		goto no
	}
	h1 = h & 0x7fc00000
	if h1 != 0x40000000 && h1 != 0x3fc00000 {
		goto no
	}
	n = 0
	// sign bit (a)
	if h&0x80000000 != 0 {
		n |= 1 << 7
	}
	// exp sign bit (b)
	if h1 == 0x3fc00000 {
		n |= 1 << 6
	}
	// rest of exp and mantissa (cd-efgh)
	n |= int((h >> 16) & 0x3f)
	//print("match %.8lux %.8lux %d\n", l, h, n);
	return n
no:
	return -1
}
