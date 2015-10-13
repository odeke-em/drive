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
	"errors"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/mattn/go-isatty"
	expirable "github.com/odeke-em/cache"
	"github.com/odeke-em/drive/config"
	"github.com/odeke-em/log"
)

var (
	ErrNoContext = errors.New("not in a drive context")
)

const (
	DriveIgnoreSuffix = ".driveignore"
)

type Options struct {
	// Depth is the number of pages/ listing recursion depth
	Depth int
	// Exports contains the formats to export your Google Docs + Sheets to
	// e.g ["csv" "txt"]
	Exports []string
	// ExportsDir is the directory to put the exported Google Docs + Sheets.
	// If not provided, will export them to the same dir as the source files are
	ExportsDir string
	// Force once set always converts NoChange into an Addition
	Force bool
	// Hidden discovers hidden paths if set
	Hidden       bool
	IgnoreRegexp *regexp.Regexp
	// IgnoreChecksum when set avoids the step
	// of comparing checksums as a final check.
	IgnoreChecksum bool
	// IgnoreConflict when set turns off the conflict resolution safety.
	IgnoreConflict bool
	// Allows listing of content in trash
	InTrash bool
	Meta    *map[string][]string
	Mount   *config.Mount
	// NoClobber when set prevents overwriting of stale content
	NoClobber bool
	// NoPrompt overwrites any prompt pauses
	NoPrompt bool
	Path     string
	// PageSize determines the number of results returned per API call
	PageSize  int64
	Recursive bool
	// Sources is a of list all paths that are
	// within the scope/path of the current gd context
	Sources []string
	// TypeMask contains the result of setting different type bits e.g
	// Folder to search only for folders etc.
	TypeMask int
	// Piped when set means to infer content to or from stdin
	Piped bool
	// Quiet when set toggles only logging of errors to stderrs as
	// well as reading from stdin in this case stdout is not logged to
	Quiet             bool
	StdoutIsTty       bool
	IgnoreNameClashes bool
	ExcludeCrudMask   CrudValue
	ExplicitlyExport  bool
	Md5sum            bool
	indexingOnly      bool
	Verbose           bool
	FixClashes        bool
}

type Commands struct {
	context *config.Context
	rem     *Remote
	opts    *Options
	log     *log.Logger

	progress      *pb.ProgressBar
	mkdirAllCache *expirable.OperationCache
}

func (opts *Options) canPrompt() bool {
	if opts == nil || !opts.StdoutIsTty {
		return false
	}
	if opts.Quiet {
		return false
	}
	return !opts.NoPrompt
}

func New(context *config.Context, opts *Options) *Commands {
	var r *Remote
	if context != nil {
		r = NewRemoteContext(context)
	}

	stdin, stdout, stderr := os.Stdin, os.Stdout, os.Stderr

	logger := log.New(stdin, stdout, stderr)

	if opts != nil {
		// should always start with /
		opts.Path = path.Clean(path.Join("/", opts.Path))

		if !opts.Force {
			ignoresPath := filepath.Join(context.AbsPath, DriveIgnoreSuffix)
			ignoreRegexp, regErr := combineIgnores(ignoresPath)

			if regErr != nil {
				logger.LogErrf("combining ignores from path %s and internally: %v\n", ignoresPath, regErr)
			}

			opts.IgnoreRegexp = ignoreRegexp
		}

		opts.StdoutIsTty = isatty.IsTerminal(stdout.Fd())

		if opts.Quiet {
			stdout = nil
		}
	}

	return &Commands{
		context:       context,
		rem:           r,
		opts:          opts,
		log:           logger,
		mkdirAllCache: expirable.New(),
	}
}

func combineIgnores(ignoresPath string) (*regexp.Regexp, error) {
	clauses, err := readCommentedFile(ignoresPath, "#")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	clauses = append(clauses, internalIgnores()...)
	if len(clauses) < 1 {
		return nil, nil
	}

	regExComp, regErr := regexp.Compile(strings.Join(clauses, "|"))
	if regErr != nil {
		return nil, regErr
	}
	return regExComp, nil
}

func (g *Commands) taskStart(tasks int64) {
	if tasks > 0 {
		g.progress = newProgressBar(tasks)
	}
}

func newProgressBar(total int64) *pb.ProgressBar {
	pbf := pb.New64(total)
	pbf.Start()
	return pbf
}

func (g *Commands) taskAdd(n int64) {
	if g.progress != nil {
		g.progress.Add64(n)
	}
}

func (g *Commands) taskFinish() {
	if g.progress != nil {
		g.progress.Finish()
	}
}
