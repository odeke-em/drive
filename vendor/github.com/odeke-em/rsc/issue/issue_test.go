// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package issue

import "testing"

func TestSearch(t *testing.T) {
	issues, err := Search("go", "all", `id:5490 reporter:rsc`, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) > 0 && issues[0].ID != 5490 {
		t.Fatalf("Search returned issue %d, want 5490", issues[0].ID)
	}
	if len(issues) != 1 {
		t.Fatalf("Search returned %d results, want 1", len(issues))
	}
	p := issues[0]
	c := p.Comment[0]
	if c.Author != "rsc@golang.org" || p.Summary != "oops" {
		t.Fatalf("Search returned Author=%q, Summary=%q, want %q, %q", c.Author, p.Summary, "rsc@golang.org", "oops")
	}
}
