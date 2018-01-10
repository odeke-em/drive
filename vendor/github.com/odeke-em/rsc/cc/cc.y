// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This grammar is derived from the C grammar in the 'ansitize'
// program, which carried this notice:
// 
// Copyright (c) 2006 Russ Cox, 
// 	Massachusetts Institute of Technology
// 
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the
// Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute,
// sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
// 
// The above copyright notice and this permission notice shall
// be included in all copies or substantial portions of the
// Software.
// 
// The software is provided "as is", without warranty of any
// kind, express or implied, including but not limited to the
// warranties of merchantability, fitness for a particular
// purpose and noninfringement.  In no event shall the authors
// or copyright holders be liable for any claim, damages or
// other liability, whether in an action of contract, tort or
// otherwise, arising from, out of or in connection with the
// software or the use or other dealings in the software.

%{
package cc

type typeClass struct {
	c Storage
	q TypeQual
	t *Type
}

type idecor struct {
	d func(*Type) (*Type, string)
	i *Init
}

%}

%union {
	abdecor func(*Type) *Type
	decl *Decl
	decls []*Decl
	decor func(*Type) (*Type, string)
	decors []func(*Type) (*Type, string)
	expr *Expr
	exprs []*Expr
	idec idecor
	idecs []idecor
	init *Init
	inits []*Init
	label *Label
	labels []*Label
	span Span
	prefix *Prefix
	prefixes []*Prefix
	stmt *Stmt
	stmts []*Stmt
	str string
	strs []string
	tc typeClass
	tk TypeKind
	typ *Type
}

%token	<str>	tokARGBEGIN
%token	<str>	tokARGEND
%token	<str>	tokAUTOLIB
%token	<str>	tokSET
%token	<str>	tokUSED

%token	<str>	tokAuto
%token	<str>	tokBreak
%token	<str>	tokCase
%token	<str>	tokChar
%token	<str>	tokConst
%token	<str>	tokContinue
%token	<str>	tokDefault
%token	<str>	tokDo
%token	<str>	tokDotDotDot
%token	<str>	tokDouble
%token	<str>	tokEnum
%token	<str>	tokError
%token	<str>	tokExtern
%token	<str>	tokFloat
%token	<str>	tokFor
%token	<str>	tokGoto
%token	<str>	tokIf
%token	<str>	tokInline
%token	<str>	tokInt
%token	<str>	tokLitChar
%token	<str>	tokLong
%token	<str>	tokName
%token	<str>	tokNumber
%token	<str>	tokOffsetof
%token	<str>	tokRegister
%token	<str>	tokReturn
%token	<str>	tokShort
%token	<str>	tokSigned
%token	<str>	tokStatic
%token	<str>	tokStruct
%token	<str>	tokSwitch
%token	<str>	tokTypeName
%token	<str>	tokTypedef
%token	<str>	tokUnion
%token	<str>	tokUnsigned
%token	<str>	tokVaArg
%token	<str>	tokVoid
%token	<str>	tokVolatile
%token	<str>	tokWhile
%token	<str>	tokString

%type	<abdecor>	abdecor abdec1
%type	<decl>	fnarg fndef edecl
%type	<decls>	decl decl_list_opt
%type	<decls>	fnarg_list fnarg_list_opt
%type	<decls>	prog xdecl topdecl
%type	<decls>	sudecl sudecl_list
%type	<decls> edecl_list
%type	<decor>	decor sudecor
%type	<decors>	sudecor_list sudecor_list_opt
%type	<expr>	expr expr_opt cexpr cexpr_opt eqexpr eqexpr_opt
%type	<exprs>	expr_list expr_list_opt
%type	<idec>	idecor
%type	<idecs>	idecor_list idecor_list_opt
%type	<init>	init binit
%type	<inits>	braced_init_list binit_list
%type	<label>	label
%type	<labels> label_list_opt
%type	<prefix>	initprefix
%type	<prefixes>	initprefix_list
%type	<stmt>	stmt block lstmt
%type	<stmts>	block1
%type	<str>	cname qname tname cqname cqtname tag tag_opt
%type	<strs>	cqname_list cqname_list_opt
%type	<strs>	cqtname_list cqtname_list_opt
%type	<strs>	qname_list qname_list_opt
%type	<strs>	string_list
%type	<tc>	typeclass
%type	<tk>	structunion
%type	<typ>	abtype type typespec

