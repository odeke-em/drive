// Copyright 2013 Google Inc. All Rights Reserved.
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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/odeke-em/drive/config"
	"github.com/odeke-em/dts/ascii-trie"
	"github.com/odeke-em/log"
)

type destination int

const (
	SelectSrc = 1 << iota
	SelectDest
)

type dirList struct {
	remote *File
	local  *File
}

func (d *dirList) Name() string {
	if d.remote != nil {
		return d.remote.Name
	}
	return d.local.Name
}

type sizeCounter struct {
	count int64
	src   int64
	dest  int64
}

func (t *sizeCounter) String() string {
	str := fmt.Sprintf("count %v", t.count)
	if t.src > 0 {
		str = fmt.Sprintf("%s src: %v", str, prettyBytes(t.src))
	}
	if t.dest > 0 {
		str = fmt.Sprintf("%s dest: %v", str, prettyBytes(t.dest))
	}
	return str
}

// Resolves the local path relative to the root directory
// Returns the path relative to the remote, the abspath on disk and an error if any
func (g *Commands) pathResolve() (relPath, absPath string, err error) {
	root := g.context.AbsPathOf("")
	absPath = g.context.AbsPathOf(g.opts.Path)
	relPath = ""

	if absPath != root {
		relPath, err = filepath.Rel(root, absPath)
		if err != nil {
			return
		}
	} else {
		var cwd string
		if cwd, err = os.Getwd(); err != nil {
			return
		}
		if cwd == root {
			relPath = ""
		} else if relPath, err = filepath.Rel(root, cwd); err != nil {
			return
		}
	}
	relPath = strings.Join([]string{"", relPath}, "/")

	return
}

func (g *Commands) resolveToLocalFile(relToRoot, fsPath string) (local *File, err error) {
	if g.opts.IgnoreRegexp != nil && g.opts.IgnoreRegexp.Match([]byte(relToRoot)) {
		err = fmt.Errorf("\n'%s' is set to be ignored yet is being processed. Use `%s` to override this\n", relToRoot, ForceKey)
		return
	}

	localinfo, statErr := os.Stat(fsPath)
	if statErr != nil && !os.IsNotExist(statErr) {
		err = statErr
		return
	}

	if localinfo != nil {
		if namedPipe(localinfo.Mode()) {
			err = fmt.Errorf("%s (%s) is a named pipe, yet not reading from it", relToRoot, fsPath)
			return
		}

		local = NewLocalFile(fsPath, localinfo)
	}

	return
}

func (g *Commands) byRemoteResolve(relToRoot, fsPath string, r *File, isPush bool) (cl []*Change, err error) {
	var l *File
	l, err = g.resolveToLocalFile(relToRoot, fsPath)
	if err != nil {
		return cl, err
	}

	return g.doChangeListRecv(relToRoot, fsPath, l, r, isPush)
}

func (g *Commands) changeListResolve(relToRoot, fsPath string, isPush bool) (cl []*Change, err error) {
	var r *File
	r, err = g.rem.FindByPath(relToRoot)
	if err != nil && err != ErrPathNotExists {
		return
	}

	return g.byRemoteResolve(relToRoot, fsPath, r, isPush)
}

func (g *Commands) doChangeListRecv(relToRoot, fsPath string, l, r *File, isPush bool) (cl []*Change, err error) {
	if l == nil && r == nil {
		err = fmt.Errorf("'%s' aka '%s' doesn't exist locally nor remotely",
			relToRoot, fsPath)
		return
	}

	return g.resolveChangeListRecv(isPush, relToRoot, relToRoot, r, l)
}

func (g *Commands) clearMountPoints() {
	if g.opts.Mount == nil {
		return
	}
	mount := g.opts.Mount
	for _, point := range mount.Points {
		point.Unmount()
	}

	if mount.CreatedMountDir != "" {
		if rmErr := os.RemoveAll(mount.CreatedMountDir); rmErr != nil {
			g.log.LogErrf("clearMountPoints removing %s: %v\n", mount.CreatedMountDir, rmErr)
		}
	}
	if mount.ShortestMountRoot != "" {
		if rmErr := os.RemoveAll(mount.ShortestMountRoot); rmErr != nil {
			g.log.LogErrf("clearMountPoints: shortestMountRoot: %v\n", mount.ShortestMountRoot, rmErr)
		}
	}
}

func (g *Commands) differ(a, b *File) bool {
	return fileDifferences(a, b, g.opts.IgnoreChecksum) == DifferNone
}

func (g *Commands) coercedMimeKey() (coerced string, ok bool) {
	if g.opts.Meta == nil {
		return
	}

	var values []string
	dict := *g.opts.Meta
	values, ok = dict[CoercedMimeKeyKey]

	if len(values) >= 1 {
		coerced = values[0]
	} else {
		ok = false
	}

	return
}

