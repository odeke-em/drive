// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.google.com/p/goplan9/draw"
	"code.google.com/p/goplan9/plan9/acme"
	"code.google.com/p/rsc/issue"
)

func acmeMode() {
	var dummy awin
	dummy.prefix = "/issue/" + *project + "/"
	if flag.NArg() > 0 {
		for _, arg := range flag.Args() {
			if !dummy.look(arg) {
				dummy.newSearch("search", arg)
			}
		}
	} else {
		dummy.look("all")
	}
	select {}
}

const (
	modeSingle = 1 + iota
	modeQuery
)

type awin struct {
	*acme.Win
	prefix   string
	mode     int
	query    string
	id       int
	data     []*issue.Issue
	tab      int
	font     *draw.Font
	fontName string
	title    string
}

var all struct {
	sync.Mutex
	m      map[string]*awin
	f      map[string]*draw.Font
	numwin int
}

func (w *awin) exit() {
	all.Lock()
	defer all.Unlock()
	if all.m[w.title] == w {
		delete(all.m, w.title)
	}
	if all.numwin--; all.numwin == 0 {
		os.Exit(0)
	}
}

func (w *awin) new(title string) *awin {
	all.Lock()
	defer all.Unlock()
	all.numwin++
	if all.m == nil {
		all.m = make(map[string]*awin)
	}
	w1 := new(awin)
	w1.title = title
	var err error
	w1.Win, err = acme.New()
	if err != nil {
		log.Printf("creating acme window: %v", err)
		time.Sleep(10 * time.Millisecond)
		w1.Win, err = acme.New()
		if err != nil {
			log.Fatalf("creating acme window again: %v", err)
		}
	}
	w1.prefix = w.prefix
	w1.Name(w1.prefix + title)
	all.m[title] = w1
	return w1
}

func (w *awin) show(title string) *awin {
	all.Lock()
	defer all.Unlock()
	if w1 := all.m[title]; w1 != nil {
		w.Ctl("show")
		return w1
	}
	return nil
}

func (w *awin) fixfont() {
	ctl := make([]byte, 1000)
	w.Seek("ctl", 0, 0)
	n, err := w.Read("ctl", ctl)
	if err != nil {
		return
	}
	f := strings.Fields(string(ctl[:n]))
	if len(f) < 8 {
		return
	}
	w.tab, _ = strconv.Atoi(f[7])
	if w.tab == 0 {
		return
	}
	name := f[6]
	if w.fontName == name {
		return
	}
	all.Lock()
	defer all.Unlock()
	if font := all.f[name]; font != nil {
		w.font = font
		w.fontName = name
		return
	}
	var disp *draw.Display = nil
	font, err := disp.OpenFont(name)
	if err != nil {
		return
	}
	if all.f == nil {
		all.f = make(map[string]*draw.Font)
	}
	all.f[name] = font
	w.font = font
}

var numRE = regexp.MustCompile(`(?m)^[0-9]+\t`)

func (w *awin) look(text string) bool {
	if text == "all" {
		if w.show("all") != nil {
			return true
		}
		w.newSearch("all", "-label:nonexistent")
		return true
	}
	if n, _ := strconv.Atoi(text); 0 < n && n < 100000 {
		if w.show(text) != nil {
			return true
		}
		w.newIssue(text, n)
		return true
	}
	if m := numRE.FindAllString(text, -1); m != nil {
		for _, s := range m {
			w.look(strings.TrimSpace(s))
		}
		return true
	}
	return false
}

func (w *awin) label(labels, text string) {
	println("label", text)
	defer close(w.blinker())
	if w.mode == modeSingle {
		w.labelOne(labels, w.id)
		w.load()
		return
	}
	if n, _ := strconv.Atoi(text); 0 < n && n < 100000 {
		w.labelOne(labels, n)
		return
	}
	if m := numRE.FindAllString(text, -1); m != nil {
		println("numre", len(m))
		for _, s := range m {
			n, _ := strconv.Atoi(strings.TrimSpace(s))
			println(n, s)
			if 0 < n && n < 100000 {
				w.labelOne(labels, n)
			}
		}
		return
	}
}

func (w *awin) labelOne(labels string, n int) {
	println("labelOne", n, labels)
	var ch Change
	ch.Label = strings.Fields(labels)
	if err := write(n, &ch); err != nil {
		w.err(fmt.Sprintf("labeling issue %d: %v\n", n, err))
	}
}

