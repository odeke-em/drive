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
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"github.com/odeke-em/statos"
)

const (
	maxNumOfConcPullTasks = 4
)

type urlMimeTypeExt struct {
	ext      string
	mimeType string
	url      string
}

type downloadArg struct {
	id              string
	path            string
	exportURL       string
	ackByteProgress bool
}

// Pull from remote if remote path exists and in a god context. If path is a
// directory, it recursively pulls from the remote if there are remote changes.
// It doesn't check if there are remote changes if isForce is set.
func (g *Commands) Pull(byId bool) (err error) {
	var cl []*Change

	g.log.Logln("Resolving...")

	spin := g.playabler()
	spin.play()

	resolver := g.pullByPath
	if byId {
		resolver = g.pullById
	}

	cl, err = resolver()
	spin.stop()

	nonConflictsPtr, conflictsPtr := g.resolveConflicts(cl, false)
	if conflictsPtr != nil {
		warnConflictsPersist(g.log, *conflictsPtr)
		return fmt.Errorf("conflicts have prevented a pull operation")
	}

	nonConflicts := *nonConflictsPtr

	ok, opMap := printChangeList(g.log, nonConflicts, !g.opts.canPrompt(), g.opts.NoClobber)
	if !ok {
		return
	}

	return g.playPullChanges(nonConflicts, g.opts.Exports, opMap)
}

func (g *Commands) PullMatches() (err error) {
	var cl []*Change
	matches, err := g.rem.FindMatches(g.opts.Path, g.opts.Sources, false)

	if err != nil {
		return err
	}

	p := g.opts.Path
	if p == "/" {
		p = ""
	}

	for match := range matches {
		if match == nil {
			continue
		}
		relToRoot := "/" + match.Name
		fsPath := g.context.AbsPathOf(relToRoot)

		ccl, cErr := g.byRemoteResolve(relToRoot, fsPath, match, false)
		if cErr != nil {
			return cErr
		}

		cl = append(cl, ccl...)
	}

	if len(cl) < 1 {
		return fmt.Errorf("no changes detected!")
	}

	nonConflictsPtr, conflictsPtr := g.resolveConflicts(cl, false)
	if conflictsPtr != nil {
		warnConflictsPersist(g.log, *conflictsPtr)
		return fmt.Errorf("conflicts have prevented a pull operation")
	}

	nonConflicts := *nonConflictsPtr

	ok, opMap := printChangeList(g.log, nonConflicts, !g.opts.canPrompt(), g.opts.NoClobber)
	if !ok {
		return nil
	}

	return g.playPullChanges(nonConflicts, g.opts.Exports, opMap)
}

