// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"code.google.com/p/rsc/c2go"
	"code.google.com/p/rsc/cc"
)

var (
	src     = flag.String("src", "", "src of search")
	dst     = flag.String("dst", "", "dst of search")
	out     = flag.String("o", "/tmp/c2go/src", "output directory")
	strip   = flag.String("", "", "strip from input paths when writing in output directory")
	inc     = flag.String("I", "", "include directory")
	fixFile = flag.String("f", "", "file containing func redefinitions")

	showGroups = flag.Bool("groups", false, "show groups")
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	args := flag.Args()
	if *inc != "" {
		cc.AddInclude(*inc)
	}
	var fixes *Config
	if *fixFile != "" {
		f, err := readConfig(*fixFile)
		if err != nil {
			log.Fatal(err)
		}
		fixes = f
	}
	if len(args) == 0 {
		flag.Usage()
	}
	var r []io.Reader
	files := args
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		r = append(r, f)
		defer f.Close()
	}
	prog, err := cc.ReadMany(files, r)
	if err != nil {
		log.Fatal(err)
	}
	renameDecls(fixes, prog)
	inferTypes(fixes, prog)
	rewriteSyntax(prog)
	rewriteTypes(prog)
	if *showGroups {
		return
	}
	rewriteLen(fixes, prog)
	fixGoTypes(prog)
	fixBools(prog)
	finalEdits(prog)
	doExports(fixes, prog)
	write(prog, files, fixes)
	for _, d := range fixes.Diffs {
		if d.Used == 0 {
			fmt.Fprintf(os.Stderr, "%s: unused diff\n", d.Line)
		}
	}
}

func readFixFile(filename string) map[string]string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]string)
	for _, match := range regexp.MustCompile(`(?ms)^func (\w+).*?^}`).FindAllStringSubmatch(string(data), -1) {
		m[match[1]] = match[0]
	}

	for _, match := range regexp.MustCompile(`(?m)^//c2go:drop (\w+)`).FindAllStringSubmatch(string(data), -1) {
		m[match[1]] = ""
	}

	return m
}

func declKey(d *cc.Decl) string {
	key := d.Name
	if t := d.OuterType; t != nil {
		name := t.Name
		if name == "" {
			name = t.Tag
		}
		if name == "" {
			name = t.String()
		}
		key = name + "." + key
	}
	if d.CurFn != nil {
		key = declKey(d.CurFn) + "." + key
	}
	return key
}

var override = map[string]*cc.Type{
	"oprrr.return": uint32Type,
	"opbra.return": uint32Type,
	"olr.return":   uint32Type,
	"olhr.return":  uint32Type,
	"olrr.return":  uint32Type,
	"osr.return":   uint32Type,
	"olhrr.return": uint32Type,
	"oshr.return":  uint32Type,
	"ofsr.return":  uint32Type,
	"osrr.return":  uint32Type,
	"oshrr.return": uint32Type,
	"omvl.return":  uint32Type,
	"ocmp.return":  uint32Type,

	"asmout.o1":  uint32Type,
	"asmout.o4":  uint32Type,
	"asmout.o5":  uint32Type,
	"asmout.o6":  uint32Type,
	"asmout.rel": &cc.Type{Kind: cc.Ptr},

	"setuintxx.wid": int64Type,

	"Link.andptr": &cc.Type{Kind: c2go.Slice},

	"chipfloat5.h": uint32Type,

	"oplook.return": &cc.Type{Kind: cc.Ptr},
	"oplook.c1":     &cc.Type{Kind: c2go.Slice},
	"oplook.c3":     &cc.Type{Kind: c2go.Slice},

	"Oprang.start": &cc.Type{Kind: c2go.Slice},

	"asmout.out": &cc.Type{Kind: c2go.Slice, Base: uint32Type},

	"span5.out": &cc.Type{Kind: cc.Array, Base: uint32Type, Width: &cc.Expr{Op: cc.Number, Text: "9"}},

	"LSym.r":  &cc.Type{Kind: c2go.Slice},
	"Prog.ft": &cc.Type{Kind: c2go.Uint8},
	"Prog.tt": &cc.Type{Kind: c2go.Uint8},
}

