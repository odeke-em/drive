// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Googlecsv prints a spreadsheet from Google Drive in CSV format.
//
// The first time googlecsv is run, it starts a web browser and asks
// the user to authenticate to Google Drive spreadsheets on its behalf.
// It saves an authentication token enabling future access to Google Drive
// spreadsheets in $HOME/.token-spreadsheets.google.com.
//
package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"code.google.com/p/rsc/oauthprompt"
)

const (
	apiClientID     = "25135650744-icpsvqtvbeo5b643tuucsdu7md2fi8tq.apps.googleusercontent.com"
	apiClientSecret = "EiK90jK03zLIr7uYcc1IOduJ"
	apiScope        = "http://docs.google.com/feeds/ https://spreadsheets.google.com/feeds/"

	download = "https://spreadsheets.google.com/feeds/download/spreadsheets/Export?exportFormat=csv&gid=0&key="
	feed     = "https://docs.google.com/feeds/documents/private/full?title="
)

var (
	client  = http.DefaultClient
	outfile = flag.String("o", "", "")
)

type DocList struct {
	XMLName xml.Name `xml:"feed"`
	ID      string   `xml:"entry>resourceId"`
}

func usage() {
	fmt.Fprintf(os.Stderr, `usage: googlecsv [-o outfile] 'spreadsheet title'

Prints the CSV content of the first spreadsheet found containing the given title as a substring.

The first time googlecsv is run, it starts a web browser and asks
the user to authenticate to Google Drive spreadsheets on its behalf.
It saves an authentication token enabling future access to Google Drive
spreadsheets in $HOME/.token-spreadsheets.google.com. 
`)
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	log.SetFlags(0)

	tr, err := oauthprompt.GoogleToken(".token-spreadsheets.google.com", apiClientID, apiClientSecret, apiScope)
	if err != nil {
		log.Fatal(err)
	}
	client = tr.Client()

	r, err := client.Get(feed + url.QueryEscape(flag.Arg(0)))
	if err != nil {
		log.Fatal(err)
	}
	if r.StatusCode != http.StatusOK {
		log.Fatalf("reading spreadsheet list: %v", r.Status)
	}
	data, err := ioutil.ReadAll(r.Body)

	var list DocList
	if err := xml.Unmarshal(data, &list); err != nil {
		log.Fatalf("parsing spreadsheet list: %v", err)
	}

	if !strings.HasPrefix(list.ID, "spreadsheet:") {
		dumpxml(bytes.NewBuffer(data))
		log.Fatalf("did not find spreadsheet - id is %s", list.ID)
	}
	r.Body.Close()

	url := download + list.ID[len("spreadsheet:"):]
	r, err = client.Get(url)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	if r.StatusCode != http.StatusOK {
		log.Fatalf("download: %v", r.Status)
	}
	defer r.Body.Close()
	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	os.Stdout.Write(data)
}

func dumpxml(r io.Reader) {
	indent := ""
	p := xml.NewDecoder(r)
	for {
		tok, err := p.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			fmt.Printf("%s<%s%s>\n", indent, t.Name.Local, fmtattr(t.Attr))
			indent += "\t"
		case xml.EndElement:
			indent = indent[:len(indent)-1]
			fmt.Printf("%s</%s>\n", indent, t.Name.Local)
		case xml.CharData:
			s := string([]byte(t))
			s = strings.TrimSpace(s)
			if s == "" {
				break
			}
			fmt.Printf("%s%s\n", indent, strings.Replace(s, "\n", "\n"+indent, -1))
		case xml.Comment:
			fmt.Printf("%s<!--%s-->\n", indent, strings.Replace(string([]byte(t)), "\n", "\n"+indent, -1))
		case xml.ProcInst:
			fmt.Printf("%s<?%s %s>\n", indent, t.Target, strings.Replace(string(t.Inst), "\n", "\n"+indent, -1))
		case xml.Directive:
			fmt.Printf("%s<!%s>\n", indent, strings.Replace(string([]byte(t)), "\n", "\n"+indent, -1))
		}
	}
}

func fmtattr(attrs []xml.Attr) string {
	s := ""
	for _, a := range attrs {
		s += fmt.Sprintf(" %q=%q", a.Name.Local, a.Value)
	}
	return s
}
