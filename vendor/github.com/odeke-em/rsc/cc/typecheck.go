// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Scoping and type checking.
// C99 standard: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1124.pdf

package cc

import (
	"fmt"
	"strconv"
	"strings"
)

type Scope struct {
	Decl map[string]*Decl
	Tag  map[string]*Type
	Next *Scope
}

func (lx *lexer) pushDecl(decl *Decl) {
	sc := lx.scope
	if sc == nil {
		panic("no scope")
	}
	if decl.Name == "" {
		return
	}
	if sc.Decl == nil {
		sc.Decl = make(map[string]*Decl)
	}
	sc.Decl[decl.Name] = decl
	if hdr := lx.declSave; hdr != nil && sc.Next == nil {
		hdr.decls = append(hdr.decls, decl)
	}
}

func (lx *lexer) lookupDecl(name string) *Decl {
	for sc := lx.scope; sc != nil; sc = sc.Next {
		decl := sc.Decl[name]
		if decl != nil {
			return decl
		}
	}
	return nil
}

func (lx *lexer) pushType(typ *Type) *Type {
	sc := lx.scope
	if sc == nil {
		panic("no scope")
	}

	if typ.Kind == Enum && typ.Decls != nil {
		for _, decl := range typ.Decls {
			lx.pushDecl(decl)
		}
	}

	if typ.Tag == "" {
		return typ
	}

	old := lx.lookupTag(typ.Tag)
	if old == nil {
		if sc.Tag == nil {
			sc.Tag = make(map[string]*Type)
		}
		sc.Tag[typ.Tag] = typ
		if hdr := lx.declSave; hdr != nil && sc.Next == nil {
			hdr.types = append(hdr.types, typ)
		}
		return typ
	}

	// merge typ into old
	if old.Kind != typ.Kind {
		lx.Errorf("conflicting tags: %s %s and %s %s", old.Kind, old.Tag, typ.Kind, typ.Tag)
		return typ
	}
	if typ.Decls != nil {
		if old.Decls != nil {
			lx.Errorf("multiple definitions for %s %s", old.Kind, old.Tag)
		}
		old.SyntaxInfo = typ.SyntaxInfo
		old.Decls = typ.Decls
	}
	return old
}

func (lx *lexer) lookupTag(name string) *Type {
	for sc := lx.scope; sc != nil; sc = sc.Next {
		typ := sc.Tag[name]
		if typ != nil {
			return typ
		}
	}
	return nil
}

func (lx *lexer) pushScope() {
	sc := &Scope{Next: lx.scope}
	lx.scope = sc
}

func (lx *lexer) popScope() {
	lx.scope = lx.scope.Next
}

func (lx *lexer) typecheck(prog *Prog) {
	for _, decl := range prog.Decls {
		lx.typecheckDecl(decl)
	}
}

func (lx *lexer) typecheckDecl(decl *Decl) {
	lx.typecheckType(decl.Type)
	if decl.Init != nil {
		lx.typecheckInit(decl.Type, decl.Init)
	}
	lx.typecheckStmt(decl.Body)
}

func (lx *lexer) typecheckStmt(stmt *Stmt) {
	if stmt == nil {
		return
	}

	lx.setSpan(stmt.Span)
	switch stmt.Op {
	case StmtDecl:
		lx.typecheckDecl(stmt.Decl)
	case StmtExpr:
		lx.typecheckExpr(stmt.Expr)
	case Empty:
		// ok
	case Block:
		for _, s := range stmt.Block {
			lx.typecheckStmt(s)
		}
	case ARGBEGIN:
		lx.Errorf("ARGBEGIN not supported")
	case Break:
		// check break context
	case Continue:
		// check continue context
	case Do:
		// push break/continue context
		lx.typecheckStmt(stmt.Body)
		lx.typecheckExpr(stmt.Expr)
	case For:
		// push break/continue context
		lx.typecheckExpr(stmt.Pre)
		lx.typecheckExpr(stmt.Expr)
		lx.typecheckExpr(stmt.Post)
		lx.typecheckStmt(stmt.Body)
	case If:
		lx.typecheckExpr(stmt.Expr)
		// check bool
		lx.typecheckStmt(stmt.Body)
		lx.typecheckStmt(stmt.Else)
	case Goto:
		// check that label exists
	case Return:
		lx.typecheckExpr(stmt.Expr)
		// check return
	case Switch:
		lx.typecheckExpr(stmt.Expr)
		lx.typecheckStmt(stmt.Body)
		// push break context
		// push switch type
	case While:
		lx.typecheckExpr(stmt.Expr)
		lx.typecheckStmt(stmt.Body)
		// push break/continue context
	}

	for _, lab := range stmt.Labels {
		lx.typecheckExpr(lab.Expr)
	}
}

