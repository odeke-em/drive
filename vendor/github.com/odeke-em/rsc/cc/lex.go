// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// A Syntax represents any syntax element.
type Syntax interface {
	// GetSpan returns the start and end position of the syntax,
	// excluding leading or trailing comments.
	// The use of a Get prefix is non-standard but avoids a conflict
	// with the field named Span in most implementations.
	GetSpan() Span

	// GetComments returns the comments attached to the syntax.
	// This method would normally be named 'Comments' but that
	// would interfere with embedding a type of the same name.
	// The use of a Get prefix is non-standard but avoids a conflict
	// with the field named Comments in most implementations.
	GetComments() *Comments
}

// SyntaxInfo contains metadata about a piece of syntax.
type SyntaxInfo struct {
	Span     Span // location of syntax in input
	Comments Comments
}

func (s *SyntaxInfo) GetSpan() Span {
	return s.Span
}

func (s *SyntaxInfo) GetComments() *Comments {
	return &s.Comments
}

// Comments collects the comments associated with a syntax element.
type Comments struct {
	Before []Comment // whole-line comments before this syntax
	Suffix []Comment // end-of-line comments after this syntax

	// For top-level syntax elements only, After lists whole-line
	// comments following the syntax.
	After []Comment
}

type lexer struct {
	// input
	start int
	byte  int
	lexInput
	pushed      []lexInput
	forcePos    Pos
	c2goComment bool // inside /*c2go ... */ comment
	comments    []Comment

	// comment assignment
	pre      []Syntax
	post     []Syntax
	enumSeen map[interface{}]bool

	// type checking state
	scope       *Scope
	includeSeen map[string]*Header

	// output
	errors []string
	prog   *Prog
	expr   *Expr
}

type Header struct {
	decls []*Decl
	types []*Type
}

func (lx *lexer) parse() {
	if lx.includeSeen == nil {
		lx.includeSeen = make(map[string]*Header)
	}
	if lx.wholeInput == "" {
		lx.wholeInput = lx.input
	}
	lx.scope = &Scope{}
	yyParse(lx)
}

type lexInput struct {
	wholeInput string
	input      string
	tok        string
	lastsym    string
	file       string
	lineno     int
	declSave   *Header
}

func (lx *lexer) pushInclude(includeLine string) {
	s := strings.TrimSpace(strings.TrimPrefix(includeLine, "#include"))
	if !strings.HasPrefix(s, "<") && !strings.HasPrefix(s, "\"") {
		lx.Errorf("malformed #include")
		return
	}
	sep := ">"
	if s[0] == '"' {
		sep = "\""
	}
	i := strings.Index(s[1:], sep)
	if i < 0 {
		lx.Errorf("malformed #include")
		return
	}
	i++

	file := s[1:i]
	origFile := file

	file, data, err := lx.findInclude(file, s[0] == '<')
	if err != nil {
		lx.Errorf("#include %s: %v", s[:i+1], err)
		return
	}

	if hdr := lx.includeSeen[file]; hdr != nil {
		for _, decl := range hdr.decls {
			// fmt.Printf("%s: replay %s\n", file, decl.Name)
			lx.pushDecl(decl)
		}
		for _, typ := range hdr.types {
			lx.pushType(typ)
		}
		return
	}

	hdr := lx.declSave
	if hdr == nil {
		hdr = new(Header)
		lx.includeSeen[file] = hdr
	}

	if extraMap[origFile] != "" {
		str := extraMap[origFile] + "\n"
		lx.pushed = append(lx.pushed, lx.lexInput)
		lx.lexInput = lexInput{
			input:      str,
			wholeInput: str,
			file:       "internal/" + origFile,
			lineno:     1,
			declSave:   hdr,
		}
	}

	if data == nil {
		return
	}

	lx.pushed = append(lx.pushed, lx.lexInput)
	str := string(append(data, '\n'))
	lx.lexInput = lexInput{
		input:      str,
		wholeInput: str,
		file:       file,
		lineno:     1,
		declSave:   hdr,
	}
}

