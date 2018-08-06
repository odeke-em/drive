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
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/cheggaaa/pb"
	"github.com/mattn/go-isatty"
	expirableCache "github.com/odeke-em/cache"
	"github.com/odeke-em/drive/config"
	"github.com/odeke-em/log"
)

var (
	ErrNoContext = errors.New("not in a drive context")
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

	// ExportsDumpToSameDirectory when set, requests that all exports be put in the
	// same directory instead of in a directory that is prefixed first by the file name
	ExportsDumpToSameDirectory bool

	// Force once set always converts NoChange into an Addition
	Force bool
	// Hidden discovers hidden paths if set
	Hidden  bool
	Ignorer func(string) bool
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
	FixClashesMode    FixClashesMode
	Match             bool
	Starred           bool
	// BaseLocal when set, during a diff uses the local file
	// as the base otherwise remote is used as the base
	BaseLocal bool
	// Destination when set is the final logical location of the
	// constituents of an operation for example a push or pull.
	// See issue #612.
	Destination                  string
	RenameMode                   RenameMode
	ExponentialBackoffRetryCount int

	Encrypter func(io.Reader) (io.Reader, error)
	Decrypter func(io.Reader) (io.ReadCloser, error)

	// AllowURLLinkedFiles when set signifies that the user
	// depending on their OS, wants us to create for them
	// clickable files where applicable.
	// See issue #697.
	AllowURLLinkedFiles bool

	// Chunksize is the size per block of data uploaded.
	// If not set, the default value from googleapi.DefaultUploadChunkSize
	// is used instead.
	// If UploadChunkSize is not set yet UploadRateLimit is, UploadChunkSize will be the same as UploadRateLimit.
	UploadChunkSize int

	// Limit the upload bandwidth to n KiB/s.
	UploadRateLimit int
}

func (opts *Options) CryptoEnabled() bool {
	return opts.Decrypter != nil || opts.Encrypter != nil
}

type Commands struct {
	context *config.Context
	rem     *Remote
	opts    *Options
	rcOpts  *Options
	log     *log.Logger

	progress      *pb.ProgressBar
	mkdirAllCache *expirableCache.OperationCache
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

func (c *Commands) DebugPrintf(fmt_ string, args ...interface{}) {
	if !((Debug() || c.opts.Verbose) && c.opts.canPreview()) {
		return
	}
	FDebugPrintf(c.log, fmt_, args...)
}

func (opts *Options) canPreview() bool {
	if opts == nil || !opts.StdoutIsTty {
		return false
	}
	if opts.Quiet {
		return false
	}
	return true
}

func rcPathChecker(absDir string) (string, error) {
	p := rcPath(absDir)
	statInfo, err := os.Stat(p)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if statInfo == nil {
		return "", os.ErrNotExist
	}
	return p, nil
}

func (opts *Options) rcPath() (string, error) {
	lastCurPath := ""
	for curPath := opts.Path; curPath != ""; curPath = path.Dir(curPath) {
		localRCP, err := rcPathChecker(curPath)
		if err == nil && localRCP != "" {
			return localRCP, nil
		}

		if false && err != nil && !os.IsNotExist(err) {
			return "", err
		}

		if lastCurPath == curPath { // Avoid getting a stalemate incase path.Dir cannot progress
			break
		}

		lastCurPath = curPath
	}

	return rcPathChecker(FsHomeDir)
}

func New(context *config.Context, opts *Options) *Commands {
	var rem *Remote
	var err error

	if context.GSAJWTConfig != nil {
		rem, err = NewRemoteContextFromServiceAccount(context.GSAJWTConfig)
	} else {
		rem, err = NewRemoteContext(context)
	}

	if err != nil {
		panic(fmt.Errorf("failed to initialize remoteContext: %v", err))
	}

	stdin, stdout, stderr := os.Stdin, os.Stdout, os.Stderr

	var logger *log.Logger = nil

	if opts == nil {
		logger = log.New(stdin, stdout, stderr)
	} else {
		if opts.Quiet {
			stdout = nil
		}

		if stdout != nil {
			opts.StdoutIsTty = isatty.IsTerminal(stdout.Fd())
		}

		if stdout == nil && opts.Piped {
			panic("piped requires stdout to be non-nil")
		}

		logger = log.New(stdin, stdout, stderr)

		// should always start with /
		opts.Path = path.Clean(path.Join("/", opts.Path))

		if !opts.Force {
			ignoresPath := filepath.Join(context.AbsPath, DriveIgnoreSuffix)
			ignorer, regErr := combineIgnores(ignoresPath)

			if regErr != nil {
				logger.LogErrf("combining ignores from path %s and internally: %v\n", ignoresPath, regErr)
			}

			opts.Ignorer = ignorer
		}

		if opts.UploadChunkSize == 0 {
			// UploadRateLimit is in KiB/s
			opts.UploadChunkSize = opts.UploadRateLimit * 1024
		}
	}

	return &Commands{
		context:       context,
		rem:           rem,
		opts:          opts,
		log:           logger,
		mkdirAllCache: expirableCache.New(),
	}
}

func (g *Commands) taskStart(tasks int64) {
	if tasks > 0 && g.opts.canPreview() {
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
