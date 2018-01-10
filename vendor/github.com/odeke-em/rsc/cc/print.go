// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

import (
	"bytes"
	"fmt"
	"strings"
)

type special int

const (
	indent special = iota
	unindent
	untab
	newline
)

type Printer struct {
	buf          bytes.Buffer
	indent       int
	html         bool
	suffix       []Comment // suffix comments to print at next newline
	hideComments bool
}

func (p *Printer) StartHTML() {
	p.buf.WriteString("<pre>")
	p.html = true
}

func (p *Printer) EndHTML() {
	p.buf.WriteString("</pre>")
}

func (p *Printer) Bytes() []byte {
	return p.buf.Bytes()
}

func (p *Printer) String() string {
	return p.buf.String()
}

type exprPrec struct {
	expr *Expr
	prec int
}

type nestBlock struct {
	stmt *Stmt
	more bool
}

type TypedName struct {
	Type *Type
	Name string
}

var htmlEscaper = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")

func (p *Printer) Print(args ...interface{}) {
	for _, arg := range args {
		switch arg := arg.(type) {
		default:
			fmt.Fprintf(&p.buf, "(?%T)", arg)
		case string:
			if p.html {
				htmlEscaper.WriteString(&p.buf, arg)
			} else {
				p.buf.WriteString(arg)
			}
		case exprPrec:
			p.printExpr(arg.expr, arg.prec)
		case *Expr:
			p.printExpr(arg, precLow)
		case *Prefix:
			p.printPrefix(arg)
		case *Init:
			p.printInit(arg)
		case *Prog:
			p.printProg(arg)
		case *Stmt:
			p.printStmt(arg)
		case *Type:
			p.printType(arg, "")
		case *Decl:
			p.printDecl(arg)
		case TypedName:
			p.printType(arg.Type, arg.Name)
		case Storage:
			p.Print(arg.String())
		case []Comment:
			for _, com := range arg {
				p.Print(com)
			}
		case Comment:
			if p.hideComments {
				break
			}
			com := arg
			if com.Suffix {
				p.suffix = append(p.suffix, com)
			} else {
				for _, line := range strings.Split(com.Text, "\n") {
					p.Print(line, newline)
				}
			}
		case nestBlock:
			if arg.stmt.Op == Block {
				p.Print(" ", arg.stmt)
			} else {
				p.Print(indent, newline, arg.stmt, unindent)
				if arg.more {
					p.Print(newline)
				}
			}
		case special:
			switch arg {
			default:
				fmt.Fprintf(&p.buf, "(?special:%d)", arg)
			case indent:
				p.indent++
			case unindent:
				p.indent--
			case untab:
				b := p.buf.Bytes()
				if len(b) > 0 && b[len(b)-1] == '\t' {
					p.buf.Truncate(len(b) - 1)
				}
			case newline:
				for _, com := range p.suffix {
					p.Print(" ", com.Text)
				}
				p.suffix = p.suffix[:0]
				p.buf.WriteString("\n")
				for i := 0; i < p.indent; i++ {
					p.buf.WriteByte('\t')
				}
			}
		}
	}
}

const (
	precNone = iota
	precArrow
	precAddr
	precMul
	precAdd
	precLsh
	precLt
	precEqEq
	precAnd
	precXor
	precOr
	precAndAnd
	precOrOr
	precCond
	precEq
	precComma
	precLow
)

var opPrec = []int{
	Add:        precAdd,
	AddEq:      precEq,
	Addr:       precAddr,
	And:        precAnd,
	AndAnd:     precAndAnd,
	AndEq:      precEq,
	Arrow:      precArrow,
	Call:       precArrow,
	Cast:       precAddr,
	CastInit:   precAddr,
	Comma:      precComma,
	Cond:       precCond,
	Div:        precMul,
	DivEq:      precEq,
	Dot:        precArrow,
	Eq:         precEq,
	EqEq:       precEqEq,
	Gt:         precLt,
	GtEq:       precLt,
	Index:      precArrow,
	Indir:      precAddr,
	Lsh:        precLsh,
	LshEq:      precEq,
	Lt:         precLt,
	LtEq:       precLt,
	Minus:      precAddr,
	Mod:        precMul,
	ModEq:      precEq,
	Mul:        precMul,
	MulEq:      precEq,
	Name:       precNone,
	Not:        precAddr,
	NotEq:      precEqEq,
	Number:     precNone,
	Offsetof:   precAddr,
	Or:         precOr,
	OrEq:       precEq,
	OrOr:       precOrOr,
	Paren:      precLow,
	Plus:       precAddr,
	PostDec:    precAddr,
	PostInc:    precAddr,
	PreDec:     precAddr,
	PreInc:     precAddr,
	Rsh:        precLsh,
	RshEq:      precEq,
	SizeofExpr: precAddr,
	SizeofType: precAddr,
	String:     precNone,
	Sub:        precAdd,
	SubEq:      precEq,
	Twid:       precAddr,
	VaArg:      precAddr,
	Xor:        precXor,
	XorEq:      precEq,
}