var stdMap = map[string]string{
	"u.h":      hdr_u_h,
	"libc.h":   hdr_libc_h,
	"stdarg.h": "",
}

var extraMap = map[string]string{
	"go.h": hdr_extra_go_h,
}

var includes []string

func AddInclude(dir string) {
	includes = append(includes, dir)
}

func (lx *lexer) findInclude(name string, std bool) (string, []byte, error) {
	if std {
		if redir, ok := stdMap[name]; ok {
			if redir == "" {
				return "", nil, nil
			}
			return "internal/" + name, []byte(redir), nil
		}
		name = "/Users/rsc/g/go/include/" + name
	}
	if !filepath.IsAbs(name) {
		name1 := filepath.Join(filepath.Dir(lx.file), name)
		if _, err := os.Stat(name1); err != nil {
			for _, dir := range includes {
				name2 := filepath.Join(dir, name)
				if _, err := os.Stat(name2); err == nil {
					name1 = name2
					break
				}
			}
		}
		name = name1
	}
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return "", nil, err
	}
	return name, data, nil
}

func (lx *lexer) pop() bool {
	if len(lx.pushed) == 0 {
		return false
	}
	lx.lexInput = lx.pushed[len(lx.pushed)-1]
	lx.pushed = lx.pushed[:len(lx.pushed)-1]
	return true
}

func (lx *lexer) pos() Pos {
	if lx.forcePos.Line != 0 {
		return lx.forcePos
	}
	return Pos{lx.file, lx.lineno, lx.byte}
}
func (lx *lexer) span() Span {
	p := lx.pos()
	return Span{p, p}
}

func (lx *lexer) setSpan(s Span) {
	lx.forcePos = s.Start
}

func span(l1, l2 Span) Span {
	if l1.Start.Line == 0 {
		return l2
	}
	if l2.Start.Line == 0 {
		return l1
	}
	return Span{l1.Start, l2.End}
}

func (lx *lexer) skip(i int) {
	lx.lineno += strings.Count(lx.input[:i], "\n")
	lx.input = lx.input[i:]
	lx.byte += i
}

func (lx *lexer) token(i int) {
	lx.tok = lx.input[:i]
	lx.skip(i)
}

func (lx *lexer) sym(i int) {
	lx.token(i)
	lx.lastsym = lx.tok
}

func (lx *lexer) comment(i int) {
	var c Comment
	c.Span.Start = lx.pos()
	c.Text = lx.input[:i]
	j := len(lx.wholeInput) - len(lx.input)
	for j > 0 && (lx.wholeInput[j-1] == ' ' || lx.wholeInput[j-1] == '\t') {
		j--
	}
	if j > 0 && lx.wholeInput[j-1] != '\n' {
		c.Suffix = true
	}
	prefix := lx.wholeInput[j : len(lx.wholeInput)-len(lx.input)]
	lines := strings.Split(c.Text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = line[len(prefix):]
		}
	}
	c.Text = strings.Join(lines, "\n")

	lx.skip(i)
	c.Span.End = lx.pos()
	lx.comments = append(lx.comments, c)
}

func isalpha(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || c == '_' || c >= 0x80 || '0' <= c && c <= '9'
}

func isspace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}

func (lx *lexer) setEnd(yy *yySymType) {
	yy.span.End = lx.pos()
}

