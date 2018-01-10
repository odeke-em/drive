// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package c2go

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"code.google.com/p/rsc/cc"
)

type PrintSpecial int

const (
	Indent PrintSpecial = iota
	Unindent
	Untab
	Newline
)

// print an error; fprintf is a bad name but helps go vet.
func fprintf(span cc.Span, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s:%d: %s\n", span.Start.File, span.Start.Line, msg)
}

type Printer struct {
	Package string
	buf     bytes.Buffer
	indent  int
	html    bool

	printed map[interface{}]bool
	suffix  []cc.Comment // suffix comments to print at next newline
}

func (p *Printer) dup(x interface{}) bool {
	if p.printed[x] {
		return true
	}
	if p.printed == nil {
		p.printed = make(map[interface{}]bool)
	}
	p.printed[x] = true
	return false
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
	expr *cc.Expr
	prec int
}

type nestBlock struct {
	stmt *cc.Stmt
	more bool
}

type unnestBlock struct {
	stmt *cc.Stmt
}

type typedInit struct {
	typ  *cc.Type
	init *cc.Init
}

var htmlEscaper = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")

func GoString(args ...interface{}) string {
	var p Printer
	p.Print(args...)
	return p.buf.String()
}

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
		case *cc.Expr:
			p.printExpr(arg, precLow)
		case *cc.Prefix:
			p.printPrefix(arg)
		case *cc.Init:
			p.printInit(nil, arg)
		case typedInit:
			p.printInit(arg.typ, arg.init)
		case *cc.Prog:
			p.printProg(arg)
		case *cc.Stmt:
			p.printStmt(arg)
		case *cc.Type:
			p.printType(arg)
		case *cc.Decl:
			p.printDecl(arg)
		case cc.Storage:
			p.Print(arg.String())
		case []cc.Comment:
			for _, com := range arg {
				p.Print(com)
			}
		case cc.Comment:
			com := arg
			if com.Suffix {
				p.suffix = append(p.suffix, com)
			} else {
				for _, line := range strings.Split(com.Text, "\n") {
					p.Print(line, Newline)
				}
			}
		case nestBlock:
			if arg.stmt.Op == cc.Block {
				p.Print(" ", arg.stmt)
			} else {
				p.Print(" {", Indent, Newline, arg.stmt, Unindent, Newline, "}")
			}
		case unnestBlock:
			if arg.stmt.Op == cc.Block {
				for i, b := range arg.stmt.Block {
					if i > 0 {
						p.Print(Newline)
					}
					p.Print(b)
				}
			} else {
				p.Print(arg.stmt)
			}
		case PrintSpecial:
			switch arg {
			default:
				fmt.Fprintf(&p.buf, "(?special:%d)", arg)
			case Indent:
				p.indent++
			case Unindent:
				p.indent--
			case Untab:
				b := p.buf.Bytes()
				if len(b) > 0 && b[len(b)-1] == '\t' {
					p.buf.Truncate(len(b) - 1)
				}
			case Newline:
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
	precCmp
	precAndAnd
	precOrOr
	precComma
	precLow
)

var opPrec = map[cc.ExprOp]int{
	cc.Add:        precAdd,
	cc.AddEq:      precLow,
	cc.Addr:       precAddr,
	cc.And:        precMul,
	cc.AndAnd:     precAndAnd,
	cc.AndEq:      precLow,
	cc.Arrow:      precArrow,
	cc.Call:       precArrow,
	cc.Cast:       precAddr,
	cc.CastInit:   precAddr,
	cc.Comma:      precComma,
	cc.Cond:       precComma,
	cc.Div:        precMul,
	cc.DivEq:      precLow,
	cc.Dot:        precArrow,
	cc.Eq:         precLow,
	cc.EqEq:       precCmp,
	cc.Gt:         precCmp,
	cc.GtEq:       precCmp,
	cc.Index:      precArrow,
	cc.Indir:      precAddr,
	cc.Lsh:        precMul,
	cc.LshEq:      precLow,
	cc.Lt:         precCmp,
	cc.LtEq:       precCmp,
	cc.Minus:      precAddr,
	cc.Mod:        precMul,
	cc.ModEq:      precLow,
	cc.Mul:        precMul,
	cc.MulEq:      precLow,
	cc.Name:       precNone,
	cc.Not:        precAddr,
	cc.NotEq:      precCmp,
	cc.Number:     precNone,
	cc.Offsetof:   precAddr,
	cc.Or:         precAdd,
	cc.OrEq:       precLow,
	cc.OrOr:       precOrOr,
	cc.Paren:      precLow,
	cc.Plus:       precAddr,
	cc.PostDec:    precAddr,
	cc.PostInc:    precAddr,
	cc.PreDec:     precAddr,
	cc.PreInc:     precAddr,
	cc.Rsh:        precMul,
	cc.RshEq:      precLow,
	cc.SizeofExpr: precAddr,
	cc.SizeofType: precAddr,
	cc.String:     precNone,
	cc.Sub:        precAdd,
	cc.SubEq:      precLow,
	cc.Twid:       precAddr,
	cc.VaArg:      precAddr,
	cc.Xor:        precAdd,
	cc.XorEq:      precLow,

	AndNot:   precMul,
	AndNotEq: precLow,
	ColonEq:  precLow,
}

