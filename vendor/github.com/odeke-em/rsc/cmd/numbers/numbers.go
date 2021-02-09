// Numbers is a text-based calculator language.
//
//	numbers [file...]
//
// Numbers executes the input files given on the command line, or else
// standard input.
//
// The input is a sequence of value definitions, such as:
//
//	Tax = Income * Tax Rate
//	Tax Rate = 5.3%
//	Income = $10,000
//
// The output shows all computed values that were not otherwise used.
// In the example above, the computation of Tax uses the other two values
// but is not itself used, so the output of the program is the value of Tax:
//
//	Tax = $530.00
//
// The computation expressed by a numbers program is not dependent on the order of
// the input lines. All permutations of the input above produce the same
// result. Of course, it is not allowed for a computation to use its own value.
//
// The printed results occur in the order their definitions appear in the input.
//
// Expressions
//
// An expression is a constant, a variable name, or a computation.
//
// Constants are numbers, percentages, or dollar amounts: 42, 5.3%, $10,000.
// The commas in large dollar amounts are required.
//
// Variable names are alphanumeric text beginning with an alphabetic character.
// They may contain embedded spaces and apostrophes.
//
// A computation expression is one of the following, where x, y, and z are
// themselves expressions.
//
//	(x)            # parenthesization
//	-x             # negation
//	x + y          # addition
//	x - y          # subtraction
//	x * y          # multiplication
//	x / y          # division
//
//	x < y          # compare less than
//	x <= y         # compare less than or equal
//	x > y          # compare greater than
//	x >= y         # compare greater than or equal
//	x == y         # compare equal
//	x != y         # compare not equal
//
//	max(x, y)      # maximum of x and y
//	min(x, y)      # minimum of x and y
//	cond(x, y, z)  # if x is true, y, otherwise z
//
// Each expression evaluates to a typed value. Constants have the types
// number, percentage, or dollar amount, depending on their form.
//
// Addition and subtraction require operands of the same type.
// The result has the same type.
//
// Multiplication requires one of its operands to be a number or percentage.
// The result has the type of the other operand.
//
// Division can operate in two ways. If the divisor is a number or percentage,
// the result has the type of the dividend. If the divisor has the same type
// as the dividend, the result is a number.
//
// Comparisons require operands of the same type.
// The result is a boolean.
//
// Max and min require operands of the same type. The result has the same type.
//
// Cond requires that the condition x be a boolean, and that y and z have
// the same type. Cond is not short-circuit: both y and z are evaluated.
// The result has the type of y and z.
//
// All values are stored as float64s.
//
// Statements
//
// The most common form of statement is the definition:
//
//	variable name = expression
//
// This defines the meaning of variable name. All variables referred to in
// expressions must have definitions. Definitions may be given in any order
// but the computation of ``expression'' must not require the use of the value
// of ``variable name.''
//
// There must not be multiple definitions for a single variable name.
// However, there is also a weak definition:
//
//	variable name ?= expression
//
// A weak definition for variable name is used only if there is no
// standard definition. This allows one program file to provide a default value
// that another program file might or might not override.
//
// In addition to these definitional statements, there are four directives:
//
//	include file-name
//
// Read additional definitions from file-name, interpreted relative to the directory
// containing the current input file.
//
//	use(x, y, z, ...)
//
// Mark the expressions as used, so that variable names they refer to are not
// printed in the program output.
//
//	print(x, y, z, ...)
//
// Print the values of the expressions in the program output, before printing the usual output.
// Printing a value counts as a use of it, so print can be used to reorder
// the output without reordering the corresponding definitions. It is also
// useful for debugging.
//
//	check(x, y, z, ...)
//
// Check that each of the expressions evaluates to the boolean value true.
// If not, print a message and exit.
//
// Examples
//
// The source code directory contains example programs that demonstrate how
// numbers might be used to compute income taxes.
//
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Op int

const (
	opConst Op = 1 + iota
	opName
	opAdd
	opSub
	opMul
	opDiv
	opMinus
	opLess
	opLessEqual
	opGreater
	opGreaterEqual
	opEqual
	opNotEqual
	opMax
	opMin
	opCond
	opAssign
	opAssignWeak
	opPrint
	opUse
	opCheck
)