func (lx *lexer) Lex(yy *yySymType) (yyt int) {
	//defer func() { println("tok", yy.str, yyt) }()
	if lx.start != 0 {
		tok := lx.start
		lx.start = 0
		return tok
	}
	*yy = yySymType{}
	defer lx.setEnd(yy)
Restart:
	yy.span.Start = lx.pos()
	in := lx.input
	if len(in) == 0 {
		if lx.pop() {
			goto Restart
		}
		return tokEOF
	}
	c := in[0]
	if isspace(c) {
		i := 1
		for i < len(in) && isspace(in[i]) {
			i++
		}
		lx.skip(i)
		goto Restart
	}

	i := 0
	switch c {
	case '#':
		i++
		for in[i] != '\n' {
			if in[i] == '\\' && in[i+1] == '\n' && i+2 < len(in) {
				i++
			}
			i++
		}
		str := in[:i]
		lx.skip(i)
		if strings.HasPrefix(str, "#include") {
			lx.pushInclude(str)
		}
		goto Restart

	case 'L':
		if in[1] != '\'' && in[1] != '"' {
			break // goes to alpha case after switch
		}
		i = 1
		fallthrough

	case '"', '\'':
		q := in[i]
		i++ // for the quote
		for ; in[i] != q; i++ {
			if in[i] == '\n' {
				what := "string"
				if q == '\'' {
					what = "character"
				}
				lx.Errorf("unterminated %s constant", what)
			}
			if in[i] == '\\' {
				i++
			}
		}
		i++ // for the quote
		lx.sym(i)
		yy.str = lx.tok
		if q == '"' {
			return tokString
		} else {
			return tokLitChar
		}

	case '.':
		if in[1] < '0' || '9' < in[1] {
			if in[1] == '.' && in[2] == '.' {
				lx.token(3)
				return tokDotDotDot
			}
			lx.token(1)
			return int(c)
		}
		fallthrough

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		for '0' <= in[i] && in[i] <= '9' || in[i] == '.' || 'A' <= in[i] && in[i] <= 'Z' || 'a' <= in[i] && in[i] <= 'z' {
			i++
		}
		lx.sym(i)
		yy.str = lx.tok
		return tokNumber

	case '/':
		switch in[1] {
		case '*':
			if strings.HasPrefix(in, "/*c2go") {
				lx.skip(6)
				lx.c2goComment = true
				goto Restart
			}
			i := 2
			for ; ; i++ {
				if i+2 <= len(in) && in[i] == '*' && in[i+1] == '/' {
					i += 2
					break
				}
				if i >= len(in) {
					lx.Errorf("unterminated /* comment")
					return tokError
				}
			}
			lx.comment(i)
			goto Restart

		case '/':
			for in[i] != '\n' {
				i++
			}
			lx.comment(i)
			goto Restart
		}
		fallthrough

	case '~', '*', '(', ')', '[', ']', '{', '}', '?', ':', ';', ',', '%', '^', '!', '=', '<', '>', '+', '-', '&', '|':
		if lx.c2goComment && in[0] == '*' && in[1] == '/' {
			lx.c2goComment = false
			lx.skip(2)
			goto Restart
		}
		if c == '-' && in[1] == '>' {
			lx.token(2)
			return tokArrow
		}
		if in[1] == '=' && tokEq[c] != 0 {
			lx.token(2)
			return int(tokEq[c])
		}
		if in[1] == c && tokTok[c] != 0 {
			if in[2] == '=' && tokTokEq[c] != 0 {
				lx.token(3)
				return int(tokTokEq[c])
			}
			lx.token(2)
			return int(tokTok[c])
		}
		lx.token(1)
		return int(c)
	}

	if isalpha(c) {
		for isalpha(in[i]) {
			i++
		}
		lx.sym(i)
		switch lx.tok {
		case "Adr":
			lx.tok = "Addr"
		case "union":
			lx.tok = "struct"
		}
		yy.str = lx.tok
		if t := tokId[lx.tok]; t != 0 {
			return int(t)
		}
		yy.decl = lx.lookupDecl(lx.tok)
		if yy.decl != nil && yy.decl.Storage&Typedef != 0 {
			t := yy.decl.Type
			for t.Kind == TypedefType && t.Base != nil {
				t = t.Base
			}
			yy.typ = &Type{Kind: TypedefType, Name: yy.str, Base: t, TypeDecl: yy.decl}
			return tokTypeName
		}
		if lx.tok == "EXTERN" {
			goto Restart
		}
		return tokName
	}

	lx.Errorf("unexpected input byte %#02x (%c)", c, c)
	return tokError
}

func (lx *lexer) Error(s string) {
	lx.Errorf("%s near %s", s, lx.lastsym)
}