var opStr = map[cc.ExprOp]string{
	cc.Add:    "+",
	cc.AddEq:  "+=",
	cc.Addr:   "&",
	cc.And:    "&",
	cc.AndAnd: "&&",
	cc.AndEq:  "&=",
	cc.Div:    "/",
	cc.DivEq:  "/=",
	cc.Eq:     "=",
	cc.EqEq:   "==",
	cc.Gt:     ">",
	cc.GtEq:   ">=",
	cc.Indir:  "*",
	cc.Lsh:    "<<",
	cc.LshEq:  "<<=",
	cc.Lt:     "<",
	cc.LtEq:   "<=",
	cc.Minus:  "-",
	cc.Mod:    "%",
	cc.ModEq:  "%=",
	cc.Mul:    "*",
	cc.MulEq:  "*=",
	cc.Not:    "!",
	cc.NotEq:  "!=",
	cc.Or:     "|",
	cc.OrEq:   "|=",
	cc.OrOr:   "||",
	cc.Plus:   "+",
	cc.PreDec: "--",
	cc.PreInc: "++",
	cc.Rsh:    ">>",
	cc.RshEq:  ">>=",
	cc.Sub:    "-",
	cc.SubEq:  "-=",
	cc.Twid:   "^",
	cc.Xor:    "^",
	cc.XorEq:  "^=",

	AndNot:   "&^",
	AndNotEq: "&^=",
	ColonEq:  ":=",
}

const (
	ExprBlock cc.ExprOp = 100000 + iota
	ExprSlice
	AndNot
	AndNotEq
	ColonEq
)