var opNames = map[Op]string{
	opConst:        "const",
	opName:         "name",
	opAdd:          "+",
	opSub:          "-",
	opMul:          "*",
	opDiv:          "/",
	opMinus:        "-",
	opLess:         "<",
	opLessEqual:    "<=",
	opGreater:      ">",
	opGreaterEqual: ">=",
	opEqual:        "==",
	opNotEqual:     "!=",
	opMax:          "max",
	opMin:          "min",
	opCond:         "cond",
	opAssign:       "=",
	opAssignWeak:   "?=",
}

func (op Op) String() string {
	s := opNames[op]
	if s == "" {
		return fmt.Sprintf("op%d", op)
	}
	return s
}

type Type int

const (
	typeError Type = iota
	typeNumber
	typeDollar
	typePercent
	typeBool
)

type Value struct {
	value float64
	typ   Type
}

func (v Value) String() string {
	switch v.typ {
	default:
		return "error"

	case typeNumber:
		return fmt.Sprintf("%g", v.value)

	case typeDollar:
		s := fmt.Sprintf("%.2f", v.value)
		dollar := s[:len(s)-3]
		cent := s[len(s)-3:]
		if cent == ".00" {
			cent = ""
		}
		comma := ""
		for len(dollar) > 3 {
			comma = "," + dollar[len(dollar)-3:] + comma
			dollar = dollar[:len(dollar)-3]
		}
		return "$" + dollar + comma + cent

	case typePercent:
		return fmt.Sprintf("%.3g%%", v.value)

	case typeBool:
		switch v.value {
		case 0:
			return "false"
		case 1:
			return "true"
		default:
			return fmt.Sprintf("bool(%g)", v.value)
		}
	}
}

