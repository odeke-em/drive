// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dbstore stores and retrieves Go data structures as rows in a SQL database.
//
// Each struct type is stored in its own table, and each field is a separate column of that table.
// This package makes it easy to store structs into such a database and to read them back out.
package dbstore

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

// A Storage records information about the data structures being stored.
// It must be initialized by one or more calls to Register before the other methods are called.
type Storage struct {
	types         []*dtype
	typeByReflect map[reflect.Type]*dtype
}

type dtype struct {
	name        string
	fields      []*field
	fieldByName map[string]*field
	keys        []*field
	rowid       *field
	fts4        bool
}

type field struct {
	key     bool
	rowid   bool
	autoinc bool
	utf8    bool
	name    string
	dbtype  string
	index   []int
}

// A Context represents the underlying SQL database.
// Typically a *sql.DB is used as the Context implementation,
// but the interface allows debugging adapters to be substituted.
type Context interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

const (
	ptrStruct = iota
	ptrPtrStruct
	ptrSliceStruct
	ptrSlicePtrStruct
)

type debugCtxt struct {
	ctxt Context
}

func (d *debugCtxt) Exec(query string, args ...interface{}) (sql.Result, error) {
	fmt.Fprintf(os.Stderr, "SQL: %s %v\n", query, args)
	return d.ctxt.Exec(query, args...)
}

func (d *debugCtxt) Query(query string, args ...interface{}) (*sql.Rows, error) {
	fmt.Fprintf(os.Stderr, "SQL: %s %v\n", query, args)
	return d.ctxt.Query(query, args...)
}

// If Debug is set to true, each Storage method will print a log of the SQL
// commands being executed.
var Debug = false

func debugContext(ctxt Context) Context {
	if Debug {
		return &debugCtxt{ctxt}
	}
	return ctxt
}

func (db *Storage) findType(val interface{}, op string) (*dtype, int, error) {
	t := reflect.TypeOf(val)
	if t.Kind() != reflect.Ptr {
		return nil, 0, fmt.Errorf("invalid type %T - must be pointer", val)
	}
	t = t.Elem()
	kind := ptrStruct
	if op == "Select" {
		if t.Kind() == reflect.Slice {
			kind = ptrSliceStruct
			t = t.Elem()
		}
		if t.Kind() == reflect.Ptr {
			kind++
			t = t.Elem()
		}
	}
	if t.Kind() != reflect.Struct {
		return nil, 0, fmt.Errorf("invalid type %T - %s should be struct", val, t.String())
	}
	dt := db.typeByReflect[t]
	if dt == nil {
		return nil, 0, fmt.Errorf("type %s not registered", t.String())
	}
	return dt, kind, nil
}

