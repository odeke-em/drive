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

	destFiles, err := g.rem.FindByPath(dest)
	if err != nil && err != ErrPathNotExists {
		return fmt.Errorf("destination: %s err: %v", dest, err)
	}

	for _, destFile := range destFiles {
		if destFile == nil {
			continue
		}

		errs := copy_(g, dest, destFile, sources, byId)
		for _, err := range errs {
			g.log.LogErrf("copy_: %s %v\n", destFile.Id, err)
		}
	}

	return nil
}

func copy_(g *Commands, dest string, destFile *File, sources []string, byId bool) (errs []error) {
	multiPaths := len(sources) > 1
	if multiPaths {
		if destFile != nil && !destFile.IsDir {
			errs = append(errs, fmt.Errorf("%s: %v", dest, ErrPathNotDir))
			return
		}
		_, err := g.remoteMkdirAll(dest)
		if err != nil {
			errs = append(errs, err)
			return
		}
	}

	srcResolver := g.rem.FindByPath
	if byId {
		srcResolver = g.rem.FindByIdMulti
	}

	done := make(chan bool)
	waitCount := uint64(0)

	for _, srcPath := range sources {
		srcFiles, srcErr := srcResolver(srcPath)
		if srcErr != nil {
			g.log.LogErrf("%s: %v\n", srcPath, srcErr)
			continue
		}

		for _, srcFile := range srcFiles {
			waitCount += 1

			go func(fromPath, toPath string, fromFile *File) {
				_, copyErr := g.copy(fromFile, toPath)
				if copyErr != nil {
					g.log.LogErrf("%s: %v\n", fromPath, copyErr)
				}
				done <- true
			}(srcPath, dest, srcFile)
		}
	}

	for i := uint64(0); i < waitCount; i += 1 {
		<-done
	}

	return nil

}

func (g *Commands) copy(src *File, destPath string) (copies []*File, errs []error) {
	if src == nil {
		errs = append(errs, fmt.Errorf("non existant src"))
		return
	}

	if !src.IsDir {
		if !src.Copyable {
			errs = append(errs, fmt.Errorf("%s is non-copyable", src.Name))
			return
		}

		destDir, destBase := g.pathSplitter(destPath)
		destParent, destParErr := g.remoteMkdirAll(destDir)

		if destParErr != nil {
			errs = append(errs, destParErr)
			return
		}

		parentId := destParent.Id
		destFiles, destErr := g.rem.FindByPath(destPath)
		if destErr != nil && destErr != ErrPathNotExists {
			errs = append(errs, destErr)
			return
		}

		for _, destFile := range destFiles {
			if destFile != nil && destFile.IsDir {
				parentId = destFile.Id
				destBase = src.Name
			}

			copy, cpErr := g.rem.copy(destBase, parentId, src)

			if copy != nil {
				copies = append(copies, copy)
			}

			if cpErr != nil {
				errs = append(errs, cpErr)
			}
		}

		return
	}

	destFile, destErr := g.remoteMkdirAll(destPath)
	if destErr != nil {
		errs = append(errs, destErr)
		return
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

	copies = append(copies, destFile)
	return
}
