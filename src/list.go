// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drive

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/odeke-em/log"
)

const (
	InTrash = 1 << iota
	Folder
	NonFolder
	Minimal
	Shared
	Owners
	CurrentVersion
)

type attribute struct {
	minimal bool
	mask    int
	parent  string
}

type traversalSt struct {
	file             *File
	headPath         string
	depth            int
	mask             int
	inTrash          bool
	explicitNoPrompt bool
	sorters          []string
	matchQuery       *matchQuery
}

func sorters(opts *Options) (sortKeys []string) {
	if opts == nil || opts.Meta == nil {
		return
	}

	meta := *(opts.Meta)
	retr, ok := meta[SortKey]
	if !ok {
		return
	}
	return retr
}

func (g *Commands) ListMatches() error {

	inTrash := trashed(g.opts.TypeMask)

	mq := g.createMatchQuery(false)

	mq.titleSearches = append(mq.titleSearches, fuzzyStringsValuePair{
		fuzzyLevel: Like, values: g.opts.Sources, inTrash: inTrash, joiner: Or,
	})

	matches, err := g.rem.FindMatches(mq)
	if err != nil {
		return err
	}

	spin := g.playabler()
	spin.play()

	traversalCount := 0

	for match := range matches {
		if match == nil {
			continue
		}

		travSt := traversalSt{
			depth:    g.opts.Depth,
			file:     match,
			headPath: g.opts.Path,
			inTrash:  g.opts.InTrash,
			mask:     g.opts.TypeMask,
			sorters:  sorters(g.opts),
		}

		traversalCount += 1

		if !g.breadthFirst(travSt, spin) {
			break
		}
	}

	spin.stop()

	if traversalCount < 1 {
		g.log.LogErrln("no matches found!")
	}

	return nil
}

func (g *Commands) createMatchQuery(exactMatch bool) *matchQuery {

	mimeQuerySearches := []fuzzyStringsValuePair{}
	titleSearches := []fuzzyStringsValuePair{}
	ownerSearches := []fuzzyStringsValuePair{}

	if g.opts.Meta != nil {
		meta := *(g.opts.Meta)
		skipMimes, sOk := meta[SkipMimeKeyKey]
		if sOk {
			mimeQuerySearches = append(mimeQuerySearches, fuzzyStringsValuePair{
				fuzzyLevel: Not, values: skipMimes, inTrash: g.opts.InTrash, joiner: And,
			})
		}

		matchMimes, mOk := meta[MatchMimeKeyKey]
		if mOk {
			mimeQuerySearches = append(mimeQuerySearches, fuzzyStringsValuePair{
				fuzzyLevel: Is, values: matchMimes, inTrash: g.opts.InTrash, joiner: Or,
			})
		}

		exactTitles, etOk := meta[ExactTitleKey]
		if etOk {
			titleSearches = append(titleSearches, fuzzyStringsValuePair{
				fuzzyLevel: Is, values: exactTitles, inTrash: g.opts.InTrash, joiner: Or,
			})
		}

		exactOwners, eoOk := meta[ExactOwnerKey]
		if eoOk {
			ownerSearches = append(ownerSearches, fuzzyStringsValuePair{
				fuzzyLevel: Is, values: exactOwners, joiner: Or,
			})
		}

		matchOwners, moOk := meta[MatchOwnerKey]
		if moOk {
			ownerSearches = append(ownerSearches, fuzzyStringsValuePair{
				fuzzyLevel: Like, values: matchOwners, joiner: Or,
			})
		}

		notOwner, soOk := meta[NotOwnerKey]
		if soOk {
			ownerSearches = append(ownerSearches, fuzzyStringsValuePair{
				fuzzyLevel: NotIn, values: notOwner, joiner: And,
			})
		}
	}

	mq := matchQuery{
		dirPath:           g.opts.Path,
		inTrash:           g.opts.InTrash,
		mimeQuerySearches: mimeQuerySearches,
		titleSearches:     titleSearches,
		ownerSearches:     ownerSearches,
	}

	return &mq
}

func (g *Commands) List(byId bool) error {
	var kvList []*keyValue

	resolver := g.rem.FindByPath
	if byId {
		resolver = g.rem.FindByIdMulti
	} else if g.opts.InTrash {
		resolver = g.rem.FindByPathTrashed
	}

	mq := g.createMatchQuery(true)

	for _, relPath := range g.opts.Sources {
		rl, rErr := resolver(relPath)
		if rErr != nil && rErr != ErrPathNotExists {
			return fmt.Errorf("%v: '%s'", rErr, relPath)
		}

		for _, r := range rl {
			if r == nil {
				g.log.LogErrf("remote: %s is nil\n", strconv.Quote(relPath))
				continue
			}

			parentPath := ""
			if !byId {
				parentPath = g.parentPather(relPath)
			} else {
				parentPath = r.Id
			}

			if remoteRootLike(parentPath) {
				parentPath = ""
			}
			if remoteRootLike(r.Name) {
				r.Name = ""
			}
			if rootLike(parentPath) {
				parentPath = ""
			}

			kvList = append(kvList, &keyValue{key: parentPath, value: r})
		}
	}

	spin := g.playabler()
	spin.play()
	for _, kv := range kvList {
		if kv == nil || kv.value == nil {
			continue
		}

		travSt := traversalSt{
			depth:      g.opts.Depth,
			file:       kv.value.(*File),
			headPath:   kv.key,
			inTrash:    g.opts.InTrash,
			mask:       g.opts.TypeMask,
			sorters:    sorters(g.opts),
			matchQuery: mq,
		}

		if !g.breadthFirst(travSt, spin) {
			break
		}
	}
	spin.stop()

	return nil
}

