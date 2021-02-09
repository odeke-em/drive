// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"code.google.com/p/rsc/cc"
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		cc.Read("<stdin>", os.Stdin)
	} else {
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}
			prog, err := cc.Read(arg, f)
			f.Close()
			if err != nil {
				log.Print(err)
				continue
			}
			var p cc.Printer
			p.StartHTML()
			p.Print(prog)
			p.EndHTML()
			fmt.Printf("%s\n", p.String())
		}
	}
}