func (g *Commands) resolveChangeListRecv(
	isPush bool, d, p string, r *File, l *File) (cl []*Change, err error) {
	var change *Change

	explicitlyRequested := hasExportLinks(r) && len(g.opts.Exports) >= 1

	if isPush {
		// Handle the case of doc files for which we don't have a direct download
		// url but have exportable links. These files should not be clobbered on push
		if hasExportLinks(r) {
			return cl, nil
		}
		change = &Change{Path: p, Src: l, Dest: r, Parent: d}
	} else {
		exportable := !g.opts.Force && hasExportLinks(r)
		if exportable && !explicitlyRequested {
			// The case when we have files that don't provide the download urls
			// but exportable links, we just need to check that mod times are the same.
			mask := fileDifferences(r, l, g.opts.IgnoreChecksum)
			if !dirTypeDiffers(mask) && !modTimeDiffers(mask) {
				return cl, nil
			}
		}
		change = &Change{Path: p, Src: r, Dest: l, Parent: d}
	}

	change.NoClobber = g.opts.NoClobber
	change.IgnoreChecksum = g.opts.IgnoreChecksum

	if explicitlyRequested {
		change.Force = true
	} else {
		change.Force = g.opts.Force
	}

	forbiddenOp := (g.opts.ExcludeCrudMask & change.crudValue()) != 0
	if forbiddenOp {
		return cl, nil
	}

	if change.Op() != OpNone {
		cl = append(cl, change)
	}

	if !g.opts.Recursive {
		return cl, nil
	}

	// TODO: handle cases where remote and local type don't match
	if !isPush && r != nil && !r.IsDir {
		return cl, nil
	}
	if isPush && l != nil && !l.IsDir {
		return cl, nil
	}

	// look-up for children
	var localChildren chan *File
	if l == nil || !l.IsDir {
		localChildren = make(chan *File)
		close(localChildren)
	} else {
		localChildren, err = list(g.context, p, g.opts.Hidden, g.opts.IgnoreRegexp)
		if err != nil {
			return
		}
	}

	var remoteChildren chan *File
	if r != nil {
		remoteChildren = g.rem.FindByParentId(r.Id, g.opts.Hidden)
	} else {
		remoteChildren = make(chan *File)
		close(remoteChildren)
	}
	dirlist, clashes := merge(remoteChildren, localChildren, g.opts.IgnoreNameClashes)

	if !g.opts.IgnoreNameClashes && len(clashes) >= 1 {
		if rootLike(p) {
			p = ""
		}

		for _, dup := range clashes {
			g.log.LogErrf("\033[91mX\033[00m %s/%v \"%v\"\n", p, dup.Name, dup.Id)
		}
		err = ErrClashesDetected
		return
	}

	// Arbitrary value. TODO: Calibrate or calculate this value
	chunkSize := 100
	srcLen := len(dirlist)
	chunkCount, remainder := srcLen/chunkSize, srcLen%chunkSize
	i := 0

	if remainder != 0 {
		chunkCount += 1
	}

	var wg sync.WaitGroup
	wg.Add(chunkCount)

	for j := 0; j < chunkCount; j += 1 {
		end := i + chunkSize
		if end >= srcLen {
			end = srcLen
		}

		go func(wg *sync.WaitGroup, isPush bool, cl *[]*Change, p string, dlist []*dirList) {
			defer wg.Done()
			for _, l := range dlist {
				// Avoiding path.Join which normalizes '/+' to '/'
				var joined string
				if p == "/" {
					joined = "/" + l.Name()
				} else {
					joined = strings.Join([]string{p, l.Name()}, "/")
				}
				childChanges, cErr := g.resolveChangeListRecv(isPush, p, joined, l.remote, l.local)
				if cErr == nil {
					*cl = append(*cl, childChanges...)
					continue
				}

				if cErr == ErrClashesDetected {
					break
				} else if cErr != ErrPathNotExists {
					g.log.LogErrf("%s: %v\n", p, cErr)
					break
				}
			}
		}(&wg, isPush, &cl, p, dirlist[i:end])

		i += chunkSize
	}
	wg.Wait()
	return cl, nil
}

