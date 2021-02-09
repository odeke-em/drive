// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Labels converts comma-separated-value records to PostScript mailing labels.
// The output is formatted for 8½"×11" sheets containing thirty 2⅝"x1" labels each,
// such as the Avery 5160.
//
//	usage: labels [options] [file...]
//
// Converts CSV records to PostScript mailing labels, using the first three fields
// of each input record as the address.
//
// The options are:
//
// 	-f font
// 		Use the named PostScript font (default Times-Roman)
// 	-m regexp
// 		Only use input records matching regexp.
// 		The text being matched is the record with commas separating fields,
// 		with no quotation marks added.
// 	-o outfile
// 		Write labels to outfile (default standard output)
// 	-p size
// 		Use text with the given point size (default 12)
// 	-v vsize
// 		Use lines of text vsize points apart (default 1.2 * text size)
// 	-x regexp
// 		Exclude input records matching regexp.
//
// If the first line of the CSV contains the text "address" (case insensitive),
// it is assumed to be a header for the spreadsheet and is skipped.
//
// Example
//
// Used with googlecsv, labels can take Google spreadsheets as input:
//
//	googlecsv 'Mailing List' | labels -f FournierMT-RegularSC > labels.ps
//
package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: labels [options] [file...]

Converts CSV records to PostScript mailing labels, using the first three fields
of each input record as the address.

The options are:

	-f font
		Use the named PostScript font (default Times-Roman)
	-m regexp
		Only use input records matching regexp.
		The text being matched is the record with commas separating fields,
		with no quotation marks added.
	-o outfile
		Write labels to outfile (default standard output)
	-p size
		Use text with the given point size (default 12)
	-v vsize
		Use lines of text vsize points apart (default 1.2 * text size)
	-x regexp
		Exclude input records matching regexp.

If the first line of the CSV contains the text "address" (case insensitive),
it is assumed to be a header for the spreadsheet and is skipped.
`)
	os.Exit(2)
}

var (
	font    = flag.String("f", "Times-Bold", "")
	outfile = flag.String("o", "", "")
	ps      = flag.Int("p", 12, "")
	vs      = flag.Int("v", 0, "")
	match   = flag.String("m", "match", "")
	exclude = flag.String("x", "exclude", "")

	matchRE   *regexp.Regexp
	excludeRE *regexp.Regexp
)

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	var input [][]string
	if flag.NArg() == 0 {
		input = readCSV("standard input", os.Stdin)
	} else {
		for _, file := range flag.Args() {
			f, err := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}
			input = append(input, readCSV(file, f)...)
			f.Close()
		}
	}

	if len(input) > 0 && strings.Contains(strings.ToLower(strings.Join(input[0], ",")), "address") {
		// assume this is a heading line
		input = input[1:]
	}

	var buf bytes.Buffer
	if *vs == 0 {
		*vs = (*ps*12 + 5) / 10
	}
	fmt.Fprintf(&buf, prolog, *font, *ps, *vs)

	nlabel := 0
	for _, line := range input {
		if len(line) == 0 {
			continue
		}
		join := strings.Join(line, ",")
		if matchRE != nil && !matchRE.MatchString(join) ||
			excludeRE != nil && excludeRE.MatchString(join) {
			continue
		}

		mark := "mark"
		for i, field := range line {
			if i >= 3 {
				break
			}
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			field = strings.Replace(field, "(", `\(`, -1)
			field = strings.Replace(field, ")", `\)`, -1)
			for _, f := range strings.Split(field, "\n") {
				f = strings.TrimSpace(f)
				if f == "" {
					continue
				}
				fmt.Fprintf(&buf, "%s (%s)", mark, f)
				mark = ""
			}
		}
		if mark == "" {
			nlabel++
			fmt.Fprintf(&buf, " label\n")
		}
	}
	fmt.Fprintf(&buf, "endlabels\n")

	if nlabel == 0 {
		log.Fatal("no labels to create")
	}

	os.Stdout.Write(buf.Bytes())
}

func readCSV(name string, r io.Reader) [][]string {
	rr := csv.NewReader(r)
	rr.FieldsPerRecord = -1
	recs, err := rr.ReadAll()
	if err != nil {
		log.Fatalf("parsing %s: %v", name, err)
	}
	return recs
}

const prolog = `%%!PS-Adobe-2.0

/numlabel 0 def
/%s findfont 
/ps %d def
/vs %d def
dup length dict begin
  {1 index /FID ne {def} {pop pop} ifelse} forall
  /Encoding ISOLatin1Encoding def
  currentdict
end
/MyFont exch definefont pop
/MyFont findfont ps scalefont setfont


/inch { 72 mul } bind def

/label {
	numlabel 3 mod 2.75 mul 0.125 add 2.625 2 div add inch
	11 numlabel 3 idiv 1 mul 0.5 add 1 2 div add sub inch
	moveto
	0 counttomark vs mul ps add vs sub -2 div rmoveto
	/max 0 def
	counttomark -1 1 {
		1 sub index stringwidth pop
		dup max gt { /max exch def } { pop } ifelse
	} for
	max 2.625 inch gt { /max 2.625 inch def } if
	max -2 div 0 rmoveto
	
	counttomark -1 1 {
		pop
		gsave 0 ps rmoveto show grestore
		0 vs rmoveto
	} for
	pop
	
	/numlabel numlabel 1 add def
	numlabel 30 ge {
		showpage
		/numlabel 0 def
	} if
} def

/endlabels {
	numlabel 0 gt { showpage } if
} def

`