// Rewrite C types to be Go types.
func rewriteTypes(prog cc.Syntax) {
	// Assign overrides to groups.
	cc.Postorder(prog, func(x cc.Syntax) {
		if d, ok := x.(*cc.Decl); ok {
			key := declKey(d)
			t := override[key]
			if t == nil {
				return
			}
			if t.Kind == cc.Array {
				// Override only applies to specific decl. Skip for now.
				return
			}
			f := flowCache[d]
			if f == nil {
				return
			}
			g := f.group
			if g.goKind != 0 || g.goType != nil {
				fmt.Fprintf(os.Stderr, "multiple overrides: %v (%p) and %v (%p)\n", key, f.group, g.goKey, g.goFlow.group)
			}
			g.goKey = key
			g.goFlow = f
			if t.Kind <= cc.Enum {
				panic("bad go type override")
			}
			if (t.Kind == cc.Ptr || t.Kind == c2go.Slice) && t.Base == nil {
				g.goKind = t.Kind
			} else {
				g.goType = t
			}
		}
	})

	// Process overrides.
	cache := make(map[*cc.Type]*cc.Type)
	for _, g := range flowGroups {
		if g.goType != nil {
			continue
		}
		if c2go.Int8 <= g.goKind && g.goKind <= c2go.Float64 {
			g.goType = &cc.Type{Kind: g.goKind}
			continue
		}
		if g.goKind == cc.Ptr || g.goKind == c2go.Slice {
			t := g.decls[0].Type
			if t == nil || t.Base == nil {
				fmt.Printf("%s: expected ptr/array/slice for %s\n", g.decls[0].Span, declKey(g.decls[0]))
				continue
			}
			g.goType = &cc.Type{Kind: g.goKind, Base: toGoType(nil, nil, t.Base, cache)}
			continue
		}
		if g.goKind != 0 {
			fmt.Printf("%s: unexpected go kind %v\n", g.goKey, g.goKind)
			continue
		}
	}

	// Process defaults.
	// Each group has a 'canonical' instance of the type
	// that we can use as the initial hint.
	for _, g := range flowGroups {
		if g.goType != nil {
			continue
		}
		if g.canon == nil {
			fmt.Printf("group missing canonical\n")
			continue
		}
		if g.isBool {
			g.goType = boolType
			continue
		}
		t := g.canon.Def()
		if cc.Char <= t.Kind && t.Kind <= cc.Enum {
			// Convert to an appropriately sized number.
			// Canon is largest rank from C; convert to Go.
			g.goType = &cc.Type{Kind: c2goKind[t.Kind]}
			continue
		}

		if t.Kind == cc.Ptr || t.Kind == cc.Array {
			// Default is convert to pointer.
			// If there are any arrays or any pointer arithmetic, convert to slice instead.
			k := cc.Ptr
			for _, d := range g.decls {
				if d.Type != nil && d.Type.Kind == cc.Array {
					k = c2go.Slice
				}
			}
			for _, f := range g.syntax {
				if f.ptrAdd {
					k = c2go.Slice
				}
			}
			if t.Base.Kind == cc.Char {
				g.goType = &cc.Type{Kind: c2go.String}
				continue
			}
			g.goType = &cc.Type{Kind: k, Base: toGoType(nil, nil, t.Base, cache)}
			continue
		}
	}

	if *showGroups {
		fmt.Printf("%d groups\n", len(flowGroups))
		for _, g := range flowGroups {
			suffix := ""
			if g.isBool {
				suffix = " [bool]"
			}
			fmt.Printf("group(%d): %v (canon %v)%s\n", len(g.decls), c2go.GoString(g.goType), c2go.GoString(g.canon), suffix)
			for i := 0; i < 2; i++ {
				for _, f := range g.syntax {
					if d, ok := f.syntax.(*cc.Decl); ok == (i == 0) {
						suffix := ""
						if ok {
							suffix = ": " + declKey(d) + " " + c2go.GoString(d.Type)
						}
						if f.ptrAdd {
							suffix += " (ptradd)"
						}
						if f.usedAsBool {
							suffix += " (bool)"
						}
						fmt.Printf("\t%s %v%s\n", f.syntax.GetSpan(), f.syntax, suffix)
					}
				}
			}
		}
	}

	// Apply grouped decisions to individual declarations.
	cc.Postorder(prog, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Decl:
			d := x
			if d.Name == "..." || d.Type == nil {
				return
			}
			if d.Name == "" && d.Type.Is(cc.Enum) && len(d.Type.Decls) > 0 {
				for _, dd := range d.Type.Decls {
					dd.Type = idealType
				}
				return
			}
			t := override[declKey(d)]
			if t != nil && t.Kind == cc.Array {
				d.Type = t
				return
			}
			f := flowCache[d]
			if f == nil {
				d.Type = toGoType(nil, d, d.Type, cache)
				if declKey(d) != "tmp" {
					fprintf(d.Span, "missing flow group for %s", d.Span, declKey(d))
				}
				return
			}
			g := f.group
			if d.Init != nil && len(d.Init.Braced) > 0 && d.Type != nil && d.Type.Kind == cc.Array {
				// Initialization of array - do not override type.
				// But if size is not given explicitly, change to slice.
				d.Type.Base = toGoType(nil, nil, d.Type.Base, cache)
				if d.Type.Width == nil {
					d.Type.Kind = c2go.Slice
				}
				return
			}
			d.Type = toGoType(g, d, d.Type, cache)
			if d.Type != nil && d.Type.Kind == cc.Func && d.Type.Base.Kind != cc.Void {
				if f != nil && f.returnValue != nil && f.returnValue.group != nil && f.returnValue.group.goType != nil {
					d.Type.Base = f.returnValue.group.goType
				}
			}

		case *cc.Expr:
			if x.Type != nil {
				t := toGoType(nil, nil, x.Type, cache)
				if t == nil {
					fprintf(x.Span, "cannot convert %v to go type\n", c2go.GoString(x.Type))
				}
				x.Type = t
			}
		}
	})
}

func toGoType(g *flowGroup, x cc.Syntax, typ *cc.Type, cache map[*cc.Type]*cc.Type) (ret *cc.Type) {
	if typ == nil {
		return nil
	}

	// Array and func implicitly convert to pointer types, so don't
	// trust the group they are in - they'll turn into pointers incorrectly.
	if g != nil && typ.Kind != cc.Array && typ.Kind != cc.Func {
		if g.goType != nil {
			return g.goType
		}
		defer func() {
			if ret != nil && ret.Kind <= cc.Enum {
				panic("bad go type override")
			}
			g.goType = ret
		}()
	}

	// Look in cache first. This cuts off recursion for self-referential types.
	// The cache only contains aggregate types - numeric types are shared
	// by many expressions in the program and we might want to translate
	// them differently in different contexts.
	if cache[typ] != nil {
		return cache[typ]
	}

	var force *cc.Type

	if d, ok := x.(*cc.Decl); ok {
		key := declKey(d)
		force = override[key]
	}

	switch typ.Kind {
	default:
		panic(fmt.Sprintf("unexpected C type %s", typ))

	case c2go.Ideal:
		return typ

	case cc.Void:
		return &cc.Type{Kind: cc.Struct} // struct{}

	case cc.Char, cc.Uchar, cc.Short, cc.Ushort, cc.Int, cc.Uint, cc.Long, cc.Ulong, cc.Longlong, cc.Ulonglong, cc.Float, cc.Double, cc.Enum:
		// TODO: Use group.
		if force != nil {
			return force
		}
		return &cc.Type{Kind: c2goKind[typ.Kind]}

	case cc.Ptr:
		t := &cc.Type{Kind: cc.Ptr}
		cache[typ] = t
		t.Base = toGoType(nil, nil, typ.Base, cache)

		if g != nil {
			if g.goKind != 0 {
				t.Kind = g.goKind
				return t
			}
			for _, f := range g.syntax {
				if f.ptrAdd || f.ptrIndex {
					t.Kind = c2go.Slice
				}
			}
		}

		if force != nil {
			if force.Base != nil {
				return force
			}
			if force.Kind == cc.Ptr || force.Kind == c2go.Slice {
				t.Kind = force.Kind
				return t
			}
		}

		if typ.Base.Kind == cc.Char {
			t.Kind = c2go.String
			t.Base = nil
			return t
		}

		return t

	case cc.Array:
		if typ.Base.Def().Kind == cc.Char {
			return &cc.Type{Kind: c2go.String}
		}
		t := &cc.Type{Kind: cc.Array, Width: typ.Width}
		cache[typ] = t
		t.Base = toGoType(nil, nil, typ.Base, cache)
		return t

	case cc.TypedefType:
		// If this is a typedef like uchar, translate the base type directly.
		def := typ.Base
		if cc.Char <= def.Kind && def.Kind <= cc.Enum {
			return toGoType(g, x, def, cache)
		}

		// Otherwise assume it is a struct or some such, and preserve the name
		// but translate the base.
		t := &cc.Type{Kind: cc.TypedefType, Name: typ.Name, TypeDecl: typ.TypeDecl}
		cache[typ] = t
		t.Base = toGoType(nil, nil, typ.Base, cache)
		return t

	case cc.Func:
		// A func Type contains Decls, and we don't fork the Decls, so don't fork the Type.
		// The Decls themselves appear in the group lists, so they'll be handled by rewriteTypes.
		// The return value has no Decl and needs to be converted.
		if !typ.Base.Is(cc.Void) {
			typ.Base = toGoType(nil, nil, typ.Base, cache)
		}
		return typ

	case cc.Struct:
		// A struct Type contains Decls, and we don't fork the Decls, so don't fork the Type.
		// The Decls themselves appear in the group lists, so they'll be handled by rewriteTypes.
		return typ
	}
}