func (g *Commands) PullPiped(byId bool) (err error) {
	resolver := g.rem.FindByPath
	if byId {
		resolver = g.rem.FindById
	}

	for _, relToRootPath := range g.opts.Sources {
		rem, err := resolver(relToRootPath)
		if err != nil {
			return fmt.Errorf("%s: %v", relToRootPath, err)
		}
		if rem == nil {
			continue
		}

		err = g.pullAndDownload(relToRootPath, os.Stdout, rem, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Commands) pullById() (cl []*Change, err error) {
	for _, srcId := range g.opts.Sources {
		rem, remErr := g.rem.FindById(srcId)
		if remErr != nil {
			return cl, fmt.Errorf("pullById: %s: %v", srcId, remErr)
		}

		if rem == nil {
			g.log.LogErrf("%s does not exist\n", srcId)
			continue
		}

		relToRootPath := filepath.Join(g.opts.Path, rem.Name)
		curAbsPath := g.context.AbsPathOf(relToRootPath)
		local, resErr := g.resolveToLocalFile(rem.Name, curAbsPath)
		if resErr != nil {
			return cl, resErr
		}

		ccl, clErr := g.doChangeListRecv(relToRootPath, curAbsPath, local, rem, false)
		if clErr != nil {
			return cl, clErr
		}
		cl = append(cl, ccl...)
	}

	return cl, nil
}

func (g *Commands) pullByPath() (cl []*Change, err error) {
	for _, relToRootPath := range g.opts.Sources {
		fsPath := g.context.AbsPathOf(relToRootPath)
		ccl, cErr := g.changeListResolve(relToRootPath, fsPath, false)
		if cErr != nil {
			return cl, cErr
		}
		if len(ccl) > 0 {
			cl = append(cl, ccl...)
		}
	}

	return cl, nil
}

func (g *Commands) pullAndDownload(relToRootPath string, fh io.Writer, rem *File, piped bool) (err error) {
	if hasExportLinks(rem) {
		return fmt.Errorf("'%s' is a GoogleDoc/Sheet document cannot be pulled from raw, only exported.\n", relToRootPath)
	}
	blobHandle, dlErr := g.rem.Download(rem.Id, "")
	if dlErr != nil {
		return dlErr
	}
	if blobHandle == nil {
		return nil
	}

	_, err = io.Copy(fh, blobHandle)
	blobHandle.Close()
	return
}

func (g *Commands) playPullChanges(cl []*Change, exports []string, opMap *map[Operation]sizeCounter) (err error) {
	var next []*Change

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

	// TODO: Only provide precedence ordering if all the other options are allowed
	// Currently noop on sorting by precedence
	if false && !g.opts.NoClobber {
		sort.Sort(ByPrecedence(cl))
	}

	go func() {
		for n := range g.rem.progressChan {
			g.taskAdd(int64(n))
		}
	}()

	for {
		if len(cl) > maxNumOfConcPullTasks {
			next, cl = cl[:maxNumOfConcPullTasks], cl[maxNumOfConcPullTasks:len(cl)]
		} else {
			next, cl = cl, []*Change{}
		}
		if len(next) == 0 {
			break
		}
		var wg sync.WaitGroup

		// play the changes
		// TODO: add timeouts

		// Defer adds until after the deletes
		for _, c := range next {
			switch c.Op() {
			case OpMod:
				wg.Add(1)
				go g.localMod(&wg, c, exports)
			case OpModConflict:
				wg.Add(1)
				go g.localMod(&wg, c, exports)
			case OpDelete:
				wg.Add(1)
				go g.localDelete(&wg, c)
			}
		}
		wg.Wait()

		for _, c := range next {
			switch c.Op() {
			case OpAdd:
				wg.Add(1)
				go g.localAdd(&wg, c, exports)
			}
		}
		wg.Wait()

	}

	g.taskFinish()
	return err
}

func (g *Commands) localMod(wg *sync.WaitGroup, change *Change, exports []string) (err error) {
	defer func() {
		if err == nil {
			src := change.Src
			index := src.ToIndex()
			wErr := g.context.SerializeIndex(index, g.context.AbsPathOf(""))

			// TODO: Should indexing errors be reported?
			if wErr != nil {
				g.log.LogErrf("serializeIndex %s: %v\n", src.Name, wErr)
			}
		}
		wg.Done()
	}()

	destAbsPath := g.context.AbsPathOf(change.Path)

	downloadPerformed := false

	// Simple heuristic to avoid downloading all the
	// content yet it could just be a modTime difference
	mask := fileDifferences(change.Src, change.Dest, change.IgnoreChecksum)
	if checksumDiffers(mask) {
		// download and replace
		if err = g.download(change, exports); err != nil {
			return
		}
		downloadPerformed = true
	}
	err = os.Chtimes(destAbsPath, change.Src.ModTime, change.Src.ModTime)

	// Update progress for the case in which you are only Chtime-ing
	// since progress for downloaded files is already handled separately
	if !downloadPerformed {
		chunks := chunkInt64(change.Src.Size)
		for n := range chunks {
			g.rem.progressChan <- n
		}
	}

	return
}

func (g *Commands) localAdd(wg *sync.WaitGroup, change *Change, exports []string) (err error) {
	defer func() {
		if err == nil && change.Src != nil {
			fileToSerialize := change.Src
			index := fileToSerialize.ToIndex()
			indexErr := g.context.SerializeIndex(index, g.context.AbsPathOf(""))

			// TODO: Should indexing errors be reported?
			if indexErr != nil {
				g.log.LogErrf("serializeIndex %s: %v\n", fileToSerialize.Name, indexErr)
			}
		}
		wg.Done()
	}()

	destAbsPath := g.context.AbsPathOf(change.Path)

	// make parent's dir if not exists
	destAbsDir := g.context.AbsPathOf(change.Parent)

	if destAbsDir != destAbsPath {
		err = os.MkdirAll(destAbsDir, os.ModeDir|0755)
		if err != nil {
			return err
		}
	}

	if change.Src.IsDir {
		return os.Mkdir(destAbsPath, os.ModeDir|0755)
	}

	// download and create
	if err = g.download(change, exports); err != nil {
		return
	}

	err = os.Chtimes(destAbsPath, change.Src.ModTime, change.Src.ModTime)
	return
}

func (g *Commands) localDelete(wg *sync.WaitGroup, change *Change) (err error) {
	defer func() {
		if err == nil {
			chunks := chunkInt64(change.Dest.Size)
			for n := range chunks {
				g.rem.progressChan <- n
			}
		}
		wg.Done()
	}()
	err = os.RemoveAll(change.Dest.BlobAt)
	return
}

func touchFile(path string) (err error) {
	var ef *os.File
	defer func() {
		if err == nil && ef != nil {
			ef.Close()
		}
	}()
	ef, err = os.Create(path)
	return
}

func (g *Commands) export(f *File, destAbsPath string, exports []string) (manifest []string, err error) {
	if len(exports) < 1 || f == nil {
		return
	}

	dirPath := sepJoin("_", destAbsPath, "exports")
	if err = os.MkdirAll(dirPath, os.ModeDir|0755); err != nil {
		return
	}

	var ok bool
	var mimeType, exportURL string

	waitables := []*urlMimeTypeExt{}

	for _, ext := range exports {
		mimeType = mimeTypeFromExt(ext)
		exportURL, ok = f.ExportLinks[mimeType]
		if !ok {
			continue
		}

		waitables = append(waitables, &urlMimeTypeExt{
			mimeType: mimeType,
			url:      exportURL,
			ext:      ext,
		})
	}

	var wg sync.WaitGroup
	wg.Add(len(waitables))

	basePath := filepath.Base(f.Name)
	baseDir := path.Join(dirPath, basePath)

	for _, exportee := range waitables {
		go func(wg *sync.WaitGroup, baseDirPath, id string, urlMExt *urlMimeTypeExt) error {
			defer func() {
				wg.Done()
			}()

			exportPath := sepJoin(".", baseDirPath, urlMExt.ext)

			// TODO: Decide if users should get to make *.desktop users even for exports
			if runtime.GOOS == OSLinuxKey && false {
				desktopEntryPath := sepJoin(".", exportPath, DesktopExtension)

				_, dentErr := f.serializeAsDesktopEntry(desktopEntryPath, urlMExt)
				if dentErr != nil {
					g.log.LogErrf("desktopEntry: %s %v\n", desktopEntryPath, dentErr)
				}
			}

			dlArg := downloadArg{
				ackByteProgress: false,
				path:            exportPath,
				id:              id,
				exportURL:       urlMExt.url,
			}

			err := g.singleDownload(&dlArg)
			if err == nil {
				manifest = append(manifest, exportPath)
			}
			return err
		}(&wg, baseDir, f.Id, exportee)
	}
	wg.Wait()
	return
}

func isLocalFile(f *File) bool {
	// TODO: Better check
	return f != nil && f.Etag == ""
}

func (g *Commands) download(change *Change, exports []string) (err error) {
	if change.Src == nil {
		return fmt.Errorf("tried to download nil change.Src")
	}

	destAbsPath := g.context.AbsPathOf(change.Path)
	if change.Src.BlobAt != "" {
		dlArg := downloadArg{
			path:            destAbsPath,
			id:              change.Src.Id,
			ackByteProgress: true,
		}

		return g.singleDownload(&dlArg)
	}

	// We need to touch the empty file to
	// ensure consistency during a push.
	if runtime.GOOS != OSLinuxKey {
		err = touchFile(destAbsPath)
		if err != nil {
			return err
		}
	} else {
		// For those our Linux kin that need .desktop files
		dirPath := g.opts.ExportsDir
		if dirPath == "" {
			dirPath = filepath.Dir(destAbsPath)
		}

		f := change.Src

		urlMExt := urlMimeTypeExt{
			url:      f.AlternateLink,
			ext:      "",
			mimeType: f.MimeType,
		}

		dirPath = filepath.Join(dirPath, f.Name)
		desktopEntryPath := sepJoin(".", dirPath, DesktopExtension)

		_, dentErr := f.serializeAsDesktopEntry(desktopEntryPath, &urlMExt)
		if dentErr != nil {
			g.log.LogErrf("desktopEntry: %s %v\n", desktopEntryPath, dentErr)
		}
	}

	if len(exports) >= 1 && hasExportLinks(change.Src) {
		exportDirPath := destAbsPath
		if g.opts.ExportsDir != "" {
			exportDirPath = path.Join(g.opts.ExportsDir, change.Src.Name)
		}

		manifest, exportErr := g.export(change.Src, exportDirPath, exports)
		if exportErr == nil {
			for _, exportPath := range manifest {
				g.log.Logf("Exported '%s' to '%s'\n", destAbsPath, exportPath)
			}
		}
		return exportErr
	}
	return
}

func (g *Commands) singleDownload(dlArg *downloadArg) (err error) {
	var fo *os.File
	fo, err = os.Create(dlArg.path)
	if err != nil {
		g.log.LogErrf("create: %s %v\n", dlArg.path, err)
		return
	}

	// close fo on exit and check for its returned error
	defer func() {
		fErr := fo.Close()
		if fErr != nil {
			g.log.LogErrf("fErr", fErr)
			err = fErr
		}
	}()

	var blob io.ReadCloser
	defer func() {
		if blob != nil {
			blob.Close()
		}
	}()

	blob, err = g.rem.Download(dlArg.id, dlArg.exportURL)
	if err != nil {
		return err
	}

	ws := statos.NewWriter(fo)

	go func() {
		commChan := ws.ProgressChan()
		if dlArg.ackByteProgress {
			for n := range commChan {
				g.rem.progressChan <- n
			}
		} else { // Just drain the progress channel
			for _ = range commChan {
				g.rem.progressChan <- 0
			}
		}
	}()

	_, err = io.Copy(ws, blob)

	return
}