var opStr = []string{
	Add:        "+",
	AddEq:      "+=",
	Addr:       "&",
	And:        "&",
	AndAnd:     "&&",
	AndEq:      "&=",
	Div:        "/",
	DivEq:      "/=",
	Eq:         "=",
	EqEq:       "==",
	Gt:         ">",
	GtEq:       ">=",
	Indir:      "*",
	Lsh:        "<<",
	LshEq:      "<<=",
	Lt:         "<",
	LtEq:       "<=",
	Minus:      "-",
	Mod:        "%",
	ModEq:      "%=",
	Mul:        "*",
	MulEq:      "*=",
	Not:        "!",
	NotEq:      "!=",
	Or:         "|",
	OrEq:       "|=",
	OrOr:       "||",
	Plus:       "+",
	PreDec:     "--",
	PreInc:     "++",
	Rsh:        ">>",
	RshEq:      ">>=",
	Sub:        "-",
	SubEq:      "-=",
	Twid:       "~",
	Xor:        "^",
	XorEq:      "^=",
	SizeofExpr: "sizeof ",
}

func (p *Printer) printExpr(x *Expr, prec int) {
	if x == nil {
		return
	}
	if p.html {
		fmt.Fprintf(&p.buf, "<span title='%s type %v'>", x.Op, x.XType)
		defer fmt.Fprintf(&p.buf, "</span>")
	}

	p.Print(x.Comments.Before)
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	var newPrec int
	if 0 <= int(x.Op) && int(x.Op) < len(opPrec) {
		newPrec = opPrec[x.Op]
	}
	if prec < newPrec {
		p.Print("(")
		defer p.Print(")")
	}
	prec = newPrec

	var str string
	if 0 <= int(x.Op) && int(x.Op) < len(opStr) {
		str = opStr[x.Op]
	}
	if str != "" {
		if x.Right != nil {
			// binary operator
			if prec == precEq {
				// right associative
				p.Print(exprPrec{x.Left, prec - 1}, " ", str, " ", exprPrec{x.Right, prec})
			} else {
				// left associative
				p.Print(exprPrec{x.Left, prec}, " ", str, " ", exprPrec{x.Right, prec - 1})
			}
		} else {
			// unary operator
			if (x.Op == Plus || x.Op == Minus || x.Op == Addr) && x.Left.Op == x.Op ||
				x.Op == Plus && x.Left.Op == PreInc ||
				x.Op == Minus && x.Left.Op == PreDec {
				prec-- // force parenthesization +(+x) not ++x
			}
			p.Print(str, exprPrec{x.Left, prec})
		}
		return
	}

	// special cases
	switch x.Op {
	default:
		p.Print(fmt.Sprintf("Expr(Op=%d)", x.Op))

	case Arrow:
		p.Print(exprPrec{x.Left, prec}, "->", x.Text)

	case Call:
		p.Print(exprPrec{x.Left, precAddr}, "(")
		for i, y := range x.List {
			if i > 0 {
				p.Print(", ")
			}
			p.printExpr(y, precComma)
		}
		p.Print(")")

	case Cast:
		p.Print("(", x.Type, ")", exprPrec{x.Left, prec})

	case CastInit:
		p.Print("(", x.Type, ")", x.Init)

	case Comma:
		for i, y := range x.List {
			if i > 0 {
				p.Print(", ")
			}
			p.printExpr(y, prec-1)
		}

	case Cond:
		p.Print(exprPrec{x.List[0], prec - 1}, " ? ", exprPrec{x.List[1], prec}, " : ", exprPrec{x.List[2], prec})

	case Dot:
		p.Print(exprPrec{x.Left, prec}, ".", x.Text)

	case Index:
		p.Print(exprPrec{x.Left, prec}, "[", exprPrec{x.Right, precLow}, "]")

	case Name, Number:
		p.Print(x.Text)

	case String:
		for i, str := range x.Texts {
			if i > 0 {
				p.Print(" ")
			}
			p.Print(str)
		}

	case Offsetof:
		p.Print("offsetof(", x.Type, ", ", exprPrec{x.Left, precComma}, ")")

	case Paren:
		p.Print("(", exprPrec{x.Left, prec}, ")")

	case PostDec:
		p.Print(exprPrec{x.Left, prec}, "--")

	case PostInc:
		p.Print(exprPrec{x.Left, prec}, "++")

	case SizeofType:
		p.Print("sizeof(", x.Type, ")")

	case VaArg:
		p.Print("va_arg(", exprPrec{x.Left, precComma}, ", ", x.Type, ")")
	}
}

func (p *Printer) printPrefix(x *Prefix) {
	if x.Dot != "" {
		p.Print(".", x.Dot)
	} else {
		p.Print("[", x.Index, "]")
	}
}