// Register records that the storage should store values with the type of val,
// which should be a pointer to a struct with exported fields.
func (db *Storage) Register(val interface{}) {
	t := reflect.TypeOf(val)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct || t.Elem().Name() == "" {
		panic(fmt.Sprintf("dbstore.Register: type %T Is not pointer to named struct"))
	}
	t = t.Elem()
	dt := &dtype{
		name:        t.PkgPath() + "." + t.Name(),
		fieldByName: make(map[string]*field),
	}
	first := true
	haveKey := false
	haveRowid := false
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("dbstore")
		if tag == "-" {
			continue
		}
		xname := f.Name
		x := strings.Split(tag, ",")
		if x[0] != "" {
			xname = x[0]
		}
		df := &field{
			name:  xname,
			index: []int{i},
		}
		for _, attr := range x[1:] {
			switch attr {
			case "fts4":
				if !first {
					panic(fmt.Sprintf("dbstore.Register %s: fts4 must be attribute on first field", dt.name))
				}
				dt.fts4 = true
			case "rowid":
				if haveKey {
					panic(fmt.Sprintf("dbstore.Register %s: cannot use rowid and key attributes in same struct", dt.name))
				}
				if haveRowid {
					panic(fmt.Sprintf("dbstore.Register %s: cannot use rowid attribute ion multiple fields", dt.name))
				}
				df.rowid = true
			case "key":
				if haveRowid {
					panic(fmt.Sprintf("dbstore.Register %s: cannot use rowid and key attributes in same struct", dt.name))
				}
				df.key = true
			case "autoinc":
				df.autoinc = true
			case "utf8":
				df.utf8 = true
				if f.Type.Kind() != reflect.String {
					panic(fmt.Sprintf("dbstore.Register %s: field %s has attr utf8 but type %s", dt.name, f.Name, f.Type))
				}
			}
		}
		if df.autoinc && !df.rowid {
			panic(fmt.Sprintf("dbstore.Register %s: cannot use autoinc without rowid attribute", dt.name))
		}
		if df.rowid && f.Type.Kind() != reflect.Int64 {
			panic(fmt.Sprintf("dbstore.Register %s: rowid attribute must be used with int64 field", dt.name))
		}

		switch f.Type.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
			df.dbtype = "integer"
			if df.rowid {
				df.dbtype += " primary key"
				if df.autoinc {
					df.dbtype += " autoincrement"
				}
			}
		case reflect.Float32, reflect.Float64:
			df.dbtype = "real"
		case reflect.Bool, reflect.String:
			// ok
		case reflect.Struct:
			if f.Type != reflect.TypeOf(time.Time{}) {
			}
			df.dbtype = "timestamp"
		case reflect.Slice:
			if f.Type.Elem() != reflect.TypeOf(byte(0)) {
			}
			df.dbtype = "blob"
		}

		if dt.fts4 {
			df.dbtype = ""
		}
		first = false

		dt.fields = append(dt.fields, df)
		dt.fieldByName[xname] = df
		if df.rowid {
			dt.rowid = df
		}
		if df.key {
			dt.keys = append(dt.keys, df)
		}
		if dt.fts4 && df.rowid {
			df.name = "docid" // required name for FTS4 rowid column
		}
	}

	if len(dt.fields) == 0 {
		panic(fmt.Sprintf("dbstore.Register %s: no fields to store", dt.name))
	}

	if len(dt.keys) == 0 {
		// retroactively make the first field a key
		dt.keys = append(dt.keys, dt.fields[0])
		dt.fields[0].key = true
	}

	db.types = append(db.types, dt)
	if db.typeByReflect == nil {
		db.typeByReflect = make(map[reflect.Type]*dtype)
	}
	db.typeByReflect[t] = dt
}

// NOTE: All these %q should really be a custom quoting mechanism
// that emits sqlite3-double-quoted escapes, but since we don't expect
// to see double quotes or backslashes in the names, using %q is fine.

// CreateTables creates the tables to hold the registered types.
// It only needs to be called when creating a new database.
// Each table is named for the type it stores, in the form "full/import/path.TypeName".
func (db *Storage) CreateTables(ctxt Context) error {
	ctxt = debugContext(ctxt)
	var buf bytes.Buffer
	for _, t := range db.types {
		buf.Reset()
		if t.fts4 {
			fmt.Fprintf(&buf, "create virtual table %q using fts4", t.name)
		} else {
			fmt.Fprintf(&buf, "create table %q", t.name)
		}

		fmt.Fprintf(&buf, " (")
		sep := ""
		for _, col := range t.fields {
			if t.fts4 && col.rowid {
				continue // fts4 rowid is implicit
			}
			fmt.Fprintf(&buf, "%s%q %s", sep, col.name, col.dbtype)
			sep = ","
		}
		if len(t.keys) > 0 && t.rowid == nil {
			fmt.Fprintf(&buf, ", unique (")
			for i, col := range t.keys {
				if i > 0 {
					fmt.Fprintf(&buf, ", ")
				}
				fmt.Fprintf(&buf, "%q", col.name)
			}
			fmt.Fprintf(&buf, ") on conflict replace")
		}
		fmt.Fprintf(&buf, ")")

		if _, err := ctxt.Exec(buf.String()); err != nil {
			return fmt.Errorf("creating table %s [%s]: %v", t.name, buf.String(), err)
		}
	}

	return nil
}

