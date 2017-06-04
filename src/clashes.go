// Copyright 2016 Google Inc. All Rights Reserved.
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
	"path/filepath"
	"strings"
)

type FixClashesMode uint8

const (
	FixClashesRename FixClashesMode = 1 + iota
	FixClashesTrash
)

func (g *Commands) ListClashes(byId bool) error {
	spin := g.playabler()
	spin.play()
	clashes, err := listClashes(g, g.opts.Sources, byId)
	spin.stop()

	if len(clashes) < 1 {
		if err == nil {
			return fmt.Errorf("no clashes exist!")
		} else {
			return err
		}
	}

	if err != nil && err != ErrClashesDetected {
		return err
	}

	warnClashesPersist(g.log, clashes)
	return nil
}

func (g *Commands) FixClashes(byId bool) error {
	spin := g.playabler()
	spin.play()
	clashes, err := listClashes(g, g.opts.Sources, byId)
	spin.stop()

	if len(clashes) < 1 {
		return err
	}

	if err != nil && err != ErrClashesDetected {
		return err
	}

	fn := g.opts.clashesHandler()
	return fn(g, clashes)
}

type clashesHandler func(g *Commands, clashes []*Change) error

// clashesHandler returns the appropriate clashes handler depending
// on the FixMode ie whether renaming or trashing is to be done.
func (opts *Options) clashesHandler() clashesHandler {
	fn := autoRenameClashes
	if opts.FixClashesMode == FixClashesTrash {
		fn = autoTrashClashes
	}

	return fn
}

func findClashesForChildren(g *Commands, parentId, relToRootPath string, depth int) (clashes []*Change, err error) {
	if depth == 0 {
		return
	}

	memoized := map[string][]*File{}
	pagePair := g.rem.FindByParentId(parentId, g.opts.Hidden)
	decrementedDepth := decrementTraversalDepth(depth)

	children := pagePair.filesChan
	discoveryOrder := []string{}
	for child := range children {
		if child == nil {
			continue
		}
		cluster, alreadyDiscovered := memoized[child.Name]
		if !alreadyDiscovered {
			discoveryOrder = append(discoveryOrder, child.Name)
		}

		cluster = append(cluster, child)
		memoized[child.Name] = cluster
	}

	// Now ensure that no error ensued during pagination/retrieval
	for rErr := range pagePair.errsChan {
		if rErr != nil {
			err = rErr
			return clashes, err
		}
	}

	separatorPrefix := relToRootPath
	if rootLike(separatorPrefix) {
		// Avoid a situation where you have Join("/", "/", "a") -> "//a"
		separatorPrefix = ""
	}

	// To preserve the discovery order
	for _, commonKey := range discoveryOrder {
		cluster, _ := memoized[commonKey]
		fullRelToRootPath := sepJoin(RemoteSeparator, separatorPrefix, commonKey)
		nameClashesPresent := len(cluster) > 1

		for _, rem := range cluster {
			if nameClashesPresent {
				change := &Change{
					Src: rem, g: g,
					Path:   fullRelToRootPath,
					Parent: g.opts.Path,
				}

				clashes = append(clashes, change)
			}

			ccl, cErr := findClashesForChildren(g, rem.Id, fullRelToRootPath, decrementedDepth)
			clashes = append(clashes, ccl...)
			if cErr != nil {
				err = combineErrors(err, cErr)
			}
		}
	}

	return
}