func (lx *lexer) typecheckType(typ *Type) {
	if typ == nil {
		return
	}

	// TODO: Is there any other work to do for type checking a type?
	switch typ.Kind {
	case Enum:
		// Give enum type to the declared names.
		// Perhaps should be done during parsing.
		for _, decl := range typ.Decls {
			if decl.Init != nil {
				lx.typecheckInit(typ, decl.Init)
			}
			decl.Type = typ
		}
	}
}

func (lx *lexer) typecheckInit(typ *Type, x *Init) {
	// TODO: Type check initializers (ugh).

	x.XType = typ
	typ = stripTypedef(typ)
	lx.setSpan(x.Span)
	if x.Braced == nil {
		lx.typecheckExpr(x.Expr)
		if x.Expr.XType == nil {
			return
		}
		if typ.Kind == Array && typ.Base.Is(Char) && x.Expr.Op == String {
			// ok to initialize char array with string
			if typ.Width == nil {
				typ.Width = x.Expr.XType.Width
			}
			return
		}
		if !canAssign(typ, x.Expr.XType, x.Expr) {
			lx.Errorf("cannot initialize %v with %v (type %v)", typ, x.Expr.XType, x.Expr)
			return
		}
		return
	}

	switch typ.Kind {
	case Array, Struct:
		// ok
	case Union:
		// C allows this but we do not.
		fallthrough
	default:
		lx.Errorf("cannot initialize type %v with braced initializer", typ)
		return
	}

	// Keep our sanity: require that either all elements have prefixes or none do.
	// This is not required by the C standard; it just makes this code more tractable.
	n := 0
	for _, elem := range x.Braced {
		if len(elem.Prefix) > 0 {
			if len(elem.Prefix) != 1 {
				lx.setSpan(elem.Span)
				lx.Errorf("unsupported compound initializer prefix")
				return
			}
			n++
		}
	}
	if n != 0 && n != len(x.Braced) {
		lx.Errorf("initializer elements must have no prefixes or all be prefixed")
		return
	}

	if n == 0 {
		// Assign elements in order.
		if typ.Kind == Array {
			// TODO: Check against typ.Width and record new typ.Width if missing
			for _, elem := range x.Braced {
				lx.typecheckInit(typ.Base, elem)
			}
			return
		}

		// Struct
		if len(x.Braced) > len(typ.Decls) {
			lx.Errorf("more initializer elements than struct fields in %v (%d > %d)", typ, len(x.Braced), len(typ.Decls))
			return
		}
		for i, elem := range x.Braced {
			decl := typ.Decls[i]
			lx.typecheckInit(decl.Type, elem)
		}
		return
	}

	// All elements have initializing prefixes.

	if typ.Kind == Array {
		for _, elem := range x.Braced {
			lx.setSpan(elem.Span)
			pre := elem.Prefix[0]
			if pre.Index == nil {
				lx.Errorf("field initializer prefix in array")
				continue
			}
			lx.typecheckExpr(pre.Index)
			// TODO: check that pre.Index is integer constant
			// TODO: record width if needed
			lx.typecheckInit(typ.Base, elem)
		}
		return
	}

	// Struct
	for _, elem := range x.Braced {
		lx.setSpan(elem.Span)
		pre := elem.Prefix[0]
		if pre.Dot == "" {
			lx.Errorf("array initializer prefix in struct")
			continue
		}
		decl := structDot(typ, pre.Dot)
		if decl == nil {
			lx.Errorf("type %v has no field .%v", typ, pre.Dot)
			continue
		}
		pre.XDecl = decl
		lx.typecheckInit(decl.Type, elem)
	}
}

