// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Dateseq prints a sequence of dates.
//
//	dateseq start end
//
// Dateseq prints dates from start to end, in the same format as the arguments.
// Accepted formats are:
//
//	yyyy-mm-dd
//	yyyy/mm/dd
//	mm/dd/yyyy
//	mm/dd
//
// Example
//
//	dateseq 2013-09-13 2013-10-05
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var formats = []string{
	"2006-01-02",
	"2006/01/02",
	"01/02/2006",
	"01/02",
}

const format = "2006-01-02"

func usage() {
	fmt.Fprintf(os.Stderr, "usage: dateseq start end\n\n")
	fmt.Fprintf(os.Stderr, "start and end must use the same date format, one of:\n")
	for _, format := range formats {
		format = strings.Replace(format, "2006", "yyyy", -1)
		format = strings.Replace(format, "01", "mm", -1)
		format = strings.Replace(format, "02", "dd", -1)
		fmt.Fprintf(os.Stderr, "\t%s\n", format)
	}
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		usage()
	}

	for _, f := range formats {
		lo, err1 := time.Parse(f, args[0])
		hi, err2 := time.Parse(f, args[1])
		if err1 != nil || err2 != nil {
			continue
		}

		for !lo.After(hi) {
			fmt.Println(lo.Format(format))
			lo = lo.AddDate(0, 0, 1)
		}
		return
	}
	usage()
}