var c2goKind = map[cc.TypeKind]cc.TypeKind{
	cc.Char:   c2go.Int8,
	cc.Uchar:  c2go.Uint8,
	cc.Short:  c2go.Int16,
	cc.Ushort: c2go.Uint16,
	cc.Int:    c2go.Int,
	cc.Uint:   c2go.Uint,
	//cc.Long:      c2go.Int32,
	cc.Long:      c2go.Int, // long is used too indiscriminately to assign 32-bit meaning to it
	cc.Ulong:     c2go.Uint32,
	cc.Longlong:  c2go.Int64,
	cc.Ulonglong: c2go.Uint64,
	cc.Float:     c2go.Float32,
	cc.Double:    c2go.Float64,
	cc.Enum:      c2go.Int,
}

// fixGoTypes fixes all the Go type mismatches.
func fixGoTypes(prog *cc.Prog) {
	did := make(map[*cc.Decl]bool)
	for _, decl := range prog.Decls {
		if did[decl] {
			continue
		}
		did[decl] = true
		if decl.Init != nil {
			fixGoTypesInit(decl, decl.Init)
		}
		if decl.Body != nil {
			t := decl.Type
			if t != nil && t.Kind == cc.Func && t.Base.Is(c2go.Int) && len(t.Decls) == 1 && t.Decls[0].Type.String() == "Fmt*" {
				fixFormatter(decl)
			}
			fixGoTypesStmt(prog, decl, decl.Body)
		}
	}
}

func fixGoTypesInit(decl *cc.Decl, x *cc.Init) {
	if x.Expr != nil {
		fixGoTypesExpr(nil, x.Expr, x.XType)
	}
	for _, init := range x.Braced {
		fixGoTypesInit(decl, init)
	}
}

func fixGoTypesStmt(prog *cc.Prog, fn *cc.Decl, x *cc.Stmt) {
	if x == nil {
		return
	}

	switch x.Op {
	case cc.StmtDecl:
		fixGoTypesExpr(fn, x.Expr, nil)

	case cc.StmtExpr:
		if x.Expr != nil && x.Expr.Op == cc.Call && x.Expr.Left.Op == cc.Name {
			switch x.Expr.Left.Text {
			case "qsort":
				fixQsort(prog, x.Expr)
				return
			case "memset":
				fixMemset(prog, fn, x)
				return
			case "free":
				x.Op = cc.Empty
				x.Expr = nil
				return
			}
		}
		fixGoTypesExpr(fn, x.Expr, nil)

	case cc.If, cc.For:
		fixGoTypesExpr(fn, x.Pre, nil)
		fixGoTypesExpr(fn, x.Post, nil)
		fixGoTypesExpr(fn, x.Expr, boolType)

	case cc.Switch:
		fixGoTypesExpr(fn, x.Expr, nil)

	case cc.Return:
		if x.Expr != nil {
			forceGoType(fn, x.Expr, fn.Type.Base)
		}
	}
	for _, stmt := range x.Block {
		fixGoTypesStmt(prog, fn, stmt)
	}
	if len(x.Block) > 0 && x.Body != nil {
		panic("block and body")
	}
	fixGoTypesStmt(prog, fn, x.Body)
	fixGoTypesStmt(prog, fn, x.Else)

	for _, lab := range x.Labels {
		// TODO: use correct type
		fixGoTypesExpr(fn, lab.Expr, nil)
	}
}

func zeroFor(targ *cc.Type) *cc.Expr {
	if targ != nil {
		k := targ.Def().Kind
		switch k {
		case c2go.String:
			return &cc.Expr{Op: cc.String, Texts: []string{`""`}}

		case c2go.Slice, cc.Ptr:
			return &cc.Expr{Op: cc.Name, Text: "nil"}

		case cc.Struct, cc.Array:
			return &cc.Expr{Op: cc.CastInit, Type: targ, Init: &cc.Init{}}

		case c2go.Bool:
			return &cc.Expr{Op: cc.Name, Text: "false"}
		}

		if c2go.Int8 <= k && k <= c2go.Float64 {
			return &cc.Expr{Op: cc.Number, Text: "0"}
		}
		return &cc.Expr{Op: cc.Number, Text: "0 /*" + targ.String() + "*/"}
	}

	return &cc.Expr{Op: cc.Number, Text: "0 /*untyped*/"}
}