func merge(remotes, locals chan *File, ignoreClashes bool) (merged []*dirList, clashesChan []*File) {
	localsTrie := asciitrie.New()
	remotesTrie := asciitrie.New()

	// TODO: Add support for FileSystems that allow same names but different files.
	for l := range locals {
		localsTrie.Set(l.Name, l)
	}

	for r := range remotes {
		list := &dirList{remote: r}

		if !ignoreClashes {
			evicted := remotesTrie.Set(r.Name, r)
			if evicted != nil {
				evictedFile, ok := evicted.(*File)
				if ok {
					clashesChan = append(clashesChan, evictedFile)
				}
				continue
			}
		}

		mem, ok := localsTrie.Get(r.Name)
		// look for local
		if ok {
			l, lOk := mem.(*File)
			if lOk && (l.IsDir == r.IsDir) {
				list.local = l
				localsTrie.Pop(r.Name)
			}
		}
		merged = append(merged, list)
	}

	// if anything left in locals, add to the dir listing
	walk := localsTrie.Walk()
	for item := range walk {
		l, ok := item.(*File)
		if !ok {
			continue
		}
		merged = append(merged, &dirList{local: l})
	}
	return
}

func reduceToSize(changes []*Change, destMask destination) (srcSize, destSize int64) {
	fromSrc := (destMask & SelectSrc) != 0
	fromDest := (destMask & SelectDest) != 0

	for _, c := range changes {
		if fromSrc && c.Src != nil {
			srcSize += c.Src.Size
		}
		if fromDest && c.Dest != nil {
			destSize += c.Dest.Size
		}
	}
	return
}

func conflict(src, dest *File, index *config.Index, push bool) bool {
	// Never been indexed means no local record.
	if index == nil {
		return false
	}

	// Check if this was only a one sided edit for a push
	if push && dest != nil && dest.ModTime.Unix() == index.ModTime {
		return false
	}

	rounded := src.ModTime.UTC().Round(time.Second)
	if rounded.Unix() != index.ModTime && src.Md5Checksum != index.Md5Checksum {
		return true
	}
	return false
}

func resolveConflicts(conflicts []*Change, push bool, indexFiler func(string) *config.Index) (resolved, unresolved []*Change) {
	if len(conflicts) < 1 {
		return
	}
	for _, ch := range conflicts {
		l, r := ch.Dest, ch.Src
		if push {
			l, r = ch.Src, ch.Dest
		}
		fileId := ""
		if l != nil {
			fileId = l.Id
		}
		if fileId == "" && r != nil {
			fileId = r.Id
		}
		if !conflict(l, r, indexFiler(fileId), push) {
			// Time to disregard this conflict if any
			if ch.Op() == OpModConflict {
				ch.IgnoreConflict = true
			}
			resolved = append(resolved, ch)
		} else {
			unresolved = append(unresolved, ch)
		}
	}
	return
}

func sift(changes []*Change) (nonConflicts, conflicts []*Change) {
	// Firstly detect the conflicting changes and if present return false
	for _, c := range changes {
		if c.Op() == OpModConflict {
			conflicts = append(conflicts, c)
		} else {
			nonConflicts = append(nonConflicts, c)
		}
	}
	return
}

func conflictsPersist(conflicts []*Change) bool {
	return len(conflicts) >= 1
}

func warnConflictsPersist(logy *log.Logger, conflicts []*Change) {
	logy.LogErrf("These %d file(s) would be overwritten. Use -%s to override this behaviour\n", len(conflicts), CLIOptionIgnoreConflict)
	for _, conflict := range conflicts {
		logy.LogErrln(conflict.Path)
	}
}

func opChangeCount(changes []*Change) map[Operation]sizeCounter {
	opMap := map[Operation]sizeCounter{}

	for _, c := range changes {
		op := c.Op()
		if op == OpNone {
			continue
		}
		counter := opMap[op]
		counter.count += 1
		if c.Src != nil && !c.Src.IsDir {
			counter.src += c.Src.Size
		}
		if c.Dest != nil && !c.Dest.IsDir {
			counter.dest += c.Dest.Size
		}
		opMap[op] = counter
	}

	return opMap
}

func previewChanges(logy *log.Logger, cl []*Change, reduce bool, opMap map[Operation]sizeCounter) {
	for _, c := range cl {
		op := c.Op()
		if op != OpNone {
			logy.Logln(c.Symbol(), c.Path)
		}
	}

	if reduce {
		for op, counter := range opMap {
			if counter.count < 1 {
				continue
			}
			_, name := op.description()
			logy.Logf("%s %s\n", name, counter.String())
		}
	}
}

func printChangeList(logy *log.Logger, changes []*Change, noPrompt bool, noClobber bool) (bool, *map[Operation]sizeCounter) {
	if len(changes) == 0 {
		logy.Logln("Everything is up-to-date.")
		return false, nil
	}
	if noPrompt {
		return true, nil
	}

	opMap := opChangeCount(changes)
	previewChanges(logy, changes, true, opMap)

	return promptForChanges(), &opMap
}