// fake operators to resolve if/else ambiguity
%left	tokShift
%left	tokElse
%left	tokTypeName
%left	'{'
%left	tokName

// real operators - usual c precedence
%left	','
%right	'=' tokAddEq tokSubEq tokMulEq tokDivEq tokModEq tokLshEq tokRshEq tokAndEq tokXorEq tokOrEq
%right	'?' ':'
%left	tokOrOr
%left	tokAndAnd
%left	'|'
%left	'^'
%left	'&'
%left	tokEqEq
%left	tokNotEq
%left	'<' '>' tokLtEq tokGtEq
%left	tokLsh tokRsh
%left	'+' '-'
%left	'*' '/' '%'
%right	tokCast
%left	'!' '~' tokSizeof tokUnary
%right	'.' '[' ']' '(' ')' tokDec tokInc tokArrow
%left	tokString

%token	startExpr startProg tokEOF

%%

top:
	startProg prog tokEOF
	{
		yylex.(*lexer).prog = &Prog{Decls: $2}
		return 0
	}
|	startExpr cexpr tokEOF
	{
		yylex.(*lexer).expr = $2
		return 0
	}

prog:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	prog xdecl
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2...)
	}
|	prog tokAUTOLIB '(' tokName ')'
	{
	}

cexpr:
	expr_list
	{
		$<span>$ = $<span>1
		if len($1) == 1 {
			$$ = $1[0]
			break
		}
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Comma, List: $1}
	}
		
expr:
	tokName
	{
		$<span>$ = $<span>1
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Name, Text: $1, XDecl: $<decl>1}
	}
|	tokNumber
	{
		$<span>$ = $<span>1
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Number, Text: $1}
	}
|	tokLitChar
	{
		$<span>$ = $<span>1
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Number, Text: $1}
	}
|	string_list
	{
		$<span>$ = $<span>1
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: String, Texts: $1}
	}
|	expr '+' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Add, Left: $1, Right: $3}
	}
|	expr '-' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Sub, Left: $1, Right: $3}
	}
|	expr '*' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Mul, Left: $1, Right: $3}
	}
|	expr '/' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Div, Left: $1, Right: $3}
	}
|	expr '%' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Mod, Left: $1, Right: $3}
	}
|	expr tokLsh expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Lsh, Left: $1, Right: $3}
	}
|	expr tokRsh expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Rsh, Left: $1, Right: $3}
	}
|	expr '<' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Lt, Left: $1, Right: $3}
	}
|	expr '>' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Gt, Left: $1, Right: $3}
	}
|	expr tokLtEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: LtEq, Left: $1, Right: $3}
	}
|	expr tokGtEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: GtEq, Left: $1, Right: $3}
	}
|	expr tokEqEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: EqEq, Left: $1, Right: $3}
	}
|	expr tokNotEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: NotEq, Left: $1, Right: $3}
	}
|	expr '&' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: And, Left: $1, Right: $3}
	}
|	expr '^' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Xor, Left: $1, Right: $3}
	}
|	expr '|' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Or, Left: $1, Right: $3}
	}
|	expr tokAndAnd expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: AndAnd, Left: $1, Right: $3}
	}
|	expr tokOrOr expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: OrOr, Left: $1, Right: $3}
	}
|	expr '?' cexpr ':' expr
	{
		$<span>$ = span($<span>1, $<span>5)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Cond, List: []*Expr{$1, $3, $5}}
	}
|	expr '=' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Eq, Left: $1, Right: $3}
	}
|	expr tokAddEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: AddEq, Left: $1, Right: $3}
	}
|	expr tokSubEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: SubEq, Left: $1, Right: $3}
	}
|	expr tokMulEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: MulEq, Left: $1, Right: $3}
	}
|	expr tokDivEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: DivEq, Left: $1, Right: $3}
	}
|	expr tokModEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: ModEq, Left: $1, Right: $3}
	}
|	expr tokLshEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: LshEq, Left: $1, Right: $3}
	}