func boolnum(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

var valueError = Value{0, typeError}
var valueTrue = Value{1, typeBool}

type Line struct {
	File   string
	Lineno int
}

func (l Line) String() string {
	return fmt.Sprintf("%s:%d", l.File, l.Lineno)
}

var nerror int

func (l Line) Errorf(format string, args ...interface{}) {
	fmt.Printf("%s: %s\n", l, fmt.Sprintf(format, args...))
	if nerror++; nerror > 20 {
		fmt.Printf("too many errors\n")
		os.Exit(1)
	}
}

type Expr struct {
	line  Line
	op    Op
	left  *Expr
	right *Expr
	list  []*Expr
	name  string
	def   *Stmt
	value Value
}

var prec = map[Op]int{
	opAdd:          3,
	opSub:          3,
	opMul:          2,
	opDiv:          2,
	opMinus:        1,
	opLess:         4,
	opLessEqual:    4,
	opGreater:      4,
	opGreaterEqual: 4,
	opEqual:        4,
	opNotEqual:     4,
}

func (x *Expr) String() string {
	switch x.op {
	case opConst:
		return x.value.String()
	case opName:
		return x.name
	case opMinus:
		return "-" + x.left.String()
	case opAdd, opSub, opMul, opDiv, opLess, opLessEqual, opGreater, opGreaterEqual, opEqual, opNotEqual:
		left := x.left.String()
		if prec[x.left.op] > prec[x.op] {
			left = "(" + left + ")"
		}
		right := x.right.String()
		if prec[x.right.op] >= prec[x.op] {
			right = "(" + right + ")"
		}
		return left + " " + x.op.String() + " " + right
	case opMax, opMin, opCond:
		s := x.op.String() + "("
		for i, arg := range x.list {
			if i > 0 {
				s += ", "
			}
			s += arg.String()
		}
		return s + ")"
	}
	return "?"
}

type Stmt struct {
	line   Line
	left   *Expr
	op     Op
	right  *Expr
	list   []*Expr
	used   bool
	hidden bool
	value  Value
	state  stmtState
}

func (s *Stmt) String() string {
	return s.left.String() + " " + s.op.String() + " " + s.right.String()
}

type stmtState int

const (
	stateNotStarted stmtState = iota
	stateComputing
	stateComputed
)

type Lexer struct {
	input  string
	file   string
	lineno int
	sym    string
}

func re(s string) *regexp.Regexp {
	return regexp.MustCompile(`\A(?:` + s + `)`)
}

var tokens = []struct {
	re  *regexp.Regexp
	val int
	fn  func(*Lexer, string, *yySymType)
}{
	{re(`include[ \t]+([^\n#]+?)[ \t]*(#.*)?\n`), _INCLUDE, include},
	{re(`(#.*)?\n `), -1, nil},
	{re(`(#.*)?\n\t`), -1, nil},
	{re(`(#.*)?\n`), '\n', nil},
	{re(`[ \t]`), -1, nil},
	{re(`min`), _MIN, nop},
	{re(`max`), _MAX, nop},
	{re(`use`), _USE, nop},
	{re(`check`), _CHECK, nop},
	{re(`print`), _PRINT, nop},
	{re(`cond`), _COND, nop},
	{re(`\pL[\pL'\-0-9]*([ \t]+[\pL0-9][\pL'\-0-9]*)*`), _EXPR, makeName},
	{re(`\$[0-9]{1,3}(,[0-9]{3})*(\.[0-9][0-9])?`), _EXPR, makeDollar},
	{re(`[0-9.]+%`), _EXPR, makePercent},
	{re(`[0-9.]+`), _EXPR, makeNumber},
	{re(`=`), '=', nil},
	{re(`\?=`), _EQQ, nil},
	{re(`<`), '<', nil},
	{re(`<=`), _LE, nil},
	{re(`>`), '>', nil},
	{re(`>=`), _GE, nil},
	{re(`==`), _EQ, nil},
	{re(`!=`), _NE, nil},
	{re(`\+`), '+', nil},
	{re(`-`), '-', nil},
	{re(`\*`), '*', nil},
	{re(`/`), '/', nil},
	{re(`\(`), '(', nil},
	{re(`\)`), ')', nil},
	{re(`,`), ',', nil},
	{re(`[^,+\-*/()<>=! \t\n]+`), 1, nop},
}

func (lx *Lexer) Lex(yy *yySymType) int {
	if len(lx.input) == 0 {
		return 0
	}
	var (
		longest    string
		longestVal int
		longestFn  func(*Lexer, string, *yySymType)
	)
	for _, tok := range tokens {
		s := tok.re.FindString(lx.input)
		if len(s) > len(longest) {
			longest = s
			longestVal = tok.val
			longestFn = tok.fn
		}
	}
	if longest == "" {
		lx.Error(fmt.Sprintf("lexer stuck at %.10q", lx.input))
		return -1
	}
	yy.line = lx.line()
	if longestFn != nil {
		lx.sym = longest
		longestFn(lx, longest, yy)
	}
	lx.input = lx.input[len(longest):]
	lx.lineno += strings.Count(longest, "\n")
	if longestVal < 0 {
		// skip
		return lx.Lex(yy)
	}
	return longestVal
}

func (lx *Lexer) Error(s string) {
	lx.line().Errorf("%s near %s", s, lx.sym)
}

func (lx *Lexer) line() Line {
	return Line{lx.file, lx.lineno}
}

func nop(*Lexer, string, *yySymType) {
	// having a function in the table
	// will make the lexer save the string
	// for use in error messages.
	// nothing more to do.
}

func include(lx *Lexer, s string, yy *yySymType) {
	file := strings.TrimSpace(strings.TrimPrefix(s, "include"))
	rfile := filepath.Join(lx.file, "..", file)
	f, err := os.Open(rfile)
	if err != nil {
		yy.line.Errorf("include %q: open %q: %v", file, rfile, err)
		os.Exit(2)
	}
	read(rfile, f)
	f.Close()
}

func makeName(lx *Lexer, s string, yy *yySymType) {
	s = strings.Replace(s, "\t", " ", -1)
	for {
		t := strings.Replace(s, "  ", " ", -1)
		if t == s {
			break
		}
		s = t
	}
	yy.expr = &Expr{line: lx.line(), op: opName, name: s}
}

func makeNumber(lx *Lexer, s string, yy *yySymType) {
	num, err := strconv.ParseFloat(s, 64)
	var value Value
	if err != nil {
		yy.line.Errorf("invalid value %s", s)
		value = valueError
	} else {
		value = Value{num, typeNumber}
	}
	yy.expr = &Expr{line: lx.line(), op: opConst, value: value}
}

func makePercent(lx *Lexer, s string, yy *yySymType) {
	num, err := strconv.ParseFloat(s[:len(s)-1], 64)
	var value Value
	if err != nil {
		yy.line.Errorf("invalid value %s", s)
		value = valueError
	} else {
		value = Value{num, typePercent}
	}
	yy.expr = &Expr{line: lx.line(), op: opConst, value: value}
}

func makeDollar(lx *Lexer, s string, yy *yySymType) {
	short := strings.Replace(s[1:], ",", "", -1)
	num, err := strconv.ParseFloat(short, 64)
	var value Value
	if err != nil {
		yy.line.Errorf("invalid value %s", s)
		value = valueError
	} else {
		value = Value{num, typeDollar}
	}
	yy.expr = &Expr{line: lx.line(), op: opConst, value: value}
}

var (
	prog []*Stmt
	def  map[string]*Stmt
	stk  []*Stmt
)

func read(name string, r io.Reader) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("reading %s: %v", name, err)
	}
	lx := &Lexer{
		input:  string(data),
		file:   name,
		lineno: 1,
	}
	yyParse(lx)
	if nerror > 0 {
		os.Exit(1)
	}
	if lx.input != "" {
		log.Fatalf("reading %s: did not consume entire file", name)
		os.Exit(1)
	}
}