func fixGoTypesExpr(fn *cc.Decl, x *cc.Expr, targ *cc.Type) (ret *cc.Type) {
	if x == nil {
		return nil
	}

	defer func() {
		x.XType = ret
	}()

	if x.Op == cc.Paren {
		return fixGoTypesExpr(fn, x.Left, targ)
	}

	// Make explicit C's implicit conversions from boolean to non-boolean and vice versa.
	switch x.Op {
	case cc.AndAnd, cc.OrOr, cc.Not, cc.EqEq, cc.Lt, cc.LtEq, cc.Gt, cc.GtEq, cc.NotEq:
		if targ != nil && targ.Kind != c2go.Bool {
			old := copyExpr(x)
			if targ.Kind == c2go.Int {
				x.Op = cc.Call
				x.Left = &cc.Expr{Op: cc.Name, Text: "bool2int"}
				x.List = []*cc.Expr{old}
				x.Right = nil
			} else {
				x.Op = cc.Cast
				x.Left = &cc.Expr{Op: cc.Call, Left: &cc.Expr{Op: cc.Name, Text: "bool2int"}, List: []*cc.Expr{old}}
				x.Type = targ
			}
			fixGoTypesExpr(fn, old, boolType)
			return targ
		}
	default:
		if targ != nil && targ.Kind == c2go.Bool {
			old := copyExpr(x)
			left := fixGoTypesExpr(fn, old, nil)
			if left != nil && left.Kind == c2go.Bool {
				return targ
			}
			if old.Op == cc.Number {
				switch old.Text {
				case "1":
					x.Op = cc.Name
					x.Text = "true"
					return targ
				case "0":
					x.Op = cc.Name
					x.Text = "false"
					return targ
				}
			}
			x.Op = cc.NotEq
			x.Left = old
			x.Right = zeroFor(left)
			return targ
		}
	}

	switch x.Op {
	default:
		panic(fmt.Sprintf("unexpected construct %v in fixGoTypesExpr - %v - %v", c2go.GoString(x), x.Op, x.Span))

	case c2go.ExprSlice:
		// inserted by rewriteLen
		left := fixGoTypesExpr(fn, x.List[0], targ)
		fixGoTypesExpr(fn, x.List[1], nil)
		fixGoTypesExpr(fn, x.List[2], nil)
		return left

	case cc.Comma:
		for i, y := range x.List {
			t := targ
			if i+1 < len(x.List) {
				t = nil
			}
			fixGoTypesExpr(fn, y, t)
		}
		return nil

	case c2go.ExprBlock:
		for _, stmt := range x.Block {
			fixGoTypesStmt(nil, fn, stmt)
		}
		return nil

	case cc.Add, cc.And, cc.Div, cc.Mod, cc.Mul, cc.Or, cc.Sub, cc.Xor:
		if x.Op == cc.Sub && isPtrSliceOrArray(x.Left.XType) && isPtrSliceOrArray(x.Right.XType) {
			left := fixGoTypesExpr(fn, x.Left, nil)
			right := fixGoTypesExpr(fn, x.Right, nil)
			if left != nil && right != nil && left.Kind != right.Kind {
				if left.Kind == c2go.Slice {
					forceConvert(fn, x.Right, right, left)
				} else {
					forceConvert(fn, x.Left, left, right)
				}
			}
			x.Left = &cc.Expr{Op: cc.Minus, Left: &cc.Expr{Op: cc.Call, Left: &cc.Expr{Op: cc.Name, Text: "cap"}, List: []*cc.Expr{x.Left}}}
			x.Right = &cc.Expr{Op: cc.Call, Left: &cc.Expr{Op: cc.Name, Text: "cap"}, List: []*cc.Expr{x.Right}}
			x.Op = cc.Add
			return intType
		}

		left := fixGoTypesExpr(fn, x.Left, targ)

		if x.Op == cc.And && x.Right.Op == cc.Twid {
			x.Op = c2go.AndNot
			x.Right = x.Right.Left
		}

		if x.Op == cc.Add && isSliceStringOrArray(left) {
			fixGoTypesExpr(fn, x.Right, nil)
			x.Op = c2go.ExprSlice
			x.List = []*cc.Expr{x.Left, x.Right, nil}
			x.Left = nil
			x.Right = nil
			if left.Kind == cc.Array {
				left = &cc.Type{Kind: c2go.Slice, Base: left.Base}
			}
			return left
		}

		right := fixGoTypesExpr(fn, x.Right, targ)
		return fixBinary(fn, x, left, right, targ)

	case cc.AddEq, cc.AndEq, cc.DivEq, cc.Eq, cc.ModEq, cc.MulEq, cc.OrEq, cc.SubEq, cc.XorEq:
		left := fixGoTypesExpr(fn, x.Left, nil)

		if x.Op == cc.AndEq && x.Right.Op == cc.Twid {
			x.Op = c2go.AndNotEq
			x.Right = x.Right.Left
		}

		if x.Op == cc.AddEq && isSliceOrString(left) {
			fixGoTypesExpr(fn, x.Right, nil)
			old := copyExpr(x.Left)
			x.Op = cc.Eq
			x.Right = &cc.Expr{Op: c2go.ExprSlice, List: []*cc.Expr{old, x.Right, nil}}
			return left
		}

		if x.Op == cc.Eq && x.Left.Op == cc.Index && sameType(x.Left.Left.XType, stringType) && c2go.GoString(x.Left.Right) == "0" && c2go.GoString(x.Right) == "0" {
			x.Left = x.Left.Left
			x.Right = &cc.Expr{Op: cc.Name, Text: `""`}
			return x.Left.XType
		}

		forceGoType(fn, x.Right, left)

		return left

	case c2go.ColonEq:
		left := fixGoTypesExpr(fn, x.Right, nil)
		x.Left.XType = left
		x.Left.XDecl.Type = left
		return left

	case cc.Addr:
		left := fixGoTypesExpr(fn, x.Left, nil)
		if left == nil {
			return nil
		}

		if targ != nil && targ.Kind == c2go.Slice && sameType(targ.Base, left) {
			l := x.Left
			l.Op = c2go.ExprSlice
			l.List = []*cc.Expr{l.Left, l.Right, nil}
			l.Left = nil
			l.Right = nil
			fixMerge(x, l)
			return targ
		}

		return &cc.Type{Kind: cc.Ptr, Base: left}

	case cc.AndAnd, cc.OrOr, cc.Not:
		fixGoTypesExpr(fn, x.Left, boolType)
		if x.Right != nil {
			fixGoTypesExpr(fn, x.Right, boolType)
		}
		return boolType

	case cc.Arrow, cc.Dot:
		left := fixGoTypesExpr(fn, x.Left, nil)

		if x.Op == cc.Arrow && isSliceOrString(left) {
			x.Left = &cc.Expr{Op: cc.Index, Left: x.Left, Right: &cc.Expr{Op: cc.Number, Text: "0"}}
		}

		return x.XDecl.Type

	case cc.Call:
		if fixPrintf(fn, x) {
			return x.XType
		}
		if fixSpecialCall(fn, x, targ) {
			return x.XType
		}
		left := fixGoTypesExpr(fn, x.Left, nil)
		for i, y := range x.List {
			if left != nil && left.Kind == cc.Func && i < len(left.Decls) {
				forceGoType(fn, y, left.Decls[i].Type)
			} else {
				fixGoTypesExpr(fn, y, nil)
			}
		}
		if left != nil && left.Kind == cc.Func {
			return left.Base
		}
		return nil

	case cc.Cast:
		fixGoTypesExpr(fn, x.Left, nil)
		return x.Type

	case cc.CastInit:
		fixGoTypesInit(nil, x.Init)
		return x.Type

	case cc.EqEq, cc.Gt, cc.GtEq, cc.Lt, cc.LtEq, cc.NotEq:
		if fixSpecialCompare(fn, x) {
			return boolType
		}
		left := fixGoTypesExpr(fn, x.Left, nil)
		if x.Right.Op == cc.Number && x.Right.Text == "0" {
			if isSliceOrPtr(left) {
				x.Right.Op = cc.Name
				x.Right.Text = "nil"
				return boolType
			}
			if left != nil && left.Kind == c2go.String {
				x.Right.Op = cc.String
				x.Right.Texts = []string{`""`}
				return boolType
			}
		}
		right := fixGoTypesExpr(fn, x.Right, nil)

		if isSliceOrArray(x.Left.XType) && isSliceOrArray(x.Right.XType) {
			x.Left = &cc.Expr{Op: cc.Minus, Left: &cc.Expr{Op: cc.Call, Left: &cc.Expr{Op: cc.Name, Text: "cap"}, List: []*cc.Expr{x.Left}}}
			x.Right = &cc.Expr{Op: cc.Minus, Left: &cc.Expr{Op: cc.Call, Left: &cc.Expr{Op: cc.Name, Text: "cap"}, List: []*cc.Expr{x.Right}}}
			return boolType
		}

		fixBinary(fn, x, left, right, nil)
		return boolType

	case cc.Index, cc.Indir:
		left := fixGoTypesExpr(fn, x.Left, nil)
		if x.Right != nil {
			fixGoTypesExpr(fn, x.Right, nil)
		}
		if left == nil {
			return nil
		}

		if isSliceOrString(left) && x.Op == cc.Indir {
			x.Op = cc.Index
			x.Right = &cc.Expr{Op: cc.Number, Text: "0"}
		}

		switch left.Kind {
		case cc.Ptr, c2go.Slice, cc.Array:
			return left.Base

		case c2go.String:
			return byteType
		}
		return nil

	case cc.Lsh, cc.Rsh:
		left := fixGoTypesExpr(fn, x.Left, targ)
		if left != nil && targ != nil && c2go.Int8 <= left.Kind && left.Kind <= c2go.Float64 && targ.Kind > left.Kind {
			forceConvert(fn, x.Left, left, targ)
			left = targ
		}
		fixShiftCount(fn, x.Right)
		return left

	case cc.LshEq, cc.RshEq:
		left := fixGoTypesExpr(fn, x.Left, nil)
		fixShiftCount(fn, x.Right)
		return left

	case cc.Name:
		if x.Text == "nelem" {
		}
		switch x.Text {
		case "T", "S", "N", "L", "P":
			x.Text = "nil"
			x.XDecl = nil
			return nil
		case "nelem":
			x.Text = "len"
			x.XDecl = nil
			fallthrough
		case "len":
			return &cc.Type{Kind: cc.Func, Base: intType}
		}
		if x.XDecl == nil {
			return nil
		}
		return x.XDecl.Type

	case cc.Number:
		return idealType

	case cc.Minus, cc.Plus, cc.Twid:
		return fixGoTypesExpr(fn, x.Left, targ)

	case cc.Offsetof:
		// TODO
		return nil

	case cc.Paren:
		return fixGoTypesExpr(fn, x.Left, targ)

	case cc.PostDec, cc.PostInc:
		left := fixGoTypesExpr(fn, x.Left, nil)

		if x.Op == cc.PostInc && isSliceOrString(left) {
			old := copyExpr(x.Left)
			x.Op = cc.Eq
			x.Right = &cc.Expr{Op: c2go.ExprSlice, List: []*cc.Expr{old, &cc.Expr{Op: cc.Number, Text: "1"}, nil}}
		}

		return nil

	case cc.SizeofExpr:
		left := fixGoTypesExpr(fn, x.Left, nil)
		if left != nil && (left.Kind == cc.Array || left.Kind == c2go.Slice) && left.Base.Def().Is(c2go.Uint8) {
			x.Op = cc.Call
			x.List = []*cc.Expr{x.Left}
			x.Left = &cc.Expr{Op: cc.Name, Text: "len"}
			return intType
		}
		return nil

	case cc.SizeofType:
		return nil

	case cc.String:
		return &cc.Type{Kind: c2go.String}

	case cc.VaArg:
		// TODO
		return nil
	}
}

