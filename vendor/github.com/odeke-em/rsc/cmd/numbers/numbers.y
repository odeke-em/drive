%{
package main
%}

%union {
	expr *Expr
	exprlist []*Expr
	stmt *Stmt
	stmtlist []*Stmt
	tok int
	op Op
	line Line
}

%left	'<' _LE '>' _GE _EQ _NE
%left	'+' '-'
%left '*' '/'
%left _UNARYMINUS

%token	'=' _EQQ

%token	<expr>	_EXPR
%token	<line>	_MAX _MIN '-' _USE _PRINT _CHECK _COND _INCLUDE

%type	<stmtlist>	prog
%type	<stmt>	stmt
%type	<expr>	expr
%type	<op>	equal
%type	<exprlist>	exprlist

%%

top:
	prog
	{prog = append(prog, $1...)}

prog:
	{$$ = nil}
|	prog stmt
	{$$ = append($1, $2)}
|	prog '\n'
	{$$ = $1}
|	prog _INCLUDE
	{$$ = $1}

stmt:
	_EXPR equal expr '\n'
	{$$ = &Stmt{line: $1.line, left: $1, op: $2, right: $3}}
|	_USE '(' exprlist ')' '\n'
	{$$ = &Stmt{line: $1, op: opUse, list: $3}}
|	_PRINT '(' exprlist ')' '\n'
	{$$ = &Stmt{line: $1, op: opPrint, list: $3}}
|	_CHECK '(' exprlist ')' '\n'
	{$$ = &Stmt{line: $1, op: opCheck, list: $3}}

equal:
	'='
	{$$ = opAssign}
|	_EQQ
	{$$ = opAssignWeak}

expr:
	_EXPR
|	'(' expr ')'
	{$$ = $2}
|	expr '+' expr
	{$$ = &Expr{line: $1.line, op: opAdd, left: $1, right: $3}}
|	expr '-' expr
	{$$ = &Expr{line: $1.line, op: opSub, left: $1, right: $3}}
|	expr '*' expr
	{$$ = &Expr{line: $1.line, op: opMul, left: $1, right: $3}}
|	expr '/' expr
	{$$ = &Expr{line: $1.line, op: opDiv, left: $1, right: $3}}
|	expr '<' expr
	{$$ = &Expr{line: $1.line, op: opLess, left: $1, right: $3}}
|	expr _LE expr
	{$$ = &Expr{line: $1.line, op: opLessEqual, left: $1, right: $3}}
|	expr '>' expr
	{$$ = &Expr{line: $1.line, op: opGreater, left: $1, right: $3}}
|	expr _GE expr
	{$$ = &Expr{line: $1.line, op: opGreaterEqual, left: $1, right: $3}}
|	expr _EQ expr
	{$$ = &Expr{line: $1.line, op: opEqual, left: $1, right: $3}}
|	expr _NE expr
	{$$ = &Expr{line: $1.line, op: opNotEqual, left: $1, right: $3}}
|	'-' expr %prec _UNARYMINUS
	{$$ = &Expr{line: $1, op: opMinus, left: $2}}
|	_MAX '(' exprlist ')'
	{$$ = &Expr{line: $1, op: opMax, list: $3}}
|	_MIN '(' exprlist ')'
	{$$ = &Expr{line: $1, op: opMin, list: $3}}
|	_COND '(' expr ',' expr ',' expr ')'
	{$$ = &Expr{line: $1, op: opCond, list: []*Expr{$3, $5, $7}}}

exprlist:
	expr
	{$$ = []*Expr{$1}}
|	exprlist ',' expr
	{$$ = append($1, $3)}

%%