func findClashesByPath(g *Commands, relToRootPath string, depth int) (clashes []*Change, err error) {
	if depth == 0 {
		return
	}

	iterCount := uint64(0)
	clashThresholdCount := uint64(1)

	cl := []*Change{}

	pagePair := g.rem.FindByPathM(relToRootPath)

	working := true
	for working {
		select {
		case err := <-pagePair.errsChan:
			if err != nil {
				return clashes, err
			}
		case rem, stillHasContent := <-pagePair.filesChan:
			if !stillHasContent {
				working = false
				break
			}
			if rem == nil {
				continue
			}

			iterCount++

			change := &Change{
				Src: rem, g: g,
				Path:   relToRootPath,
				Parent: g.opts.Path,
			}

			cl = append(cl, change)
		}
	}

	if iterCount > clashThresholdCount {
		clashes = append(clashes, cl...)
	}

	// Check for any errors that were
	// encountered during remote file retrieval
	for pageErr := range pagePair.errsChan {
		if pageErr == nil {
			return clashes, pageErr
		}
	}

	decrementedDepth := decrementTraversalDepth(depth)
	if decrementedDepth == 0 {
		return
	}

	for _, change := range cl {
		ccl, cErr := findClashesForChildren(g, change.Src.Id, change.Path, decrementedDepth)
		if cErr != nil {
			err = combineErrors(err, cErr)
		}

		clashes = append(clashes, ccl...)
	}

	return
}

func listClashes(g *Commands, sources []string, byId bool) (clashes []*Change, err error) {
	for _, relToRootPath := range g.opts.Sources {
		cclashes, cErr := findClashesByPath(g, relToRootPath, g.opts.Depth)

		clashes = append(clashes, cclashes...)

		if cErr != nil {
			err = combineErrors(err, cErr)
		}
	}

	return
}

func autoRenameClashes(g *Commands, clashes []*Change) error {
	clashesMap := map[string][]*Change{}

	for _, clash := range clashes {
		group := clashesMap[clash.Path]
		if clash.Src == nil {
			continue
		}
		clashesMap[clash.Path] = append(group, clash)
	}

	renames := make([]renameOp, 0, len(clashesMap)) // we will have at least len(clashesMap) renames

	for commonPath, group := range clashesMap {
		ext := filepath.Ext(commonPath)
		name := strings.TrimSuffix(commonPath, ext)
		nextIndex := 0

		for i, n := 1, len(group); i < n; i++ { // we can leave first item with original name
			var newName string
			for {
				newName = fmt.Sprintf("%v_%d%v", name, nextIndex, ext)
				nextIndex++

				dupCheck, err := g.rem.FindByPath(newName)
				if err != nil && err != ErrPathNotExists {
					return err
				}

				if dupCheck == nil {
					newName = filepath.Base(newName)
					break
				}
			}

			r := renameOp{newName: newName, change: group[i], originalPath: commonPath}
			renames = append(renames, r)
		}
	}

	if g.opts.canPrompt() {
		g.log.Logln("Some clashes found, we can fix them by following renames:")
		for _, r := range renames {
			g.log.Logf("%v %v -> %v\n", r.originalPath, r.change.Src.Id, r.newName)
		}
		status := promptForChanges("Proceed with the changes [Y/n] ? ")
		if !accepted(status) {
			return ErrClashFixingAborted
		}
	}

	var composedError error

	quot := customQuote
	for _, r := range renames {
		message := fmt.Sprintf("Renaming %s %s -> %s\n", quot(r.originalPath), quot(r.change.Src.Id), quot(r.newName))
		_, err := g.rem.rename(r.change.Src.Id, r.newName)
		if err == nil {
			g.log.Log(message)
			continue
		}

		composedError = reComposeError(composedError, fmt.Sprintf("%s err: %v", message, err))
	}

	return composedError
}

func autoTrashClashes(g *Commands, clashes []*Change) error {
	for _, c := range clashes {
		// Let's coerce this change to a deletion
		c.Dest = c.Src
		c.Src = nil
	}

	if g.opts.canPrompt() {
		g.log.Logln("Some clashes found, trash them all?")
		for _, c := range clashes {
			g.log.Logln(c.Symbol(), c.Path, c.Dest.Id)
		}
		if status := promptForChanges(); !accepted(status) {
			return status.Error()
		}
	}

	opt := trashOpt{
		toTrash:   true,
		permanent: false,
	}
	return g.playTrashChangeList(clashes, &opt)
}