|	expr tokRshEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: RshEq, Left: $1, Right: $3}
	}
|	expr tokAndEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: AndEq, Left: $1, Right: $3}
	}
|	expr tokXorEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: XorEq, Left: $1, Right: $3}
	}
|	expr tokOrEq expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: OrEq, Left: $1, Right: $3}
	}
|	'*' expr	%prec tokUnary
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Indir, Left: $2}
	}
|	'&' expr	%prec tokUnary
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Addr, Left: $2}
	}
|	'+' expr	%prec tokUnary
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Plus, Left: $2}
	}
|	'-' expr	%prec tokUnary
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Minus, Left: $2}
	}
|	'!' expr
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Not, Left: $2}
	}
|	'~' expr
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Twid, Left: $2}
	}
|	tokInc expr
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: PreInc, Left: $2}
	}
|	tokDec expr
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: PreDec, Left: $2}
	}
|	tokSizeof expr	%prec tokSizeof
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: SizeofExpr, Left: $2}
	}
|	tokSizeof '(' abtype ')'	%prec tokSizeof
	{
		$<span>$ = span($<span>1, $<span>4)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: SizeofType, Type: $3}
	}
|	tokOffsetof '(' abtype ',' expr ')'	%prec tokSizeof
	{
		$<span>$ = span($<span>1, $<span>6)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Offsetof, Type: $3, Left: $5}
	}
|	'(' abtype ')' expr	%prec tokCast
	{
		$<span>$ = span($<span>1, $<span>4)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Cast, Type: $2, Left: $4}
	}
|	'(' abtype ')' braced_init_list	%prec tokCast
	{
		$<span>$ = span($<span>1, $<span>4)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: CastInit, Type: $2, Init: &Init{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Braced: $4}}
	}
|	'(' cexpr ')'
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Paren, Left: $2}
	}		
|	expr '(' expr_list_opt ')'
	{
		$<span>$ = span($<span>1, $<span>4)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Call, Left: $1, List: $3}
	}
|	expr '[' cexpr ']'
	{
		$<span>$ = span($<span>1, $<span>4)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Index, Left: $1, Right: $3}
	}
|	expr tokInc
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: PostInc, Left: $1}
	}
|	expr tokDec
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: PostDec, Left: $1}
	}
|	tokVaArg '(' expr ',' abtype ')'
	{
		$<span>$ = span($<span>1, $<span>6)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: VaArg, Left: $3, Type: $5}
	}

block1:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	block1 decl
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = $1
		for _, d := range $2 {
			$$ = append($$, &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: StmtDecl, Decl: d})
		}
	}
|	block1 lstmt
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2)
	}

block:
	'{'
	{
		yylex.(*lexer).pushScope()
	}
	block1 '}'
	{
		$<span>$ = span($<span>1, $<span>4)
		yylex.(*lexer).popScope()
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Block, Block: $3}
	}

label:
	tokCase expr ':'
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Label{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Case, Expr: $2}
	}
|	tokDefault ':'
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Label{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Default}
	}
|	tokName ':'
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Label{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: LabelName, Name: $1}
	}

lstmt:
	label_list_opt stmt
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = $2
		$$.Labels = $1
	}
		
stmt:
	';'
	{
		$<span>$ = $<span>1
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Empty}
	}
|	tokUSED '(' cexpr ')' ';'
	{
		$<span>$ = $<span>1
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Empty}
	}
|	tokSET '(' cexpr ')' ';'
	{
		$<span>$ = $<span>1
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Empty}
	}			
|	block
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	cexpr ';'
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: StmtExpr, Expr: $1}
	}	
|	tokARGBEGIN block1 tokARGEND
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: ARGBEGIN, Block: $2}
	}
|	tokBreak ';'
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Break}
	}
|	tokContinue ';'
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Continue}
	}
|	tokDo lstmt tokWhile '(' cexpr ')' ';'
	{
		$<span>$ = span($<span>1, $<span>7)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Do, Body: $2, Expr: $5}
	}
|	tokFor '(' cexpr_opt ';' cexpr_opt ';' cexpr_opt ')' lstmt
	{
		$<span>$ = span($<span>1, $<span>9)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, 
			Op: For,
			Pre: $3,
			Expr: $5,
			Post: $7,
			Body: $9,
		}
	}