func (lx *lexer) Errorf(format string, args ...interface{}) {
	lx.errors = append(lx.errors, fmt.Sprintf("%s: %s", lx.span(), fmt.Sprintf(format, args...)))
}

type Pos struct {
	File string
	Line int
	Byte int
}

type Span struct {
	Start Pos
	End   Pos
}

func (l Span) String() string {
	return fmt.Sprintf("%s:%d", l.Start.File, l.Start.Line)
}

type Comment struct {
	Span
	Text   string
	Suffix bool
}

var tokEq = [256]int32{
	'*': tokMulEq,
	'/': tokDivEq,
	'+': tokAddEq,
	'-': tokSubEq,
	'%': tokModEq,
	'^': tokXorEq,
	'!': tokNotEq,
	'=': tokEqEq,
	'<': tokLtEq,
	'>': tokGtEq,
	'&': tokAndEq,
	'|': tokOrEq,
}

var tokTok = [256]int32{
	'<': tokLsh,
	'>': tokRsh,
	'=': tokEqEq,
	'+': tokInc,
	'-': tokDec,
	'&': tokAndAnd,
	'|': tokOrOr,
}

var tokTokEq = [256]int32{
	'<': tokLshEq,
	'>': tokRshEq,
}

var tokId = map[string]int32{
	"auto":     tokAuto,
	"break":    tokBreak,
	"case":     tokCase,
	"char":     tokChar,
	"const":    tokConst,
	"continue": tokContinue,
	"default":  tokDefault,
	"do":       tokDo,
	"double":   tokDouble,
	"else":     tokElse,
	"enum":     tokEnum,
	"extern":   tokExtern,
	"float":    tokFloat,
	"for":      tokFor,
	"goto":     tokGoto,
	"if":       tokIf,
	"inline":   tokInline,
	"int":      tokInt,
	"long":     tokLong,
	"offsetof": tokOffsetof,
	"register": tokRegister,
	"return":   tokReturn,
	"short":    tokShort,
	"signed":   tokSigned,
	"sizeof":   tokSizeof,
	"static":   tokStatic,
	"struct":   tokStruct,
	"switch":   tokSwitch,
	"typedef":  tokTypedef,
	"union":    tokUnion,
	"unsigned": tokUnsigned,
	"va_arg":   tokVaArg,
	"void":     tokVoid,
	"volatile": tokVolatile,
	"while":    tokWhile,

	"ARGBEGIN": tokARGBEGIN,
	"ARGEND":   tokARGEND,
	"AUTOLIB":  tokAUTOLIB,
	"USED":     tokUSED,
	"SET":      tokSET,
}

// Comment assignment.
// We build two lists of all subexpressions, preorder and postorder.
// The preorder list is ordered by start location, with outer expressions first.
// The postorder list is ordered by end location, with outer expressions last.
// We use the preorder list to assign each whole-line comment to the syntax
// immediately following it, and we use the postorder list to assign each
// end-of-line comment to the syntax immediately preceding it.

// enum walks the expression adding it and its subexpressions to the pre list.
// The order may not reflect the order in the input.
func (lx *lexer) enum(x Syntax) {
	switch x := x.(type) {
	default:
		panic(fmt.Errorf("order: unexpected type %T", x))
	case nil:
		return
	case *Expr:
		if x == nil {
			return
		}
		lx.enum(x.Left)
		lx.enum(x.Right)
		for _, y := range x.List {
			lx.enum(y)
		}
	case *Init:
		if x == nil {
			return
		}
		lx.enum(x.Expr)
		for _, y := range x.Braced {
			lx.enum(y)
		}
	case *Prog:
		if x == nil {
			return
		}
		for _, y := range x.Decls {
			lx.enum(y)
		}
	case *Stmt:
		if x == nil {
			return
		}
		for _, y := range x.Labels {
			lx.enum(y)
		}
		lx.enum(x.Pre)
		lx.enum(x.Expr)
		lx.enum(x.Post)
		lx.enum(x.Body)
		lx.enum(x.Else)
		for _, y := range x.Block {
			lx.enum(y)
		}
	case *Label:
		// ok
	case *Decl:
		if x == nil {
			return
		}
		if lx.enumSeen[x] {
			return
		}
		lx.enumSeen[x] = true
		lx.enum(x.Type)
		lx.enum(x.Init)
		lx.enum(x.Body)
	case *Type:
		if x == nil {
			return
		}
		lx.enum(x.Base)
		for _, y := range x.Decls {
			lx.enum(y)
		}
		return // do not record type itself, just inner decls
	}
	lx.pre = append(lx.pre, x)
}

