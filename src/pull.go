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

	"github.com/odeke-em/drive/config"
	"github.com/odeke-em/statos"
)

var (
	maxConcPulls = DefaultMaxProcs
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
func (g *Commands) Pull(byId bool) error {
	cl, clashes, err := pullLikeResolve(g, byId)

	if len(clashes) >= 1 {
		warnClashesPersist(g.log, clashes)
		return ErrClashesDetected
	}

	if err != nil {
		return err
	}

	nonConflictsPtr, conflictsPtr := g.resolveConflicts(cl, false)
	if conflictsPtr != nil {
		warnConflictsPersist(g.log, *conflictsPtr)
		return fmt.Errorf("conflicts have prevented a pull operation")
	}

	nonConflicts := *nonConflictsPtr

	clArg := changeListArg{
		logy:      g.log,
		changes:   nonConflicts,
		noPrompt:  !g.opts.canPrompt(),
		noClobber: g.opts.NoClobber,
	}

	ok, opMap := printChangeList(&clArg)
	if !ok {
		return nil
	}

	return g.playPullChanges(nonConflicts, g.opts.Exports, opMap)
}

func pullLikeResolve(g *Commands, byId bool) (cl, clashes []*Change, err error) {
	g.log.Logln("Resolving...")

	spin := g.playabler()
	spin.play()
	defer spin.stop()

	resolver := g.pullByPath
	if byId {
		resolver = g.pullById
	}

	return resolver()
}

func pullLikeMatchesResolver(g *Commands) (cl, clashes []*Change, err error) {
	mq := matchQuery{
		dirPath: g.opts.Path,
		inTrash: false,
		titleSearches: []fuzzyStringsValuePair{
			{fuzzyLevel: Like, values: g.opts.Sources},
		},
	}
	matches, err := g.rem.FindMatches(&mq) // g.opts.Path, g.opts.Sources, false)

	if err != nil {
		return
	}

	p := g.opts.Path
	if p == "/" {
		p = ""
	}

	combiner := func(from []*Change, to *[]*Change, done chan bool) {
		*to = append(*to, from...)
		done <- true
	}

	for match := range matches {
		if match == nil {
			continue
		}
		relToRoot := "/" + match.Name
		fsPath := g.context.AbsPathOf(relToRoot)

		ccl, cclashes, cErr := g.byRemoteResolve(relToRoot, fsPath, match, false)
		if cErr != nil {
			err = cErr
			return
		}

		done := make(chan bool, 2)
		go combiner(ccl, &cl, done)
		go combiner(cclashes, &clashes, done)

		<-done
	}

	return
}

func (g *Commands) PullMatches() (err error) {
	cl, clashes, err := pullLikeMatchesResolver(g)

	if len(clashes) >= 1 {
		warnClashesPersist(g.log, clashes)
		return ErrClashesDetected
	}

	if err != nil {
		return err
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

	clArg := changeListArg{
		logy:      g.log,
		changes:   nonConflicts,
		noPrompt:  !g.opts.canPrompt(),
		noClobber: g.opts.NoClobber,
	}

	ok, opMap := printChangeList(&clArg)
	if !ok {
		return nil
	}

	return g.playPullChanges(nonConflicts, g.opts.Exports, opMap)
}

func (g *Commands) PullPiped(byId bool) (err error) {
	resolver := g.rem.FindByPath
	if byId {
		resolver = g.rem.FindByIdMulti
	}

	for _, relToRootPath := range g.opts.Sources {
		remotes, err := resolver(relToRootPath)

		if err != nil {
			return fmt.Errorf("%s: %v", relToRootPath, err)
		}

		for _, rem := range remotes {
			if rem == nil {
				continue
			}

			err = g.pullAndDownload(relToRootPath, os.Stdout, rem, true)
			if err != nil {
				g.log.LogErrf("pullPiped:: %s %s: %v\n", relToRootPath, rem.Id, err)
			}
		}
	}

	return nil
}

func (g *Commands) pullById() (cl, clashes []*Change, err error) {
	for _, srcId := range g.opts.Sources {
		rem, remErr := g.rem.FindById(srcId)
		if remErr != nil {
			return cl, clashes, fmt.Errorf("pullById: %s: %v", srcId, remErr)
		}

		if rem == nil {
			g.log.LogErrf("%s does not exist\n", srcId)
			continue
		}

		relToRootPath := filepath.Join(g.opts.Path, rem.Name)
		curAbsPath := g.context.AbsPathOf(relToRootPath)
		local, resErr := g.resolveToLocalFile(rem.Name, curAbsPath)
		if resErr != nil {
			return cl, clashes, resErr
		}

		ccl, cclashes, clErr := g.doChangeListRecv(relToRootPath, curAbsPath, local, rem, false)
		if clErr != nil {
			if clErr != ErrClashesDetected {
				return cl, clashes, clErr
			} else {
				clashes = append(clashes, cclashes...)
			}
		}
		cl = append(cl, ccl...)
	}

	if len(clashes) >= 1 {
		err = ErrClashesDetected
	}

	return cl, clashes, err
}

func (g *Commands) pullByPath() (cl, clashes []*Change, err error) {
	for _, relToRootPath := range g.opts.Sources {
		fsPath := g.context.AbsPathOf(relToRootPath)
		ccl, cclashes, cErr := g.changeListResolve(relToRootPath, fsPath, false)
		if cErr != nil {
			if cErr != ErrClashesDetected {
				return cl, clashes, cErr
			} else {
				clashes = append(clashes, cclashes...)
			}
		}
		if len(ccl) > 0 {
			cl = append(cl, ccl...)
		}
	}

	if len(clashes) >= 1 {
		err = ErrClashesDetected
	}

	return cl, clashes, err
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

	sort.Sort(ByPrecedence(cl))

	go func() {
		for n := range g.rem.progressChan {
			g.taskAdd(int64(n))
		}
	}()

	nMax := len(cl)
	doneAck := make(chan bool)

	maxConcPulls := maxProcs()

	loader := make(chan *Change, maxConcPulls)
	waiter := make(chan bool, maxConcPulls)

	for i := 0; i < maxConcPulls; i++ {
		waiter <- true
	}

	go func() {
		defer close(loader)

		for _, c := range cl {
			if c == nil {
				doneAck <- true
				continue
			}

			<-waiter
			loader <- c
		}
	}()

	canPrintSteps := g.opts.Verbose && g.opts.canPrompt()

	go func() {
		for ch := range loader {
			var fn func(*Change, []string) error = nil

			op := ch.Op()
			switch op {
			case OpMod:
				fn = g.localMod
			case OpModConflict:
				fn = g.localMod
			case OpAdd:
				fn = g.localAdd
			case OpDelete:
				fn = g.localDelete
			case OpIndexAddition:
				fn = g.localAddIndex
			}

			if fn == nil {
				g.log.LogErrf("pull: cannot find operator for %v", op)
				doneAck <- true
				waiter <- true
				continue
			}

			go func(c *Change, f func(*Change, []string) error) {
				if canPrintSteps {
					g.log.Logln("\033[01mPull::Started", c.Path, "\033[00m")
				}

				if err := f(c, exports); err != nil {
					g.log.LogErrf("pull: %s err: %v\n", c.Path, err)
				}

				if canPrintSteps {
					g.log.Logln("\033[04mPull::Done", c.Path, "\033[00m")
				}

				doneAck <- true
				waiter <- true
			}(ch, fn)
		}
	}()

	for i := 0; i < nMax; i++ {
		<-doneAck
	}

	g.taskFinish()
	return err
}

func (g *Commands) localAddIndex(change *Change, conform []string) (err error) {
	f := change.Src
	defer func() {
		if f != nil {
			chunks := chunkInt64(change.Src.Size)
			for n := range chunks {
				g.rem.progressChan <- n
			}
		}
	}()

	return g.createIndex(f)
}

func (g *Commands) localMod(change *Change, exports []string) (err error) {
	defer func() {
		if err == nil {
			src := change.Src
			indexErr := g.createIndex(src)
			// TODO: Should indexing errors be reported?
			if indexErr != nil {
				g.log.LogErrf("localMod:createIndex %s: %v\n", src.Name, indexErr)
			}
		}
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

func (g *Commands) localAdd(change *Change, exports []string) (err error) {
	defer func() {
		if err == nil && change.Src != nil {
			fileToSerialize := change.Src

			indexErr := g.createIndex(fileToSerialize)
			// TODO: Should indexing errors be reported?
			if indexErr != nil {
				g.log.LogErrf("localAdd:createIndex %s: %v\n", fileToSerialize.Name, indexErr)
			}
		}
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
		err = os.Mkdir(destAbsPath, os.ModeDir|0755)
		if os.IsExist(err) {
			err = nil
		}
		return err
	}

	// download and create
	if err = g.download(change, exports); err != nil {
		return
	}

	err = os.Chtimes(destAbsPath, change.Src.ModTime, change.Src.ModTime)
	return
}

func (g *Commands) localDelete(change *Change, conform []string) (err error) {
	defer func() {
		if err == nil {
			chunks := chunkInt64(change.Dest.Size)
			for n := range chunks {
				g.rem.progressChan <- n
			}

			dest := change.Dest
			index := dest.ToIndex()
			rmErr := g.context.RemoveIndex(index, g.context.AbsPathOf(""))
			// For the sake of files missing remotely yet present locally and might not have a FileId
			if rmErr != nil && rmErr != config.ErrEmptyFileIdForIndex {
				g.log.LogErrf("localDelete removing index for: \"%s\" at \"%s\" %v\n", dest.Name, dest.BlobAt, rmErr)
			}
		}
	}()

	err = os.RemoveAll(change.Dest.BlobAt)
	if err != nil {
		g.log.LogErrf("localDelete: \"%s\" %v\n", change.Dest.BlobAt, err)
	}

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

	n := len(waitables)
	doneAck := make(chan bool, n)

	basePath := filepath.Base(f.Name)
	baseDir := path.Join(dirPath, basePath)

	for _, exportee := range waitables {
		go func(baseDirPath, id string, urlMExt *urlMimeTypeExt) error {
			defer func() {
				doneAck <- true
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
		}(baseDir, f.Id, exportee)
	}

	for i := 0; i < n; i++ {
		<-doneAck
	}

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
	if err = touchFile(destAbsPath); err != nil {
		return err
	}

	// For our Linux kin that need .desktop files
	if runtime.GOOS == OSLinuxKey {
		f := change.Src

		urlMExt := urlMimeTypeExt{
			url:      f.AlternateLink,
			ext:      "",
			mimeType: f.MimeType,
		}

		desktopEntryPath := sepJoin(".", destAbsPath, DesktopExtension)

		_, dentErr := f.serializeAsDesktopEntry(desktopEntryPath, &urlMExt)
		if dentErr != nil {
			g.log.LogErrf("desktopEntry: %s %v\n", desktopEntryPath, dentErr)
		}
	}

	canExport := len(exports) >= 1 && hasExportLinks(change.Src)
	if !canExport {
		return nil
	}

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