|	tokGoto tag ';'
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Goto, Text: $2}
	}
|	tokIf '(' cexpr ')' lstmt	%prec tokShift
	{
		$<span>$ = span($<span>1, $<span>5)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: If, Expr: $3, Body: $5}
	}
|	tokIf '(' cexpr ')' lstmt tokElse lstmt
	{
		$<span>$ = span($<span>1, $<span>7)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: If, Expr: $3, Body: $5, Else: $7}
	}
|	tokReturn cexpr_opt ';'
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Return, Expr: $2}
	}
|	tokSwitch '(' cexpr ')' lstmt
	{
		$<span>$ = span($<span>1, $<span>5)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Switch, Expr: $3, Body: $5}
	}
|	tokWhile '(' cexpr ')' lstmt
	{
		$<span>$ = span($<span>1, $<span>5)
		$$ = &Stmt{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: While, Expr: $3, Body: $5}
	}

// Abstract declarator - abdec1 includes the slot where the name would go
abdecor:
	{
		$<span>$ = Span{}
		$$ = func(t *Type) *Type { return t}
	}
|	'*' qname_list_opt abdecor
	{
		$<span>$ = span($<span>1, $<span>3)
		_, q, _ := splitTypeWords($2)
		abdecor := $3
		$$ = func(t *Type) *Type {
			return abdecor(&Type{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Kind: Ptr, Base: t, Qual: q})
		}
	}
|	abdec1
	{
		$<span>$ = $<span>1
		$$ = $1
	}

abdec1:
	abdec1 '(' fnarg_list_opt ')'
	{
		$<span>$ = span($<span>1, $<span>4)
		abdecor := $1
		decls := $3
		span := $<span>$
		for _, decl := range decls {
			t := decl.Type
			if t != nil {
				if t.Kind == TypedefType && t.Base != nil {
					t = t.Base
				}
				if t.Kind == Array {
					if t.Width == nil {
						t = t.Base
					}
					decl.Type = &Type{Kind: Ptr, Base: t}
				}
			}
		}
		$$ = func(t *Type) *Type {
			return abdecor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Func, Base: t, Decls: decls})
		}
	}
|	abdecor '[' expr_opt ']'
	{
		$<span>$ = span($<span>1, $<span>4)
		abdecor := $1
		span := $<span>$
		expr := $3
		$$ = func(t *Type) *Type {
			return abdecor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Array, Base: t, Width: expr})
		}
			
	}	
|	'(' abdecor ')'	
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = $2
	}

// Concrete declarator
decor:
	tag
	{
		$<span>$ = $<span>1
		name := $1
		$$ = func(t *Type) (*Type, string) { return t, name }
	}		
|	'*' qname_list_opt decor
	{
		$<span>$ = span($<span>1, $<span>3)
		_, q, _ := splitTypeWords($2)
		decor := $3
		span := $<span>$
		$$ = func(t *Type) (*Type, string) {
			return decor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Ptr, Base: t, Qual: q})
		}
	}
|	'(' decor ')'
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = $2
	}
|	decor '(' fnarg_list_opt ')'
	{
		$<span>$ = span($<span>1, $<span>4)
		decor := $1
		decls := $3
		span := $<span>$
		$$ = func(t *Type) (*Type, string) {
			return decor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Func, Base: t, Decls: decls})
		}
	}
|	decor '[' expr_opt ']'
	{
		$<span>$ = span($<span>1, $<span>4)
		decor := $1
		span := $<span>$
		expr := $3
		$$ = func(t *Type) (*Type, string) {
			return decor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Array, Base: t, Width: expr})
		}	
	}	

// Function argument
fnarg:
	tokName
	{
		$<span>$ = $<span>1
		$$ = &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: $1}
	}
|	type abdecor
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Type: $2($1)}
	}	
|	type decor
	{
		$<span>$ = span($<span>1, $<span>2)
		typ, name := $2($1)
		$$ = &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: name, Type: typ}
	}
|	tokDotDotDot
	{
		$<span>$ = $<span>1
		$$ = &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: "..."}
	}