func stripTypedef(t *Type) *Type {
	if t != nil && t.Kind == TypedefType && t.Base != nil {
		t = t.Base
	}
	return t
}

func isInt(t *Type) bool {
	t = stripTypedef(t)
	return Char <= t.Kind && t.Kind <= Ulonglong || t.Kind == Enum
}

func isPtr(t *Type) bool {
	t = stripTypedef(t)
	return t.Kind == Ptr || t.Kind == Array
}

func ptrBase(t *Type) *Type {
	t = stripTypedef(t)
	if t == nil || (t.Kind != Ptr && t.Kind != Array) {
		return nil
	}
	return t.Base
}

func toPtr(t *Type) *Type {
	t1 := stripTypedef(t)
	if t1.Kind == Ptr {
		return t
	}
	if t1.Kind == Array {
		return &Type{Kind: Ptr, Base: t1.Base}
	}
	return nil
}

func isArith(t *Type) bool {
	t = stripTypedef(t)
	return Char <= t.Kind && t.Kind <= Enum
}

func isScalar(t *Type) bool {
	t = stripTypedef(t)
	return Char <= t.Kind && t.Kind <= Ptr
}

func (t *Type) Is(k TypeKind) bool {
	t = stripTypedef(t)
	return t != nil && t.Kind == k
}

func (t *Type) Def() *Type {
	return stripTypedef(t)
}

func isNull(x *Expr) bool {
	for x != nil && x.Op == Paren {
		x = x.Left
	}
	return x != nil && x.Op == Number && x.Text == "0"
}

func isVoidPtr(t *Type) bool {
	return ptrBase(t).Is(Void)
}

func (t *Type) IsPtrVoid() bool {
	return isVoidPtr(t)
}

func isCompatPtr(t1, t2 *Type) bool {
	return isCompat(ptrBase(t1), ptrBase(t2))
}

// This is not correct; see C99 §6.2.7.
func isCompat(t1, t2 *Type) bool {
	t1 = stripTypedef(t1)
	t2 = stripTypedef(t2)
	if t1 == nil || t2 == nil || t1.Kind != t2.Kind {
		return false
	}
	if t1 == t2 {
		return true
	}
	switch t1.Kind {
	default:
		// arithmetic
		return true
	case Ptr, Array:
		return isCompat(ptrBase(t1), ptrBase(t2))
	case Struct, Union, Enum:
		return t1.Tag != "" && t1.Tag == t2.Tag
	case Func:
		if len(t1.Decls) != len(t2.Decls) || !isCompat(t1.Base, t2.Base) {
			return false
		}
		for i, d1 := range t1.Decls {
			d2 := t2.Decls[i]
			if d1.Type == nil && d1.Name == "..." {
				if d2.Type == nil && d2.Name == "..." {
					continue
				}
				return false
			}
			if !isCompat(d1.Type, d2.Type) {
				return false
			}
		}
		return true
	}
}

// TODO
func compositePtr(t1, t2 *Type) *Type {
	return toPtr(t1)
}

func canAssign(l, r *Type, rx *Expr) bool {
	switch {
	case isArith(l) && isArith(r):
		// ok
	case isCompat(l, r):
		// ok
	case isCompatPtr(l, r):
		// ok
	case isPtr(l) && isPtr(r) && (isVoidPtr(l) || isVoidPtr(r)):
		// ok
	case isPtr(l) && isNull(rx):
		// ok
		rx.XType = toPtr(l)
	case isPtr(l) && isCompat(ptrBase(l), r):
		// ok
	case isVoidPtr(l) && r.Is(Func), isVoidPtr(r) && l.Is(Func):
		// ok
	case isPtr(l) && ptrBase(l).Is(Func) && r.Is(Func): // && isCompat(ptrBase(l), r):
		if !isCompat(ptrBase(l), r) {
			fmt.Printf("not compat: %v and %v (%v)\n", ptrBase(l), r, rx)
		}
		// ok
	default:
		return false
	}
	return true
}

