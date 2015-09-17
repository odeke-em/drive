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
	"path/filepath"
)

type moveOpt struct {
	src      string
	dest     string
	byId     bool
	srcFile  *File
	destFile *File
}

func (g *Commands) Move(byId bool) (err error) {
	argc := len(g.opts.Sources)
	if argc < 2 {
		return fmt.Errorf("move: expected <src> [src...] <dest>, instead got: %v", g.opts.Sources)
	}

	rest, dest := g.opts.Sources[:argc-1], g.opts.Sources[argc-1]

	for _, src := range rest {
		prefix := commonPrefix(src, dest)

		// Trying to nest a parent into its child
		if prefix == src {
			return fmt.Errorf("%s cannot be nested into %s", src, dest)
		}

		opt := moveOpt{
			src:  src,
			dest: dest,
			byId: byId,
		}

		errs := g.move(&opt)
		for _, err := range errs {
			if err != nil {
				// TODO: Actually throw the error? Impact on UX if thrown?
				fmt.Printf("move: %s: %v\n", src, err)
			}
		}
	}

	return nil
}

func (g *Commands) move(opt *moveOpt) (errs []error) {

	srcResolver := g.rem.FindByPath
	if opt.byId {
		srcResolver = g.rem.FindByIdMulti
	}

	sources, err := srcResolver(opt.src)
	if err != nil {
		errs = append(errs, fmt.Errorf("src('%s') %v", opt.src, err))
		return
	}

	newParents, parErr := g.rem.FindByPath(opt.dest)
	if parErr != nil {
		errs = append(errs, fmt.Errorf("dest: '%s' %v", opt.dest, parErr))
		return
	}

	for _, src := range sources {

		for _, newParent := range newParents {
			if newParent == nil || !newParent.IsDir {
				errs = append(errs, fmt.Errorf("dest: '%s' must be an existant folder", opt.dest))
				continue
			}

			opt.srcFile = src
			opt.destFile = newParent
			if err := move_(g, opt); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return
}

func move_(g *Commands, opt *moveOpt) error {
	var err error

	remSrc := opt.srcFile
	newParent := opt.destFile

	if remSrc == nil {
		return fmt.Errorf("src: '%s' could not be found", opt.src)
	}

	if !opt.byId {
		parentPath := g.parentPather(opt.src)
		oldParents, parErr := g.rem.FindByPath(parentPath)
		if parErr != nil && parErr != ErrPathNotExists {
			return parErr
		}

		for _, oldParent := range oldParents {
			// TODO: If oldParent is not found, retry since it may have been moved temporarily at least
			if oldParent != nil && oldParent.Id == newParent.Id {
				return fmt.Errorf("src and dest are the same srcParentId %s destParentId %s",
					customQuote(oldParent.Id), customQuote(newParent.Id))
			}
		}
	}

	newFullPath := filepath.Join(opt.dest, remSrc.Name)

	// Check for a duplicate
	var dupChecks []*File
	dupChecks, err = g.rem.FindByPath(newFullPath)
	if err != nil && err != ErrPathNotExists {
		return err
	}

	for _, dup := range dupChecks {
		if dup == nil {
			continue
		}

		if dup.Id == remSrc.Id { // Trying to move to self
			return fmt.Errorf("move: trying to move fileId:%s to self fileId:%s", customQuote(dup.Id), customQuote(remSrc.Id))
		}

		if !g.opts.Force {
			return fmt.Errorf("%s already exists. Use `%s` flag to override this behaviour", newFullPath, ForceKey)
		}
	}

	// Avoid self-nesting
	if remSrc.Id == newParent.Id {
		return fmt.Errorf("move: cannot move '%s' to itself", opt.src)
	}

	if err = g.rem.insertParent(remSrc.Id, newParent.Id); err != nil {
		return err
	}

	if opt.byId { // TODO: Also take out this current parent
		return nil
	}
	return g.removeParent(remSrc.Id, opt.src)
}

func (g *Commands) removeParent(fileId, relToRootPath string) error {
	parentPath := g.parentPather(relToRootPath)
	parents, pErr := g.rem.FindByPath(parentPath)
	if pErr != nil {
		return pErr
	}

	for _, parent := range parents {
		if parent == nil {
			return fmt.Errorf("non existant parent '%s' for src", parentPath)
		}

		if err := g.rem.removeParent(fileId, parent.Id); err != nil {
			g.log.LogErrf("removeParent:: %s %s %v\n", fileId, relToRootPath, err)
		}
	}

	return nil
}

func (g *Commands) Rename(byId bool) error {
	if len(g.opts.Sources) < 2 {
		return fmt.Errorf("rename: expecting <src> <newname>")
	}

	src := g.opts.Sources[0]
	resolver := g.rem.FindByPath
	if byId {
		resolver = g.rem.FindByIdMulti
	}

	remoteSources, err := resolver(src)
	if err != nil {
		return fmt.Errorf("%s: %v", src, err)
	}

	for _, remSrc := range remoteSources {
		if remSrc == nil {
			g.log.LogErrf("%s does not exist", src)
		}

		if err = rename_(g, src, remSrc, byId); err != nil {
			g.log.LogErrf("%s %v\n", src, err)
		}
	}

	return nil
}

func rename_(g *Commands, src string, remSrc *File, byId bool) error {
	var parentPath string
	if !byId {
		parentPath = g.parentPather(src)
	} else {
		parentPath = g.opts.Path
	}

	newName := g.opts.Sources[1]
	urlBoundName := urlToPath(newName, true)
	newFullPath := filepath.Join(parentPath, urlBoundName)

	dupChecks, err := g.rem.FindByPath(newFullPath)

	if err == nil && len(dupChecks) >= 1 {
		for _, dup := range dupChecks {
			if dup.Id == remSrc.Id { // Trying to rename self
				continue
			}

			if !g.opts.Force {
				g.log.LogErrf("%s already exists. Use `%s` flag to override this behaviour", newFullPath, ForceKey)
				continue
			}

			if _, err = g.rem.rename(remSrc.Id, newName); err != nil {
				g.log.LogErrf("rename: %s %s %v\n", remSrc.Id, newName, err)
			}
		}
	}

	return err
}
