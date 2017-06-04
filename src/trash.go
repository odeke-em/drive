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
	// "path/filepath"
)

type trashOpt struct {
	permanent bool
	toTrash   bool
	byId      bool
}

func (g *Commands) Trash(byId bool) (err error) {
	opt := trashOpt{
		toTrash:   true,
		permanent: false,
		byId:      byId,
	}
	return g.reduceForTrash(g.opts.Sources, &opt)
}

func (g *Commands) Delete(byId bool) (err error) {
	opt := trashOpt{
		toTrash:   true,
		permanent: true,
		byId:      byId,
	}
	return g.reduceForTrash(g.opts.Sources, &opt)
}

func (g *Commands) Untrash(byId bool) (err error) {
	opt := trashOpt{
		toTrash:   false,
		permanent: false,
		byId:      byId,
	}
	return g.reduceForTrash(g.opts.Sources, &opt)
}

func (g *Commands) EmptyTrash() error {
	rootFile, err := g.rem.FindByPath("/")
	if err != nil {
		return err
	}

	spin := g.playabler()
	spin.play()
	defer spin.stop()

	if g.opts.canPrompt() {
		travSt := traversalSt{
			depth:            -1,
			file:             rootFile,
			headPath:         "/",
			inTrash:          true,
			mask:             g.opts.TypeMask,
			explicitNoPrompt: true,
		}

		if !g.breadthFirst(travSt, spin) {
			return nil
		}

		g.log.Logln("This operation is irreversible. Empty trash! ")

		if status := promptForChanges(); !accepted(status) {
			g.log.Logln("Aborted emptying trash")
			return status.Error()
		}
	}

	err = g.rem.EmptyTrash()
	if err == nil {
		g.log.Logln("Successfully emptied trash")
	}
	return err
}

func (g *Commands) trasher(relToRoot string, opt *trashOpt) (*Change, error) {
	var file *File
	if relToRoot == "/" && opt.toTrash {
		return nil, immutableAttemptErr(fmt.Errorf("Will not try to trash root."))
	}
	resolver := g.rem.FindByPathTrashed
	if opt.byId {
		resolver = g.rem.FindById
	} else if opt.toTrash {
		resolver = g.rem.FindByPath
	}

	file, err := resolver(relToRoot)
	if err != nil {
		return nil, err
	}

	if opt.byId {
		if file.Labels != nil {
			if file.Labels.Trashed == opt.toTrash {
				return nil, illogicalStateErr(fmt.Errorf("toTrash=%v set yet already file.Trash=%v", opt.toTrash, file.Labels.Trashed))
			}
		}
		relToRoot = fmt.Sprintf("%s (%s)", relToRoot, file.Name)
	}

	change := &Change{Path: relToRoot, g: g}
	if opt.toTrash {
		change.Dest = file
	} else {
		change.Src = file
	}
	return change, nil
}

func (g *Commands) trashByMatch(inTrash, permanent bool) error {
	mq := matchQuery{
		dirPath: g.opts.Path,
		inTrash: false,
		titleSearches: []fuzzyStringsValuePair{
			{fuzzyLevel: Like, values: g.opts.Sources, inTrash: inTrash},
		},
	}

	var cl []*Change
	p := g.opts.Path
	if p == "/" {
		p = ""
	}

	pagePair := g.rem.FindMatches(&mq)
	errsChan := pagePair.errsChan
	matches := pagePair.filesChan

	working := true
	for working {
		select {
		case err := <-errsChan:
			if err != nil {
				return err
			}
		case match, stillHasContent := <-matches:
			if !stillHasContent {
				working = false
				break
			}
			if match == nil {
				continue
			}
			ch := &Change{Path: p + "/" + match.Name, g: g}
			if inTrash {
				ch.Src = match
			} else {
				ch.Dest = match
			}
			cl = append(cl, ch)
		}
	}

	if len(cl) < 1 {
		return noMatchesFoundErr(fmt.Errorf("no matches found!"))
	}

	clArg := changeListArg{
		logy:       g.log,
		changes:    cl,
		noClobber:  false,
		noPrompt:   !g.opts.canPrompt(),
		canPreview: g.opts.canPreview(),
	}

	status, _ := printChangeList(&clArg)
	if !accepted(status) {
		return status.Error()
	}

	toTrash := !inTrash
	opt := trashOpt{
		toTrash:   toTrash,
		permanent: permanent,
	}

	return g.playTrashChangeList(cl, &opt)
}

func (g *Commands) TrashByMatch() error {
	return g.trashByMatch(false, false)
}

func (g *Commands) UntrashByMatch() error {
	return g.trashByMatch(true, false)
}

func (g *Commands) DeleteByMatch() error {
	return g.trashByMatch(false, true)
}

func (g *Commands) reduceForTrash(args []string, opt *trashOpt) error {
	var cl []*Change
	for i, relToRoot := range args {
		c, cErr := g.trasher(relToRoot, opt)
		g.DebugPrintf("[reduceForTrash] #%d: relToRoot: %s\n", i, relToRoot)
		if cErr != nil {
			g.log.LogErrf("\033[91m'%s': %v\033[00m\n", relToRoot, cErr)
		} else if c != nil {
			cl = append(cl, c)
		}
	}

	clArg := changeListArg{
		logy:       g.log,
		changes:    cl,
		noClobber:  false,
		canPreview: g.opts.canPreview(),
		noPrompt:   !g.opts.canPrompt(),
	}

	status, _ := printChangeList(&clArg)
	if !accepted(status) {
		return status.Error()
	}

	if opt.permanent && g.opts.canPrompt() {
		status := promptForChanges("This operation is irreversible. Continue [Y/n] ")
		if !accepted(status) {
			return status.Error()
		}
	}
	return g.playTrashChangeList(cl, opt)
}

func (g *Commands) playTrashChangeList(cl []*Change, opt *trashOpt) (err error) {
	trashSize, unTrashSize := reduceToSize(cl, SelectDest|SelectSrc)
	g.taskStart(trashSize + unTrashSize)

	var fn func(*Change) error
	if opt.permanent {
		fn = g.remoteDelete
		g.DebugPrintf("[playTrashChangeList] isPermanentOp. Selecting remoteDelete\n")
	} else {
		fn = g.remoteUntrash
		if opt.toTrash {
			fn = g.remoteTrash
		}
		g.DebugPrintf("[playTrashChangeList/nonPermanentOp]: toTrash: %v\n", opt.toTrash)
	}

	for i, c := range cl {
		op := c.Op()
		g.DebugPrintf("[playTrashChangeList] #%d op: %v change: %#v\n", i, op, c)
		if c.Op() == OpNone {
			continue
		}

		cErr := fn(c)
		if cErr != nil {
			g.log.LogErrln(cErr)
		}
	}

	g.taskFinish()
	return err
}