func run() {
	// Build list of effective definitions.
	def = make(map[string]*Stmt)
	for _, stmt := range prog {
		switch stmt.op {
		case opAssign, opAssignWeak:
			if stmt.left.op != opName {
				stmt.line.Errorf("cannot define %v", stmt.left)
				continue
			}
			if old := def[stmt.left.name]; old != nil {
				if old.op != opAssignWeak {
					stmt.line.Errorf("cannot overwrite definition of %q at %v", stmt.left.name, old.line)
					continue
				}
				old.hidden = true
			}
			def[stmt.left.name] = stmt
		}
	}

	// Build topologically sorted list of subexpressions
	// and record used names.
	for _, stmt := range prog {
		switch stmt.op {
		case opAssign, opAssignWeak:
			if stmt.hidden {
				continue
			}
			stmt.compute()

		case opCheck, opPrint, opUse:
			for _, x := range stmt.list {
				x.compute()
			}
		}
	}

	if nerror > 0 {
		os.Exit(1)
	}

	// Execute non-assignment statements
	// and print the values of any unused assignments.
	for _, stmt := range prog {
		switch stmt.op {
		case opAssign, opAssignWeak:
			if !stmt.hidden && !stmt.used {
				fmt.Printf("%v = %v\n", stmt.left.name, stmt.value)
			}

		case opCheck:
			for _, x := range stmt.list {
				if x.value != valueTrue {
					x.line.Errorf("check %v failed", x)
				}
			}

		case opPrint:
			for _, x := range stmt.list {
				fmt.Printf("%v = %v\n", x, x.value)
			}
		}
	}
}

func (s *Stmt) compute() Value {
	if s.state == stateComputing {
		if stk[len(stk)-1] == s {
			s.line.Errorf("circular definition: %s depends on itself", s.left.name)
			return valueError
		}
		err := "circular definition:\n"
		mystk := stk
		for len(mystk) > 0 && mystk[0] != s {
			mystk = mystk[1:]
		}
		for _, s := range mystk {
			err += "\t" + s.left.name + " depends on\n"
		}
		err += "\t" + s.left.name
		s.line.Errorf("%s", err)
		return valueError
	}

	if s.state == stateComputed {
		return s.value
	}

	stk = append(stk, s)
	s.state = stateComputing
	s.value = s.right.compute()
	s.state = stateComputed
	stk = stk[:len(stk)-1]
	return s.value
}