var (
	boolType   = &cc.Type{Kind: c2go.Bool}
	byteType   = &cc.Type{Kind: c2go.Byte}
	intType    = &cc.Type{Kind: c2go.Int}
	uintType   = &cc.Type{Kind: c2go.Uint}
	uint32Type = &cc.Type{Kind: c2go.Uint32}
	int64Type  = &cc.Type{Kind: c2go.Int64}
	uint64Type = &cc.Type{Kind: c2go.Uint64}
	idealType  = &cc.Type{Kind: c2go.Ideal}
	stringType = &cc.Type{Kind: c2go.String}
)

func forceGoType(fn *cc.Decl, x *cc.Expr, targ *cc.Type) {
	actual := fixGoTypesExpr(fn, x, targ)
	forceConvert(fn, x, actual, targ)
}

func forceConvert(fn *cc.Decl, x *cc.Expr, actual, targ *cc.Type) {
	if isNumericConst(x) && targ != nil {
		switch targ.Kind {
		case cc.Ptr, c2go.Slice:
			if x.Op == cc.Number && x.Text == "0" {
				x.Op = cc.Name
				x.Text = "nil"
				x.XType = targ
			}
		case c2go.String:
			if x.Op == cc.Number && x.Text == "0" {
				x.Op = cc.Name
				x.Text = `""`
				x.XType = targ
			}

		}
		return
	}

	if actual == nil || targ == nil {
		return
	}

	if actual.Kind == c2go.Ideal && c2go.Int8 <= targ.Kind && targ.Kind <= c2go.Float64 {
		return
	}

	if x != nil && x.Op == cc.Name && x.Text == "nil" {
		if targ.Kind == cc.Func || targ.Kind == cc.Ptr || targ.Kind == c2go.Slice {
			return
		}
	}

	// Func conversions are never useful.
	// If the func types are different, the conversion will fail;
	// if not, the conversion is unnecessary.
	// Either way the conversion is an eyesore.
	if targ.Kind == cc.Func || targ.Kind == cc.Ptr && targ.Base.Kind == cc.Func {
		return
	}

	if actual.Kind == c2go.Bool && c2go.Int8 <= targ.Kind && targ.Kind <= c2go.Float64 {
		old := copyExpr(x)
		if targ.Kind == c2go.Int {
			x.Op = cc.Call
			x.Left = &cc.Expr{Op: cc.Name, Text: "bool2int"}
			x.List = []*cc.Expr{old}
			x.Right = nil
		} else {
			x.Op = cc.Cast
			x.Left = &cc.Expr{Op: cc.Call, Left: &cc.Expr{Op: cc.Name, Text: "bool2int"}, List: []*cc.Expr{old}}
			x.Type = targ
		}
		return
	}

	if actual.Kind == cc.Array && targ.Kind == c2go.Slice && sameType(actual.Base, targ.Base) {
		old := copyExpr(x)
		x.Op = c2go.ExprSlice
		x.List = []*cc.Expr{old, nil, nil}
		x.Left = nil
		x.Right = nil
		return
	}

	if actual.Kind == c2go.Slice && targ.Kind == cc.Ptr && sameType(actual.Base, targ.Base) {
		old := copyExpr(x)
		x.Op = cc.Addr
		x.Left = &cc.Expr{Op: cc.Index, Left: old, Right: &cc.Expr{Op: cc.Number, Text: "0"}}
		return
	}

	if !sameType(actual, targ) {
		old := copyExpr(x)
		// for debugging:
		// old = &cc.Expr{Op: cc.Cast, Left: old, Type: actual, XType: actual}
		x.Op = cc.Cast
		x.Left = old
		x.Right = nil
		x.List = nil
		x.Type = targ
		x.XType = targ
		if actual.Kind == cc.Array && targ.Kind == c2go.Slice {
			x.Op = c2go.ExprSlice
			x.List = []*cc.Expr{old, nil, nil}
			x.Left = nil
			x.Type = nil
		}
	}
}

func isNumericConst(x *cc.Expr) bool {
	switch x.Op {
	case cc.Number:
		return true
	case cc.Name:
		// TODO
	case cc.Add, cc.And, cc.Div, cc.Mod, cc.Mul, cc.Or, cc.Sub, cc.Xor, cc.Lsh, cc.Rsh:
		return isNumericConst(x.Left) && isNumericConst(x.Right)
	case cc.Plus, cc.Minus, cc.Twid:
		return isNumericConst(x.Left)
	}
	return false
}

func fixShiftCount(fn *cc.Decl, x *cc.Expr) {
	typ := fixGoTypesExpr(fn, x, nil)
	if typ == nil {
		return
	}
	switch typ.Kind {
	case c2go.Uint8, c2go.Uint16, c2go.Uint32, c2go.Uint64, c2go.Uint, c2go.Uintptr, c2go.Byte:
		return
	}
	if typ.Kind == c2go.Int64 {
		forceConvert(fn, x, typ, uint64Type)
		return
	}
	forceConvert(fn, x, typ, uintType)
}

