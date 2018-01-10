// +build ignore

package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"code.google.com/p/rsc/c2go"
	"code.google.com/p/rsc/cc"
)

func tryPrintf(curfn *cc.Decl, x *cc.Expr, name string, fmtpos int, newName string) bool {
	if (x.Left.Text == name || strings.Contains(name, ".") && x.Left.String() == name) && len(x.List) >= fmtpos+1 && x.List[fmtpos].Op == cc.String && len(x.List[fmtpos].Texts) == 1 {
		fixPrintFormat(curfn, x.List[fmtpos], x.List[fmtpos+1:])
		x.Left.Text = newName
		x.Left.XDecl = nil
		return true
	}
	return false
}

func fixPrintf(curfn *cc.Decl, x *cc.Expr) bool {
	if x.Op != cc.Call {
		return false
	}
	if tryPrintf(curfn, x, "sprint", 1, "fmt.Sprintf") {
		targ := x.List[0]
		x.List = x.List[1:]
		x.Right = copyExpr(x)
		x.List = nil
		x.Left = targ
		x.Op = cc.Eq

		// sprint(strchr(s, 0), "hello") => s += "hello"
		if targ.Op == cc.Call && len(targ.List) == 2 && targ.Left.Text == "strchr" && targ.List[1].Text == "0" {
			x.Left = targ.List[0]
			x.Op = cc.AddEq
		} else if targ.Op == cc.Add && targ.Right.Op == cc.Call && targ.Right.Left.Text == "strlen" && len(targ.Right.List) == 1 && c2go.GoString(targ.Left) == c2go.GoString(targ.Right.List[0]) {
			x.Left = targ.Left
			x.Op = cc.AddEq
		}
		return true
	}
	if tryPrintf(curfn, x, "snprint", 2, "fmt.Sprintf") {
		targ := x.List[0]
		x.List = x.List[2:]
		x.Right = copyExpr(x)
		x.List = nil
		x.Left = targ
		x.Op = cc.Eq
		return true
	}
	if tryPrintf(curfn, x, "fmtprint", 1, "fmt.Sprintf") {
		targ := x.List[0]
		x.List = x.List[1:]
		x.Right = copyExpr(x)
		x.List = nil
		x.Left = targ
		x.Op = cc.AddEq
		return true
	}
	if tryPrintf(curfn, x, "print", 0, "fmt.Printf") {
		return true
	}
	if tryPrintf(curfn, x, "sysfatal", 0, "log.Fatalf") {
		return true
	}
	if tryPrintf(curfn, x, "ctxt.diag", 0, "ctxt.diag") {
		return true
	}
	if tryPrintf(curfn, x, "Bprint", 1, "fmt.Fprintf") {
		return true
	}
	if tryPrintf(curfn, x, "smprint", 0, "fmt.Sprintf") {
		return true
	}

	return false
}

