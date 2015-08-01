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
	"sync"
	"time"

	"github.com/odeke-em/drive/config"
)

const (
	FetchById = iota
	FetchMatches
	Fetch
)

func setIndexingOnlyOption(g *Commands) {
	g.opts.indexingOnly = true
}

func (g *Commands) Prune() (err error) {
	setIndexingOnlyOption(g)

	deletions, delErr := g.pruneStaleIndices()
	if delErr != nil {
		return delErr
	}

	for staleFileId := range deletions {
		delErr := g.context.PopIndicesKey(staleFileId)

		if delErr != nil {
			g.log.LogErrf("fetch removing index for: \"%s\" %v\n", staleFileId, delErr)
		}
	}

	return nil
}

func (g *Commands) Fetch() (err error) {
	return g.fetch(Fetch)
}

func (g *Commands) FetchById() (err error) {
	return g.fetch(FetchById)
}

func (g *Commands) FetchMatches() (err error) {
	return g.fetch(FetchMatches)
}

func (g *Commands) fetch(fetchOp int) (err error) {
	setIndexingOnlyOption(g)

	err = g.context.CreateIndicesBucket()
	if err != nil {
		return err
	}

	var cl []*Change
	switch fetchOp {
	case FetchById:
		cl, _, err = pullLikeResolve(g, true)
	case FetchMatches:
		cl, _, err = pullLikeMatchesResolver(g)
	default:
		cl, _, err = pullLikeResolve(g, false)
	}

	if err != nil {
		return err
	}

	clArg := changeListArg{
		logy:      g.log,
		changes:   cl,
		noPrompt:  !g.opts.canPrompt(),
		noClobber: g.opts.NoClobber,
	}

	ok, opMap := printFetchChangeList(&clArg)
	if !ok {
		return nil
	}

	return g.playFetchChanges(cl, opMap)
}

func loneCountRegister(wg *sync.WaitGroup, progress chan int) {
	progress <- 1
	wg.Done()
	return
}

func (g *Commands) playFetchChanges(cl []*Change, opMap *map[Operation]sizeCounter) (err error) {
	changeCount := len(cl)

	g.taskStart(int64(changeCount))

	progressDone := make(chan bool, 1)

	var wg sync.WaitGroup
	wg.Add(changeCount)

	go func() {
		for n := range g.rem.progressChan {
			g.taskAdd(int64(n))
		}
		progressDone <- true
	}()

	ticker := time.Tick(1e9 / 10)

	for _, c := range cl {
		switch c.Op() {
		case OpDelete:
			go g.removeIndex(&wg, c.Dest)
		case OpNone:
			loneCountRegister(&wg, g.rem.progressChan)
		default:
			go g.addIndex(&wg, c.Src)
		}

		<-ticker
	}

	wg.Wait()
	close(g.rem.progressChan)

	<-progressDone
	g.taskFinish()

	return nil
}

func (g *Commands) addIndex(wg *sync.WaitGroup, f *File) (err error) {
	defer loneCountRegister(wg, g.rem.progressChan)

	indexErr := g.createIndex(f)
	// TODO: Should indexing errors be reported?
	if indexErr != nil {
		g.log.LogErrf("addIndex %s: %v\n", f.Name, indexErr)
	}

	return indexErr
}

func (g *Commands) removeIndex(wg *sync.WaitGroup, f *File) (err error) {
	err = g.context.CreateIndicesBucket()
	if err != nil {
		return err
	}

	defer loneCountRegister(wg, g.rem.progressChan)
	if f.Id == "" {
		return
	}

	index := f.ToIndex()
	rmErr := g.context.RemoveIndex(index, g.context.AbsPathOf(""))

	// TODO: Should indexing errors be reported?
	if rmErr != nil {
		g.log.LogErrf("removeIndex %s: %v\n", f.Name, rmErr)
	}

	return rmErr
}

func queryIdify(ids map[string]*File) string {
	idified := []string{}

	for id, _ := range ids {
		idified = append(idified, fmt.Sprintf("(id=%s)", strconv.Quote(id)))
	}

	return sepJoin(" or ", idified...)
}

func mapifyFiles(g *Commands, ids map[string]bool) (map[string]*File, error) {
	delivery := make(chan *File)
	mapping := make(map[string]*File)

	for id, _ := range ids {
		go func(qId string) {
			f, err := g.rem.FindById(qId)
			if false && err != nil && err != ErrPathNotExists {
				g.log.LogErrln(id, err, "x")
			}
			delivery <- f
		}(id)
	}

	doneCount := len(ids)
	for i := 0; i < doneCount; i += 1 {
		f := <-delivery
		if f == nil {
			continue
		}

		mapping[f.Id] = f
	}

	return mapping, nil
}

func (g *Commands) listIndicesKeys() (chan string, error) {
	return g.context.ListKeys(g.context.AbsPathOf(""), config.IndicesKey)
}

func (g *Commands) pruneStaleIndices() (deletions chan string, err error) {
	var listing chan string
	listing, err = g.listIndicesKeys()

	deletions = make(chan string)

	if err != nil {
		close(deletions)
		return
	}

	go func() {

		spin := g.playabler()

		defer func() {
			spin.stop()
			close(deletions)
		}()

		queriesPerRequest := 10
		tick := time.Tick(time.Duration(1e9 / queriesPerRequest))

		iterating := true

		doneCount := uint64(0)
		totalDeletions := uint64(0)

		for iterating {
			spin.play()

			localIds := map[string]bool{}
			i := 0
			for i < queriesPerRequest {
				i += 1
				fileId, ok := <-listing
				if !ok {
					iterating = false
					break
				}

				localIds[fileId] = true
			}

			if len(localIds) < 1 {
				break
			}

			mapping, mErr := mapifyFiles(g, localIds)
			if mErr != nil {
				g.log.LogErrf("syncIndices: %v\n", mErr)
				continue
			}

			delCount := 0
			for localId, _ := range localIds {
				_, ok := mapping[localId]
				if !ok {
					delCount += 1
					deletions <- localId
				}
			}

			doneCount += uint64(i)
			totalDeletions += uint64(delCount)

			spin.pause()
			deletionsReport := ""
			if delCount >= 1 {
				deletionsReport = fmt.Sprintf("%v/%v to be deleted", delCount, i)
			}

			if totalDeletions >= 1 {
				deletionsReport = fmt.Sprintf("%s(%v deletions so far)", deletionsReport, totalDeletions)
			}

			g.log.LogErrf("\rprune: %s %v index items processed so far\r", deletionsReport, doneCount)
			<-tick
		}
	}()

	return
}

func (g *Commands) createIndex(f *File) (err error) {
	if f == nil {
		return config.ErrDerefNilIndex
	}
	index := f.ToIndex()
	return g.context.SerializeIndex(index)
}

func printFetchChangeList(clArg *changeListArg) (bool, *map[Operation]sizeCounter) {
	if len(clArg.changes) == 0 {
		clArg.logy.Logln("Everything is up-to-date.")
		return false, nil
	}
	if clArg.noPrompt {
		return true, nil
	}

	logy := clArg.logy
	changes := clArg.changes
	op := OpIndexAddition
	_, description := op.description()

	for _, c := range changes {
		op := c.Op()
		if op != OpNone {
			logy.Logln(c.Symbol(), c.Path)
		}
	}

	logy.Logf("%s %d\n", description, len(changes))

	return promptForChanges(), nil
}