// Insert inserts the value into the database.
func (db *Storage) Insert(ctxt Context, val interface{}) error {
	ctxt = debugContext(ctxt)
	t, _, err := db.findType(val, "Insert")
	if err != nil {
		return err
	}

	var args []interface{}
	rval := reflect.ValueOf(val).Elem()
	var rowcol *field
	for _, col := range t.fields {
		rv := rval.FieldByIndex(col.index)
		if col.rowid && rv.Int() == 0 {
			rowcol = col
			args = append(args, nil)
		} else {
			args = append(args, rval.FieldByIndex(col.index).Interface())
		}
	}

	var buf bytes.Buffer
	if t.fts4 {
		fmt.Fprintf(&buf, "update %q", t.name)
		fmt.Fprintf(&buf, " set ")
		sep := ""
		var args1 []interface{}
		for i, col := range t.fields {
			if !col.key {
				fmt.Fprintf(&buf, "%s%q = ?", sep, col.name)
				sep = ", "
				args1 = append(args1, args[i])
			}
		}
		fmt.Fprintf(&buf, " where ")
		sep = ""
		for i, col := range t.fields {
			if col.key {
				fmt.Fprintf(&buf, "%s%q = ?", sep, col.name)
				sep = ", "
				args1 = append(args1, args[i])
			}
		}

		res, err := ctxt.Exec(buf.String(), args1...)
		if err != nil {
			return err
		}
		count, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		// fall through to ordinary insert command
		buf.Reset()
	}

	fmt.Fprintf(&buf, "insert or replace into %q (", t.name)
	for i, col := range t.fields {
		if i > 0 {
			fmt.Fprintf(&buf, ", ")
		}
		fmt.Fprintf(&buf, "%q", col.name)
	}
	fmt.Fprintf(&buf, ") values (")
	for i := range t.fields {
		if i > 0 {
			fmt.Fprintf(&buf, ", ")
		}
		fmt.Fprintf(&buf, "?")
	}
	fmt.Fprintf(&buf, ")")

	res, err := ctxt.Exec(buf.String(), args...)
	if err != nil {
		return err
	}
	if rowcol != nil {
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		rval.FieldByIndex(rowcol.index).SetInt(id)
	}
	return nil
}

// Delete deletes the value from the database.
// The unique identification fields in val must be set.
//
// Delete executes a command like:
//	delete from Structs
//	where Key1 = val.Key1 and Key2 = val.Key2
func (db *Storage) Delete(ctxt Context, val interface{}) error {
	ctxt = debugContext(ctxt)
	t, _, err := db.findType(val, "Delete")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	var args []interface{}
	rval := reflect.ValueOf(val).Elem()

	fmt.Fprintf(&buf, "delete from %q where ", t.name)
	sep := ""
	for _, col := range t.fields {
		if !col.key {
			continue
		}
		args = append(args, rval.FieldByIndex(col.index).Interface())
		fmt.Fprintf(&buf, "%s%q = ?", sep, col.name)
		sep = " and "
	}

	_, err = ctxt.Exec(buf.String(), args...)
	return err
}

// Read reads the named columns from the database into val.
// The key fields in val must already be set.
//
// Read executes a command like:
//	select columns from Structs
//	where Key1 = val.Key1 AND Key2 = val.Key2
func (db *Storage) Read(ctxt Context, val interface{}, columns ...string) error {
	ctxt = debugContext(ctxt)
	t, _, err := db.findType(val, "Read")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	var args, scanargs []interface{}
	rval := reflect.ValueOf(val).Elem()

	want := make(map[string]bool)
	for _, name := range columns {
		want[name] = true
	}

	fmt.Fprintf(&buf, "select ")
	sep := ""
	var fixes []func()
	for _, col := range t.fields {
		if !want[col.name] && !want["ALL"] {
			continue
		}
		delete(want, col.name)
		if col.key {
			continue // already set
		}
		fmt.Fprintf(&buf, "%s%q", sep, col.name)
		sep = ", "
		scanargs = append(scanargs, rval.FieldByIndex(col.index).Addr().Interface())
		if col.utf8 {
			fixes = append(fixes, func() {
				v := rval.FieldByIndex(col.index)
				s := v.String()
				if !utf8.ValidString(s) {
					v.SetString(string([]rune(s)))
				}
			})
		}
	}
	if sep == "" {
		// nothing to select, but want to provide error if not there.
		// select count of rows.
		fmt.Fprintf(&buf, "count(*)")
		scanargs = append(scanargs, new(int))
	}

	delete(want, "ALL")
	if len(want) != 0 {
		// some column wasn't found
		for _, name := range columns {
			if want[name] {
				return fmt.Errorf("unknown column %q", name)
			}
		}
	}

	fmt.Fprintf(&buf, " from %q where ", t.name)
	sep = ""
	for _, col := range t.fields {
		if !col.key {
			continue
		}
		fmt.Fprintf(&buf, "%s%q = ?", sep, col.name)
		sep = " and "
		args = append(args, rval.FieldByIndex(col.index).Interface())
	}

	rows, err := ctxt.Query(buf.String(), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return ErrNotFound
	}

	if err := rows.Scan(scanargs...); err != nil {
		return err
	}

	for _, fix := range fixes {
		fix()
	}

	return nil
}

