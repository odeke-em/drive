// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbstore

import (
	"database/sql"
	"testing"
	"time"

	_ "code.google.com/p/gosqlite/sqlite3"
)

type Data1 struct {
	Create string
	Table  string
	On     string
}

func TestBasic(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	storage := new(Storage)
	storage.Register(new(Data1))
	if err := storage.CreateTables(db); err != nil {
		t.Fatal(err)
	}

	d1a := Data1{Create: "key", Table: "hello, world", On: "other world"}
	if err := storage.Insert(db, &d1a); err != nil {
		t.Fatal(err)
	}

	var d1b Data1
	if err := storage.Select(db, &d1b, ""); err != nil {
		t.Fatal(err)
	}

	if d1b != d1a {
		t.Errorf("wrong value from Select: have %+v want %+v", d1b, d1a)
	}

	var d1c Data1
	d1c.Create = "key"
	if err := storage.Read(db, &d1c, "Table"); err != nil {
		t.Fatal(err)
	}
	if d1c.Table != d1a.Table || d1c.On != "" {
		t.Errorf("wrong value from Read: have %q, %q want %q, %q", d1c.Table, d1c.On, d1a.Table, "")
	}

	d1a.On = "new"
	d1a.Table = "hi"
	if err := storage.Write(db, &d1a, "On"); err != nil {
		t.Fatal(err)
	}

	var d1d Data1
	d1d.Create = "key"
	if err := storage.Read(db, &d1c, "Table"); err != nil {
		t.Fatal(err)
	}
	if d1c.Table != "hello, world" || d1c.On != "" {
		t.Errorf("wrong value from Read: have %q, %q want %q, %q", d1c.Table, d1c.On, "hello, world", "")
	}
}

func TestRowidInsert(t *testing.T) {
	Debug = true
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	storage := new(Storage)
	storage.Register(new(Msg))
	if err := storage.CreateTables(db); err != nil {
		t.Fatal(err)
	}

	var t1 Msg
	t1.X = 123
	if err := storage.Insert(db, &t1); err != nil {
		t.Fatal(err)
	}
	t1.X = 234
	if err := storage.Insert(db, &t1); err != nil {
		t.Fatal(err)
	}

	var all []Msg
	if err := storage.Select(db, &all, "order by X"); err != nil {
		t.Fatal(err)
	}

	if len(all) != 2 || all[0].X != 123 || all[1].X != 234 {
		t.Fatal("wrong results: %v", all)
	}
}

type Msg struct {
	X int64 `dbstore:",rowid"`
	Y time.Time
	Z bool
}