func fixPrintFormat(curfn *cc.Decl, fx *cc.Expr, args []*cc.Expr) {
	for _, arg := range args {
		fixGoTypesExpr(curfn, arg, nil)
		cc.Preorder(arg, func(x cc.Syntax) {
			if x, ok := x.(*cc.Expr); ok && x.Op == cc.Name && strings.HasPrefix(x.Text, "bigP") {
				x.Text = "p"
				x.XDecl = nil
			}
		})
	}

	format, err := strconv.Unquote(fx.Texts[0])
	if err != nil {
		fprintf(fx.Span, "cannot parse quoted string: %v", err)
		return
	}

	suffix := ""

	var buf bytes.Buffer
	start := 0
	narg := 0
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		buf.WriteString(format[start:i])
		start = i
		i++
		for i < len(format) {
			c := format[i]
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '#', '-', '.', ',', ' ', 'h', 'l', 'u':
				i++
				continue
			}
			break
		}
		if i >= len(format) {
			fprintf(fx.Span, "print format ends mid-verb")
			return
		}
		flags, verb := format[start:i], format[i]
		start = i + 1

		allFlags := flags
		_ = allFlags

		flags = strings.Replace(flags, "h", "", -1)
		flags = strings.Replace(flags, "l", "", -1)
		flags = strings.Replace(flags, "u", "", -1)

		if j := strings.Index(flags, "#0"); j >= 0 && verb == 'x' {
			k := j + 2
			for k < len(flags) && '0' <= flags[k] && flags[k] <= '9' {
				k++
			}
			n, _ := strconv.Atoi(flags[j+2 : k])
			flags = flags[:j+2] + fmt.Sprint(n-2) + flags[k:]
		}

		convert := ""
		switch verb {
		default:
			fprintf(fx.Span, "unrecognized format %s%c", flags, verb)
			buf.WriteString("%")
			buf.WriteString(flags)
			buf.WriteString(string(verb))

		case 'f', 'e', 'E', 'g', 'G', 's', 'c', 'p':
			// usual meanings
			buf.WriteString(flags)
			buf.WriteString(string(verb))

		case 'x', 'X', 'o', 'd', 'b':
			// usual meanings, but force unsigned if u is given
			buf.WriteString(flags)
			buf.WriteString(string(verb))
			if narg >= len(args) || !strings.Contains(allFlags, "u") {
				break
			}
			arg := args[narg]
			if t := arg.XType.Def(); t != nil {
				switch t.Kind {
				case c2go.Int8:
					convert = "uint8"
				case c2go.Int16:
					convert = "uint16"
				case c2go.Int32:
					convert = "uint32"
				case c2go.Int64:
					convert = "uint64"
				case c2go.Int:
					convert = "uint"
				}
			}

		case 'A': // address
			if allFlags != "%" {
				fprintf(fx.Span, "format %s%c", allFlags, verb)
			}
			buf.WriteString("%v")
			convert = "Aconv" + suffix

		case 'L':
			if allFlags != "%" {
				fprintf(fx.Span, "format %s%c", allFlags, verb)
			}
			buf.WriteString("%v")
			if narg >= len(args) {
				break
			}
			arg := args[narg]
			if (arg.Op == cc.Dot || arg.Op == cc.Arrow) && arg.Text == "lineno" {
				arg.Text = "Line"
				arg.XDecl = nil
				args[narg] = &cc.Expr{Op: cc.Call, Left: arg, XType: stringType}
				break
			}
			convert = "ctxt.Line"

		case '@':
			if allFlags != "%" {
				fprintf(fx.Span, "format %s%c", allFlags, verb)
			}
			buf.WriteString("%v")
			convert = "RAconv" + suffix

		case 'D':
			if allFlags != "%" && allFlags != "%l" {
				fprintf(fx.Span, "format %s%c", allFlags, verb)
			}
			buf.WriteString("%v")
			if narg >= len(args) {
				break
			}
			flag := &cc.Expr{Op: cc.Name, Text: "0"}
			if strings.Contains(allFlags, "l") {
				flag.Text = "fmtLong"
			}
			arg := args[narg]
			args[narg] = &cc.Expr{
				Op:   cc.Call,
				Left: &cc.Expr{Op: cc.Name, Text: "Dconv" + suffix},
				List: []*cc.Expr{
					&cc.Expr{Op: cc.Name, Text: "p"},
					flag,
					arg,
				},
				XType: stringType,
			}

		case 'M':
			if allFlags != "%" {
				fprintf(fx.Span, "format %s%c", allFlags, verb)
			}
			buf.WriteString("%v")
			convert = "Mconv" + suffix

		case 'R':
			if allFlags != "%" {
				fprintf(fx.Span, "format %s%c", allFlags, verb)
			}
			buf.WriteString("%v")
			convert = "Rconv" + suffix

		case '$':
			if allFlags != "%" {
				fprintf(fx.Span, "format %s%c", allFlags, verb)
			}
			buf.WriteString("%q")

		case 'P':
			if allFlags != "%" {
				fprintf(fx.Span, "format %s%c", allFlags, verb)
			}
			buf.WriteString("%v")
		}

		if convert != "" && narg < len(args) {
			arg := args[narg]
			args[narg] = &cc.Expr{Op: cc.Call, Left: &cc.Expr{Op: cc.Name, Text: convert}, List: []*cc.Expr{arg}, XType: stringType}
		}

		if verb != 'r' {
			narg++
		}
	}
	buf.WriteString(format[start:])

	fx.Texts[0] = strconv.Quote(buf.String())
}

func fixFormatter(fn *cc.Decl) {
	// Find va_arg assignment.
	var arg *cc.Expr
	var argType *cc.Type
	var ps []*cc.Expr
	cc.Preorder(fn.Body, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Expr:
			if x.Op == cc.Name && strings.HasPrefix(x.Text, "bigP") {
				ps = append(ps, x)
			}
		case *cc.Stmt:
			stmt := x
			if stmt.Op != cc.StmtExpr {
				return
			}
			expr := stmt.Expr
			if expr.Op != cc.Eq {
				return
			}
			if expr.Left.Op == cc.Name && strings.HasPrefix(expr.Left.Text, "bigP") {
				stmt.Op = cc.Empty
				stmt.Expr = nil
				return
			}
			if expr.Op != cc.Eq || expr.Right.Op != cc.VaArg {
				return
			}
			if arg != nil {
				fprintf(fn.Span, "multiple va_arg in formatter")
			}
			arg = expr.Left
			argType = expr.Right.Type
			stmt.Op = cc.Empty
		}
	})

	if arg == nil {
		fprintf(fn.Span, "missing va_arg in formatter")
		return
	}

	fp := fn.Type.Decls[0]
	fp.Type = stringType
	fn.Type.Decls[0] = arg.XDecl
	fn.Type.Base = stringType

	if strings.HasPrefix(fn.Name, "Dconv") && len(ps) > 0 {
		pd := &cc.Decl{Name: "p", Type: ps[0].XDecl.Type}
		fd := &cc.Decl{Name: "flag", Type: intType}
		for _, p := range ps {
			p.XDecl = pd
		}
		fn.Type.Decls = []*cc.Decl{
			pd,
			fd,
			arg.XDecl,
		}
	}

	cc.Preorder(fn.Body, func(x cc.Syntax) {
		switch x := x.(type) {
		case *cc.Stmt:
			if x.Op == cc.StmtDecl && x.Decl == arg.XDecl {
				x.Decl = fp
			}

			if x.Op == cc.Return && x.Expr != nil && x.Expr.Text == "0" {
				x.Expr = &cc.Expr{Op: cc.Name, Text: fp.Name, XDecl: fp}
			}

		case *cc.Expr:
			if x.Op == cc.Arrow && x.Text == "flags" {
				x.Op = cc.Name
				x.Text = "flag"
				x.XDecl = nil
				x.Left = nil
			}

			if x.Op == cc.Name {
				switch x.Text {
				case "FmtLong":
					x.Text = "fmtLong"
					x.XDecl = nil
				}
			}
		}
	})

}
