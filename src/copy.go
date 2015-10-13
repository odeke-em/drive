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
	"errors"
	"fmt"
)

var ErrPathNotDir = errors.New("not a directory")

type copyArgs struct {
	destPath string
	src      *File
	dest     *File
}

func (g *Commands) Copy(byId bool) error {
	argc := len(g.opts.Sources)
	if argc < 2 {
		return fmt.Errorf("expecting src [src1....] dest got: %v", g.opts.Sources)
	}

	g.log.Logln("Processing...")

	spin := g.playabler()
	spin.play()
	defer spin.stop()

	end := argc - 1
	sources, dest := g.opts.Sources[:end], g.opts.Sources[end]

	destFile, err := g.rem.FindByPath(dest)
	if err != nil && err != ErrPathNotExists {
		return fmt.Errorf("destination: %s err: %v", dest, err)
	}

	multiPaths := len(sources) > 1
	if multiPaths {
		if destFile != nil && !destFile.IsDir {
			return fmt.Errorf("%s: %v", dest, ErrPathNotDir)
		}
		_, err := g.remoteMkdirAll(dest)
		if err != nil {
			return err
		}
	}

	srcResolver := g.rem.FindByPath
	if byId {
		srcResolver = g.rem.FindById
	}

	done := make(chan bool)
	waitCount := uint64(0)

	for _, srcPath := range sources {
		srcFile, srcErr := srcResolver(srcPath)
		if srcErr != nil {
			g.log.LogErrf("%s: %v\n", srcPath, srcErr)
			continue
		}

		waitCount += 1

		go func(fromPath, toPath string, fromFile *File) {
			_, copyErr := g.copy(fromFile, toPath)
			if copyErr != nil {
				g.log.LogErrf("%s: %v\n", fromPath, copyErr)
			}
			done <- true
		}(srcPath, dest, srcFile)
	}

	for i := uint64(0); i < waitCount; i += 1 {
		<-done
	}

	return nil
}

func (g *Commands) copy(src *File, destPath string) (*File, error) {
	if src == nil {
		return nil, fmt.Errorf("non existent src")
	}

	if !src.IsDir {
		if !src.Copyable {
			return nil, fmt.Errorf("%s is non-copyable", src.Name)
		}

		destDir, destBase := g.pathSplitter(destPath)
		destParent, destParErr := g.remoteMkdirAll(destDir)

		if destParErr != nil {
			return nil, destParErr
		}

		parentId := destParent.Id
		destFile, destErr := g.rem.FindByPath(destPath)
		if destErr != nil && destErr != ErrPathNotExists {
			return nil, destErr
		}
		if destFile != nil && destFile.IsDir {
			parentId = destFile.Id
			destBase = src.Name
		}
		return g.rem.copy(destBase, parentId, src)
	}

	destFile, destErr := g.remoteMkdirAll(destPath)
	if destErr != nil {
		return nil, destErr
	}

	children := g.rem.findChildren(src.Id, false)

	for child := range children {
		// TODO: add concurrency after retry scheme is added
		// because could suffer from rate limit restrictions
		chName := sepJoin("/", destPath, child.Name)
		_, chErr := g.copy(child, chName)

		if chErr != nil {
			g.log.LogErrf("copy: %s: %v\n", chName, chErr)
		}
	}

	return destFile, nil
}