func (lx *lexer) toBool(x *Expr) *Type {
	if x.XType == nil {
		return nil
	}
	t := stripTypedef(x.XType)
	if Char <= t.Kind && t.Kind <= Ptr || t.Kind == Enum {
		return BoolType
	}
	lx.Errorf("cannot use %v (type %v) in boolean context", x, x.XType)
	return nil
}

// The "usual arithmetic conversions".
func promote2(l, r *Type) *Type {
	l = promote1(l)
	r = promote1(r)

	// if mixed signedness, make l signed and r unsigned.
	// specifically, if l is unsigned, swap with r.
	if (l.Kind-Char)&1 == 1 {
		l, r = r, l
	}

	switch {
	case l.Kind == Void || r.Kind == Void || l.Kind > Double || r.Kind > Double:
		return nil
	case l.Kind == r.Kind:
		return l
	// double wins.
	case l.Kind == Double:
		return l
	case r.Kind == Double:
		return r
	// float wins.
	case l.Kind == Float:
		return l
	case r.Kind == Float:
		return r
	// if both signed or both unsigned, higher kind wins.
	case (l.Kind-Char)&1 == (r.Kind-Char)&1:
		if l.Kind < r.Kind {
			return r
		}
		return l
	// mixed signedness: l is signed, r is unsigned (see above).
	// if unsigned higher kind than signed, unsigned wins.
	case r.Kind >= l.Kind:
		return r
	// signed is higher kind than unsigned (l.Kind > r.Kind).
	// if signed bigger than unsigned, signed wins.
	// only possible way this isn't true
	case (l.Kind-Char)/2 > (r.Kind-Char)/2 && (l.Kind != Long || r.Kind != Uint):
		return l
	// otherwise, use unsigned type corresponding to the signed type.
	default:
		return &Type{Kind: l.Kind + 1}
	}
	panic(fmt.Sprintf("missing case in promote2(%v, %v)", l, r))
}

func promote1(l *Type) *Type {
	l = stripTypedef(l)
	if Char <= l.Kind && l.Kind <= Ushort || l.Kind == Enum {
		l = IntType
	}
	return l
}

func structDot(t *Type, name string) *Decl {
	if t == nil || (t.Kind != Struct && t.Kind != Union) {
		return nil
	}
	for _, decl := range t.Decls {
		if decl.Name == name {
			return decl
		}
		if decl.Name == "" {
			d := structDot(decl.Type, name)
			if d != nil {
				return d
			}
		}
	}
	return nil
}

func (lx *lexer) parseChar1(text string) (val byte, wid int, ok bool) {
	if text[0] != '\\' {
		return text[0], 1, true
	}
	if len(text) == 1 {
		lx.Errorf("truncated escape sequence in character or string constant")
		return
	}
	switch text[1] {
	case 'a':
		return 7, 2, true
	case 'b':
		return 8, 2, true
	case 'f':
		return 12, 2, true
	case 'n':
		return 10, 2, true
	case 'r':
		return 13, 2, true
	case 't':
		return 9, 2, true
	case 'v':
		return 11, 2, true
	case '\'', '"', '?', '\\':
		return text[1], 2, true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		i := 2
		v := int(text[1] - '0')
		for i < 4 && i < len(text) && '0' <= text[i] && text[i] <= '7' {
			v = v*8 + int(text[i]-'0')
			i++
		}
		if v >= 256 {
			lx.Errorf("octal escape %s out of range", text[:i])
			return
		}
		return byte(v), i, true
	case 'x':
		i := 2
		v := 0
		for i < len(text) && ishex(text[i]) {
			v = v*16 + unhex(text[i])
			i++
		}
		if i-2 > 2 {
			lx.Errorf("hexadecimal escape %s out of range", text[:i])
			return
		}
		if i == 0 {
			lx.Errorf("hexadecimal escape %s missing digits", text[:i])
			return
		}
		return byte(v), i, true

	default:
		lx.Errorf("invalid escape sequence %s", text[:2])
	}
	return
}

