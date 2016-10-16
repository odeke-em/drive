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
	"os"
	"path/filepath"
)

type moveOpt struct {
	src        string
	dest       string
	byId       bool
	keepParent bool
}

type RenameMode uint

const (
	RenameNone RenameMode = 1 << iota
	RenameRemote
	RenameLocal
)

func canRenameLocal(rmode RenameMode) bool {
	return (rmode & RenameLocal) != 0
}

func canRenameRemote(rmode RenameMode) bool {
	return (rmode & RenameRemote) != 0
}

func cannotRename(rmode RenameMode) bool {
	return !(canRenameLocal(rmode) || canRenameRemote(rmode))
}

func (g *Commands) Move(byId, keepParent bool ) error {
	argc := len(g.opts.Sources)
	if argc < 2 {
		return invalidArgumentsErr(fmt.Errorf("move: expected <src> [src...] <dest>, instead got: %v", g.opts.Sources))
	}

	rest, dest := g.opts.Sources[:argc-1], g.opts.Sources[argc-1]

	var composedError error = nil

	for _, src := range rest {
		prefix := commonPrefix(src, dest)

		// Trying to nest a parent into its child
		if prefix == src {
			return illogicalStateErr(fmt.Errorf("%s cannot be nested into %s", src, dest))
		}

		opt := moveOpt{
			src:        src,
			dest:       dest,
			byId:       byId,
			keepParent: keepParent,
		}

		if err := g.move(&opt); err != nil {
			message := fmt.Sprintf("move: %s: %v", src, err)
			composedError = reComposeError(composedError, message)
		}
	}

	return composedError
}

func (g *Commands) move(opt *moveOpt) (err error) {
	var newParent, remSrc *File

	srcResolver := g.rem.FindByPath
	if opt.byId {
		srcResolver = g.rem.FindById
	}

	if remSrc, err = srcResolver(opt.src); err != nil {
		return makeErrorWithStatus(fmt.Sprintf("src('%s')", opt.src), err, StatusRemoteLookupFailed)
	}

	if remSrc == nil {
		return nonExistantRemoteErr(fmt.Errorf("src: '%s' could not be found", opt.src))
	}

	if newParent, err = g.rem.FindByPath(opt.dest); err != nil {
		return remoteLookupErr(fmt.Errorf("dest: '%s' %v", opt.dest, err))
	}

	if newParent == nil || !newParent.IsDir {
		return illogicalStateErr(fmt.Errorf("dest: '%s' must be an existent folder", opt.dest))
	}

	if !opt.byId {
		parentPath := g.parentPather(opt.src)
		oldParent, parErr := g.rem.FindByPath(parentPath)
		if parErr != nil && parErr != ErrPathNotExists {
			return parErr
		}

		// TODO: If oldParent is not found, retry since it may have been moved temporarily at least
		if oldParent != nil && oldParent.Id == newParent.Id {
			return illogicalStateErr(fmt.Errorf("src and dest are the same srcParentId %s destParentId %s",
				customQuote(oldParent.Id), customQuote(newParent.Id)))
		}
	}

	newFullPath := filepath.Join(opt.dest, remSrc.Name)

	// Check for a duplicate
	var dupCheck *File
	dupCheck, err = g.rem.FindByPath(newFullPath)
	if err != nil && err != ErrPathNotExists {
		return err
	}

	if dupCheck != nil {
		if dupCheck.Id == remSrc.Id { // Trying to move to self
			return illogicalStateErr(fmt.Errorf("move: trying to move fileId:%s to self fileId:%s", customQuote(dupCheck.Id), customQuote(remSrc.Id)))
		}
		if !g.opts.Force {
			return overwriteAttemptedErr(fmt.Errorf("%s already exists. Use `%s` flag to override this behaviour", newFullPath, ForceKey))
		}
	}

	// Avoid self-nesting
	if remSrc.Id == newParent.Id {
		return illogicalStateErr(fmt.Errorf("move: cannot move '%s' to itself", opt.src))
	}

	if err = g.rem.insertParent(remSrc.Id, newParent.Id); err != nil {
		return err
	}

	if opt.byId { // TODO: Also take out this current parent
		return nil
	}

	if !opt.keepParent {
		err = g.removeParent(remSrc.Id, opt.src)
	}

	return err
}

func (g *Commands) removeParent(fileId, relToRootPath string) error {
	parentPath := g.parentPather(relToRootPath)
	parent, pErr := g.rem.FindByPath(parentPath)
	if pErr != nil {
		return pErr
	}
	if parent == nil {
		return illogicalStateErr(fmt.Errorf("non existent parent '%s' for src", parentPath))
	}
	return g.rem.removeParent(fileId, parent.Id)
}

func (g *Commands) renameLocal(oldRelToRootPath, newRelToRootPath string) error {
	oldLocalPath := g.context.AbsPathOf(oldRelToRootPath)
	newLocalPath := g.context.AbsPathOf(newRelToRootPath)

	return os.Rename(oldLocalPath, newLocalPath)
}

func (g *Commands) renameRemote(remSrcId, parentPath, newName string) error {
	urlBoundName := urlToPath(newName, true)
	newFullPath := filepath.Join(parentPath, urlBoundName)

	dupCheck, err := g.rem.FindByPath(newFullPath)

	if err == nil && dupCheck != nil {
		if dupCheck.Id == remSrcId { // Trying to rename self
			return nil
		}
		if !g.opts.Force {
			return overwriteAttemptedErr(fmt.Errorf("%s already exists. Use `%s` flag to override this behaviour", newFullPath, ForceKey))
		}
	}

	_, err = g.rem.rename(remSrcId, newName)
	return err
}

func (g *Commands) Rename(byId bool) error {
	if len(g.opts.Sources) < 2 {
		return invalidArgumentsErr(fmt.Errorf("rename: expecting <src> <newname>"))
	}

	if cannotRename(g.opts.RenameMode) {
		return illogicalStateErr(fmt.Errorf("no rename mode set, set either `local` or `remote` mode"))
	}

	src := g.opts.Sources[0]
	resolver := g.rem.FindByPath
	if byId {
		resolver = g.rem.FindById
	}
	remSrc, err := resolver(src)
	if err != nil {
		return fmt.Errorf("%s: %v", src, err)
	}
	if remSrc == nil {
		return illogicalStateErr(fmt.Errorf("%s does not exist", src))
	}

	var parentPath string
	if !byId {
		parentPath = g.parentPather(src)
	} else {
		parentPath = g.opts.Path
	}

	newName := g.opts.Sources[1]
	if canRenameRemote(g.opts.RenameMode) {
		if err = g.renameRemote(remSrc.Id, parentPath, newName); err != nil {
			return err
		}
	}

	if canRenameLocal(g.opts.RenameMode) {
		// Otherwise time to mirror the same changes locally
		oldRelToRootPath := sepJoin(RemoteSeparator, parentPath, remSrc.Name)
		newRelToRootPath := sepJoin(RemoteSeparator, parentPath, newName)

		if err = g.renameLocal(oldRelToRootPath, newRelToRootPath); err != nil {
			return err
		}
	}

	return err
}