func (w *awin) newIssue(title string, id int) {
	w = w.new(title)
	w.mode = modeSingle
	w.id = id
	w.Ctl("cleartag")
	w.Fprintf("tag", " Get Put Look ")
	go w.load()
	go w.loop()
}

func (w *awin) newSearch(title, query string) {
	w = w.new(title)
	w.mode = modeQuery
	w.query = query
	w.Ctl("cleartag")
	w.Fprintf("tag", " Get Search ")
	w.Write("body", []byte("Loading..."))
	go w.load()
	go w.loop()
}

func (w *awin) blinker() chan struct{} {
	c := make(chan struct{})
	go func() {
		t := time.NewTicker(300 * time.Millisecond)
		defer t.Stop()
		dirty := false
		for {
			select {
			case <-t.C:
				dirty = !dirty
				if dirty {
					w.Ctl("dirty")
				} else {
					w.Ctl("clean")
				}
			case <-c:
				if dirty {
					w.Ctl("clean")
				}
				return
			}
		}
	}()
	return c
}

func (w *awin) clear() {
	w.Addr(",")
	w.Write("data", nil)
}

func (w *awin) load() {
	w.fixfont()

	switch w.mode {
	case modeSingle:
		data, err := fetch(fmt.Sprintf("id:%d", w.id), true)
		w.clear()
		if err != nil {
			w.Write("body", []byte(err.Error()))
			break
		}
		var buf bytes.Buffer
		printFull(&buf, data)
		fmt.Fprintf(&buf, "\nNew Comment:\n\n")
		w.Write("body", buf.Bytes())
		w.Ctl("clean")
		w.data = data
	case modeQuery:
		w.query = strings.Replace(w.query, ": ", ":", -1)
		w.query = strings.TrimSpace(w.query)
		data, err := fetch(w.query, false)
		w.clear()
		if err != nil {
			w.Write("body", []byte(err.Error()))
			break
		}
		if w.title == "search" {
			w.Fprintf("body", "Search: %s\n\n", w.query)
		}
		w.printList(data)
		w.Ctl("clean")
		w.data = data
	}
	w.Addr("0")
	w.Ctl("dot=addr")
	w.Ctl("show")
}

func (w *awin) err(s string) {
	if !strings.HasSuffix(s, "\n") {
		s = s + "\n"
	}
	w1 := w.show("+Errors")
	if w1 == nil {
		w1 = w.new("+Errors")
	}
	w1.Fprintf("body", "%s", s)
	w1.Addr("$")
	w1.Ctl("dot=addr")
	w1.Ctl("show")
}

func diff(line, field, old string) string {
	old = strings.TrimSpace(old)
	line = strings.TrimSpace(strings.TrimPrefix(line, field))
	if old == line {
		return ""
	}
	return line
}

func diffUpdate(line, field, old string) string {
	old = strings.TrimSpace(old)
	line = strings.TrimSpace(strings.TrimPrefix(line, field))
	if old == line {
		return ""
	}
	if line == "" {
		return "-" + old
	}
	return line
}

func diffUpdates(line, field string, old []string) []string {
	line = strings.TrimSpace(strings.TrimPrefix(line, field))
	had := make(map[string]bool)
	for _, f := range old {
		had[f] = true
	}
	var d []string
	kept := make(map[string]bool)
	for _, f := range strings.Fields(line) {
		f = strings.TrimSuffix(f, ",")
		if had[f] {
			kept[f] = true
		} else {
			d = append(d, f)
		}
	}
	for _, f := range old {
		if had[f] && !kept[f] {
			d = append(d, "-"+f)
		}
	}
	return d
}

