// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.google.com/p/rsc/scancab/database"
	"code.google.com/p/rsc/websrv"
)

var (
	addr  = flag.String("addr", ":5555", "http port")
	creat = flag.Bool("c", false, "create database")
	dir   = flag.String("dir", ".", "directory containing scans")
	lib   = flag.String("lib", "", "library directory")

	db   *database.DB
	tmpl *template.Template
)

func main() {
	pkg, _ := build.Import("code.google.com/p/rsc/scancab", "", build.FindOnly)
	if pkg.Dir != "" {
		*lib = filepath.Join(pkg.Dir, "lib")
	}

	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
	}

	file := filepath.Join(*dir, "scancab.db")
	var err error
	if *creat {
		db, err = database.Create(file)
	} else {
		db, err = database.Open(file)
	}
	if err != nil {
		log.Fatal(err)
	}

	tmpl = template.New("all").Funcs(funcMap)
	if _, err := tmpl.ParseGlob(filepath.Join(*lib, "template/*.html")); err != nil {
		log.Fatal(err)
	}

	go watch()

	http.Handle("/", websrv.Handler)
	http.Handle("/scan/static/", http.StripPrefix("/scan/static/", http.FileServer(http.Dir(filepath.Join(*lib, "static")))))

	websrv.Handle("/scan/list", new(listRequest))
	websrv.Handle("/scan/incoming", new(incomingRequest))
	websrv.Handle("/scan/incoming-count", new(incomingCountRequest))
	websrv.Handle("/scan/doc/{DocID}", new(show1))
	websrv.Handle("/scan/doc/{DocID}/edit", new(editRequest))
	websrv.Handle("/scan/doc/{DocID}/pdf", new(showPDF))
	websrv.Handle("/scan/doc/{DocID}/thumb", new(thumb))

	websrv.Serve(*addr)
}

func sendJson(ctxt *websrv.Context, obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	ctxt.Header().Set("Content-Type", "application/json")
	ctxt.Write(data)
}

type listRequest struct {
	Start int
	N     int
	Q     string
}

func (req *listRequest) Get(ctxt *websrv.Context) {
	if req.N == 0 {
		req.N = 10
	}
	var docs []*database.Doc
	var err error
	if req.Q != "" {
		docs, err = db.Search(req.Q, "created desc", req.Start, req.N)
	} else {
		docs, err = db.Enum("created desc", req.Start, req.N)
	}
	if err != nil {
		panic(err)
	}
	sendJson(ctxt, docs)
}

type incomingCountRequest struct{}

func (*incomingCountRequest) Get(ctxt *websrv.Context) {
	docs, err := db.Pending("created desc", 0, 1000)
	if err != nil {
		panic(err)
	}
	sendJson(ctxt, len(docs))
}

type incomingRequest struct {
	N int
}

func (req *incomingRequest) Get(ctxt *websrv.Context) {
	if req.N == 0 {
		req.N = 10
	}
	docs, err := db.Pending("created desc", 0, req.N)
	if err != nil {
		panic(err)
	}

	sendJson(ctxt, docs)
}

var timeFormats = []string{
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04",
	"2006-01-02",
	"2006-01-02T15:04:05.999999999Z",
	"2006-01-02T15:04Z",
	"2006-01-02",
}

type docTime struct {
	Parsed time.Time
}

func (d *docTime) Parse(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		d.Parsed = time.Time{}
		return nil
	}
	for _, f := range timeFormats {
		if t, err := time.Parse(f, s); err == nil {
			d.Parsed = t
			return nil
		}
	}
	return fmt.Errorf("bad time: %q", s)
}

type editRequest struct {
	DocID
	Tags     string
	Desc     string
	Time     docTime
	Due      docTime
	Text     string
	Location string
}

func (req *editRequest) Get(ctxt *websrv.Context) {
	runTemplate(ctxt, "edit.html", req.Doc)
}

func (req *editRequest) Post(ctxt *websrv.Context) {
	doc := req.Doc

	fmt.Println(ctxt.Request.Form)

	doc.Time = req.Time.Parsed
	if doc.Time.IsZero() {
		doc.Time = doc.Created
	}

	doc.Due = req.Due.Parsed
	doc.Tags = req.Tags
	doc.Desc = req.Desc
	doc.Text = req.Text
	doc.Location = req.Location

	if err := db.Write(doc); err != nil {
		panic(err)
	}

	ctxt.Redirect("/scan/doc/" + strconv.FormatInt(req.Doc.ID, 10))
}