func (g *Commands) ListShared() (err error) {
	var kvList []*keyValue

	for _, relPath := range g.opts.Sources {
		sharedRemotes, sErr := g.rem.FindByPathShared(relPath)
		if sErr != nil {
			g.log.LogErrf("%v: '%s'\n", sErr, relPath)
			return sErr
		}

		parentPath := g.parentPather(relPath)

		if remoteRootLike(parentPath) {
			parentPath = ""
		}

		if rootLike(parentPath) {
			parentPath = ""
		}

		for s := range sharedRemotes {
			if remoteRootLike(s.Name) {
				s.Name = ""
			}

			if s == nil {
				continue
			}

			kvList = append(kvList, &keyValue{key: parentPath, value: s})
		}
	}

	spin := g.playabler()
	spin.play()
	for _, kv := range kvList {
		if kv == nil || kv.value == nil {
			continue
		}

		travSt := traversalSt{
			depth:    g.opts.Depth,
			file:     kv.value.(*File),
			headPath: kv.key,
			inTrash:  g.opts.InTrash,
			mask:     g.opts.TypeMask,
		}

		if !g.breadthFirst(travSt, spin) {
			break
		}
	}
	spin.stop()
	return
}

func (f *File) pretty(logy *log.Logger, opt attribute) {
	fmtdPath := sepJoin("/", opt.parent, f.Name)

	if opt.minimal {
		logy.Logf("%s ", fmtdPath)
	} else {
		if f.IsDir {
			logy.Logf("d")
		} else {
			logy.Logf("-")
		}
		if f.Shared {
			logy.Logf("s")
		} else {
			logy.Logf("-")
		}

		if f.UserPermission != nil {
			logy.Logf(" %-10s ", f.UserPermission.Role)
		}
	}

	if owners(opt.mask) && len(f.OwnerNames) >= 1 {
		logy.Logf(" %s ", strings.Join(f.OwnerNames, " & "))
	}

	if version(opt.mask) {
		logy.Logf(" v%d", f.Version)
	}

	if !opt.minimal {
		logy.Logf(" %-10s\t%-10s\t\t%-20s\t%-50s\n", prettyBytes(f.Size), f.Id, f.ModTime, fmtdPath)
	} else {
		logy.Logln()
	}
}

func (g *Commands) breadthFirst(travSt traversalSt, spin *playable) bool {

	opt := attribute{
		minimal: isMinimal(g.opts.TypeMask),
		mask:    travSt.mask,
	}

	opt.parent = ""
	if travSt.headPath != "/" {
		opt.parent = travSt.headPath
	}

	f := travSt.file
	if !f.IsDir {
		f.pretty(g.log, opt)
		return true
	}

	// New head path
	if !(rootLike(opt.parent) && rootLike(f.Name)) {
		opt.parent = sepJoin("/", opt.parent, f.Name)
	}

	// A depth of < 0 means traverse as deep as you can
	if travSt.depth == 0 {
		// At the end of the line, this was successful.
		return true
	} else if travSt.depth > 0 {
		travSt.depth -= 1
	}

	expr := buildExpression(f.Id, travSt.mask, travSt.inTrash)

	if travSt.matchQuery != nil {
		exprExtra := travSt.matchQuery.Stringer()
		expr = sepJoinNonEmpty(" and ", fmt.Sprintf("(%s)", expr), exprExtra)
	}

	req := g.rem.service.Files.List()
	req.Q(expr)
	req.MaxResults(g.opts.PageSize)

	spin.pause()

	canPrompt := !travSt.explicitNoPrompt
	if canPrompt {
		canPrompt = g.opts.canPrompt()
	}

	fileChan := reqDoPage(req, g.opts.Hidden, canPrompt)

	spin.play()

	onlyFiles := (g.opts.TypeMask & NonFolder) != 0

	iterCount := uint64(0)

	var collector []*File

	for file := range fileChan {
		if file == nil {
			return false
		}

		if isHidden(file.Name, g.opts.Hidden) {
			continue
		}

		collector = append(collector, file)
	}

	if len(travSt.sorters) >= 1 {
		collector = g.sort(collector, travSt.sorters...)
	}

	var children []*File
	for _, file := range collector {
		if file.IsDir {
			children = append(children, file)
		}

		// The case in which only directories wanted is covered by the buildExpression clause
		// reason being that only folder are allowed to be roots, including the only files clause
		// would result in incorrect traversal since non-folders don't have children.
		// Just don't print it, however, the folder will still be explored.
		if onlyFiles && file.IsDir {
			continue
		}
		file.pretty(g.log, opt)
		iterCount += 1
	}

	if !travSt.inTrash && !g.opts.InTrash {
		for _, file := range children {
			childSt := traversalSt{
				depth:            travSt.depth,
				file:             file,
				headPath:         opt.parent,
				inTrash:          travSt.inTrash,
				mask:             g.opts.TypeMask,
				explicitNoPrompt: travSt.explicitNoPrompt,
				sorters:          travSt.sorters,
				matchQuery:       travSt.matchQuery,
			}

			if !g.breadthFirst(childSt, spin) {
				return false
			}
		}
		return true
	}

	return iterCount >= 1
}

func isMinimal(mask int) bool {
	return (mask & Minimal) != 0
}

func owners(mask int) bool {
	return (mask & Owners) != 0
}

func version(mask int) bool {
	return (mask & CurrentVersion) != 0
}

func shared(mask int) bool {
	return (mask & Shared) != 0
}

func trashed(mask int) bool {
	return (mask & InTrash) != 0
}