func ishex(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

func unhex(c byte) int {
	if '0' <= c && c <= '9' {
		return int(c) - '0'
	}
	if 'a' <= c && c <= 'f' {
		return int(c) - 'a' + 10
	}
	if 'A' <= c && c <= 'F' {
		return int(c) - 'A' + 10
	}
	return -1
}

func (lx *lexer) parseChar(text string) (val byte, ok bool) {
	if len(text) < 3 || text[0] != '\'' || text[len(text)-1] != '\'' {
		lx.Errorf("invalid character constant %v", text)
		return 0, false
	}
	val, wid, ok := lx.parseChar1(text[1 : len(text)-1])
	if !ok {
		return 0, false
	}
	if wid != len(text)-2 {
		lx.Errorf("invalid character constant %v - multiple characters", text)
		return 0, false
	}
	return val, true
}

func (lx *lexer) parseString(text string) (val string, ok bool) {
	if len(text) < 2 || text[0] != '"' || text[len(text)-1] != '"' {
		lx.Errorf("invalid string constant %v", text)
		return "", false
	}
	tval := text[1 : len(text)-1]
	var bval []byte
	for len(tval) > 0 {
		ch, wid, ok := lx.parseChar1(tval)
		if !ok {
			return "", false
		}
		bval = append(bval, ch)
		tval = tval[wid:]
	}
	return string(bval), true
}

func (lx *lexer) typecheckExpr(x *Expr) {
	if x == nil {
		return
	}

	if x.Op != Offsetof {
		lx.typecheckExpr(x.Left)
	}
	lx.typecheckExpr(x.Right)
	for _, y := range x.List {
		lx.typecheckExpr(y)
	}
	lx.typecheckType(x.Type)

	lx.setSpan(x.Span)
	switch x.Op {
	default:
		panic("missing typecheck " + x.Op.String())

	case Add:
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		switch {
		case isPtr(l) && isInt(r):
			x.XType = toPtr(l)
		case isInt(l) && isPtr(r):
			x.XType = toPtr(r)
		default:
			lx.typecheckArith(x)
		}

	case Addr:
		t := x.Left.XType
		if t == nil {
			break
		}
		if isPtr(t) {
			t = toPtr(t)
		}
		x.XType = &Type{Kind: Ptr, Base: t}

	case AddEq:
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		if isPtr(l) && isInt(r) {
			x.XType = toPtr(l)
			break
		}
		lx.typecheckArithEq(x)

	case And, Mod, Or, Xor:
		// int & int
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		if !isInt(l) || !isInt(r) {
			lx.Errorf("invalid bitwise op of %v (type %v) and %v (type %v)", x.Left, l, x.Right, r)
			break
		}
		x.XType = promote2(l, r)

	case AndAnd, OrOr:
		// bool && bool
		l := lx.toBool(x.Left)
		r := lx.toBool(x.Right)
		if l == nil || r == nil {
			break
		}
		x.XType = BoolType

	case AndEq, ModEq, OrEq, XorEq:
		// int &= int
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		if !isInt(l) || !isInt(r) {
			lx.Errorf("invalid bitwise op of %v (type %v) and %v (type %v)", x.Left, l, x.Right, r)
			break
		}
		x.XType = l

	case Arrow:
		t := x.Left.XType
		if t == nil {
			lx.Errorf("arrow missing type %s", x.Left)
			break
		}
		if !isPtr(t) {
			lx.Errorf("invalid -> of non-pointer %v (type %v)", x.Left, t)
			break
		}
		t = stripTypedef(ptrBase(t))
		if t.Kind != Struct && t.Kind != Union {
			lx.Errorf("invalid -> of pointer to non-struct/union %v (type %v)", x.Left, t)
			break
		}
		d := structDot(t, x.Text)
		if d == nil {
			lx.Errorf("unknown field %v->%v", t, x.Text)
			break
		}
		x.XDecl = d
		x.XType = d.Type

	case Call:
		t := x.Left.XType
		if t == nil {
			break
		}
		if isPtr(t) {
			t = ptrBase(t)
		}
		t = stripTypedef(t)
		if t.Kind != Func {
			lx.Errorf("invalid call of %v (type %v)", x.Left, x.Left.XType)
			break
		}
		x.XType = t.Base
		for i := 0; i < len(x.List) || i < len(t.Decls); i++ {
			if i >= len(t.Decls) {
				lx.Errorf("too many arguments to call")
				break
			}
			d := t.Decls[i]
			if d.Name == "..." {
				break
			}
			if i >= len(x.List) {
				if len(x.List) == 0 && len(t.Decls) == 1 && t.Decls[0].Type.Is(Void) {
					break
				}
				lx.Errorf("not enough arguments to call")
				break
			}
			if x.List[i].XType != nil && !canAssign(d.Type, x.List[i].XType, x.List[i]) {
				lx.Errorf("cannot assign %v (type %v) to %v in call", x.List[i], x.List[i].XType, d.Type)
			}
		}

	case Cast:
		// NOTE: Assuming cast is valid.
		x.XType = x.Type

	case CastInit:
		lx.typecheckInit(x.Type, x.Init)
		x.XType = x.Type

	case Comma:
		x.XType = x.List[len(x.List)-1].XType

	case Cond:
		c, l, r := lx.toBool(x.List[0]), x.List[1].XType, x.List[2].XType
		if c == nil || l == nil || r == nil {
			break
		}
		switch {
		default:
			lx.Errorf("incompatible branches %v (type %v) and %v (type %v) in conditional", x.List[1], l, x.List[2], r)
		case isArith(l) && isArith(r):
			x.XType = promote2(l, r)
		case l == r:
			x.XType = l
		case isCompatPtr(l, r):
			x.XType = compositePtr(l, r)
		case isPtr(l) && isNull(x.List[1]):
			x.XType = toPtr(l)
			x.List[1].XType = x.XType
		case isPtr(r) && isNull(x.List[0]):
			x.XType = toPtr(r)
			x.List[0].XType = x.XType
		case isPtr(l) && isVoidPtr(r):
			x.XType = r
		case isPtr(r) && isVoidPtr(l):
			x.XType = l
		}

	case Div, Mul:
		lx.typecheckArith(x)

	case DivEq, MulEq:
		lx.typecheckArithEq(x)

	case Dot:
		t := x.Left.XType
		if t == nil {
			break
		}
		t = stripTypedef(t)
		if t.Kind != Struct && t.Kind != Union {
			lx.Errorf("invalid . of non-struct/union %v (type %v)", x.Left, t)
			break
		}
		d := structDot(t, x.Text)
		if d == nil {
			lx.Errorf("unknown field %v.%v", t, x.Text)
			break
		}
		x.XDecl = d
		x.XType = d.Type

	case Eq:
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		x.XType = l
		if !canAssign(l, r, x.Right) {
			lx.Errorf("invalid assignment %v (type %v) = %v (typ %v)", x.Left, l, x.Right, r)
			break
		}

	case EqEq, NotEq, Gt, GtEq, Lt, LtEq:
		x.XType = BoolType
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		if isArith(l) && isArith(r) {
			if x.Left.Op != Number && x.Right.Op == Number {
				x.Right.XType = x.Left.XType
			}
			if x.Right.Op != Number && x.Left.Op == Number {
				x.Left.XType = x.Right.XType
			}
			break
		}
		if isCompatPtr(l, r) {
			break
		}
		if x.Op == EqEq || x.Op == NotEq {
			if isPtr(l) {
				if isNull(x.Right) || isVoidPtr(r) {
					x.Right.XType = toPtr(l)
					break
				}
			}
			if isPtr(r) {
				if isNull(x.Left) || isVoidPtr(l) {
					x.Left.XType = toPtr(r)
					break
				}
			}
		}
		lx.Errorf("invalid comparison of %v (type %v) and %v (type %v)", x.Left, l, x.Right, r)

	case Index:
		// ptr[int]
		// int[ptr]
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		switch {
		case isPtr(l) && isInt(r):
			x.XType = ptrBase(l)
		case isInt(l) && isPtr(r):
			x.XType = ptrBase(r)
		default:
			lx.Errorf("invalid index %v (types %v, %v)", x, l, r, isPtr(l), isInt(r), r.Kind, r.Base, r.Base.Kind)
		}

	case Indir:
		// *ptr
		t := x.Left.XType
		if t == nil {
			break
		}
		if !isPtr(t) {
			lx.Errorf("invalid indirect of non-pointer %v (type %v)", x.Left, t)
			break
		}
		x.XType = ptrBase(t)

	case Lsh, Rsh:
		// int << int
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		if !isInt(l) || !isInt(r) {
			lx.Errorf("invalid shift of %v (type %v) and %v (type %v)", x.Left, l, x.Right, r)
			break
		}
		x.XType = promote1(l)

	case LshEq, RshEq:
		// int <<= int
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		if !isInt(l) || !isInt(r) {
			lx.Errorf("invalid shift of %v (type %v) and %v (type %v)", x.Left, l, x.Right, r)
			break
		}
		x.XType = l

	case Minus, Plus:
		// -int
		// -float
		t := x.Left.XType
		if t == nil {
			break
		}
		if !isArith(t) {
			lx.Errorf("invalid ± of %v (type %v)", x, t)
			break
		}
		x.XType = promote1(t)

	case Name:
		if x.XDecl == nil {
			lx.Errorf("undefined: %s", x.Text)
			break
		}
		//	XXX this happens for enums
		//	if x.XDecl.Type == nil {
		//		lx.Errorf("missing type for defined variable: %s", x.Text)
		//	}
		x.XType = x.XDecl.Type

	case Not:
		// !bool
		lx.toBool(x.Left)
		x.XType = BoolType

	case Number:
		num := x.Text
		if num[0] == '\'' {
			// character constant
			_, _ = lx.parseChar(num)
			x.XType = IntType
			break
		}

		if strings.Contains(num, ".") || !strings.HasPrefix(num, "0x") && strings.ContainsAny(num, "eE") {
			// floating point
			num = strings.TrimRight(num, "fFlL")
			suf := x.Text[len(num):]
			f, err := strconv.ParseFloat(num, 64)
			if err != nil {
				lx.Errorf("invalid floating point constant %v", x.Text)
				break
			}
			_ = f // TODO use this
			x.XType = DoubleType
			switch suf {
			case "":
			case "f", "F":
				x.XType = FloatType
			default:
				lx.Errorf("unsupported floating point constant suffix %v", x.Text)
			}
			break
		}

		// integer
		num = strings.TrimRight(num, "uUlL")
		suf := x.Text[len(num):]
		i, err := strconv.ParseUint(num, 0, 64)
		if err != nil {
			lx.Errorf("invalid integer constant %v", x.Text)
			break
		}
		_ = i // TODO use this
		has := strings.Contains
		suf = strings.ToUpper(suf)
		switch {
		case has(suf, "U") && has(suf, "LL"):
			x.XType = UlonglongType
		case has(suf, "U") && has(suf, "L"):
			if uint64(uint32(i)) == i {
				x.XType = UlongType
			} else {
				x.XType = UlonglongType
			}
		case has(suf, "U"):
			if uint64(uint32(i)) == i {
				x.XType = UintType
			} else {
				x.XType = UlonglongType
			}
		case has(suf, "LL"):
			if int64(i) >= 0 {
				x.XType = LonglongType
			} else {
				lx.Errorf("integer constant %v overflows signed long long", x.Text)
			}
		case has(suf, "L"):
			if int32(i) >= 0 && uint64(int32(i)) == i {
				x.XType = LongType
			} else if int64(i) >= 0 {
				x.XType = LonglongType
			} else {
				lx.Errorf("integer constant %v overflows signed long long", x.Text)
			}
		default:
			if int32(i) >= 0 && uint64(int32(i)) == i {
				x.XType = IntType
			} else if int64(i) >= 0 {
				x.XType = LonglongType
			} else {
				lx.Errorf("integer constant %v overflows signed long long", x.Text)
			}
		}

	case Offsetof:
		x.XType = LongType
		if x.Left.Op != Name {
			lx.Errorf("offsetof field too complicated")
		}
		d := structDot(x.Type, x.Left.Text)
		if d == nil {
			lx.Errorf("unknown field %v.%v", x.Type, x.Text)
		}

	case Paren:
		// (non-void)
		t := x.Left.XType
		if t == nil {
			break
		}
		if t.Kind == Void {
			lx.Errorf("cannot parenthesize void expression")
			break
		}
		x.XType = t

	case PostDec, PostInc, PreDec, PreInc:
		// int--
		// float--
		// ptr--
		t := x.Left.XType
		if t == nil {
			break
		}
		if !isArith(t) && !isPtr(t) {
			lx.Errorf("cannot increment/decrement %v (type %v)", x.Left, t)
			break
		}
		x.XType = t

	case SizeofExpr:
		x.XType = LongType

	case SizeofType:
		x.XType = LongType

	case String:
		// string list
		var str []string
		ok := true
		for _, text := range x.Texts {
			s, sok := lx.parseString(text)
			if !sok {
				ok = false
			}
			str = append(str, s)
		}
		if !ok {
			break
		}
		s := strings.Join(str, "")
		_ = s // TODO use this
		x.XType = &Type{Kind: Array, Width: &Expr{Op: Number, Text: fmt.Sprint(len(s) + 1)}, Base: CharType}

	case Sub:
		l, r := x.Left.XType, x.Right.XType
		if l == nil || r == nil {
			break
		}
		switch {
		case isPtr(l) && isInt(r):
			x.XType = toPtr(l)
		case isCompatPtr(l, r):
			x.XType = LongType
		default:
			lx.typecheckArith(x)
		}

	case SubEq:
		// ptr -= int
		lx.typecheckArithEq(x)

	case Twid:
		// ~int
		t := x.Left.XType
		if t == nil {
			break
		}
		if !isInt(t) {
			lx.Errorf("invalid ~ of %v (type %v)", x, t)
			break
		}
		x.XType = promote1(t)

	case VaArg:
		// va_arg(arg, int)
		t := x.Left.XType
		if t == nil {
			break
		}
		if t.Name != "va_list" {
			lx.Errorf("va_arg takes va_list, have %v (type %v)", x.Left, t)
		}
		x.XType = x.Type
	}
}

func (lx *lexer) typecheckArith(x *Expr) {
	// int + int
	// float + float
	l, r := x.Left.XType, x.Right.XType
	if l == nil || r == nil {
		return
	}
	if !isArith(l) || !isArith(r) {
		lx.Errorf("invalid arithmetic op of %v (type %v) and %v (type %v)", x.Left, l, x.Right, r)
		return
	}
	x.XType = promote2(l, r)
}

func (lx *lexer) typecheckArithEq(x *Expr) {
	// int + int
	// float + float
	l, r := x.Left.XType, x.Right.XType
	if l == nil || r == nil {
		return
	}
	if !isArith(l) || !isArith(r) {
		lx.Errorf("invalid arithmetic op of %v (type %v) and %v (type %v)", x.Left, l, x.Right, r)
		return
	}
	x.XType = l
}
