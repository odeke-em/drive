// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"debug/goobj"
	"encoding/json"
	"log"
	"os"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	pkg, err := goobj.Parse(f, "main")
	if err != nil {
		log.Fatal(err)
	}
	js, err := json.MarshalIndent(pkg, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.WriteString(string(js) + "\n")
}