func fixBinary(fn *cc.Decl, x *cc.Expr, left, right, targ *cc.Type) *cc.Type {
	if left == nil || right == nil {
		return nil
	}

	if left.Kind != c2go.Ideal && (left.Kind < c2go.Int8 || left.Kind > c2go.Float64) {
		return nil
	}
	if right.Kind != c2go.Ideal && (right.Kind < c2go.Int8 || right.Kind > c2go.Float64) {
		return nil
	}

	if targ != nil && (targ.Kind < c2go.Int8 || targ.Kind > c2go.Float64) {
		return nil
	}

	// Want to do everything at as high a precision as possible for as long as possible.
	// If target is wider, convert early.
	// If target is narrower, don't convert at all - let next step do it.
	// Must make left and right match.
	// Convert to largest of three.
	t := left
	if t.Kind == c2go.Ideal || t.Kind < right.Kind && right.Kind != c2go.Ideal {
		t = right
	}
	if targ != nil && (t.Kind == c2go.Ideal || t.Kind < targ.Kind && targ.Kind != c2go.Ideal) {
		t = targ
	}
	if !sameType(t, left) {
		forceConvert(fn, x.Left, left, t)
	}
	if !sameType(t, right) {
		forceConvert(fn, x.Right, right, t)
	}
	return t
}

func sameType(t, u *cc.Type) bool {
	t = t.Def()
	u = u.Def()
	if t == u {
		return true
	}
	if t == nil || u == nil {
		return false
	}
	if t.Kind != u.Kind {
		return false
	}
	if t.Name != "" || u.Name != "" {
		return t.Name == u.Name
	}
	if !sameType(t.Base, u.Base) || len(t.Decls) != len(u.Decls) {
		return false
	}
	for i, td := range t.Decls {
		ud := u.Decls[i]
		if !sameType(td.Type, ud.Type) || t.Kind == cc.Struct && td.Name != ud.Name {
			return false
		}
	}
	return true
}

func isSliceOrString(typ *cc.Type) bool {
	return typ != nil && (typ.Kind == c2go.Slice || typ.Kind == c2go.String)
}

func isSliceStringOrArray(typ *cc.Type) bool {
	return typ != nil && (typ.Kind == c2go.Slice || typ.Kind == c2go.String || typ.Kind == cc.Array)
}

func isSliceOrPtr(typ *cc.Type) bool {
	return typ != nil && (typ.Kind == c2go.Slice || typ.Kind == cc.Ptr)
}

func isPtrSliceOrArray(typ *cc.Type) bool {
	return typ != nil && (typ.Kind == cc.Ptr || typ.Kind == cc.Array || typ.Kind == c2go.Slice)
}

func isSliceOrArray(typ *cc.Type) bool {
	return typ != nil && (typ.Kind == c2go.Slice || typ.Kind == cc.Array)
}

func fixSpecialCall(fn *cc.Decl, x *cc.Expr, targ *cc.Type) bool {
	if x.Left.Op != cc.Name {
		return false
	}
	switch x.Left.Text {
	case "memmove":
		if len(x.List) != 3 {
			fprintf(x.Span, "unsupported %v", x)
			return false
		}
		siz := x.List[2]
		if siz.Op == cc.Number && siz.Text == "4" {
			obj1, obj1Type := objIndir(fn, x.List[0])
			obj2, obj2Type := objIndir(fn, x.List[1])
			if obj1Type == nil || obj2Type == nil {
				fprintf(x.Span, "unsupported %v - missing types", x)
				return true
			}
			if (obj1Type.Kind == c2go.Uint32 || obj1Type.Kind == c2go.Int32) && obj2Type.Kind == c2go.Float32 {
				x.Op = cc.Eq
				x.Left = obj1
				x.Right = &cc.Expr{
					Op: cc.Call,
					Left: &cc.Expr{Op: cc.Name,
						Text: "math.Float32bits",
					},
					List: []*cc.Expr{obj2},
				}
				x.XType = uint32Type
				return true
			}
			fprintf(x.Span, "unsupported %v - size 4 type %v %v", x, c2go.GoString(obj1Type), c2go.GoString(obj2Type))
		}
		if siz.Op == cc.Number && siz.Text == "8" {
			obj1, obj1Type := objIndir(fn, x.List[0])
			obj2, obj2Type := objIndir(fn, x.List[1])
			if obj1Type == nil || obj2Type == nil {
				fprintf(x.Span, "unsupported %v - missing types", x)
				return true
			}
			if (obj1Type.Kind == c2go.Uint64 || obj1Type.Kind == c2go.Int64) && obj2Type.Kind == c2go.Float64 {
				x.Op = cc.Eq
				x.Left = obj1
				x.Right = &cc.Expr{
					Op: cc.Call,
					Left: &cc.Expr{Op: cc.Name,
						Text: "math.Float64bits",
					},
					List: []*cc.Expr{obj2},
				}
				x.XType = uint64Type
				return true
			}
			fprintf(x.Span, "unsupported %v - size 8 type %v %v", x, c2go.GoString(obj1Type), c2go.GoString(obj2Type))
		}
		if siz.Op == cc.SizeofExpr {
			obj1Type := fixGoTypesExpr(fn, x.List[0], nil)
			obj2Type := fixGoTypesExpr(fn, x.List[1], nil)
			sizeType := fixGoTypesExpr(fn, siz.Left, nil)
			if obj1Type == nil || obj2Type == nil {
				fprintf(x.Span, "unsupported %v - bad types", x)
				return true
			}
			if obj2Type.Kind == cc.Array && sameType(obj2Type, sizeType) || obj2Type.Kind == c2go.Slice && c2go.GoString(x.List[1]) == c2go.GoString(siz.Left) {
				x.Left.Text = "copy"
				x.Left.XDecl = nil
				x.List = x.List[:2]
				return true
			}
			fprintf(x.Span, "unsupported %v - not array %v %v", x, c2go.GoString(obj2Type), c2go.GoString(sizeType))
			return true
		}
		left := fixGoTypesExpr(fn, x.List[0], nil)
		right := fixGoTypesExpr(fn, x.List[1], nil)
		fixGoTypesExpr(fn, siz, nil)
		if isSliceOrArray(left) && isSliceOrArray(right) && left.Base.Is(c2go.Uint8) && right.Base.Is(c2go.Uint8) {
			x.Left.Text = "copy"
			x.Left.XDecl = nil
			if x.List[1].Op == c2go.ExprSlice && x.List[1].List[1] == nil {
				x.List[1].List[2] = siz
			} else {
				x.List[1] = &cc.Expr{Op: c2go.ExprSlice, List: []*cc.Expr{x.List[1], nil, siz}}
			}
			x.List = x.List[:2]
			return true
		}
		fprintf(x.Span, "unsupported %v (%v %v)", x, c2go.GoString(left), c2go.GoString(right))
		return true

	case "mal", "malloc", "emallocz":
		if len(x.List) != 1 {
			fprintf(x.Span, "unsupported %v - too many args", x)
			return false
		}
		siz := x.List[0]
		var count *cc.Expr
		if siz.Op == cc.Mul {
			count = siz.Left
			siz = siz.Right
		}
		var typ *cc.Type
		switch siz.Op {
		default:
			typ = byteType
			count = siz

		case cc.SizeofExpr:
			typ = fixGoTypesExpr(fn, siz.Left, nil)
			if typ == nil {
				fprintf(siz.Span, "failed to type check %v", siz.Left)
			}

		case cc.SizeofType:
			typ = siz.Type
			if typ == nil {
				fprintf(siz.Span, "sizeoftype missing type")
			}
		}
		if typ == nil {
			fprintf(x.Span, "unsupported %v - cannot understand type", x)
			return true
		}
		if count == nil {
			x.Left.Text = "new"
			x.Left.XDecl = nil
			x.List = []*cc.Expr{&cc.Expr{Op: cc.Name, Text: c2go.GoString(typ)}}
			x.XType = &cc.Type{Kind: cc.Ptr, Base: typ}
		} else {
			x.Left.Text = "make"
			x.Left.XDecl = nil
			x.List = []*cc.Expr{
				&cc.Expr{Op: cc.Name, Text: "[]" + c2go.GoString(typ)},
				count,
			}
			x.XType = &cc.Type{Kind: c2go.Slice, Base: typ}
		}
		return true

	case "strdup", "estrdup":
		if len(x.List) != 1 {
			fprintf(x.Span, "unsupported %v - too many args", x)
			return false
		}
		fixGoTypesExpr(fn, x.List[0], stringType)
		fixMerge(x, x.List[0])
		x.XType = stringType
		return true

	case "strcpy", "strcat", "fmtstrcpy":
		if len(x.List) != 2 {
			fprintf(x.Span, "unsupported %v - too many args", x)
			return false
		}
		fixGoTypesExpr(fn, x.List[0], nil)
		fixGoTypesExpr(fn, x.List[1], stringType)
		x.Op = cc.Eq
		if x.Left.Text == "strcat" || x.Left.Text == "fmtstrcpy" {
			x.Op = cc.AddEq
		}
		x.Left = x.List[0]
		x.Right = x.List[1]
		x.XType = stringType
		return true

	case "strlen":
		x.Left.Text = "len"
		x.Left.XDecl = nil
		x.XType = intType
		return true

	case "TUP", "CASE":
		if len(x.List) != 2 {
			fprintf(x.Span, "unsupported %v - too many args", x)
			return false
		}
		left := fixGoTypesExpr(fn, x.List[0], targ)
		right := fixGoTypesExpr(fn, x.List[1], targ)
		x.Op = cc.Or
		x.Left = &cc.Expr{Op: cc.Lsh, Left: x.List[0], Right: &cc.Expr{Op: cc.Number, Text: "16"}, XType: left}
		x.Right = x.List[1]
		x.List = nil
		x.XType = fixBinary(fn, x, left, right, targ)
		return true
	}

	return false
}