func (p *Printer) printInit(x *Init) {
	p.Print(x.Comments.Before)
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	if len(x.Prefix) > 0 {
		for _, pre := range x.Prefix {
			p.Print(pre)
		}
		p.Print(" = ")
	}
	if x.Expr != nil {
		p.printExpr(x.Expr, precComma)
	} else {
		nl := len(x.Braced) > 0 && x.Braced[0].Span.Start.Line != x.Braced[len(x.Braced)-1].Span.End.Line
		p.Print("{")
		if nl {
			p.Print(indent)
		}
		for i, y := range x.Braced {
			if i > 0 {
				p.Print(",")
			}
			if nl {
				p.Print(newline)
			} else if i > 0 {
				p.Print(" ")
			}
			p.Print(y)
		}
		if nl {
			p.Print(unindent, newline)
		}
		p.Print("}")
	}

	for _, com := range x.Comments.After {
		p.Print(com)
	}
}

func (p *Printer) printProg(x *Prog) {
	p.Print(x.Comments.Before)
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	for _, decl := range x.Decls {
		p.Print(decl, newline)
	}
}

func (p *Printer) printStmt(x *Stmt) {
	if len(x.Labels) > 0 {
		p.Print(untab, unindent, x.Comments.Before, indent, "\t")
		for _, lab := range x.Labels {
			p.Print(untab, unindent, lab.Comments.Before, indent, "\t")
			p.Print(untab)
			switch {
			case lab.Name != "":
				p.Print(lab.Name)
			case lab.Expr != nil:
				p.Print("case ", lab.Expr)
			default:
				p.Print("default")
			}
			p.Print(":", lab.Comments.Suffix, newline)
		}
	} else {
		p.Print(x.Comments.Before)
	}
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	switch x.Op {
	case ARGBEGIN:
		p.Print("ARGBEGIN{", indent, newline, x.Body, unindent, newline, "}ARGEND")

	case Block:
		p.Print("{", indent)
		for _, b := range x.Block {
			p.Print(newline, b)
		}
		p.Print(unindent, newline, "}")

	case Break:
		p.Print("break;")

	case Continue:
		p.Print("continue;")

	case Do:
		p.Print("do", nestBlock{x.Body, true}, " while(", x.Expr, ");")

	case Empty:
		p.Print(";")

	case For:
		p.Print("for(", x.Pre, ";")
		if x.Expr != nil {
			p.Print(" ")
		}
		p.Print(x.Expr, ";")
		if x.Post != nil {
			p.Print(" ")
		}
		p.Print(x.Post, ")", nestBlock{x.Body, false})

	case If:
		p.Print("if(", x.Expr, ")", nestBlock{x.Body, x.Else != nil})
		if x.Else != nil {
			if x.Body.Op == Block {
				p.Print(" ")
			}
			p.Print("else", nestBlock{x.Else, false})
		}

	case Goto:
		p.Print("goto ", x.Text, ";")

	case Return:
		if x.Expr == nil {
			p.Print("return;")
		} else {
			p.Print("return ", x.Expr, ";")
		}

	case StmtDecl:
		p.Print(x.Decl, ";")

	case StmtExpr:
		p.Print(x.Expr, ";")

	case Switch:
		p.Print("switch(", x.Expr, ")", nestBlock{x.Body, false})

	case While:
		p.Print("while(", x.Expr, ")", nestBlock{x.Body, false})
	}
}

func (p *Printer) printType(x *Type, name string) {
	// Shouldn't happen but handle in case it does.
	p.Print(x.Comments.Before)
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	switch x.Kind {
	case Ptr:
		p.printType(x.Base, "*"+name)
	case Array:
		if strings.HasPrefix(name, "*") {
			name = "(" + name + ")"
		}
		if x.Width == nil {
			p.printType(x.Base, name+"[]")
		} else {
			p.printType(x.Base, name+"["+x.Width.String()+"]")
		}
	case Func:
		var pp Printer
		if strings.HasPrefix(name, "*") {
			name = "(" + name + ")"
		}
		pp.Print(name, "(")
		for i, decl := range x.Decls {
			if i > 0 {
				pp.Print(", ")
			}
			pp.Print(decl)
		}
		pp.Print(")")
		p.printType(x.Base, pp.String())

	default:
		p.Print(x.String())
		i := 0
		for i < len(name) && name[i] == '*' {
			i++
		}
		if i < len(name) && name[i] != '\n' {
			p.Print(" ")
		}
		p.Print(name)
	}
}

func (p *Printer) printDecl(x *Decl) {
	p.Print(x.Comments.Before)
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	if x.Storage != 0 {
		p.Print(x.Storage, " ")
	}
	if x.Type == nil {
		p.Print(x.Name)
	} else {
		name := x.Name
		if x.Type.Kind == Func && x.Body != nil {
			name = "\n" + name
		}
		p.Print(TypedName{x.Type, name})
		if x.Name == "" {
			switch x.Type.Kind {
			case Struct, Union, Enum:
				p.Print(" {", indent)
				for _, decl := range x.Type.Decls {
					p.Print(newline, decl)
				}
				p.Print(unindent, newline, "}")
			}
		}
	}
	if x.Init != nil {
		p.Print(" = ", x.Init)
	}
	if x.Body != nil {
		p.Print(newline, x.Body)
	}
}
