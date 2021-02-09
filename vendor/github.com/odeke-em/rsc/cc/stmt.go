// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

type Stmt struct {
	SyntaxInfo
	Op     StmtOp
	Pre    *Expr
	Expr   *Expr
	Post   *Expr
	Decl   *Decl
	Body   *Stmt
	Else   *Stmt
	Block  []*Stmt
	Labels []*Label
	Text   string
	Type   *Type
}

type StmtOp int

const (
	_ StmtOp = iota
	StmtDecl
	StmtExpr
	Empty
	Block
	ARGBEGIN
	Break
	Continue
	Do
	For
	If
	Goto
	Return
	Switch
	While
)

type Label struct {
	SyntaxInfo
	Op   LabelOp
	Expr *Expr
	Name string
}

type LabelOp int

const (
	_ LabelOp = iota
	Case
	Default
	LabelName
)