func fixMemset(prog *cc.Prog, fn *cc.Decl, stmt *cc.Stmt) {
	x := stmt.Expr
	if len(x.List) != 3 || x.List[1].Op != cc.Number || x.List[1].Text != "0" {
		fprintf(x.Span, "unsupported %v - nonzero", x)
		return
	}

	if x.List[2].Op == cc.SizeofExpr || x.List[2].Op == cc.SizeofType {
		obj, objType := objIndir(fn, x.List[0])
		if !matchSize(fn, obj, objType, x.List[2]) {
			fprintf(x.Span, "unsupported %v - wrong size", x)
			return
		}

		x.Op = cc.Eq
		x.Left = obj
		x.Right = zeroFor(objType)
		x.List = nil
		return
	}

	siz := x.List[2]
	var count *cc.Expr
	var objType *cc.Type
	if siz.Op == cc.Mul {
		count = siz.Left
		siz = siz.Right
		if siz.Op != cc.SizeofExpr && siz.Op != cc.SizeofType {
			fprintf(x.Span, "unsupported %v - wrong array size", x)
			return
		}

		switch siz.Op {
		case cc.SizeofExpr:
			p := unparen(siz.Left)
			if p.Op != cc.Indir && p.Op != cc.Index || !sameType(p.Left.XType, x.List[0].XType) {
				fprintf(x.Span, "unsupported %v - wrong size", x)
			}
			objType = fixGoTypesExpr(fn, x.List[0], nil)
		case cc.SizeofType:
			objType = fixGoTypesExpr(fn, x.List[0], nil)
			if !sameType(siz.Type, objType.Base) {
				fprintf(x.Span, "unsupported %v - wrong size", x)
			}
		}
	} else {
		count = siz
		objType = fixGoTypesExpr(fn, x.List[0], nil)
		if !objType.Base.Is(c2go.Byte) && !objType.Base.Is(c2go.Uint8) {
			fprintf(x.Span, "unsupported %v - wrong size form for non-byte type", x)
			return
		}
	}

	// Found it. Replace with zeroing for loop.
	stmt.Op = cc.For
	stmt.Pre = &cc.Expr{
		Op: cc.Eq,
		Left: &cc.Expr{
			Op:    cc.Name,
			Text:  "i",
			XType: intType,
		},
		Right: &cc.Expr{
			Op:    cc.Number,
			Text:  "0",
			XType: intType,
		},
		XType: boolType,
	}
	stmt.Expr = &cc.Expr{
		Op: cc.Lt,
		Left: &cc.Expr{
			Op:    cc.Name,
			Text:  "i",
			XType: intType,
		},
		Right: count,
		XType: boolType,
	}
	stmt.Post = &cc.Expr{
		Op: cc.PostInc,
		Left: &cc.Expr{
			Op:    cc.Name,
			Text:  "i",
			XType: intType,
		},
		XType: intType,
	}
	stmt.Body = &cc.Stmt{
		Op: cc.Block,
		Block: []*cc.Stmt{
			{
				Op: cc.StmtExpr,
				Expr: &cc.Expr{
					Op: cc.Eq,
					Left: &cc.Expr{
						Op:    cc.Index,
						Left:  x.List[0],
						Right: &cc.Expr{Op: cc.Name, Text: "i"},
					},
					Right: zeroFor(objType.Base),
				},
			},
		},
	}
	return
}