// Initialized declarator
idecor:
	decor
	{
		$<span>$ = $<span>1
		$$ = idecor{$1, nil}
	}
|	decor '=' init
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = idecor{$1, $3}
	}

// Class words
cname:
	tokAuto
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokStatic
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokExtern
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokTypedef
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokRegister
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokInline
	{
		$<span>$ = $<span>1
		$$ = $1
	}

// Qualifier words
qname:
	tokConst
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokVolatile
	{
		$<span>$ = $<span>1
		$$ = $1
	}

// Type words
tname:
	tokChar
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokShort
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokInt
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokLong
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokSigned
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokUnsigned
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokFloat
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokDouble
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokVoid
	{
		$<span>$ = $<span>1
		$$ = $1
	}

cqname:
	cname
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	qname
	{
		$<span>$ = $<span>1
		$$ = $1
	}

cqtname:
	cqname
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tname
	{
		$<span>$ = $<span>1
		$$ = $1
	}

// Type specifier but not a tname
typespec:
	tokTypeName
	{
		$<span>$ = $<span>1
		$$ = $<typ>1
		if $$ == nil {
			$$ = &Type{Kind: TypedefType, Name: $<str>1}
		}
	}

// Types annotated with class info.
//	typeclass:
//		cqname* typespec cqname*
//	|	cqname* tname cqtname*
//	|	cqname+
// except LALR(1) can't handle that.
typeclass:
	cqname_list %prec tokShift
	{
		$<span>$ = $<span>1
		$$.c, $$.q, $$.t = splitTypeWords(append($1, "int"))
	}
|	cqname_list typespec cqname_list_opt
	{
		$<span>$ = span($<span>1, $<span>3)
		$$.c, $$.q, _ = splitTypeWords(append($1, $3...))
		$$.t = $2
	}
|	cqname_list tname cqtname_list_opt
	{
		$<span>$ = span($<span>1, $<span>3)
		$1 = append($1, $2)
		$1 = append($1, $3...)
		$$.c, $$.q, $$.t = splitTypeWords($1)
	}
|	typespec cqname_list_opt
	{
		$<span>$ = span($<span>1, $<span>2)
		$$.c, $$.q, _ = splitTypeWords($2)
		$$.t = $1
	}
|	tname cqtname_list_opt
	{
		$<span>$ = span($<span>1, $<span>2)
		var ts []string
		ts = append(ts, $1)
		ts = append(ts, $2...)
		$$.c, $$.q, $$.t = splitTypeWords(ts)
	}

// Types without class info (check for class in higher level)
type:
	typeclass
	{
		$<span>$ = $<span>1
		if $1.c != 0 {
			yylex.(*lexer).Errorf("%v not allowed here", $1.c)
		}
		if $1.q != 0 && $1.q != Const && $1.q != Volatile {
			yylex.(*lexer).Errorf("%v ignored here (TODO)?", $1.q)
		}
		$$ = $1.t
	}

abtype:
	type abdecor
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = $2($1)
	}

// Declaration (finally)
decl:
	typeclass idecor_list_opt ';'
	{
		lx := yylex.(*lexer)
		$<span>$ = span($<span>1, $<span>3)
		// TODO: use $1.q
		$$ = nil
		for _, idec := range $2 {
			typ, name := idec.d($1.t)
			d := &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: name, Type: typ, Storage: $1.c, Init: idec.i}
			lx.pushDecl(d);
			$$ = append($$, d);
		}
		if $2 == nil {
			d := &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: "", Type: $1.t, Storage: $1.c}
			lx.pushDecl(d);
			$$ = append($$, d)
		}
	}

topdecl:
	typeclass idecor_list_opt ';'
	{
		lx := yylex.(*lexer)
		$<span>$ = span($<span>1, $<span>3)
		// TODO: use $1.q
		$$ = nil
		for _, idec := range $2 {
			typ, name := idec.d($1.t)
			d := lx.lookupDecl(name)
			if d == nil {
				d = &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: name, Type: typ, Storage: $1.c, Init: idec.i}
				lx.pushDecl(d)
			} else {
				d.Span = $<span>$
				if idec.i != nil {
					d.Init = idec.i
				}
			}
			$$ = append($$, d);
		}
		if $2 == nil {
			d := &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: "", Type: $1.t, Storage: $1.c}
			lx.pushDecl(d);
			$$ = append($$, d)
		}
	}