type DocID struct {
	Doc *database.Doc
}

func (d *DocID) Parse(s string) error {
	id, _ := strconv.ParseInt(s, 0, 64)
	doc, err := db.SearchID(id)
	if err != nil {
		return err
	}
	d.Doc = doc
	return nil
}

type show1 struct {
	DocID
}

func (s *show1) Get(ctxt *websrv.Context) {
	sendJson(ctxt, s.Doc)
	//runTemplate(ctxt, "show.html", s.Doc)
}

type showPDF struct {
	DocID
}

func (s *showPDF) Get(ctxt *websrv.Context) {
	ctxt.ServeFile(filepath.Join(*dir, s.Doc.File))
}

type thumb struct {
	DocID
	DPI  int `websrv:"dpi"`
	Page int `websrv:"page"`
}

func (t *thumb) Get(ctxt *websrv.Context) {
	png, err := db.Thumb(t.Doc, t.Page, t.DPI)
	if err != nil {
		panic(err)
	}
	ctxt.Write(png)
}

var funcMap = template.FuncMap{}

func runTemplate(w io.Writer, name string, data interface{}) {
	tmpl = template.New("all").Funcs(funcMap)
	if _, err := tmpl.ParseGlob(filepath.Join(*lib, "template/*.html")); err != nil {
		log.Fatal(err)
	}
	t := tmpl.Lookup(name)
	if t == nil {
		panic("no such template: " + name)
	}
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func watch() {
	known := make(map[string]bool)
	for {
		// TODO: Use fsnotify, or else have a "scan now" button.
		fis, err := ioutil.ReadDir(*dir)
		if err != nil {
			// TODO: save and display on home page
			log.Print(err)
			continue
		}
		// TODO: have a map of known files.
		for _, fi := range fis {
			name := fi.Name()
			if !strings.HasSuffix(name, ".pdf") && !strings.HasSuffix(name, ".PDF") {
				continue
			}
			if known[name] {
				continue
			}
			known[name] = true
			_, err := db.SearchFile(name)
			if err != database.ErrNotFound {
				continue
			}
			addfile(name)
		}
		time.Sleep(1 * time.Minute)
	}
}

var pdfPrefix = []byte("%PDF-")
var gs sync.Mutex

func addfile(name string) {
	println("addfile", name)
	path := filepath.Join(*dir, name)

	f, err := os.Open(path)
	if err != nil {
		// TODO: save
		log.Print(err)
		return
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		log.Print(err)
		return
	}

	buf := make([]byte, 1024)
	io.ReadFull(f, buf)
	f.Seek(0, 0)

	if !bytes.HasPrefix(buf, pdfPrefix) {
		// TODO: save
		log.Printf("%s is not a PDF", path)
		return
	}

	gs.Lock()
	defer gs.Unlock()
	cmd := exec.Command("gs", "-sDEVICE=nullpage", "-dQUIET", "-dBATCH", "-")
	cmd.Stdin = strings.NewReader("(" + path + ") (r) file GS_PDF_ProcSet begin pdfdict begin pdfopen begin pdfpagecount ==\n")
	out, err := cmd.CombinedOutput()
	out = bytes.TrimSpace(out)
	if len(out) == 0 && err != nil {
		log.Printf("failed to parse PDF with ghostscript: %v", err)
		return
	}

	npg, err := strconv.Atoi(string(out))
	if err != nil {
		log.Printf("failed to parse PDF with ghostscript:\n%s", out)
		return
	}

	h := sha1.New()
	size, err := io.Copy(h, f)
	if err != nil {
		// TODO: save
		log.Print(err)
		return
	}

	// TODO: extract text into doc.Text
	// invoke ps2ascii?

	doc := &database.Doc{
		File:    name,
		SHA1:    fmt.Sprintf("%x", h.Sum(nil)),
		Pages:   npg,
		Size:    size,
		Created: st.ModTime().UTC(),
	}

	println("WRITE")
	if err := db.Write(doc); err != nil {
		log.Print(err)
		return
	}

	db.Thumb(doc, 1, 72)
}
