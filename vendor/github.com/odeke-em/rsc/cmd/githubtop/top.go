// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Dump Github top languages.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
)

var (
	langRE = regexp.MustCompile(`<li><a href="(/languages/[^"]+)">([^<>]+)</a></li>`)
	rankRE = regexp.MustCompile(`is the #([0-9]+) most popular language on GitHub`)

	mostPopular = []byte(`is <strong>the most</strong> popular language on GitHub`)
)

type Lang struct {
	Rank int
	Name string
}

type byRank []Lang

func (x byRank) Len() int           { return len(x) }
func (x byRank) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x byRank) Less(i, j int) bool { return x[i].Rank < x[j].Rank }

func main() {
	data := get("https://github.com/languages")
	var all []Lang
	for _, m := range langRE.FindAllSubmatch(data, -1) {
		link := "https://github.com" + string(m[1])
		name := string(m[2])
		data := get(link)
		m := rankRE.FindSubmatch(data)
		var rank int
		if m == nil {
			if !bytes.Contains(data, mostPopular) {
				log.Printf("cannot find rank on %s", link)
				continue
			}
			rank = 1
		} else {
			rank, _ = strconv.Atoi(string(m[1]))
		}
		all = append(all, Lang{rank, name})
	}
	sort.Sort(byRank(all))
	for _, lang := range all {
		fmt.Printf("%2d\t%s\n", lang.Rank, lang.Name)
	}
}

func get(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("reading %s: %v", url, err)
	}
	return data
}
