// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package issue provides access to the Google Code Issue Tracker API.
package issue

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// A Meta holds issue metadata such as summary and owner.
type Meta struct {
	Summary   string
	Status    string
	Duplicate int // if Status == "Duplicate"
	Owner     string
	CC        []string
	Label     []string
}

// An Issue represents a single issue on the tracker.
// The initial report is Comment[0] and is always present.
type Issue struct {
	ID int
	Meta
	Comment []*Comment
}

// A Comment represents a single comment on an issue.
type Comment struct {
	Author string
	Time   time.Time
	Meta   // changes made by this comment
	Text   string
}

type _Feed struct {
	Entry []_Entry `xml:"entry"`
}

type _Entry struct {
	ID        string    `xml:"id"`
	Title     string    `xml:"title"`
	Published time.Time `xml:"published"`
	Content   string    `xml:"content"`
	Updates   []_Update `xml:"updates"`
	Author    struct {
		Name string `xml:"name"`
	} `xml:"author"`
	Owner      string   `xml:"owner>username"`
	Status     string   `xml:"status"`
	Label      []string `xml:"label"`
	MergedInto string   `xml:"mergedInto"`
	CC         []string `xml:"cc>username"`

	Dir      string
	Number   int
	Comments []_Entry
}

type _Update struct {
	Summary    string   `xml:"summary"`
	Owner      string   `xml:"ownerUpdate"`
	Label      string   `xml:"label"`
	Status     string   `xml:"status"`
	MergedInto string   `xml:"mergedInto"`
	CC         []string `xml:"cc>username"`
}

var xmlDebug = false

// Search queries for issues on the tracker for the given project (for example, "go").
// The can string is typically "open" (search only open issues) or "all" (search all issues).
// The format of the can string and the query are documented at
// https://code.google.com/p/support/wiki/IssueTrackerAPI.
func Search(project, can, query string, detail bool, client *http.Client) ([]*Issue, error) {
	if client == nil {
		client = http.DefaultClient
	}
	q := url.Values{
		"q":           {query},
		"max-results": {"1000"},
		"can":         {can},
	}
	u := "https://code.google.com/feeds/issues/p/" + project + "/issues/full?" + q.Encode()
	r, err := client.Get(u)
	if err != nil {
		return nil, err
	}

	if xmlDebug {
		io.Copy(os.Stdout, r.Body)
		r.Body.Close()
		return nil, nil
	}

	var feed _Feed
	err = xml.NewDecoder(r.Body).Decode(&feed)
	r.Body.Close()
	if err != nil {
		return nil, err
	}

	var issues []*Issue
	for i := range feed.Entry {
		e := &feed.Entry[i]
		id := e.ID
		if i := strings.Index(id, "id="); i >= 0 {
			id = id[:i+len("id=")]
		}
		n, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("invalid issue ID %q", id)
		}
		dup, _ := strconv.Atoi(e.MergedInto)
		p := &Issue{
			ID: n,
			Meta: Meta{
				Summary:   strings.Replace(e.Title, "\n", " ", -1),
				Status:    e.Status,
				Duplicate: dup,
				Owner:     e.Owner,
				CC:        e.CC,
				Label:     e.Label,
			},
			Comment: []*Comment{
				{
					Author: e.Author.Name,
					Time:   e.Published,
					Text:   html.UnescapeString(e.Content),
				},
			},
		}
		issues = append(issues, p)
		if detail {
			u := "https://code.google.com/feeds/issues/p/" + project + "/issues/" + id + "/comments/full"
			r, err := client.Get(u)
			if err != nil {
				return nil, err
			}

			var feed _Feed
			err = xml.NewDecoder(r.Body).Decode(&feed)
			r.Body.Close()
			if err != nil {
				return nil, err
			}

			for i := range feed.Entry {
				e := &feed.Entry[i]
				c := &Comment{
					Author: strings.TrimPrefix(e.Title, "Comment by "),
					Time:   e.Published,
					Text:   html.UnescapeString(e.Content),
				}
				p.Comment = append(p.Comment, c)
				for _, up := range e.Updates {
					if up.Summary != "" {
						c.Meta.Summary = up.Summary
					}
					if up.Owner != "" {
						c.Meta.Owner = up.Owner
					}
					if up.Status != "" {
						c.Meta.Status = up.Status
					}
					if up.MergedInto != "" {
						c.Meta.Duplicate, _ = strconv.Atoi(up.MergedInto)
					}
					if up.Label != "" {
						c.Meta.Label = append(c.Meta.Label, up.Label)
					}
					c.Meta.CC = append(c.Meta.CC, up.CC...)
				}
			}
		}
	}

	sort.Sort(BySummary(issues))
	return issues, nil
}

type BySummary []*Issue

func (x BySummary) Len() int           { return len(x) }
func (x BySummary) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x BySummary) Less(i, j int) bool { return x[i].Summary < x[j].Summary }

type ByID []*Issue

func (x ByID) Len() int           { return len(x) }
func (x ByID) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByID) Less(i, j int) bool { return x[i].ID < x[j].ID }