func (lx *lexer) order(prog *Prog) {
	lx.enumSeen = make(map[interface{}]bool)
	lx.enum(prog)
	sort.Sort(byStart(lx.pre))
	lx.post = make([]Syntax, len(lx.pre))
	copy(lx.post, lx.pre)
	sort.Sort(byEnd(lx.post))
}

type byStart []Syntax

func (x byStart) Len() int      { return len(x) }
func (x byStart) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x byStart) Less(i, j int) bool {
	pi := x[i].GetSpan()
	pj := x[j].GetSpan()
	// Order by start byte, leftmost first,
	// and break ties by choosing outer before inner.
	if pi.Start.Byte != pj.Start.Byte {
		return pi.Start.Byte < pj.Start.Byte
	}
	return pi.End.Byte > pj.End.Byte
}

type byEnd []Syntax

func (x byEnd) Len() int      { return len(x) }
func (x byEnd) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x byEnd) Less(i, j int) bool {
	pi := x[i].GetSpan()
	pj := x[j].GetSpan()
	// Order by end byte, leftmost first,
	// and break ties by choosing inner before outer.
	if pi.End.Byte != pj.End.Byte {
		return pi.End.Byte < pj.End.Byte
	}
	return pi.Start.Byte > pj.Start.Byte
}

// assignComments attaches comments to nearby syntax.
func (lx *lexer) assignComments() {
	// Generate preorder and postorder lists.
	lx.order(lx.prog)

	// Split into whole-line comments and suffix comments.
	var line, suffix []Comment
	for _, com := range lx.comments {
		if com.Suffix {
			suffix = append(suffix, com)
		} else {
			line = append(line, com)
		}
	}

	// Assign line comments to syntax immediately following.
	for _, x := range lx.pre {
		start := x.GetSpan().Start
		xcom := x.GetComments()
		for len(line) > 0 && start.Byte >= line[0].Start.Byte {
			xcom.Before = append(xcom.Before, line[0])
			line = line[1:]
		}
	}

	// Remaining line comments go at end of file.
	lx.prog.Comments.After = append(lx.prog.Comments.After, line...)

	// Assign suffix comments to syntax immediately before.
	for i := len(lx.post) - 1; i >= 0; i-- {
		x := lx.post[i]

		// Do not assign suffix comments to call, list, end-of-list, or whole file.
		// Instead assign them to the last argument, element, or rule.
		/*
			switch x.(type) {
			case *CallExpr, *ListExpr, *End, *File:
				continue
			}
		*/

		// Do not assign suffix comments to something that starts
		// on an earlier line, so that in
		//
		//	tags = [ "a",
		//		"b" ], # comment
		//
		// we assign the comment to "b" and not to tags = [ ... ].
		span := x.GetSpan()
		start, end := span.Start, span.End
		if start.Line != end.Line {
			continue
		}
		xcom := x.GetComments()
		for len(suffix) > 0 && end.Byte <= suffix[len(suffix)-1].Start.Byte {
			xcom.Suffix = append(xcom.Suffix, suffix[len(suffix)-1])
			suffix = suffix[:len(suffix)-1]
		}
	}

	// We assigned suffix comments in reverse.
	// If multiple suffix comments were appended to the same
	// expression node, they are now in reverse. Fix that.
	for _, x := range lx.post {
		reverseComments(x.GetComments().Suffix)
	}

	// Remaining suffix comments go at beginning of file.
	lx.prog.Comments.Before = append(lx.prog.Comments.Before, suffix...)
}

// reverseComments reverses the []Comment list.
func reverseComments(list []Comment) {
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
}
