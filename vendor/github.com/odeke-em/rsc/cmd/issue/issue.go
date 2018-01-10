// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/rsc/issue"
	"code.google.com/p/rsc/oauthprompt"
)

var auth struct {
	APIClientID     string
	APIClientSecret string
}

const Version = "Go1.2"

var aflag = flag.Bool("a", false, "run in acme mode")
var project = flag.String("p", "go", "code.google.com project identifier")

func usage() {
	fmt.Fprintf(os.Stderr, `usage: issue [-a] [-p project] [query]

If query is a single number, prints the full history for the issue.
Otherwise, prints a table of matching results.
The special query 'go1' is shorthand for 'Priority-Go1'.

The -a flag runs as an Acme window, making the query optional.
`)
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)

	if *aflag {
		if err := login(); err != nil {
			log.Fatal(err)
		}
		acmeMode()
		return
	}

	if flag.NArg() != 1 {
		usage()
	}

	full := false
	q := flag.Arg(0)
	n, _ := strconv.Atoi(q)
	if n != 0 {
		q = "id:" + q
		full = true
	}
	if q == "go1" {
		q = "label:Priority-Go1"
	}

	data, err := fetch(q, full)
	if err != nil {
		log.Fatal(err)
	}

	if full {
		printFull(os.Stdout, data)
	} else {
		printList(os.Stdout, data)
	}
}

func fetch(query string, full bool) ([]*issue.Issue, error) {
	can := "open"
	if full {
		can = "all"
	}
	return issue.Search(*project, can, query, full, client)
}

type Change struct {
	Summary string
	Status  string
	Owner   string
	Label   []string
	CC      []string
	Comment string
}

func write(id int, ch *Change) error {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version='1.0' encoding='UTF-8'?>
<entry xmlns='http://www.w3.org/2005/Atom' xmlns:issues='http://schemas.google.com/projecthosting/issues/2009'>
  <content type='html'>`)
	xml.Escape(&buf, []byte(ch.Comment))
	buf.WriteString(`</content>
  <author>
    <name>ignored</name>
  </author>
  <issues:updates>
`)
	tag := func(t, data string) {
		buf.WriteString(`    ` + t)
		xml.Escape(&buf, []byte(data))
		buf.WriteString(`</` + t[1:])
	}

	if ch.Summary != "" {
		tag("<issues:summary>", ch.Summary)
	}
	if ch.Status != "" {
		status := ch.Status
		merge := ""
		if strings.HasPrefix(status, "Duplicate ") {
			merge = strings.TrimPrefix(status, "Duplicate ")
			status = "Duplicate"
		}
		tag("<issues:status>", status)
		if merge != "" {
			tag("<issues:mergedIntoUpdate>", merge)
		}
	}
	if ch.Owner != "" {
		tag("<issues:ownerUpdate>", ch.Owner)
	}
	for _, l := range ch.Label {
		tag("<issues:label>", l)
	}
	for _, cc := range ch.CC {
		tag("<issues:ccUpdate>", cc)
	}
	buf.WriteString(`
  </issues:updates>
</entry>
`)

	// Done with XML!

	u := "https://code.google.com/feeds/issues/p/" + *project + "/issues/" + fmt.Sprint(id) + "/comments/full"
	req, err := http.NewRequest("POST", u, &buf)
	if err != nil {
		return fmt.Errorf("write: %v", err)
	}
	req.Header.Set("Content-Type", "application/atom+xml")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("write: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		buf.Reset()
		io.Copy(&buf, resp.Body)
		return fmt.Errorf("write: %v\n%s", resp.Status, buf.String())
	}
	return nil
}

func printFull(w io.Writer, list []*issue.Issue) {
	for _, e := range list {
		fmt.Fprintf(w, "Summary: %s\n", e.Summary)
		fmt.Fprintf(w, "Status: %s", e.Status)
		if e.Status == "Duplicate" {
			fmt.Fprintf(w, " %d", e.Duplicate)
		}
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Owner: %s\n", e.Owner)
		fmt.Fprintf(w, "CC:")
		for _, cc := range e.CC {
			fmt.Fprintf(w, " %s", cc)
		}
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Labels:")
		for _, l := range e.Label {
			fmt.Fprintf(w, " %s", l)
		}
		fmt.Fprintf(w, "\n")

		for i, c := range e.Comment {
			what := "Reported"
			if i > 0 {
				what = "Comment"
			}
			fmt.Fprintf(w, "\n%s by %s (%s)\n", what, c.Author, c.Time.Format("2006-01-02 15:04:05"))
			if c.Summary != "" {
				fmt.Fprintf(w, "\tSummary: %s\n", c.Summary)
			}
			if c.Owner != "" {
				fmt.Fprintf(w, "\tOwner: %s\n", c.Owner)
			}
			if c.Status != "" {
				fmt.Fprintf(w, "\tStatus: %s\n", c.Status)
			}
			for _, l := range c.Label {
				fmt.Fprintf(w, "\tLabel: %s\n", l)
			}
			for _, l := range c.CC {
				fmt.Fprintf(w, "\tLabel: %s\n", l)
			}
			if c.Text != "" {
				fmt.Fprintf(w, "\n\t%s\n", wrap(html.UnescapeString(c.Text), "\t"))
			}
		}
	}
}

func printList(w io.Writer, list []*issue.Issue) {
	for _, e := range list {
		fmt.Fprintf(w, "%v\t%v\n", e.ID, e.Summary)
	}
}

func wrap(t string, prefix string) string {
	out := ""
	t = strings.Replace(t, "\r\n", "\n", -1)
	lines := strings.Split(t, "\n")
	for i, line := range lines {
		if i > 0 {
			out += "\n" + prefix
		}
		s := line
		for len(s) > 70 {
			i := strings.LastIndex(s[:70], " ")
			if i < 0 {
				i = 69
			}
			i++
			out += s[:i] + "\n" + prefix
			s = s[i:]
		}
		out += s
	}
	return out
}

var client = http.DefaultClient

func login() error {
	if false {
		data, err := ioutil.ReadFile("../../authblob")
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &auth); err != nil {
			return err
		}
	}

	auth.APIClientID = "993255737644.apps.googleusercontent.com"
	auth.APIClientSecret = "kjB02zudLVECBmJdKVMaZluI"

	tr, err := oauthprompt.GoogleToken(".token-code.google.com", auth.APIClientID, auth.APIClientSecret, "https://code.google.com/feeds/issues")
	if err != nil {
		return err
	}
	client = tr.Client()
	return nil
}
