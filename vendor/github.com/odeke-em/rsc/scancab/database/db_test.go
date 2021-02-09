// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	db, err := Create(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	docs, err := db.Enum("ID", 0, 10)
	if err != nil {
		t.Fatalf("enum on empty db: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("enum on empty db found %d docs", len(docs))
	}

	docs, err = db.Search("text", "ID", 0, 10)
	if err != nil {
		t.Fatalf("search on empty db: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("search on empty db found %d docs", len(docs))
	}

	doc := &Doc{
		File:     "file.pdf",
		SHA1:     "abcdef",
		Created:  time.Now().UTC(),
		Time:     time.Now().UTC(),
		Due:      time.Now().UTC(),
		Desc:     "description here",
		Tags:     "Tag1 Tag2 tag3",
		Location: "file cabinet",
	}
	if err := db.Write(doc); err != nil {
		t.Fatalf("write: %v", err)
	}
	if doc.ID == 0 {
		t.Fatalf("Write did not set ID")
	}

	docs, err = db.Enum("ID", 0, 10)
	if err != nil {
		t.Fatalf("enum on single-doc db: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("enum on single-doc db found %d docs", len(docs))
	}
	if *doc != *docs[0] {
		t.Fatalf("enum returned wrong doc:\n%+v\n%+v", *doc, *docs[0])
	}

	docs, err = db.Search("Desc:description", "ID", 0, 10)
	if err != nil {
		t.Fatalf("search for description: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("search for description found %d docs", len(docs))
	}
	if *doc != *docs[0] {
		t.Fatalf("search for description returned wrong doc:\n%+v\n%+v", *doc, *docs[0])
	}

	docs, err = db.Search("tags:Tag1 tags:tag3", "ID", 0, 10)
	if err != nil {
		t.Fatalf("search for tags: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("search for tags found %d docs", len(docs))
	}
	if *doc != *docs[0] {
		t.Fatalf("search for tags returned wrong doc:\n%+v\n%+v", *doc, *docs[0])
	}

	if err := db.Write(doc); err != nil {
		t.Fatalf("rewrite: %v", err)
	}

	docs, err = db.Search("tags:Tag1 tags:tag4", "ID", 0, 10)
	if err != nil {
		t.Fatalf("impossible search for tags: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("impossible search for tags found %d docs", len(docs))
	}

	if err := db.Delete(doc); err != nil {
		t.Fatalf("cannot delete doc: %v", err)
	}

	docs, err = db.Enum("ID", 0, 10)
	if err != nil {
		t.Fatalf("enum on newly empty db: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("enum on newly empty db found %d docs", len(docs))
	}

	docs, err = db.Search("description", "ID", 0, 10)
	if err != nil {
		t.Fatalf("search on newly empty db: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("search on newly empty db found %d docs", len(docs))
	}

}