xdecl:
	topdecl
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	fndef
	{
		$<span>$ = $<span>1
		$$ = []*Decl{$1}
	}
|	tokExtern tokString '{' prog '}'
	{
		$$ = $4
	}

fndef:
	typeclass decor decl_list_opt 
	{
		lx := yylex.(*lexer)
		typ, name := $2($1.t)
		if typ.Kind != Func {
			yylex.(*lexer).Errorf("invalid function definition")
			return 0
		}
		d := lx.lookupDecl(name)
		if d == nil {
			d = &Decl{Name: name, Type: typ, Storage: $1.c}
			lx.pushDecl(d);
		} else {
			d.Type = typ
		}
		$<decl>$ = d
		lx.pushScope()
		for _, decl := range typ.Decls {
			lx.pushDecl(decl);
		}
	}
	block
	{
		yylex.(*lexer).popScope();
		$<span>$ = span($<span>1, $<span>5)
		$$ = $<decl>4
		$$.Span = $<span>$
		if $3 != nil {
			yylex.(*lexer).Errorf("cannot use pre-prototype definitions")
		}
		$$.Body = $5
	}

tag:
	tokName
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tokTypeName
	{
		$<span>$ = $<span>1
		$$ = $1
	}

// struct/union
structunion:
	tokStruct
	{
		$<span>$ = $<span>1
		$$ = Struct
	}
|	tokUnion
	{
		$<span>$ = $<span>1
		$$ = Union
	}

sudecor:
	decor
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	tag_opt ':' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		name := $1
		expr := $3
		$$ = func(t *Type) (*Type, string) {
			t.Width = expr
			return t, name
		}
	}

sudecl:
	type sudecor_list_opt ';'
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = nil
		for _, decor := range $2 {
			typ, name := decor($1)
			$$ = append($$, &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: name, Type: typ})
		}
		if $2 == nil {
			$$ = append($$, &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Type: $1})
		}
	}

typespec:
	structunion tag
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = yylex.(*lexer).pushType(&Type{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Kind: $1, Tag: $2})
	}
|	structunion tag_opt '{' sudecl_list '}'
	{
		$<span>$ = span($<span>1, $<span>5)
		$$ = yylex.(*lexer).pushType(&Type{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Kind: $1, Tag: $2, Decls: $4})
	}

initprefix:
	'.' tag
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = &Prefix{Span: $<span>$, Dot: $2}
	}

expr:
	expr tokArrow tag
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Arrow, Left: $1, Text: $3}
	}
|	expr '.' tag
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Expr{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Op: Dot, Left: $1, Text: $3}
	}

// enum
typespec:
	tokEnum tag
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = yylex.(*lexer).pushType(&Type{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Kind: Enum, Tag: $2})
	}
|	tokEnum tag_opt '{' edecl_list comma_opt '}'
	{
		$<span>$ = span($<span>1, $<span>6)
		$$ = yylex.(*lexer).pushType(&Type{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Kind: Enum, Tag: $2, Decls: $4})
	}

edecl:
	tokName eqexpr_opt
	{
		$<span>$ = span($<span>1, $<span>2)
		var x *Init
		if $2 != nil {
			x = &Init{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Expr: $2}
		}
		$$ = &Decl{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Name: $1, Init: x}
		yylex.(*lexer).pushDecl($$);
	}

eqexpr:
	'=' expr
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = $2
	}

// initializers
init:
	expr
	{
		$<span>$ = $<span>1
		$$ = &Init{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Expr: $1}
	}
|	braced_init_list
	{
		$<span>$ = $<span>1
		$$ = &Init{SyntaxInfo: SyntaxInfo{Span: $<span>$}, Braced: $1}
	}

braced_init_list:
	'{' '}'
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = []*Init{}
	}
|	'{' binit_list binit '}'
	{
		$<span>$ = span($<span>1, $<span>4)
		$$ = append($2, $3)
	}