func (x *Expr) compute() Value {
	switch x.op {
	case opConst:
		// x.value already set

	case opName:
		stmt := def[x.name]
		if stmt == nil {
			x.line.Errorf("unknown value %q", x.name)
			return valueError
		}
		stmt.used = true
		x.value = stmt.compute()

	case opAdd:
		v1, v2, typ := x.binary()
		x.value = Value{v1.value + v2.value, typ}

	case opSub:
		v1, v2, typ := x.binary()
		x.value = Value{v1.value - v2.value, typ}

	case opMul:
		v1, v2, typ := x.binary()
		x.value = Value{v1.value * v2.value, typ}

	case opDiv:
		v1, v2, typ := x.binary()
		if typ == typeError {
			x.value = Value{0, typ}
			break
		}
		if v2.value == 0 {
			x.line.Errorf("division by zero")
			x.value = valueError
			break
		}
		x.value = Value{v1.value / v2.value, typ}

	case opMinus:
		v1 := x.left.compute()
		x.value = Value{-v1.value, v1.typ}

	case opLess:
		v1, v2, typ := x.binary()
		x.value = Value{boolnum(v1.value < v2.value), typ}

	case opLessEqual:
		v1, v2, typ := x.binary()
		x.value = Value{boolnum(v1.value <= v2.value), typ}

	case opGreater:
		v1, v2, typ := x.binary()
		x.value = Value{boolnum(v1.value > v2.value), typ}

	case opGreaterEqual:
		v1, v2, typ := x.binary()
		x.value = Value{boolnum(v1.value >= v2.value), typ}

	case opEqual:
		v1, v2, typ := x.binary()
		x.value = Value{boolnum(v1.value == v2.value), typ}

	case opNotEqual:
		v1, v2, typ := x.binary()
		x.value = Value{boolnum(v1.value != v2.value), typ}

	case opMax, opMin:
		v1 := x.list[0].compute()
		for _, arg := range x.list[1:] {
			v2 := arg.compute()
			if v2.typ == typeError {
				v1.typ = typeError
			}
			if v1.typ != v2.typ && v1.typ != typeError {
				x.line.Errorf("cannot compute %v of %v and %v (mismatched types)", x.op, v1, v2)
			}
			if x.op == opMax {
				if v1.value < v2.value {
					v1.value = v2.value
				}
			} else {
				if v1.value > v2.value {
					v1.value = v2.value
				}
			}
		}
		x.value = v1
	case opCond:
		// Not short circuit, at least for now, so that every run
		// type-checks the entire program.
		cond := x.list[0].compute()
		if cond.typ == typeError {
			x.value = valueError
			break
		}
		if cond.typ != typeBool {
			x.line.Errorf("condition must be bool, not %v", cond.typ)
			x.value = valueError
			break
		}
		v1 := x.list[1].compute()
		v2 := x.list[2].compute()
		if v1.typ == typeError || v2.typ == typeError {
			x.value = valueError
			break
		}
		if v1.typ != v2.typ {
			x.line.Errorf("cannot use %v and %v as branches of conditional (mismatched types)", v1, v2)
			x.value = valueError
			break
		}
		if cond.value != 0 {
			x.value = v1
		} else {
			x.value = v2
		}
	}

	return x.value
}

func (x *Expr) binary() (v1, v2 Value, typ Type) {
	v1 = x.left.compute()
	v2 = x.right.compute()
	typ = v1.typ
	if v1.typ == typeError || v2.typ == typeError {
		return valueError, valueError, typeError
	}
	switch x.op {
	case opAdd, opSub:
		if v1.typ != v2.typ {
			goto BadType
		}

	case opLess, opLessEqual, opGreater, opGreaterEqual, opEqual, opNotEqual:
		if v1.typ != v2.typ {
			goto BadType
		}
		typ = typeBool

	case opMul, opDiv:
		if v1.typ == v2.typ {
			typ = typeNumber
			break
		}
		if v2.typ == typePercent {
			v2.value /= 100
			v2.typ = typeNumber
		}
		if v2.typ == typeNumber {
			break
		}
		if x.op == opMul {
			if v1.typ == typePercent {
				v1.value /= 100
				v1.typ = typeNumber
			}
			if v1.typ == typeNumber {
				typ = v2.typ
				break
			}
		}
		goto BadType
	}
	return v1, v2, typ

BadType:
	x.line.Errorf("cannot compute %v %v %v (mismatched types)", v1, x.op, v2)
	return valueError, valueError, typeError
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		read("<stdin>", os.Stdin)
	} else {
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}
			read(arg, f)
			f.Close()
		}
	}
	run()
}
