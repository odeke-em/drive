// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

import "testing"

var exprTests = []string{
	"x",
	"123",
	"1.4",
	"'z'",
	`"abc" "def"`,
	"x + y",
	"x * y",
	"x / y",
	"x % y",
	"x << y",
	"x >> y",
	"x < y",
	"x > y",
	"x <= y",
	"x >= y",
	"x == y",
	"x != y",
	"x & y",
	"x ^ y",
	"x | y",
	"x && y",
	"x || y",
	"x ? y : z",
	"x = y",
	"x += y",
	"x -= y",
	"x *= y",
	"x /= y",
	"x %= y",
	"x <<= y",
	"x >>= y",
	"x &= y",
	"x ^= y",
	"x |= y",
	"*x",
	"&x",
	"+x",
	"-x",
	"!x",
	"~x",
	"++x",
	"--x",
	"sizeof x",
	"sizeof(int)",
	"offsetof(int, x)",
	"(int)x",
	"(int){}",
	"(int){x}",
	"(x, y, z)",
	"x, y, z",
	"f(x, y, z)",
	"x[y]",
	"x++",
	"x--",
	"va_arg(x, int)",
}

func TestPrintExpr(t *testing.T) {
	for _, str := range exprTests {
		x, err := ParseExpr(str)
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		out := x.String()
		if out != str {
			t.Errorf("ParseExpr(%#q).String() = %#q, want original input", str, out)
		}
	}
}