func (p *Printer) printExpr(x *cc.Expr, prec int) {
	if x == nil {
		return
	}
	if p.html {
		fmt.Fprintf(&p.buf, "<span title='%s type %v'>", x.Op, x.XType)
		defer fmt.Fprintf(&p.buf, "</span>")
	}

	p.Print(x.Comments.Before)
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	newPrec := opPrec[x.Op]
	if prec < newPrec {
		p.Print("(")
		defer p.Print(")")
	}
	prec = newPrec

	str := opStr[x.Op]
	if str != "" {
		if x.Right != nil {
			// binary operator
			// left associative
			p.Print(exprPrec{x.Left, prec}, " ", str, " ", exprPrec{x.Right, prec - 1})
		} else {
			// unary operator
			if (x.Op == cc.Plus || x.Op == cc.Minus || x.Op == cc.Addr) && x.Left.Op == x.Op ||
				x.Op == cc.Plus && x.Left.Op == cc.PreInc ||
				x.Op == cc.Minus && x.Left.Op == cc.PreDec {
				prec-- // force parenthesization +(+x) not ++x
			}
			p.Print(str, exprPrec{x.Left, prec})
		}
		return
	}

	// special cases
	switch x.Op {
	default:
		panic(fmt.Sprintf("printExpr missing case for %v", x.Op))

	case ExprBlock:
		if len(x.Block) == 0 {
			break
		}
		p.Print("(func(){")
		for i, stmt := range x.Block {
			if i > 0 {
				p.Print("; ")
			}
			p.Print(stmt)
		}
		p.Print("})()")

	case ExprSlice:
		p.Print(x.List[0], "[")
		if x.List[1] != nil {
			p.Print(x.List[1])
		}
		p.Print(":")
		if x.List[2] != nil {
			p.Print(x.List[2])
		}
		p.Print("]")

	case cc.Arrow:
		name := x.Text
		if x.XDecl != nil {
			name = x.XDecl.Name
		}
		p.Print(exprPrec{x.Left, prec}, ".", name)

	case cc.Call:
		p.Print(exprPrec{x.Left, precAddr}, "(")
		for i, y := range x.List {
			if i > 0 {
				p.Print(", ")
			}
			p.printExpr(y, precComma)
		}
		p.Print(")")

	case cc.Cast:
		if x.Type.Kind == cc.Ptr || x.Type.Kind == cc.Func {
			p.Print("(", x.Type, ")(", exprPrec{x.Left, precLow}, ")")
		} else {
			p.Print(x.Type, "(", exprPrec{x.Left, precLow}, ")")
		}

	case cc.CastInit:
		if x.Type.Kind == cc.Ptr || x.Type.Kind == cc.Func {
			p.Print("(", x.Type, ")", x.Init)
		} else {
			p.Print(x.Type, x.Init)
		}

	case cc.Comma:
		if len(x.List) == 0 {
			break
		}
		p.Print("(func() {")
		for i, y := range x.List {
			if i > 0 {
				p.Print("; ")
			}
			p.printExpr(y, precLow)
		}
		p.Print("}())")

	case cc.Cond:
		p.Print("TERNARY(", exprPrec{x.List[0], prec - 1}, ", ", exprPrec{x.List[1], prec}, ", ", exprPrec{x.List[2], prec}, ")")

	case cc.Dot:
		name := x.XDecl.Name
		p.Print(exprPrec{x.Left, prec}, ".", name)

	case cc.Index:
		p.Print(exprPrec{x.Left, prec}, "[", exprPrec{x.Right, precLow}, "]")

	case cc.Name:
		name := x.Text
		if x.XDecl != nil {
			name = x.XDecl.Name
			if x.XDecl.GoPackage != "" && p.Package != "" && x.XDecl.GoPackage != p.Package {
				name = path.Base(x.XDecl.GoPackage) + "." + name
			}
		}
		p.Print(name)

	case cc.Number:
		p.Print(strings.TrimRight(x.Text, "LlUu"))

	case cc.SizeofExpr:
		p.Print("sizeof(", x.Left, ")")

	case cc.String:
		for i, str := range x.Texts {
			if i > 0 {
				p.Print(" + ")
			}
			p.Print(str)
		}

	case cc.Offsetof:
		p.Print("offsetof(", x.Type, ", ", exprPrec{x.Left, precComma}, ")")

	case cc.Paren:
		p.Print(exprPrec{x.Left, prec})

	case cc.PostDec:
		p.Print(exprPrec{x.Left, prec}, "--")

	case cc.PostInc:
		p.Print(exprPrec{x.Left, prec}, "++")

	case cc.SizeofType:
		p.Print("sizeof(", x.Type, ")")

	case cc.VaArg:
		p.Print("va_arg(", exprPrec{x.Left, precComma}, ", ", x.Type, ")")
	}
}

func (p *Printer) printPrefix(x *cc.Prefix) {
	if x.Dot != "" {
		p.Print(x.XDecl.Name, ": ")
	} else {
		p.Print(x.Index, ": ")
	}
}

func (p *Printer) printInit(typ *cc.Type, x *cc.Init) {
	p.Print(x.Comments.Before)
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	if len(x.Prefix) > 0 {
		for _, pre := range x.Prefix {
			p.Print(pre)
		}
	}
	if x.Expr != nil {
		if x.Expr.Op == cc.Number && (typ.Is(cc.Ptr) || typ.Is(Slice)) {
			p.Print("nil")
			return
		}
		p.printExpr(x.Expr, precComma)
		return
	}

	nl := len(x.Braced) > 0 && x.Braced[0].Span.Start.Line != x.Braced[len(x.Braced)-1].Span.End.Line
	if typ != nil {
		p.printType(typ)
	}
	p.Print("{")
	if nl {
		p.Print(Indent)
	}
	warned := false
	for i, y := range x.Braced {
		if nl {
			p.Print(Newline)
		} else if i > 0 {
			p.Print(" ")
		}
		var subtyp *cc.Type
		if typ != nil {
			if typ.Is(cc.Struct) && i < len(typ.Def().Decls) && len(y.Prefix) == 0 {
				subtyp = typ.Def().Decls[i].Type
			} else if typ.Is(cc.Struct) && len(y.Prefix) == 1 && y.Prefix[0].XDecl != nil {
				subtyp = y.Prefix[0].XDecl.Type
			} else if typ.Is(cc.Array) || typ.Is(Slice) {
				subtyp = typ.Def().Base
			} else if !warned {
				warned = true
				fprintf(x.Span, "too many fields in braced initializer of %s", GoString(typ))
			}
		}
		p.printInit(subtyp, y)
		p.Print(",")
	}
	if typ != nil && typ.Is(cc.Struct) && len(x.Braced) > 0 && len(x.Braced[0].Prefix) == 0 && len(x.Braced) < len(typ.Def().Decls) {
		for i := len(x.Braced); i < len(typ.Def().Decls); i++ {
			subtyp := typ.Def().Decls[i].Type
			if subtyp.Is(cc.Ptr) || subtyp.Is(Slice) {
				p.Print(" nil,")
			} else if subtyp.Is(cc.Array) {
				p.Print(" ", subtyp, "{},")
			} else {
				p.Print(" 0,")
			}
		}
	}
	if nl {
		p.Print(Unindent, Newline)
	}
	p.Print("}")
}

