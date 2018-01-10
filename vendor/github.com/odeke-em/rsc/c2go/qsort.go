// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"code.google.com/p/rsc/c2go"
	"code.google.com/p/rsc/cc"
)

func unparen(x *cc.Expr) *cc.Expr {
	for x != nil && x.Op == cc.Paren {
		x = x.Left
	}
	return x
}

func fixQsort(prog *cc.Prog, x *cc.Expr) {
	if len(x.List) != 4 {
		fprintf(x.Span, "unsupported %v - wrong arg count", x)
		return
	}
	if x.List[2].Op != cc.SizeofExpr || unparen(x.List[2].Left).String() != unparen(x.List[0]).String()+"[0]" {
		fprintf(x.Span, "unsupported %v - wrong elem size %v vs %v", x, unparen(x.List[2].Left).String(), unparen(x.List[0]).String()+"[0]")
		return
	}
	if x.List[3].Op != cc.Name || x.List[3].XDecl == nil {
		fprintf(x.Span, "unsupported %v - unknown comparison function", x)
		return
	}
	cmp := x.List[3].XDecl.Name
	decl := x.List[3].XDecl

	typ := fixQsortCmp(decl)
	if typ == nil {
		return
	}

	x.Left.Text = "sort.Sort"
	x.Left.XDecl = nil
	x.List = []*cc.Expr{
		&cc.Expr{
			Op:   cc.Call,
			Left: &cc.Expr{Op: cc.Name, Text: cmp},
			List: []*cc.Expr{
				{
					Op: c2go.ExprSlice,
					List: []*cc.Expr{
						x.List[0],
						nil,
						x.List[1],
					},
				},
			},
		},
	}

	typeDecl := &cc.Decl{
		GoPackage:  decl.GoPackage,
		SyntaxInfo: cc.SyntaxInfo{Span: decl.Span},
		Name:       cmp,
		Storage:    cc.Typedef,
		Type: &cc.Type{
			Kind: c2go.Slice,
			Base: typ,
		},
	}

	lenFunc := &cc.Decl{
		GoPackage:  decl.GoPackage,
		SyntaxInfo: cc.SyntaxInfo{Span: decl.Span},
		Name:       "(x " + cmp + ") Len",
		Type: &cc.Type{
			Kind: cc.Func,
			Base: intType,
		},
		Body: &cc.Stmt{
			Op: cc.Block,
			Block: []*cc.Stmt{
				{
					Op: cc.Return,
					Expr: &cc.Expr{
						Op:   cc.Name,
						Text: "len(x)",
					},
				},
			},
		},
	}

	swapFunc := &cc.Decl{
		GoPackage:  decl.GoPackage,
		SyntaxInfo: cc.SyntaxInfo{Span: decl.Span},
		Name:       "(x " + cmp + ") Swap",
		Type: &cc.Type{
			Kind: cc.Func,
			Base: &cc.Type{Kind: cc.Void},
			Decls: []*cc.Decl{
				{Name: "i", Type: &cc.Type{Kind: cc.TypedefType}},
				{Name: "j", Type: intType},
			},
		},
		Body: &cc.Stmt{
			Op: cc.Block,
			Block: []*cc.Stmt{
				{
					Op: cc.StmtExpr,
					Expr: &cc.Expr{
						Op:   cc.Name,
						Text: "x[i], x[j] = x[j], x[i]",
					},
				},
			},
		},
	}

	for i, d := range prog.Decls {
		if d == decl {
			prog.Decls = append(append(prog.Decls[:i:i], typeDecl, lenFunc, swapFunc), prog.Decls[i:]...)
			return
		}
	}
	prog.Decls = append(prog.Decls[:len(prog.Decls):len(prog.Decls)], typeDecl, lenFunc, swapFunc)
}

// isGoVoidPtr reports whether t is a void* or the Go translation of a void* (*struct{}).
func isGoVoidPtr(t *cc.Type) bool {
	if t == nil || t.Kind != cc.Ptr {
		return false
	}
	t = t.Base
	return t.Kind == cc.Void || t.Kind == cc.Struct && len(t.Decls) == 0
}