|	'{' binit_list binit ',' '}'
	{
		$<span>$ = span($<span>1, $<span>5)
		$$ = append($2, $3)
	}

binit_list:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	binit_list binit ','
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = append($1, $2)
	}

binit:
	init
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	initprefix_list eq_opt init
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = $3
		$$.Prefix = $1
	}

initprefix:
	'[' expr ']'
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = &Prefix{Span: $<span>$, Index: $2}
	}

eq_opt:
	{
		$<span>$ = Span{}
	}
|	'='
	{
		$<span>$ = $<span>1
	}

comma_opt:
	{
		$<span>$ = Span{}
	}
|	','
	{
		$<span>$ = $<span>1
	}

// Special notations - should be created implicitly
// if we ever finish the yacc replacement.

initprefix_list:
	initprefix
	{
		$<span>$ = $<span>1
		$$ = []*Prefix{$1}
	}
|	initprefix_list initprefix
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2)
	}

tag_opt:
	{
		$<span>$ = Span{}
		$$ = ""
	}
|	tag
	{
		$<span>$ = $<span>1
		$$ = $1
	}

cexpr_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	cexpr
	{
		$<span>$ = $<span>1
		$$ = $1
	}

expr_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	expr
	{
		$<span>$ = $<span>1
		$$ = $1
	}

expr_list:
	expr
	{
		$<span>$ = $<span>1
		$$ = []*Expr{$1}
	}
|	expr_list ',' expr
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = append($1, $3)
	}

expr_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	expr_list
	{
		$<span>$ = $<span>1
		$$ = $1
	}

decl_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	decl_list_opt decl
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2...)
	}

label_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	label_list_opt label
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2)
	}

fnarg_list:
	fnarg
	{
		$<span>$ = $<span>1
		$$ = []*Decl{$1}
	}
|	fnarg_list ',' fnarg
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = append($1, $3)
	}

fnarg_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	fnarg_list
	{
		$<span>$ = $<span>1
		$$ = $1
	}

idecor_list:
	idecor
	{
		$<span>$ = $<span>1
		$$ = []idecor{$1}
	}
|	idecor_list ',' idecor
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = append($1, $3)
	}

idecor_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	idecor_list
	{
		$<span>$ = $<span>1
		$$ = $1
	}

qname_list:
	qname
	{
		$<span>$ = $<span>1
		$$ = []string{$1}
	}
|	qname_list qname
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2)
	}

qname_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	qname_list
	{
		$<span>$ = $<span>1
		$$ = $1
	}

cqname_list:
	cqname
	{
		$<span>$ = $<span>1
		$$ = []string{$1}
	}
|	cqname_list cqname
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2)
	}

cqname_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	cqname_list
	{
		$<span>$ = $<span>1
		$$ = $1
	}

cqtname_list:
	cqtname
	{
		$<span>$ = $<span>1
		$$ = []string{$1}
	}
|	cqtname_list cqtname
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2)
	}

cqtname_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	cqtname_list
	{
		$<span>$ = $<span>1
		$$ = $1
	}

sudecor_list:
	sudecor
	{
		$<span>$ = $<span>1
		$$ = nil
		$$ = append($$, $1)
	}
|	sudecor_list ',' sudecor
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = append($1, $3)
	}

sudecor_list_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	sudecor_list
	{
		$<span>$ = $<span>1
		$$ = $1
	}

sudecl_list:
	sudecl
	{
		$<span>$ = $<span>1
		$$ = $1
	}
|	sudecl_list sudecl
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2...)
	}

eqexpr_opt:
	{
		$<span>$ = Span{}
		$$ = nil
	}
|	eqexpr
	{
		$<span>$ = $<span>1
		$$ = $1
	}

edecl_list:
	edecl
	{
		$<span>$ = $<span>1
		$$ = []*Decl{$1}
	}
|	edecl_list ',' edecl
	{
		$<span>$ = span($<span>1, $<span>3)
		$$ = append($1, $3)
	}

string_list:
	tokString
	{
		$<span>$ = $<span>1
		$$ = []string{$1}
	}
|	string_list tokString
	{
		$<span>$ = span($<span>1, $<span>2)
		$$ = append($1, $2)
	}