func (p *Printer) printProg(x *cc.Prog) {
	p.Print(x.Comments.Before)
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	for _, decl := range x.Decls {
		p.Print(decl, Newline)
	}
}

const (
	BlockNoBrace cc.StmtOp = 100000 + iota
	ForRange
)

func (p *Printer) printStmt(x *cc.Stmt) {
	if len(x.Labels) > 0 {
		p.Print(Untab, Unindent, x.Comments.Before, Indent, "\t")
		for i := 0; i < len(x.Labels); i++ {
			lab := x.Labels[i]
			p.Print(Untab, Unindent, lab.Comments.Before, Indent, "\t")
			p.Print(Untab)
			switch lab.Op {
			case cc.LabelName:
				p.Print(lab.Name)
			case cc.Case:
				p.Print("case ", lab.Expr)
				for i+1 < len(x.Labels) && x.Labels[i+1].Op == cc.Case {
					p.Print(", ", lab.Comments.Suffix, Newline)
					i++
					lab = x.Labels[i]
					p.Print(lab.Comments.Before, lab.Expr)
				}
			case cc.Default:
				p.Print("default")
			}
			p.Print(":", lab.Comments.Suffix, Newline)
		}
	} else {
		p.Print(x.Comments.Before)
	}
	defer p.Print(x.Comments.Suffix, x.Comments.After)

	switch x.Op {
	case cc.ARGBEGIN:
		p.Print("ARGBEGIN{", Indent, Newline, x.Body, Unindent, Newline, "}ARGEND")

	case cc.Block:
		p.Print("{", Indent)
		for _, b := range x.Block {
			p.Print(Newline, b)
		}
		p.Print(Unindent, Newline, "}")

	case BlockNoBrace:
		for i, b := range x.Block {
			if i > 0 {
				p.Print(Newline)
			}
			p.Print(b)
		}

	case cc.Break:
		p.Print("break")

	case cc.Continue:
		p.Print("continue")

	case cc.Do:
		p.Print("for ")
		if x.Pre != nil {
			p.Print(x.Pre, "; ; ", x.Pre, " ")
		}
		p.Print("{", Indent, Newline, unnestBlock{x.Body}, Newline, "if ", &cc.Expr{Op: cc.Not, Left: x.Expr}, " {", Indent, Newline, "break", Unindent, Newline, "}", Unindent, Newline, "}")

	case cc.Empty:
		// ok

	case cc.For:
		p.Print("for ", x.Pre, "; ", x.Expr, "; ", x.Post, nestBlock{x.Body, false})

	case ForRange:
		p.Print("for ", x.Pre, " = range ", x.Post, nestBlock{x.Body, false})

	case cc.If:
		p.Print("if ")
		if x.Pre != nil {
			p.Print(x.Pre, "; ")
		}
		p.Print(x.Expr, nestBlock{x.Body, x.Else != nil})
		if x.Else != nil {
			if x.Else.Op == cc.If {
				p.Print(" else ", x.Else)
			} else {
				p.Print(" else", nestBlock{x.Else, false})
			}
		}

	case cc.Goto:
		p.Print("goto ", x.Text)

	case cc.Return:
		if x.Expr == nil {
			p.Print("return")
		} else {
			p.Print("return ", x.Expr)
		}

	case cc.StmtDecl:
		p.Print(x.Decl)

	case cc.StmtExpr:
		p.Print(x.Expr)

	case cc.Switch:
		p.Print("switch ", x.Expr, nestBlock{x.Body, false})

	case cc.While:
		p.Print("for ")
		if x.Pre != nil {
			p.Print(x.Pre, "; ")
		}
		p.Print(x.Expr)
		if x.Pre != nil {
			p.Print("; ", x.Pre, " ")
		}
		p.Print(nestBlock{x.Body, false})
	}
}

