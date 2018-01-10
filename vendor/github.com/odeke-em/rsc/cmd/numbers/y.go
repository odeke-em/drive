//line numbers.y:2
package main

import __yyfmt__ "fmt"

//line numbers.y:2
//line numbers.y:5
type yySymType struct {
	yys      int
	expr     *Expr
	exprlist []*Expr
	stmt     *Stmt
	stmtlist []*Stmt
	tok      int
	op       Op
	line     Line
}

const _LE = 57346
const _GE = 57347
const _EQ = 57348
const _NE = 57349
const _UNARYMINUS = 57350
const _EQQ = 57351
const _EXPR = 57352
const _MAX = 57353
const _MIN = 57354
const _USE = 57355
const _PRINT = 57356
const _CHECK = 57357
const _COND = 57358
const _INCLUDE = 57359

var yyToknames = []string{
	" <",
	"_LE",
	" >",
	"_GE",
	"_EQ",
	"_NE",
	" +",
	" -",
	" *",
	" /",
	"_UNARYMINUS",
	" =",
	"_EQQ",
	"_EXPR",
	"_MAX",
	"_MIN",
	"_USE",
	"_PRINT",
	"_CHECK",
	"_COND",
	"_INCLUDE",
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line numbers.y:101

//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 30
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 150

var yyAct = []int{

	24, 19, 66, 44, 65, 44, 23, 17, 20, 21,
	42, 16, 41, 22, 46, 44, 18, 40, 15, 38,
	39, 25, 26, 45, 44, 43, 44, 14, 13, 47,
	48, 49, 50, 51, 52, 53, 54, 55, 56, 64,
	63, 61, 1, 60, 10, 62, 3, 58, 59, 32,
	33, 34, 35, 36, 37, 28, 29, 30, 31, 6,
	30, 31, 7, 8, 9, 2, 5, 4, 68, 0,
	70, 11, 12, 69, 32, 33, 34, 35, 36, 37,
	28, 29, 30, 31, 32, 33, 34, 35, 36, 37,
	28, 29, 30, 31, 28, 29, 30, 31, 67, 0,
	0, 0, 0, 0, 0, 0, 0, 71, 32, 33,
	34, 35, 36, 37, 28, 29, 30, 31, 32, 33,
	34, 35, 36, 37, 28, 29, 30, 31, 0, 0,
	0, 57, 0, 0, 0, 0, 0, 0, 0, 27,
	32, 33, 34, 35, 36, 37, 28, 29, 30, 31,
}
var yyPact = []int{

	-1000, -1000, 42, -1000, -1000, -1000, 56, 2, 1, -8,
	-10, -1000, -1000, -10, -10, -10, 114, -1000, -10, -10,
	-9, -14, -16, -2, 136, -4, -13, -1000, -10, -10,
	-10, -10, -10, -10, -10, -10, -10, -10, 104, -1000,
	-10, -10, -10, 16, -10, 15, 14, 48, 48, -1000,
	-1000, 84, 84, 84, 84, 84, 84, -1000, -23, -25,
	70, -1000, 136, -1000, -1000, -1000, -1000, -10, 45, -10,
	80, -1000,
}
var yyPgo = []int{

	0, 65, 46, 0, 44, 6, 42,
}
var yyR1 = []int{

	0, 6, 1, 1, 1, 1, 2, 2, 2, 2,
	4, 4, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 5, 5,
}
var yyR2 = []int{

	0, 1, 0, 2, 2, 2, 4, 5, 5, 5,
	1, 1, 1, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 2, 4, 4, 8, 1, 3,
}
var yyChk = []int{

	-1000, -6, -1, -2, 25, 24, 17, 20, 21, 22,
	-4, 15, 16, 26, 26, 26, -3, 17, 26, 11,
	18, 19, 23, -5, -3, -5, -5, 25, 10, 11,
	12, 13, 4, 5, 6, 7, 8, 9, -3, -3,
	26, 26, 26, 27, 28, 27, 27, -3, -3, -3,
	-3, -3, -3, -3, -3, -3, -3, 27, -5, -5,
	-3, 25, -3, 25, 25, 27, 27, 28, -3, 28,
	-3, 27,
}
var yyDef = []int{

	2, -2, 1, 3, 4, 5, 0, 0, 0, 0,
	0, 10, 11, 0, 0, 0, 0, 12, 0, 0,
	0, 0, 0, 0, 28, 0, 0, 6, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 24,
	0, 0, 0, 0, 0, 0, 0, 14, 15, 16,
	17, 18, 19, 20, 21, 22, 23, 13, 0, 0,
	0, 7, 29, 8, 9, 25, 26, 0, 0, 0,
	0, 27,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	25, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	26, 27, 12, 10, 28, 11, 3, 13, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	4, 15, 6,
}
var yyTok2 = []int{

	2, 3, 5, 7, 8, 9, 14, 16, 17, 18,
	19, 20, 21, 22, 23, 24,
}
var yyTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

const yyFlag = -1000

func yyTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(yyToknames) {
		if yyToknames[c-4] != "" {
			return yyToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yylex1(lex yyLexer, lval *yySymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		c = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			c = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		c = yyTok3[i+0]
		if c == char {
			c = yyTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %U %s\n", uint(char), yyTokname(c))
	}
	return c
}

func yyParse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yychar), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar = yylex1(yylex, &yylval)
	}
	yyn += yychar
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yychar { /* valid shift */
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yychar {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error("syntax error")
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf("saw %s\n", yyTokname(yychar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}
			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		//line numbers.y:35
		{
			prog = append(prog, yyS[yypt-0].stmtlist...)
		}
	case 2:
		//line numbers.y:38
		{
			yyVAL.stmtlist = nil
		}
	case 3:
		//line numbers.y:40
		{
			yyVAL.stmtlist = append(yyS[yypt-1].stmtlist, yyS[yypt-0].stmt)
		}
	case 4:
		//line numbers.y:42
		{
			yyVAL.stmtlist = yyS[yypt-1].stmtlist
		}
	case 5:
		//line numbers.y:44
		{
			yyVAL.stmtlist = yyS[yypt-1].stmtlist
		}
	case 6:
		//line numbers.y:48
		{
			yyVAL.stmt = &Stmt{line: yyS[yypt-3].expr.line, left: yyS[yypt-3].expr, op: yyS[yypt-2].op, right: yyS[yypt-1].expr}
		}
	case 7:
		//line numbers.y:50
		{
			yyVAL.stmt = &Stmt{line: yyS[yypt-4].line, op: opUse, list: yyS[yypt-2].exprlist}
		}
	case 8:
		//line numbers.y:52
		{
			yyVAL.stmt = &Stmt{line: yyS[yypt-4].line, op: opPrint, list: yyS[yypt-2].exprlist}
		}
	case 9:
		//line numbers.y:54
		{
			yyVAL.stmt = &Stmt{line: yyS[yypt-4].line, op: opCheck, list: yyS[yypt-2].exprlist}
		}
	case 10:
		//line numbers.y:58
		{
			yyVAL.op = opAssign
		}
	case 11:
		//line numbers.y:60
		{
			yyVAL.op = opAssignWeak
		}
	case 12:
		yyVAL.expr = yyS[yypt-0].expr
	case 13:
		//line numbers.y:65
		{
			yyVAL.expr = yyS[yypt-1].expr
		}
	case 14:
		//line numbers.y:67
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opAdd, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 15:
		//line numbers.y:69
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opSub, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 16:
		//line numbers.y:71
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opMul, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 17:
		//line numbers.y:73
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opDiv, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 18:
		//line numbers.y:75
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opLess, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 19:
		//line numbers.y:77
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opLessEqual, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 20:
		//line numbers.y:79
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opGreater, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 21:
		//line numbers.y:81
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opGreaterEqual, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 22:
		//line numbers.y:83
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opEqual, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 23:
		//line numbers.y:85
		{
			yyVAL.expr = &Expr{line: yyS[yypt-2].expr.line, op: opNotEqual, left: yyS[yypt-2].expr, right: yyS[yypt-0].expr}
		}
	case 24:
		//line numbers.y:87
		{
			yyVAL.expr = &Expr{line: yyS[yypt-1].line, op: opMinus, left: yyS[yypt-0].expr}
		}
	case 25:
		//line numbers.y:89
		{
			yyVAL.expr = &Expr{line: yyS[yypt-3].line, op: opMax, list: yyS[yypt-1].exprlist}
		}
	case 26:
		//line numbers.y:91
		{
			yyVAL.expr = &Expr{line: yyS[yypt-3].line, op: opMin, list: yyS[yypt-1].exprlist}
		}
	case 27:
		//line numbers.y:93
		{
			yyVAL.expr = &Expr{line: yyS[yypt-7].line, op: opCond, list: []*Expr{yyS[yypt-5].expr, yyS[yypt-3].expr, yyS[yypt-1].expr}}
		}
	case 28:
		//line numbers.y:97
		{
			yyVAL.exprlist = []*Expr{yyS[yypt-0].expr}
		}
	case 29:
		//line numbers.y:99
		{
			yyVAL.exprlist = append(yyS[yypt-2].exprlist, yyS[yypt-0].expr)
		}
	}
	goto yystack /* stack new state and value */
}