func fixSpecialCompare(fn *cc.Decl, x *cc.Expr) bool {
	if x.Right.Op != cc.Number || x.Right.Text != "0" || x.Left.Op != cc.Call || x.Left.Left.Op != cc.Name {
		return false
	}

	call := x.Left
	switch call.Left.Text {
	case "memcmp":
		if len(call.List) != 3 {
			fprintf(x.Span, "unsupported %v", x)
			return false
		}
		obj1, obj1Type := objIndir(fn, call.List[0])
		obj2, obj2Type := objIndir(fn, call.List[1])
		if obj1Type == nil || !sameType(obj1Type, obj2Type) {
			fprintf(x.Span, "unsupported %v", call)
			return true
		}

		if !matchSize(fn, obj1, obj1Type, call.List[2]) && !matchSize(fn, obj2, obj2Type, call.List[2]) {
			fprintf(x.Span, "unsupported %v - wrong size", call)
			return true
		}

		x.Left = obj1
		x.Right = obj2
		x.List = nil
		x.XType = boolType
		return true

	case "strcmp":
		if len(call.List) != 2 {
			fprintf(x.Span, "unsupported %v", x)
			return false
		}
		obj1 := call.List[0]
		obj2 := call.List[1]

		x.Left = obj1
		x.Right = obj2
		x.List = nil
		x.XType = boolType
		return true
	}

	return false
}

func objIndir(fn *cc.Decl, x *cc.Expr) (*cc.Expr, *cc.Type) {
	objType := fixGoTypesExpr(fn, x, nil)
	obj := x
	if obj.XType != nil && obj.XType.Kind == cc.Array {
		// obj stays as is
	} else if obj.Op == cc.Addr {
		obj = obj.Left
		if objType != nil {
			objType = objType.Base
		}
	} else {
		obj = &cc.Expr{Op: cc.Indir, Left: obj}
		if objType != nil {
			objType = objType.Base
		}
	}
	if objType == nil {
		objType = obj.XType
	}
	return obj, objType
}

func matchSize(fn *cc.Decl, obj *cc.Expr, objType *cc.Type, siz *cc.Expr) bool {
	switch siz.Op {
	default:
		return false

	case cc.SizeofType:
		// ok if sizeof type of first arg
		return sameType(siz.Type, objType)

	case cc.SizeofExpr:
		// ok if sizeof *firstarg
		y := siz.Left
		if y.Op == cc.Paren {
			y = y.Left
		}
		return obj.String() == y.String()
	}
}

func fixBools(prog *cc.Prog) {
	cc.Preorder(prog, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Expr:
			if x.Op == cc.Not {
				switch x.Left.Op {
				case cc.OrOr, cc.AndAnd, cc.EqEq, cc.NotEq, cc.LtEq, cc.GtEq, cc.Lt, cc.Gt, cc.Paren:
					x.Left = negate(x.Left)
					fixMerge(x, x.Left)
				}
			}
		}
	})
}

func negate(x *cc.Expr) *cc.Expr {
	switch x.Op {
	case cc.Paren:
		return negate(x.Left)
	case cc.OrOr:
		x.Op = cc.AndAnd
		x.Left = negate(x.Left)
		x.Right = negate(x.Right)
	case cc.AndAnd:
		x.Op = cc.OrOr
		x.Left = negate(x.Left)
		x.Right = negate(x.Right)
	case cc.EqEq:
		x.Op = cc.NotEq
	case cc.NotEq:
		x.Op = cc.EqEq
	case cc.Lt:
		x.Op = cc.GtEq
	case cc.GtEq:
		x.Op = cc.Lt
	case cc.Gt:
		x.Op = cc.LtEq
	case cc.LtEq:
		x.Op = cc.Gt
	default:
		x = &cc.Expr{Op: cc.Not, Left: x}
	}
	return x
}

func rewriteLen(cfg *Config, prog *cc.Prog) {
	cc.Postorder(prog, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Expr:
			// Assign to len, generated by len rewrite. Change to reslice.
			if x.Op == cc.Eq && x.Left.Op == cc.Call && x.Left.Left.Op == cc.Name && x.Left.Left.Text == "len" && len(x.Left.List) > 0 {
				arg := x.Left.List[0]
				x.Left = arg
				x.Right = &cc.Expr{
					Op:   c2go.ExprSlice,
					List: []*cc.Expr{arg, nil, x.Right},
				}
				return
			}

			if (x.Op == cc.Arrow || x.Op == cc.Dot) && x.XDecl != nil {
				k := declKey(x.XDecl)
				name := cfg.Len[k]
				if name == "" {
					return
				}
				d := x.XDecl
				if d.OuterType == nil {
					fmt.Fprintf(os.Stderr, "found use of %s but missing type\n", k)
					return
				}
				t := d.OuterType
				var other *cc.Decl
				for _, dd := range t.Decls {
					if dd.Name == name {
						other = dd
						break
					}
				}
				if other == nil {
					fmt.Fprintf(os.Stderr, "found use of %s but cannot find field %s\n", k, name)
					return
				}
				left := x.Left
				x.Op = cc.Call
				x.Left = &cc.Expr{
					Op:    cc.Name,
					Text:  "len",
					XType: &cc.Type{Kind: cc.Func, Base: intType},
				}
				x.List = []*cc.Expr{
					{
						Op:    cc.Dot,
						Left:  left,
						Text:  name,
						XDecl: other,
					},
				}
			}
		}
	})

	cc.Postorder(prog, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Type:
			out := x.Decls[:0]
			for _, d := range x.Decls {
				k := declKey(d)
				if cfg.Len[k] == "" && !cfg.Delete[k] {
					out = append(out, d)
				}
			}
			x.Decls = out
		}
	})
}

func finalEdits(prog *cc.Prog) {
	cc.Postorder(prog, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Expr:
			if x.Op == cc.Arrow && x.Text == "prg" && c2go.GoString(x.Left) == "ctxt.arch" {
				x.Left = x.Left.Left // drop arch.
			}
		}
	})
}

func doExports(cfg *Config, prog *cc.Prog) {
	for _, d := range cfg.TopDecls {
		if shouldExport(cfg, d.Name) {
			exportDecl(d)
		}
		pkg := d.GoPackage
		if pkg == "" {
			continue
		}
		cc.Preorder(d, func(x cc.Syntax) {
			switch x := x.(type) {
			case *cc.Expr:
				if x.Op == cc.Name && x.XDecl != nil && x.XDecl.GoPackage != "" && x.XDecl.GoPackage != pkg {
					exportDecl(x.XDecl)
				}
			case *cc.Type:
				if x.Kind == cc.TypedefType && x.TypeDecl != nil && x.TypeDecl.GoPackage != "" && x.TypeDecl.GoPackage != pkg {
					exportDecl(x.TypeDecl)
					x.Name = x.TypeDecl.Name
				}
			}
		})
	}
}

func shouldExport(cfg *Config, name string) bool {
	for _, s := range cfg.Exports {
		if s == name {
			return true
		}
	}
	return false
}

func exportDecl(d *cc.Decl) {
	d.Name = exportName(d.Name)
	if d.Storage&cc.Typedef != 0 && d.Type.Kind == cc.Struct {
		for _, dd := range d.Type.Decls {
			exportDecl(dd)
			if dd.Name == "U" {
				for _, dd := range dd.Type.Decls {
					exportDecl(dd)
				}
			}
		}
	}
}

func exportName(name string) string {
	_, size := utf8.DecodeRuneInString(name)
	return strings.ToUpper(name[:size]) + name[size:]
}