const (
	Bool cc.TypeKind = 100000 + iota
	Int8
	Uint8
	Byte
	Int16
	Uint16
	Int
	Uint
	Int32
	Rune
	Uint32
	Uintptr
	Int64
	Uint64
	Float32
	Float64
	Ideal
	String
	Slice
)

func (p *Printer) printType(t *cc.Type) {
	if t == nil {
		p.Print("<nil type>")
		return
	}

	// Shouldn't happen but handle in case it does.
	p.Print(t.Comments.Before)
	defer p.Print(t.Comments.Suffix, t.Comments.After)

	if t == cc.BoolType {
		p.Print("bool")
		return
	}
	if typemap[t.Kind] != "" {
		p.Print(typemap[t.Kind])
		return
	}

	switch t.Kind {
	default:
		p.Print("C.", t.String()) // hope for the best

	case Slice:
		p.Print("[]", t.Base)

	case String:
		p.Print("string")

	case cc.Struct:
		if len(t.Decls) == 0 {
			p.Print("struct{}")
			break
		}
		p.Print("struct {", Indent)
		p.printStructBody(t)
		p.Print(Unindent, Newline, "}")

	case cc.Enum:
		if t.Tag != "" {
			p.Print(t.Tag)
		} else {
			p.Print("int")
		}

	case cc.TypedefType:
		if t.Base != nil && typemap[t.Base.Kind] != "" && strings.ToLower(t.Name) == t.Name {
			p.Print(typemap[t.Base.Kind])
			return
		}
		if t.TypeDecl != nil && t.TypeDecl.GoPackage != "" && p.Package != "" && t.TypeDecl.GoPackage != p.Package {
			p.Print(path.Base(t.TypeDecl.GoPackage) + "." + t.Name)
			break
		}
		p.Print(t.Name)

	case cc.Ptr:
		if t.Base.Is(cc.Func) {
			p.Print(t.Base)
			return
		}
		if t.Base.Is(cc.Void) {
			p.Print("*[0]byte")
			return
		}
		p.Print("*", t.Base)

	case cc.Func:
		p.Print("func(")
		for i, arg := range t.Decls {
			if i > 0 {
				p.Print(", ")
			}
			if arg.Name == "..." {
				p.Print("...interface{}")
				continue
			}
			if arg.Name == "" && arg.Type.Is(cc.Void) {
				continue
			}
			p.Print(arg.Type)
		}
		p.Print(")")
		if !t.Base.Is(cc.Void) {
			p.Print(" ", t.Base)
		}

	case cc.Array:
		if t.Width == nil {
			p.Print("[XXX]", t.Base)
			return
		}
		p.Print("[", t.Width, "]", t.Base)
	}
}

var typemap = map[cc.TypeKind]string{
	Bool:    "bool",
	Int8:    "int8",
	Uint8:   "uint8",
	Int16:   "int16",
	Uint16:  "uint16",
	Rune:    "rune",
	Byte:    "byte",
	Int:     "int",
	Uint:    "uint",
	Uintptr: "uintptr",
	Int32:   "int32",
	Uint32:  "uint32",
	Int64:   "int64",
	Uint64:  "uint64",
	Float32: "float32",
	Float64: "float64",
}

func (p *Printer) oldprintDecl(x *cc.Decl) {
	if x.Storage != 0 {
		p.Print(x.Storage, " ")
	}
	if x.Type == nil || x.Init != nil && len(x.Init.Braced) > 0 {
		p.Print(x.Name)
	} else {
		name := x.Name
		if x.Type.Kind == cc.Func && x.Body != nil {
			name = "\n" + name
		}
		p.Print(cc.TypedName{x.Type, name})
		if x.Name == "" {
			switch x.Type.Kind {
			case cc.Struct, cc.Union, cc.Enum:
				p.Print(" {", Indent)
				for _, decl := range x.Type.Decls {
					p.Print(Newline, decl)
				}
				p.Print(Unindent, Newline, "}")
			}
		}
	}
	if x.Init != nil {
		p.Print(" = ", typedInit{x.Type, x.Init})
	}
	if x.Body != nil {
		p.Print(Newline, x.Body)
	}
}

