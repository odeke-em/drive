// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Scancab is an implementation of a document scanning cabinet.
// A scanning cabinet is like a file cabinet but stores digital scans of documents
// instead of paper copies.
//
// Scancab watches the given directory for PDF files to include in the cabinet.
// Each PDF is assumed to be a separate document: there is no facility for
// grouping multiple PDFs into a single document or for splitting a large PDF
// into multiple documents.
//
// When a new PDF is found, it shows up in the incoming list and can be filed
// by adding metadata. The per-document metadata is:
//
//	file name (immutable)
//	SHA1 hash of content (immutable)
//	file creation time (immutable)
//	document time
//	due date
//	tags
//	description
//	ocr text
//	physical location
//
// Scancab stores metadata and page thumbnails in two sqlite3 databases
// named metadata.db and thumbnail.db, in the same directory as the PDFs.
//
// This program was inspired by github.com/bradfitz/scanningcabinet,
// but the implementation is new. The most significant difference is the
// use of local disk and a local web server instead of App Engine.
package main
