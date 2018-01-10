// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package database implements the database for the scanning cabinet.
package database

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "code.google.com/p/gosqlite/sqlite3"
	"code.google.com/p/rsc/dbstore"
)

// A DB holds metadata for the scanning cabinet.
// The actual scans are stored separately, as ordinary files.
type DB struct {
	meta  *sql.DB
	store *dbstore.Storage
	file  string
}

// Open opens the database in the named file.
func Open(name string) (*DB, error) {
	_, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	meta, err := sql.Open("sqlite3", name)
	if err != nil {
		return nil, err
	}
	store := new(dbstore.Storage)
	store.Register(new(Doc))
	store.Register(new(thumb))
	store.Register(new(text))
	return &DB{meta, store, name}, nil
}

// A Doc is a single document in the database.
type Doc struct {
	ID      int64 `dbstore:",rowid,autoinc"`
	File    string
	SHA1    string
	Pages   int
	Size    int64
	Created time.Time

	// user-editable
	Time     time.Time
	Due      time.Time
	Tags     string
	Desc     string
	Text     string
	Location string
}

type thumb struct {
	ID   int64 `dbstore:",key"`
	Page int   `dbstore:",key"`
	DPI  int   `dbstore:",key"`
	PNG  []byte
	Used time.Time
}

type text struct {
	ID   int64 `dbstore:",key,fts4"`
	Desc string
	Text string
	Tags string
}

// Create creates a new database in the named file.
// The file must not exist.
func Create(name string) (*DB, error) {
	_, err := os.Stat(name)
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("create %s: already exists", name)
	}

	meta, err := sql.Open("sqlite3", name)
	if err != nil {
		return nil, err
	}
	store := new(dbstore.Storage)
	store.Register(new(Doc))
	store.Register(new(thumb))
	store.Register(new(text))

	if err := store.CreateTables(meta); err != nil {
		return nil, err
	}

	return &DB{meta, store, name}, nil
}

// Write adds doc to the database, overwriting any existing
// entry with the same doc.ID.
func (db *DB) Write(doc *Doc) error {
	tx, err := db.meta.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op if tx.Commit called already

	id := doc.ID
	if err := db.store.Insert(tx, doc); err != nil {
		doc.ID = id
		return err
	}
	txt := text{
		ID:   doc.ID,
		Desc: strings.ToLower(doc.Desc),
		Text: strings.ToLower(doc.Text),
		Tags: strings.ToLower(doc.Tags),
	}
	if err := db.store.Insert(tx, &txt); err != nil {
		doc.ID = id
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Delete deletes the document (or any document with the same doc.ID)
// from the database.
func (db *DB) Delete(doc *Doc) error {
	if doc.ID == 0 {
		return fmt.Errorf("doc never written")
	}

	tx, err := db.meta.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op after Commit

	if err := db.store.Delete(tx, doc); err != nil {
		return err
	}
	txt := text{ID: doc.ID}
	if err := db.store.Delete(tx, &txt); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Search returns documents matching query, an SQLITE full-text search.
// The results are ordered by 'sortBy', and at most count results are returned
// after skipping the first offset results.
func (db *DB) Search(query, sortBy string, offset, count int) ([]*Doc, error) {
	var docs []*Doc
	err := db.store.Select(db.meta, &docs, `where "ID" in (select "ID" from "code.google.com/p/rsc/scancab/database.text" where "code.google.com/p/rsc/scancab/database.text" match ?) order by `+sortBy+` limit ? offset ?`, query, count, offset)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// Enum enumerates all the documents in the database,
// sorting them by 'sortBy', and then returning at most count documents
// after skipping offset.
func (db *DB) Enum(sortBy string, offset, count int) ([]*Doc, error) {
	var docs []*Doc
	err := db.store.Select(db.meta, &docs, `order by `+sortBy+` limit ? offset ?`, count, offset)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// Pending enumerates the unfiled documents in the database,
// sorting them by 'sortBy', and then returning at most count documents
// after skipping offset.
func (db *DB) Pending(sortBy string, offset, count int) ([]*Doc, error) {
	var docs []*Doc
	err := db.store.Select(db.meta, &docs, `where "Time" = ? order by `+sortBy+` limit ? offset ?`, time.Time{}, count, offset)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

var ErrNotFound = errors.New("file not found")

// SearchFile searches the database for the document describing the named file.
func (db *DB) SearchFile(name string) (*Doc, error) {
	var doc *Doc
	err := db.store.Select(db.meta, &doc, `where "File" = ?`, name)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, ErrNotFound
	}
	return doc, nil
}

// SearchID searches the database for the document with the given ID.
func (db *DB) SearchID(id int64) (*Doc, error) {
	var doc *Doc
	err := db.store.Select(db.meta, &doc, `where "ID" = ?`, id)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

var pngHeader = []byte("\x89PNG")

// Thumb returns a PNG thumbnail for the given page of the given document.
// The thumbnail shows at most the top three inches of the document, at the given dots-per-inch.
// It requires Ghostscript to be installed in the executable search path as 'gs'.
func (db *DB) Thumb(doc *Doc, page, dpi int) ([]byte, error) {
	if page < 1 || page > doc.Pages {
		return nil, fmt.Errorf("page %d not in range [%d, %d]", page, 1, doc.Pages)
	}

	now := time.Now().UTC()

	b := thumb{ID: doc.ID, Page: page, DPI: dpi, Used: now}
	if err := db.store.Read(db.meta, &b, "PNG"); err == nil {
		db.store.Write(db.meta, &b, "Used")
		return b.PNG, nil
	}

	file := filepath.Join(filepath.Dir(db.file), doc.File)
	if _, err := os.Stat(file); err != nil {
		return nil, err
	}

	pg := strconv.Itoa(page)

	cmd := exec.Command("gs", "-sDEVICE=bbox", "-dQUIET", "-dBATCH", "-dNOPAUSE", "-dSAFER", file)
	data, err := cmd.CombinedOutput()
	f := strings.Fields(string(data))
	if len(f) == 0 || f[0] != "%%BoundingBox:" {
		return nil, fmt.Errorf("invoking gs: %v\n%s", err, data)
	}
	x1, _ := strconv.Atoi(f[1])
	y1, _ := strconv.Atoi(f[2])
	x2, _ := strconv.Atoi(f[3])
	y2, _ := strconv.Atoi(f[4])

	show := 3
	dy := show * 72
	if dy > y2-y1 {
		dy = y2 - y1
	}

	_ = x1
	_ = x2

	gflag := fmt.Sprintf("-g%dx%d", dpi*17/2, dpi*dy/72)
	cflag := fmt.Sprintf("<</Install {0 %d translate}>> setpagedevice", -(y2 - dy))

	cmd = exec.Command("gs", "-sDEVICE=png16m", "-dFirstPage="+pg, "-dLastPage="+pg, "-sOutputFile=-", "-r"+strconv.Itoa(dpi), "-dQUIET", "-dBATCH", "-dNOPAUSE", "-dSAFER", "-dGraphicsAlphaBits=4", "-dTextAlphaBits=4", "-dUseTrimBox", gflag, "-c", cflag, "-f", file)
	cmd.Stderr = os.Stderr
	data, err = cmd.Output()
	if len(data) == 0 && err != nil {
		return nil, fmt.Errorf("invoking gs: %v", err)
	}

	if !bytes.HasPrefix(data, pngHeader) {
		return nil, fmt.Errorf("gs: [%v]\n%s", err, data)
	}

	b.PNG = data
	db.store.Insert(db.meta, &b)
	return b.PNG, nil
}