func (p *Printer) printDecl(decl *cc.Decl) {
	if p.dup(decl) {
		return
	}

	p.Print(decl.Comments.Before)
	defer p.Print(decl.Comments.Suffix, decl.Comments.After)

	t := decl.Type
	if decl.Storage&cc.Typedef != 0 {
		if t.Kind == cc.Struct || t.Kind == cc.Union || t.Kind == cc.Union {
			if t.Tag == "" {
				t.Tag = decl.Name
			} else if decl.Name != t.Tag {
				fprintf(decl.Span, "typedef %s and tag %s do not match", decl.Name, t.Tag)
			}
			if t.Kind == cc.Enum {
				p.printEnumDecl(t)
			} else {
				p.printStructDecl(t)
			}
			return
		}
		p.Print("type ", decl.Name, " ", decl.Type)
		return
	}

	if decl.Name == "" {
		switch t.Kind {
		case cc.Struct, cc.Union:
			p.printStructDecl(t)
			return
		case cc.Enum:
			p.printEnumDecl(t)
			return
		}
		if Bool <= t.Kind && t.Kind <= Float64 && t.Decls != nil {
			// was an enum
			p.printEnumDecl(t)
			return
		}
		fprintf(decl.Span, "empty declaration of type %s", GoString(t))
		return
	}

	if t.Kind == cc.Func {
		p.printFuncDecl(decl)
		return
	}

	if decl.Init != nil && len(decl.Init.Braced) > 0 {
		p.Print("var ", decl.Name, " = ", typedInit{decl.Type, decl.Init})
		return
	}

	p.Print("var ", decl.Name, " ", decl.Type)
	if decl.Init != nil {
		p.Print(" = ", decl.Init)
	}
}

func (p *Printer) printStructDecl(t *cc.Type) {
	if p.dup(t) {
		return
	}
	if t.Kind == cc.Union {
		fprintf(t.Span, "cannot convert union")
		return
	}
	p.Print("type ", t.Tag, " struct {", Indent)
	p.printStructBody(t)
	p.Print(Unindent, Newline, "}")
}

func (p *Printer) printStructBody(t *cc.Type) {
	for _, decl := range t.Decls {
		if decl.Name == "" {
			// Hope this is a struct definition.
			if decl.Type.Kind != cc.Struct {
				fprintf(decl.Span, "unnamed non-struct field of type %v", decl.Type)
				continue
			}
			p.printStructBody(decl.Type)
			continue
		}
		p.Print(Newline, decl.Name, " ", decl.Type)
	}
}

func (p *Printer) printEnumDecl(t *cc.Type) {
	typeSuffix := ""
	if t.Tag != "" {
		typeSuffix = " " + t.Tag
		p.Print("type ", t.Tag, " int", Newline)
		//	fprintf(t.Span, "cannot handle enum tags")
		//	return
	}
	p.Print("const (", Indent)
	for i, decl := range t.Decls {
		p.Print(Newline, decl.Name)
		if decl.Init == nil && i == 0 {
			if len(t.Decls) >= 2 && t.Decls[1].Init == nil {
				p.Print(typeSuffix, " = iota")
			} else {
				p.Print(typeSuffix, " = 0")
			}
		} else if decl.Init != nil {
			p.Print(typeSuffix, " = ", decl.Init.Expr)
			if i+1 < len(t.Decls) && t.Decls[i+1].Init == nil {
				p.Print(" + iota")
				if i > 0 {
					p.Print(fmt.Sprintf("-%d", i))
				}
			}
		}
	}
	p.Print(Unindent, Newline, ")")
}

func (p *Printer) printFuncDecl(decl *cc.Decl) {
	if decl.Body == nil {
		// wait for definition
		return
	}
	p.Print("func ", decl.Name, "(")
	for i, arg := range decl.Type.Decls {
		if arg.Type.Is(cc.Void) {
			continue
		}
		if i > 0 {
			p.Print(", ")
		}
		if arg.Name == "..." {
			p.Print("args ...interface{}")
			continue
		}
		p.Print(arg.Name, " ", arg.Type)
	}
	p.Print(")")
	if !decl.Type.Base.Is(cc.Void) {
		p.Print(" ", decl.Type.Base)
	}
	p.Print(" ", decl.Body, Newline)
}