// Write writes the named columns from val into the database.
// The key fields in val must already be set and the value must already exist.
//
// Write executes a command like:
//	update Structs
//	set column1 = val.Column1, column2 = val.Column2
//	where Key1 = val.Key1 AND Key2 = val.Key2
func (db *Storage) Write(ctxt Context, val interface{}, columns ...string) error {
	ctxt = debugContext(ctxt)
	t, _, err := db.findType(val, "Write")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	var args []interface{}
	rval := reflect.ValueOf(val).Elem()

	want := make(map[string]bool)
	for _, name := range columns {
		want[name] = true
	}

	fmt.Fprintf(&buf, "update %q set ", t.name)
	sep := ""
	for _, col := range t.fields {
		if !want[col.name] {
			continue
		}
		delete(want, col.name)
		if col.key {
			continue // already set
		}
		fmt.Fprintf(&buf, "%s%q = ?", sep, col.name)
		sep = ", "
		args = append(args, rval.FieldByIndex(col.index).Interface())
	}
	if sep == "" {
		// nothing to set, but want to provide error if not there.
		return db.Read(ctxt, val)
	}

	if len(want) != 0 {
		// some column wasn't found
		for _, name := range columns {
			if want[name] {
				return fmt.Errorf("unknown column %q", name)
			}
		}
	}

	fmt.Fprintf(&buf, " where ")
	sep = ""
	for _, col := range t.fields {
		if !col.key {
			continue
		}
		fmt.Fprintf(&buf, "%s%q = ?", sep, col.name)
		sep = " and "
		args = append(args, rval.FieldByIndex(col.index).Interface())
	}

	res, err := ctxt.Exec(buf.String(), args...)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNotFound
	}

	return nil
}

// ErrNotFound is the error returned by Read, Select, and Write when
// there are no matching records in the database.
var ErrNotFound = errors.New("database record not found")

// Select executes a select command to read one or more rows into val.
// To read values of type Example, val may take any of these types:
//	*Example - read a single Example, returning ErrNotFound if not found
//	**Example - allocate and read a single Example, setting it to nil if not found
//	*[]Example - read a slice of Examples
//	*[]*Example - read a slice of Examples
//
// Select executes a command like
//	select Key1, Key2, Field3, Field4 from Structs
//	<query here>
func (db *Storage) Select(ctxt Context, val interface{}, query string, args ...interface{}) error {
	ctxt = debugContext(ctxt)
	t, kind, err := db.findType(val, "Select")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "select ")
	sep := ""
	var indexes [][]int
	for _, col := range t.fields {
		fmt.Fprintf(&buf, "%s%q", sep, col.name)
		sep = ", "
		indexes = append(indexes, col.index)
	}
	fmt.Fprintf(&buf, " from %q %s", t.name, query)

	rows, err := ctxt.Query(buf.String(), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	rval := reflect.ValueOf(val).Elem()
	rval.Set(reflect.Zero(rval.Type()))
	switch kind {
	case ptrStruct:
		if !rows.Next() {
			return ErrNotFound
		}
		return scan1(rows, t, rval)

	case ptrPtrStruct:
		if !rows.Next() {
			rval.Set(reflect.Zero(rval.Type()))
			return nil
		}
		rval.Set(reflect.New(rval.Type().Elem()))
		return scan1(rows, t, rval.Elem())

	case ptrSliceStruct:
		for rows.Next() {
			n := rval.Len()
			rval.Set(reflect.Append(rval, reflect.Zero(rval.Type().Elem())))
			if err := scan1(rows, t, rval.Index(n)); err != nil {
				return err
			}
		}
		return nil

	case ptrSlicePtrStruct:
		for rows.Next() {
			n := rval.Len()
			rval.Set(reflect.Append(rval, reflect.New(rval.Type().Elem().Elem())))
			if err := scan1(rows, t, rval.Index(n).Elem()); err != nil {
				return err
			}
		}
		return nil
	}

	panic("dbstore: internal error: unexpected kind")
}

func scan1(rows *sql.Rows, t *dtype, rval reflect.Value) error {
	var args []interface{}
	for _, col := range t.fields {
		args = append(args, rval.FieldByIndex(col.index).Addr().Interface())
	}
	if err := rows.Scan(args...); err != nil {
		return err
	}

	for _, col := range t.fields {
		if col.utf8 {
			v := rval.FieldByIndex(col.index)
			s := v.String()
			if !utf8.ValidString(s) {
				v.SetString(string([]rune(s)))
			}
		}
	}

	return nil
}