func fixQsortCmp(decl *cc.Decl) *cc.Type {
	ftyp := decl.Type
	if ftyp.Kind != cc.Func || len(ftyp.Decls) != 2 || !isGoVoidPtr(ftyp.Decls[0].Type) || !isGoVoidPtr(ftyp.Decls[1].Type) {
		fprintf(decl.Span, "invalid qsort cmp function %v - wrong args", c2go.GoString(ftyp))
		return nil
	}

	a1, a2 := ftyp.Decls[0], ftyp.Decls[1]
	var eq1, eq2, p1, p2 *cc.Expr
	var indir1, indir2 bool

	cc.Preorder(decl.Body, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Expr:
			if x.Op != cc.Eq {
				return
			}
			r := x.Right
			if r.Op == cc.Indir {
				r = r.Left
			}
			if r.Op == cc.Cast && r.Left.Op == cc.Name {
				if r.Left.XDecl == a1 && p1 == nil {
					p1 = x.Left
					eq1 = x
					indir1 = r != x.Right
				}
				if r.Left.XDecl == a2 && p2 == nil {
					p2 = x.Left
					eq2 = x
					indir1 = r != x.Right
				}
			}
		}
	})

	if p1 == nil || p2 == nil {
		fprintf(decl.Span, "invalid qsort cmp function - cannot find arg extraction")
		return nil
	}

	if !sameType(p1.XType, p2.XType) {
		fprintf(decl.Span, "invalid qsort cmp function - different arg types %v and %v", c2go.GoString(p1.XType), c2go.GoString(p2.XType))
		return nil
	}
	if indir1 != indir2 {
		fprintf(decl.Span, "invalid qsort cmp function - different arg indirection")
		return nil
	}

	typ := p1.XType
	if !indir1 {
		if typ.Def().Kind != cc.Ptr {
			fprintf(decl.Span, "invalid qsort cmp function - arg ptr cast to non-ptr %v", c2go.GoString(typ))
			return nil
		}
		typ = typ.Def().Base
	}

	// Have all the information. Committed.
	// Rewrite to take x, i, j, use x[i] for p1, x[j] for p2,
	// take address of x[i], x[j] if there was no indirect,
	// replace all return z with return z < 0.
	cmp := decl.Name
	decl.Name = "(x " + cmp + ") Less"
	decl.Type = &cc.Type{
		Kind: cc.Func,
		Base: boolType,
		Decls: []*cc.Decl{
			{Name: "i", Type: &cc.Type{Kind: cc.TypedefType}},
			{Name: "j", Type: intType},
		},
	}

	prefix := ""
	if !indir1 {
		prefix = "&"
	}
	eq1.Right = &cc.Expr{Op: cc.Name, Text: prefix + "x[i]", XType: p1.XType}
	eq2.Right = &cc.Expr{Op: cc.Name, Text: prefix + "x[j]", XType: p1.XType}

	cc.Preorder(decl.Body, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Stmt:
			if x.Op == cc.Return && x.Expr != nil {
				ret := x.Expr
				// Pick off 0, -1, +1.
				// Otherwise rewrite ret to ret < 0.
				switch ret.Op {
				case cc.Minus, cc.Plus:
					if ret.Left.Op == cc.Number {
						ret.Op = cc.Name
						if ret.Op == cc.Plus {
							ret.Text = "true"
						} else {
							ret.Text = "false"
						}
						ret.Left = nil
						ret.XType = boolType
						return
					}
				case cc.Number:
					ret.Op = cc.Name
					ret.Text = "false"
					ret.XType = boolType
					return
				}
				x.Expr = &cc.Expr{Op: cc.Lt, Left: ret, Right: &cc.Expr{Op: cc.Number, Text: "0"}, XType: boolType}
				return
			}
		}
	})

	return typ
}
