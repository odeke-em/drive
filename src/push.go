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
	"io/ioutil"
	"os"
	"os/signal"
	gopath "path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/odeke-em/drive/config"
)

var mkdirAllMu = sync.Mutex{}

// Pushes to remote if local path exists and in a gd context. If path is a
// directory, it recursively pushes to the remote if there are local changes.
// It doesn't check if there are local changes if isForce is set.
func (g *Commands) Push() (err error) {
	defer g.clearMountPoints()

	root := g.context.AbsPathOf("")
	var cl []*Change

	g.log.Logln("Resolving...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	spin := g.playabler()
	spin.play()

	// To Ensure mount points are cleared in the event of external exceptions
	go func() {
		_ = <-c
		spin.stop()
		g.clearMountPoints()
		os.Exit(1)
	}()

	// TODO: Look at clashes?
	clashes := []*Change{}

	for _, relToRootPath := range g.opts.Sources {
		fsPath := g.context.AbsPathOf(relToRootPath)
		ccl, cclashes, cErr := g.changeListResolve(relToRootPath, fsPath, true)
		if cErr != nil {
			if cErr == ErrClashesDetected {
				clashes = append(clashes, cclashes...)
				continue
			} else {
				spin.stop()
				return cErr
			}
		}
		if len(ccl) > 0 {
			cl = append(cl, ccl...)
		}
	}

	if len(clashes) >= 1 {
		warnClashesPersist(g.log, clashes)
		return ErrClashesDetected
	}

	mount := g.opts.Mount
	if mount != nil {
		for _, mt := range mount.Points {
			ccl, _, cerr := lonePush(g, root, mt.Name, mt.MountPath)
			if cerr == nil {
				cl = append(cl, ccl...)
			}
		}
	}

	spin.stop()

	nonConflictsPtr, conflictsPtr := g.resolveConflicts(cl, true)
	if conflictsPtr != nil {
		warnConflictsPersist(g.log, *conflictsPtr)
		return fmt.Errorf("conflicts have prevented a push operation")
	}

	nonConflicts := *nonConflictsPtr

	pushSize, modSize := reduceToSize(cl, SelectDest|SelectSrc)

	// TODO: Handle compensation from deletions and modifications
	if false {
		pushSize -= modSize
	}

	// Warn about (near) quota exhaustion
	quotaStatus, quotaErr := g.QuotaStatus(pushSize)
	if quotaErr != nil {
		return quotaErr
	}

	unSafe := false
	switch quotaStatus {
	case AlmostExceeded:
		g.log.LogErrln("\033[92mAlmost exceeding your drive quota\033[00m")
	case Exceeded:
		g.log.LogErrln("\033[91mThis change will exceed your drive quota\033[00m")
		unSafe = true
	}
	if unSafe {
		g.log.LogErrf(" projected size: (%d) %s\n", pushSize, prettyBytes(pushSize))
		if !promptForChanges() {
			return
		}
	}

	clArg := changeListArg{
		logy:      g.log,
		changes:   nonConflicts,
		noPrompt:  !g.opts.canPrompt(),
		noClobber: g.opts.NoClobber,
	}

	ok, opMap := printChangeList(&clArg)
	if !ok {
		return
	}

	return g.playPushChanges(nonConflicts, opMap)
}

func (g *Commands) resolveConflicts(cl []*Change, push bool) (*[]*Change, *[]*Change) {
	if g.opts.IgnoreConflict {
		return &cl, nil
	}

	nonConflicts, conflicts := sift(cl)
	resolved, unresolved := resolveConflicts(conflicts, push, g.deserializeIndex)
	if conflictsPersist(unresolved) {
		return &resolved, &unresolved
	}

	for _, ch := range unresolved {
		resolved = append(resolved, ch)
	}

	for _, ch := range resolved {
		nonConflicts = append(nonConflicts, ch)
	}
	return &nonConflicts, nil
}

func (g *Commands) PushPiped() (err error) {
	// Cannot push asynchronously because the push order must be maintained
	for _, relToRootPath := range g.opts.Sources {
		remotes, resErr := g.rem.FindByPath(relToRootPath)
		if resErr != nil && resErr != ErrPathNotExists {
			return resErr
		}

		if len(remotes) < 1 {
			continue
		}

		// TODO: Decide if pushing to all those remotes should proceed
		rem := remotes[0]

		if rem != nil && !g.opts.Force {
			return fmt.Errorf("%s already exists remotely, use `%s` to override this behaviour.\n", relToRootPath, ForceKey)
		}

		if hasExportLinks(rem) {
			return fmt.Errorf("'%s' is a GoogleDoc/Sheet document cannot be pushed to raw.\n", relToRootPath)
		}

		base := filepath.Base(relToRootPath)
		local := fauxLocalFile(base)
		if rem == nil {
			rem = local
		}

		parentPath := g.parentPather(relToRootPath)
		parents, pErr := g.rem.FindByPath(parentPath)
		var parent *File

		if pErr == nil && len(parents) >= 1 {
			parent = parents[0]
		} else {
			spin := g.playabler()
			spin.play()
			parent, pErr = g.remoteMkdirAll(parentPath)
			spin.stop()
			if pErr != nil || parent == nil {
				g.log.LogErrf("%s: %v", relToRootPath, pErr)
				return
			}
		}

		args := upsertOpt{
			parentId:       parent.Id,
			fsAbsPath:      relToRootPath,
			src:            rem,
			dest:           rem,
			mask:           g.opts.TypeMask,
			nonStatable:    true,
			ignoreChecksum: g.opts.IgnoreChecksum,
		}

		rem, _, rErr := g.rem.upsertByComparison(os.Stdin, &args)
		if rErr != nil {
			g.log.LogErrf("%s: %v\n", relToRootPath, rErr)
			return rErr
		}

		if rem == nil {
			continue
		}

		index := rem.ToIndex()
		wErr := g.context.SerializeIndex(index)

		// TODO: Should indexing errors be reported?
		if wErr != nil {
			g.log.LogErrf("serializeIndex %s: %v\n", rem.Name, wErr)
		}
	}
	return
}

func (g *Commands) deserializeIndex(identifier string) *config.Index {
	index, err := g.context.DeserializeIndex(identifier)
	if err != nil {
		return nil
	}
	return index
}

func translateOpToChanger(g *Commands, c *Change) func(*Change) error {
	var fn func(*Change) error = nil

	op := c.Op()
	switch op {
	case OpMod:
		fn = g.remoteMod
	case OpModConflict:
		fn = g.remoteMod
	case OpAdd:
		fn = g.remoteAdd
	case OpDelete:
		fn = g.remoteTrash
	}
	return fn
}

func (g *Commands) playPushChanges(cl []*Change, opMap *map[Operation]sizeCounter) (err error) {
	if opMap == nil {
		result := opChangeCount(cl)
		opMap = &result
	}

	totalSize := int64(0)
	ops := *opMap
	for _, counter := range ops {
		totalSize += counter.src
	}

	g.taskStart(totalSize)

	defer close(g.rem.progressChan)

	go func() {
		for n := range g.rem.progressChan {
			g.taskAdd(int64(n))
		}
	}()

	type workPair struct {
		fn  func(*Change) error
		arg *Change
	}

	n := maxProcs()
	bench := make(chan *workPair, n)
	ackChan := make(chan bool, n)

	// Fill up ackChan initially with pass tokens
	for i := 0; i < n; i++ {
		ackChan <- true
	}

	throttle := time.Tick(time.Duration(1e9 / n))
	canPrintSteps := g.opts.Verbose && g.opts.canPrompt()

	sort.Sort(ByPrecedence(cl))

	doneCount := len(cl)
	done := make(chan bool, doneCount)

	go func() {
		defer close(bench)
		for i, c := range cl {
			if c == nil {
				done <- true
				g.log.LogErrf("BUGON:: push: nil change found for change index %d\n", i)
				continue
			}

			fn := translateOpToChanger(g, c)

			if fn == nil {
				g.log.LogErrf("push: cannot find operator for %v", c.Op())
				done <- true
				continue
			}

			bench <- &workPair{fn: fn, arg: c}
			<-ackChan
		}
	}()

	go func() {
		for wp := range bench {
			fn, c := wp.fn, wp.arg

			go func() {
				if canPrintSteps {
					g.log.Logln("\033[01mPush::Started", c.Path, "\033[00m")
				}

				if err := fn(c); err != nil {
					g.log.LogErrf("push: %s err: %v\n", c.Path, err)
				}

				if canPrintSteps {
					g.log.Logln("\033[04mPush::Done", c.Path, "\033[00m")
				}

				<-throttle

				done <- true
				ackChan <- true
			}()
		}
	}()

	for i := 0; i < doneCount; i++ {
		<-done
	}

	g.taskFinish()
	return err
}

func lonePush(g *Commands, parent, absPath, path string) (cl, clashes []*Change, err error) {
	remotes, err := g.rem.FindByPath(absPath)
	if err != nil && err != ErrPathNotExists {
		return
	}

	var r, l *File

	if len(remotes) >= 1 {
		r = remotes[0]
	}

	localinfo, _ := os.Stat(path)
	if localinfo != nil {
		l = NewLocalFile(path, localinfo)
	}

	clr := &changeListResolve{
		push:   true,
		dir:    parent,
		base:   absPath,
		remote: r,
		local:  l,
	}

	return g.resolveChangeListRecv(clr)
}

func (g *Commands) pathSplitter(absPath string) (dir, base string) {
	p := strings.Split(absPath, "/")
	pLen := len(p)
	base = p[pLen-1]
	p = append([]string{"/"}, p[:pLen-1]...)
	dir = gopath.Join(p...)
	return
}

func (g *Commands) parentPather(absPath string) string {
	dir, _ := g.pathSplitter(absPath)
	return dir
}

func (g *Commands) remoteMod(change *Change) (err error) {
	if change.Dest == nil && change.Src == nil {
		err = fmt.Errorf("bug on: both dest and src cannot be nil")
		g.log.LogErrln(err)
		return err
	}

	absPath := g.context.AbsPathOf(change.Path)

	if change.Src != nil && change.Src.IsDir {
		needsMkdirAll := change.Dest == nil || change.Src.Id == ""
		if needsMkdirAll {
			if destFile, _ := g.remoteMkdirAll(change.Path); destFile != nil {
				change.Src.Id = destFile.Id
			}
		}
	}

	if change.Dest != nil && change.Src != nil && change.Src.Id == "" {
		change.Src.Id = change.Dest.Id // TODO: bad hack
	}

	var parent *File
	parentPath := g.parentPather(change.Path)
	parent, err = g.remoteMkdirAll(parentPath)

	if err != nil {
		g.log.LogErrf("remoteMod/remoteMkdirAll: `%s` got %v\n", parentPath, err)
		return err
	}

	if parent == nil {
		err = errCannotMkdirAll(parentPath)
		g.log.LogErrln(err)
		return
	}

	args := upsertOpt{
		parentId:       parent.Id,
		fsAbsPath:      absPath,
		src:            change.Src,
		dest:           change.Dest,
		mask:           g.opts.TypeMask,
		ignoreChecksum: g.opts.IgnoreChecksum,
	}

	coercedMimeKey, ok := g.coercedMimeKey()
	if ok {
		args.mimeKey = coercedMimeKey
	} else if args.src != nil && !args.src.IsDir { // Infer it from the extension
		args.mimeKey = filepath.Ext(args.src.Name)
	}

	rem, err := g.rem.UpsertByComparison(&args)
	if err != nil {
		g.log.LogErrf("%s: %v\n", change.Path, err)
		return
	}
	if rem == nil {
		return
	}
	index := rem.ToIndex()
	wErr := g.context.SerializeIndex(index)

	// TODO: Should indexing errors be reported?
	if wErr != nil {
		g.log.LogErrf("serializeIndex %s: %v\n", rem.Name, wErr)
	}
	return
}

func (g *Commands) remoteAdd(change *Change) (err error) {
	return g.remoteMod(change)
}

func (g *Commands) remoteUntrash(change *Change) (err error) {
	target := change.Src
	defer func() {
		g.taskAdd(target.Size)
	}()

	err = g.rem.Untrash(target.Id)
	if err != nil {
		return
	}

	index := target.ToIndex()
	wErr := g.context.SerializeIndex(index)

	// TODO: Should indexing errors be reported?
	if wErr != nil {
		g.log.LogErrf("serializeIndex %s: %v\n", target.Name, wErr)
	}
	return
}

func remoteRemover(g *Commands, change *Change, fn func(string) error) (err error) {
	defer func() {
		g.taskAdd(change.Dest.Size)
	}()

	err = fn(change.Dest.Id)
	if err != nil {
		return
	}

	if change.Dest.IsDir {
		mkdirAllMu.Lock()
		g.mkdirAllCache.Remove(change.Path)
		mkdirAllMu.Unlock()
	}

	index := change.Dest.ToIndex()
	err = g.context.RemoveIndex(index, g.context.AbsPathOf(""))

	if err != nil {
		if change.Src != nil {
			g.log.LogErrf("%s \"%s\": remove indexfile %v\n", change.Path, change.Dest.Id, err)
		}
	}
	return
}

func (g *Commands) remoteTrash(change *Change) error {
	return remoteRemover(g, change, g.rem.Trash)
}

func (g *Commands) remoteDelete(change *Change) error {
	return remoteRemover(g, change, g.rem.Delete)
}

func (g *Commands) remoteMkdirAll(d string) (file *File, err error) {
	mkdirAllMu.Lock()

	cachedValue, ok := g.mkdirAllCache.Get(d)

	if ok && cachedValue != nil {
		castF, castOk := cachedValue.Value().(*File)
		// g.log.Logln("CacheHit", d, castF, castOk)
		if castOk && castF != nil {
			mkdirAllMu.Unlock()
			return castF, nil
		}
	}

	// Try the lookup one last time in case a coroutine raced us to it.
	retrFiles, retryErr := g.rem.FindByPath(d)

	var retrFile *File

	if len(retrFiles) >= 1 {
		retrFile = retrFiles[0]
	}

	if retryErr != nil && retryErr != ErrPathNotExists {
		mkdirAllMu.Unlock()
		return retrFile, retryErr
	}

	if retrFile != nil {
		mkdirAllMu.Unlock()
		return retrFile, nil
	}

	parDirPath, last := remotePathSplit(d)

	parents, parentErr := g.rem.FindByPath(parDirPath)

	var parent *File
	if len(parents) >= 1 {
		parent = parents[0]
	}

	if parentErr != nil && parentErr != ErrPathNotExists {
		mkdirAllMu.Unlock()
		return parent, parentErr
	}

	mkdirAllMu.Unlock()
	if parent == nil {
		parent, parentErr = g.remoteMkdirAll(parDirPath)
		if parentErr != nil || parent == nil {
			return parent, parentErr
		}
	}

	mkdirAllMu.Lock()
	defer mkdirAllMu.Unlock()

	g.mkdirAllCache.Put(parDirPath, newExpirableCacheValue(parent))

	remoteFile := &File{
		IsDir:   true,
		Name:    last,
		ModTime: time.Now(),
	}

	args := upsertOpt{
		parentId: parent.Id,
		src:      remoteFile,
	}

	cur, curErr := g.rem.UpsertByComparison(&args)
	if curErr != nil {
		return cur, curErr
	}

	if cur == nil {
		return cur, ErrPathNotExists
	}

	index := cur.ToIndex()
	wErr := g.context.SerializeIndex(index)

	// TODO: Should indexing errors be reported?
	if wErr != nil {
		g.log.LogErrf("serializeIndex %s: %v\n", cur.Name, wErr)
	}

	g.mkdirAllCache.Put(d, newExpirableCacheValue(cur))
	return cur, nil
}

func namedPipe(mode os.FileMode) bool {
	return (mode & os.ModeNamedPipe) != 0
}

func symlink(mode os.FileMode) bool {
	return (mode & os.ModeSymlink) != 0
}

func list(context *config.Context, p string, hidden bool, ignore *regexp.Regexp) (fileChan chan *File, err error) {
	absPath := context.AbsPathOf(p)
	var f []os.FileInfo
	f, err = ioutil.ReadDir(absPath)
	fileChan = make(chan *File)
	if err != nil {
		close(fileChan)
		return
	}

	go func() {
		for _, file := range f {
			fileName := file.Name()
			if fileName == config.GDDirSuffix {
				continue
			}
			if isHidden(fileName, hidden) {
				continue
			}

			resPath := gopath.Join(absPath, fileName)
			if anyMatch(ignore, fileName, resPath) {
				continue
			}

			// TODO: (@odeke-em) decide on how to deal with isFifo
			if namedPipe(file.Mode()) {
				fmt.Fprintf(os.Stderr, "%s (%s) is a named pipe, not reading from it\n", p, resPath)
				continue
			}

			if !symlink(file.Mode()) {
				fileChan <- NewLocalFile(resPath, file)
			} else {
				var symResolvPath string
				symResolvPath, err = filepath.EvalSymlinks(resPath)
				if err != nil {
					continue
				}

				if anyMatch(ignore, symResolvPath) {
					continue
				}

				var symInfo os.FileInfo
				symInfo, err = os.Stat(symResolvPath)
				if err != nil {
					continue
				}

				lf := NewLocalFile(symResolvPath, symInfo)
				// Retain the original name as appeared in
				// the manifest instead of the resolved one
				lf.Name = fileName
				fileChan <- lf
			}
		}
		close(fileChan)
	}()
	return
}