func (w *awin) put() {
	defer close(w.blinker())
	switch w.mode {
	case modeSingle:
		e := w.data[0]
		var ch Change
		data, err := w.ReadAll("body")
		if err != nil {
			w.err(fmt.Sprintf("Put: %v", err))
			return
		}
		sdata := string(data)
		off := 0
		for _, line := range strings.SplitAfter(sdata, "\n") {
			off += len(line)
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			switch {
			case strings.HasPrefix(line, "Summary:"):
				ch.Summary = diff(line, "Summary:", e.Summary)

			case strings.HasPrefix(line, "Status:"):
				status := e.Status
				if status == "Duplicate" {
					status += " " + fmt.Sprint(e.Duplicate)
				}
				ch.Status = diff(line, "Status:", status)

			case strings.HasPrefix(line, "Owner:"):
				ch.Owner = diffUpdate(line, "Owner:", e.Owner)

			case strings.HasPrefix(line, "CC:"):
				ch.CC = diffUpdates(line, "CC:", e.CC)

			case strings.HasPrefix(line, "Labels:"):
				ch.Label = diffUpdates(line, "Labels:", e.Label)

			default:
				w.err(fmt.Sprintf("Put: unknown summary line: %s", line))
			}
		}

		i := strings.Index(sdata, "\nReported by ")
		if i >= off {
			ch.Comment = strings.TrimSpace(sdata[off:i])
		}

		tag := "\nNew Comment:"
		i = strings.Index(sdata, tag)
		if i >= 0 {
			s := strings.TrimSpace(sdata[i+len(tag):])
			if s != "" {
				if ch.Comment != "" {
					ch.Comment += "\n\n"
				}
				ch.Comment = s
			}
		}
		if err := write(w.id, &ch); err != nil {
			w.err(fmt.Sprintf("Put: %v", err))
			return
		}
		w.load()

	case modeQuery:
		w.err("cannot Put issue list")
	}
}

func (w *awin) loadText(e *acme.Event) {
	if len(e.Text) == 0 && e.Q0 < e.Q1 {
		w.Addr("#%d,#%d", e.Q0, e.Q1)
		data, err := w.ReadAll("xdata")
		if err != nil {
			w.err(err.Error())
		}
		e.Text = data
	}
}

func (w *awin) selection() string {
	w.Ctl("addr=dot")
	data, err := w.ReadAll("xdata")
	if err != nil {
		w.err(err.Error())
	}
	return string(data)
}

func (w *awin) loop() {
	defer w.exit()
	for e := range w.EventChan() {
		switch e.C2 {
		case 'x', 'X': // execute
			println("x", string(e.Text))
			cmd := strings.TrimSpace(string(e.Text))
			if cmd == "Get" {
				w.load()
				break
			}
			if cmd == "Put" {
				w.put()
				break
			}
			if cmd == "Del" {
				w.Ctl("del")
				break
			}
			if strings.HasPrefix(cmd, "Search ") {
				w.newSearch("search", strings.TrimSpace(strings.TrimPrefix(cmd, "Search")))
				break
			}
			if strings.HasPrefix(cmd, "Label ") {
				text := w.selection()
				w.label(strings.TrimSpace(strings.TrimPrefix(cmd, "Label")), text)
				break
			}
			w.WriteEvent(e)
		case 'l', 'L': // look
			w.loadText(e)
			if !w.look(string(e.Text)) {
				w.WriteEvent(e)
			}
		}
	}
}

func (w *awin) printTabbed(rows [][]string) {
	var wid []int

	if w.font != nil {
		for _, row := range rows {
			for len(wid) < len(row) {
				wid = append(wid, 0)
			}
			for i, col := range row {
				n := w.font.StringWidth(col)
				if wid[i] < n {
					wid[i] = n
				}
			}
		}
	}

	var buf bytes.Buffer
	for _, row := range rows {
		for i, col := range row {
			buf.WriteString(col)
			if i == len(row)-1 {
				break
			}
			if w.font == nil || w.tab == 0 {
				buf.WriteString("\t")
				continue
			}
			pos := w.font.StringWidth(col)
			for pos <= wid[i] {
				buf.WriteString("\t")
				pos += w.tab - pos%w.tab
			}
		}
		buf.WriteString("\n")
	}

	w.Write("body", buf.Bytes())
}

func (w *awin) printList(entries []*issue.Issue) {
	var rows [][]string
	var buf bytes.Buffer
	for _, e := range entries {
		buf.Reset()
		buf.WriteString(e.Summary)
		if len(e.Label) > 0 {
			buf.WriteString(" [")
			for i, l := range e.Label {
				if i > 0 {
					buf.WriteString(" ")
				}
				buf.WriteString(l)
			}
			buf.WriteString("]")
		}
		rows = append(rows, []string{fmt.Sprint(e.ID), buf.String()})
	}
	w.printTabbed(rows)
}
