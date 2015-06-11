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

// Package contains the main entry point of gd.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/odeke-em/command"
	"github.com/odeke-em/drive/config"
	"github.com/odeke-em/drive/gen"
	"github.com/odeke-em/drive/src"
)

var context *config.Context
var DefaultMaxProcs = runtime.NumCPU()

func bindCommandWithAliases(key, description string, cmd command.Cmd, requiredFlags []string) {
	command.On(key, description, cmd, requiredFlags)
	aliases, ok := drive.Aliases[key]
	if ok {
		for _, alias := range aliases {
			command.On(alias, description, cmd, requiredFlags)
		}
	}
}

func main() {
	maxProcs, err := strconv.ParseInt(os.Getenv("GOMAXPROCS"), 10, 0)
	if err != nil || maxProcs < 1 {
		maxProcs = int64(DefaultMaxProcs)
	}
	runtime.GOMAXPROCS(int(maxProcs))

	bindCommandWithAliases(drive.AboutKey, drive.DescAbout, &aboutCmd{}, []string{})
	bindCommandWithAliases(drive.CopyKey, drive.DescCopy, &copyCmd{}, []string{})
	bindCommandWithAliases(drive.DiffKey, drive.DescDiff, &diffCmd{}, []string{})
	bindCommandWithAliases(drive.EmptyTrashKey, drive.DescEmptyTrash, &emptyTrashCmd{}, []string{})
	bindCommandWithAliases(drive.FeaturesKey, drive.DescFeatures, &featuresCmd{}, []string{})
	bindCommandWithAliases(drive.InitKey, drive.DescInit, &initCmd{}, []string{})
	bindCommandWithAliases(drive.HelpKey, drive.DescHelp, &helpCmd{}, []string{})

	bindCommandWithAliases(drive.ListKey, drive.DescList, &listCmd{}, []string{})
	bindCommandWithAliases(drive.MoveKey, drive.DescMove, &moveCmd{}, []string{})
	bindCommandWithAliases(drive.PullKey, drive.DescPull, &pullCmd{}, []string{})
	bindCommandWithAliases(drive.PushKey, drive.DescPush, &pushCmd{}, []string{})
	bindCommandWithAliases(drive.PubKey, drive.DescPublish, &publishCmd{}, []string{})
	bindCommandWithAliases(drive.RenameKey, drive.DescRename, &renameCmd{}, []string{})
	bindCommandWithAliases(drive.QuotaKey, drive.DescQuota, &quotaCmd{}, []string{})
	bindCommandWithAliases(drive.ShareKey, drive.DescShare, &shareCmd{}, []string{})
	bindCommandWithAliases(drive.StatKey, drive.DescStat, &statCmd{}, []string{})
	bindCommandWithAliases(drive.Md5sumKey, drive.DescMd5sum, &md5SumCmd{}, []string{})
	bindCommandWithAliases(drive.UnshareKey, drive.DescUnshare, &unshareCmd{}, []string{})
	bindCommandWithAliases(drive.TouchKey, drive.DescTouch, &touchCmd{}, []string{})
	bindCommandWithAliases(drive.TrashKey, drive.DescTrash, &trashCmd{}, []string{})
	bindCommandWithAliases(drive.UntrashKey, drive.DescUntrash, &untrashCmd{}, []string{})
	bindCommandWithAliases(drive.DeleteKey, drive.DescDelete, &deleteCmd{}, []string{})
	bindCommandWithAliases(drive.UnpubKey, drive.DescUnpublish, &unpublishCmd{}, []string{})
	bindCommandWithAliases(drive.VersionKey, drive.Version, &versionCmd{}, []string{})

	command.DefineHelp(&helpCmd{})
	command.ParseAndRun()
}

type helpCmd struct {
	args []string
}

func (cmd *helpCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *helpCmd) Run(args []string) {
	drive.ShowDescriptions(args...)
	exitWithError(nil)
}

type featuresCmd struct{}

func (cmd *featuresCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *featuresCmd) Run(args []string) {
	context, path := discoverContext(args)
	exitWithError(drive.New(context, &drive.Options{
		Path: path,
	}).About(drive.AboutFeatures))
}

type versionCmd struct{}

func (cmd *versionCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *versionCmd) Run(args []string) {
	fmt.Printf("drive version: %s\n%s\n", drive.Version, generated.PkgInfo)
	exitWithError(nil)
}

type initCmd struct{}

func (cmd *initCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *initCmd) Run(args []string) {
	exitWithError(drive.New(initContext(args), nil).Init())
}

type quotaCmd struct{}

func (cmd *quotaCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *quotaCmd) Run(args []string) {
	context, path := discoverContext(args)
	exitWithError(drive.New(context, &drive.Options{
		Path: path,
	}).About(drive.AboutQuota))
}

type listCmd struct {
	byId        *bool
	hidden      *bool
	pageCount   *int
	recursive   *bool
	files       *bool
	directories *bool
	depth       *int
	pageSize    *int64
	longFmt     *bool
	noPrompt    *bool
	shared      *bool
	inTrash     *bool
	version     *bool
	matches     *bool
	owners      *bool
	quiet       *bool
	sort        *string
}

func (cmd *listCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.depth = fs.Int(drive.DepthKey, 1, "maximum recursion depth")
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "list all paths even hidden ones")
	cmd.files = fs.Bool("f", false, "list only files")
	cmd.directories = fs.Bool("d", false, "list all directories")
	cmd.longFmt = fs.Bool("l", false, "long listing of contents")
	cmd.pageSize = fs.Int64("p", 100, "number of results per pagination")
	cmd.shared = fs.Bool("shared", false, "show files that are shared with me")
	cmd.inTrash = fs.Bool(drive.TrashedKey, false, "list content in the trash")
	cmd.version = fs.Bool("version", false, "show the number of times that the file has been modified on \n\t\tthe server even with changes not visible to the user")
	cmd.noPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before pagination")
	cmd.owners = fs.Bool("owners", false, "shows the owner names per file")
	cmd.recursive = fs.Bool("r", false, "recursively list subdirectories")
	cmd.sort = fs.String(drive.SortKey, "", drive.DescSort)
	cmd.matches = fs.Bool(drive.MatchesKey, false, "list by prefix")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "list by id instead of path")

	return fs
}

func (cmd *listCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, (*cmd.byId || *cmd.matches))

	typeMask := 0
	if *cmd.directories {
		typeMask |= drive.Folder
	}
	if *cmd.shared {
		typeMask |= drive.Shared
	}
	if *cmd.owners {
		typeMask |= drive.Owners
	}
	if *cmd.version {
		typeMask |= drive.CurrentVersion
	}
	if *cmd.files {
		typeMask |= drive.NonFolder
	}
	if *cmd.inTrash {
		typeMask |= drive.InTrash
	}
	if !*cmd.longFmt {
		typeMask |= drive.Minimal
	}

	depth := *cmd.depth
	if *cmd.recursive {
		depth = drive.InfiniteDepth
	}

	meta := map[string][]string{
		drive.SortKey: drive.NonEmptyTrimmedStrings(*cmd.sort),
	}

	options := drive.Options{
		Depth:     depth,
		Hidden:    *cmd.hidden,
		InTrash:   *cmd.inTrash,
		PageSize:  *cmd.pageSize,
		Path:      path,
		NoPrompt:  *cmd.noPrompt,
		Recursive: *cmd.recursive,
		Sources:   sources,
		TypeMask:  typeMask,
		Quiet:     *cmd.quiet,
		Meta:      &meta,
	}

	if *cmd.shared {
		exitWithError(drive.New(context, &options).ListShared())
	} else if *cmd.matches {
		exitWithError(drive.New(context, &options).ListMatches())
	} else {
		exitWithError(drive.New(context, &options).List(*cmd.byId))
	}
}

type md5SumCmd struct {
	byId      *bool
	depth     *int
	hidden    *bool
	recursive *bool
	quiet     *bool
}

func (cmd *md5SumCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.depth = fs.Int(drive.DepthKey, 1, "maximum recursion depth")
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "discover hidden paths")
	cmd.recursive = fs.Bool("r", false, "recursively discover folders")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "stat by id instead of path")
	return fs
}

func (cmd *md5SumCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.byId)
	
	depth := *cmd.depth
	if *cmd.recursive {
		depth = drive.InfiniteDepth
	}

	opts := drive.Options{
		Hidden:    *cmd.hidden,
		Path:      path,
		Recursive: *cmd.recursive,
		Sources:   sources,
		Quiet:     *cmd.quiet,
		Depth:     depth,
		Md5sum:    true,		
	}

	if *cmd.byId {
		exitWithError(drive.New(context, &opts).StatById())
	} else {
		exitWithError(drive.New(context, &opts).Stat())
	}
}

type statCmd struct {
	byId      *bool
	depth     *int
	hidden    *bool
	recursive *bool
	quiet     *bool
	md5sum    *bool
}

func (cmd *statCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.depth = fs.Int(drive.DepthKey, 1, "maximum recursion depth")
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "discover hidden paths")
	cmd.recursive = fs.Bool("r", false, "recursively discover folders")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "stat by id instead of path")
	cmd.md5sum = fs.Bool(drive.Md5sumKey, false, "produce output compatible with md5sum(1)")
	return fs
}

func (cmd *statCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.byId)

	depth := *cmd.depth
	if *cmd.recursive {
		depth = drive.InfiniteDepth
	}

	opts := drive.Options{
		Hidden:    *cmd.hidden,
		Path:      path,
		Recursive: *cmd.recursive,
		Sources:   sources,
		Quiet:     *cmd.quiet,
		Depth:     depth,
		Md5sum:    *cmd.md5sum,
	}

	if *cmd.byId {
		exitWithError(drive.New(context, &opts).StatById())
	} else {
		exitWithError(drive.New(context, &opts).Stat())
	}
}

type pullCmd struct {
	byId              *bool
	exportsDir        *string
	export            *string
	excludeOps        *string
	force             *bool
	hidden            *bool
	matches           *bool
	noPrompt          *bool
	noClobber         *bool
	recursive         *bool
	ignoreChecksum    *bool
	ignoreConflict    *bool
	piped             *bool
	quiet             *bool
	ignoreNameClashes *bool
}

func (cmd *pullCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noClobber = fs.Bool("no-clobber", false, "prevents overwriting of old content")
	cmd.export = fs.String(
		"export", "", "comma separated list of formats to export your docs + sheets files")
	cmd.recursive = fs.Bool("r", true, "performs the pull action recursively")
	cmd.noPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before applying the pull action")
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows pulling of hidden paths")
	cmd.force = fs.Bool("force", false, "forces a pull even if no changes present")
	cmd.ignoreChecksum = fs.Bool(drive.CLIOptionIgnoreChecksum, true, drive.DescIgnoreChecksum)
	cmd.ignoreConflict = fs.Bool(drive.CLIOptionIgnoreConflict, false, drive.DescIgnoreConflict)
	cmd.ignoreNameClashes = fs.Bool(drive.CLIOptionIgnoreNameClashes, false, drive.DescIgnoreNameClashes)
	cmd.exportsDir = fs.String("export-dir", "", "directory to place exports")
	cmd.matches = fs.Bool(drive.MatchesKey, false, "search by prefix")
	cmd.piped = fs.Bool("piped", false, "if true, read content from stdin")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.excludeOps = fs.String(drive.CLIOptionExcludeOperations, "", drive.DescExcludeOps)
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "pull by id instead of path")

	return fs
}

func (cmd *pullCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, (*cmd.byId || *cmd.matches))

	excludes := drive.NonEmptyTrimmedStrings(strings.Split(*cmd.excludeOps, ",")...)
	excludeCrudMask := drive.CrudAtoi(excludes...)
	if excludeCrudMask == drive.AllCrudOperations {
		exitWithError(fmt.Errorf("all CRUD operations forbidden"))
	}

	// Filter out empty strings.
	exports := drive.NonEmptyTrimmedStrings(strings.Split(*cmd.export, ",")...)

	options := &drive.Options{
		Exports:           uniqOrderedStr(exports),
		ExportsDir:        strings.Trim(*cmd.exportsDir, " "),
		Force:             *cmd.force,
		Hidden:            *cmd.hidden,
		IgnoreChecksum:    *cmd.ignoreChecksum,
		IgnoreConflict:    *cmd.ignoreConflict,
		NoPrompt:          *cmd.noPrompt,
		NoClobber:         *cmd.noClobber,
		Path:              path,
		Recursive:         *cmd.recursive,
		Sources:           sources,
		Piped:             *cmd.piped,
		Quiet:             *cmd.quiet,
		IgnoreNameClashes: *cmd.ignoreNameClashes,
		ExcludeCrudMask:   excludeCrudMask,
	}

	if *cmd.matches {
		exitWithError(drive.New(context, options).PullMatches())
	} else if *cmd.piped {
		exitWithError(drive.New(context, options).PullPiped(*cmd.byId))
	} else {
		exitWithError(drive.New(context, options).Pull(*cmd.byId))
	}
}

type pushCmd struct {
	noClobber   *bool
	hidden      *bool
	force       *bool
	noPrompt    *bool
	recursive   *bool
	piped       *bool
	mountedPush *bool
	// convert when set tells Google drive to convert the document into
	// its appropriate Google Docs format
	convert *bool
	// ocr when set indicates that Optical Character Recognition should be
	// attempted on .[gif, jpg, pdf, png] uploads
	ocr               *bool
	ignoreChecksum    *bool
	ignoreConflict    *bool
	ignoreNameClashes *bool
	quiet             *bool
	coercedMimeKey    *string
	excludeOps        *string
}

func (cmd *pushCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noClobber = fs.Bool("no-clobber", false, "allows overwriting of old content")
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows pushing of hidden paths")
	cmd.recursive = fs.Bool("r", true, "performs the push action recursively")
	cmd.noPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before applying the push action")
	cmd.force = fs.Bool("force", false, "forces a push even if no changes present")
	cmd.mountedPush = fs.Bool("m", false, "allows pushing of mounted paths")
	cmd.convert = fs.Bool("convert", false, "toggles conversion of the file to its appropriate Google Doc format")
	cmd.ocr = fs.Bool("ocr", false, "if true, attempt OCR on gif, jpg, pdf and png uploads")
	cmd.piped = fs.Bool("piped", false, "if true, read content from stdin")
	cmd.ignoreChecksum = fs.Bool(drive.CLIOptionIgnoreChecksum, true, drive.DescIgnoreChecksum)
	cmd.ignoreConflict = fs.Bool(drive.CLIOptionIgnoreConflict, false, drive.DescIgnoreConflict)
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.coercedMimeKey = fs.String(drive.CoercedMimeKeyKey, "", "the mimeType you are trying to coerce this file to be")
	cmd.ignoreNameClashes = fs.Bool(drive.CLIOptionIgnoreNameClashes, false, drive.DescIgnoreNameClashes)
	cmd.excludeOps = fs.String(drive.CLIOptionExcludeOperations, "", drive.DescExcludeOps)
	return fs
}

func (cmd *pushCmd) Run(args []string) {
	if *cmd.mountedPush {
		cmd.pushMounted(args)
	} else {
		sources, context, path := preprocessArgs(args)

		options := cmd.createPushOptions()
		options.Path = path
		options.Sources = sources

		if *cmd.piped {
			exitWithError(drive.New(context, options).PushPiped())
		} else {
			exitWithError(drive.New(context, options).Push())
		}
	}
}

type touchCmd struct {
	byId      *bool
	hidden    *bool
	recursive *bool
	matches   *bool
	quiet     *bool
}

func (cmd *touchCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows pushing of hidden paths")
	cmd.recursive = fs.Bool("r", false, "toggles recursive touching")
	cmd.matches = fs.Bool(drive.MatchesKey, false, "search by prefix and touch")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "share by id instead of path")
	return fs
}

func (cmd *touchCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.matches || *cmd.byId)

	opts := drive.Options{
		Hidden:    *cmd.hidden,
		Path:      path,
		Recursive: *cmd.recursive,
		Sources:   sources,
		Quiet:     *cmd.quiet,
	}

	if *cmd.matches {
		exitWithError(drive.New(context, &opts).TouchByMatch())
	} else {
		exitWithError(drive.New(context, &opts).Touch(*cmd.byId))
	}
}

func (cmd *pushCmd) createPushOptions() *drive.Options {
	mask := drive.OptNone
	if *cmd.convert {
		mask |= drive.OptConvert
	}
	if *cmd.ocr {
		mask |= drive.OptOCR
	}

	meta := map[string][]string{
		drive.CoercedMimeKeyKey: drive.NonEmptyTrimmedStrings(*cmd.coercedMimeKey),
	}

	excludes := drive.NonEmptyTrimmedStrings(strings.Split(*cmd.excludeOps, ",")...)
	excludeCrudMask := drive.CrudAtoi(excludes...)
	if excludeCrudMask == drive.AllCrudOperations {
		exitWithError(fmt.Errorf("all CRUD operations forbidden yet asking to push"))
	}

	return &drive.Options{
		Force:             *cmd.force,
		Hidden:            *cmd.hidden,
		IgnoreChecksum:    *cmd.ignoreChecksum,
		IgnoreConflict:    *cmd.ignoreConflict,
		NoClobber:         *cmd.noClobber,
		NoPrompt:          *cmd.noPrompt,
		Recursive:         *cmd.recursive,
		Piped:             *cmd.piped,
		Quiet:             *cmd.quiet,
		Meta:              &meta,
		TypeMask:          mask,
		ExcludeCrudMask:   excludeCrudMask,
		IgnoreNameClashes: *cmd.ignoreNameClashes,
	}
}

func (cmd *pushCmd) pushMounted(args []string) {
	argc := len(args)

	var contextArgs, rest, sources []string

	if !*cmd.mountedPush {
		contextArgs = args
	} else {
		// Expectation is that at least one path has to be passed in
		if argc < 2 {
			cwd, cerr := os.Getwd()
			if cerr != nil {
				contextArgs = []string{cwd}
			}
			rest = args
		} else {
			rest = args[:argc-1]
			contextArgs = args[argc-1:]
		}
	}

	rest = drive.NonEmptyStrings(rest...)
	context, path := discoverContext(contextArgs)
	contextAbsPath, err := filepath.Abs(path)
	exitWithError(err)

	if path == "." {
		path = ""
	}

	mount, auxSrcs := config.MountPoints(path, contextAbsPath, rest, *cmd.hidden)

	root := context.AbsPathOf("")

	sources, err = relativePathsOpt(root, auxSrcs, true)
	exitWithError(err)

	options := cmd.createPushOptions()
	options.Mount = mount
	options.Sources = sources

	exitWithError(drive.New(context, options).Push())
}

type aboutCmd struct {
	features *bool
	quota    *bool
	filesize *bool
	quiet    *bool
}

func (cmd *aboutCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.features = fs.Bool("features", false, "gives information on features present on this drive")
	cmd.quota = fs.Bool("quota", false, "prints out quota information for this drive")
	cmd.filesize = fs.Bool("filesize", false, "prints out information about file sizes e.g the max upload size for a specific file size")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	return fs
}

func (cmd *aboutCmd) Run(args []string) {
	_, context, _ := preprocessArgs(args)

	mask := drive.AboutNone
	if *cmd.features {
		mask |= drive.AboutFeatures
	}
	if *cmd.quota {
		mask |= drive.AboutQuota
	}
	if *cmd.filesize {
		mask |= drive.AboutFileSizes
	}
	exitWithError(drive.New(context, &drive.Options{
		Quiet: *cmd.quiet,
	}).About(mask))
}

type diffCmd struct {
	hidden         *bool
	ignoreChecksum *bool
	quiet          *bool
}

func (cmd *diffCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows pulling of hidden paths")
	cmd.ignoreChecksum = fs.Bool(drive.CLIOptionIgnoreChecksum, true, drive.DescIgnoreChecksum)
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	return fs
}

func (cmd *diffCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		Recursive:      true,
		Path:           path,
		Hidden:         *cmd.hidden,
		Sources:        sources,
		IgnoreChecksum: *cmd.ignoreChecksum,
		Quiet:          *cmd.quiet,
	}).Diff())
}

type publishCmd struct {
	hidden *bool
	quiet  *bool
	byId   *bool
}

type unpublishCmd struct {
	hidden *bool
	quiet  *bool
	byId   *bool
}

func (cmd *unpublishCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows pulling of hidden paths")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "unpublish by id instead of path")
	return fs
}

func (cmd *unpublishCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.byId)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.quiet,
	}).Unpublish(*cmd.byId))
}

type emptyTrashCmd struct {
	noPrompt *bool
	quiet    *bool
}

func (cmd *emptyTrashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before emptying the trash")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	return fs
}

func (cmd *emptyTrashCmd) Run(args []string) {
	_, context, _ := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		NoPrompt: *cmd.noPrompt,
		Quiet:    *cmd.quiet,
	}).EmptyTrash())
}

type deleteCmd struct {
	hidden  *bool
	matches *bool
	quiet   *bool
	byId    *bool
}

func (cmd *deleteCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows trashing hidden paths")
	cmd.matches = fs.Bool(drive.MatchesKey, false, "search by prefix and delete")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "delete by id instead of path")
	return fs
}

func (cmd *deleteCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.matches || *cmd.byId)

	opts := drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.quiet,
	}

	if !*cmd.matches {
		exitWithError(drive.New(context, &opts).Delete(*cmd.byId))
	} else {
		exitWithError(drive.New(context, &opts).DeleteByMatch())
	}
}

type trashCmd struct {
	hidden  *bool
	matches *bool
	quiet   *bool
	byId    *bool
}

func (cmd *trashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows trashing hidden paths")
	cmd.matches = fs.Bool(drive.MatchesKey, false, "search by prefix and trash")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "trash by id instead of path")
	return fs
}

func (cmd *trashCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.matches || *cmd.byId)
	opts := drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.quiet,
	}

	if !*cmd.matches {
		exitWithError(drive.New(context, &opts).Trash(*cmd.byId))
	} else {
		exitWithError(drive.New(context, &opts).TrashByMatch())
	}
}

type copyCmd struct {
	quiet     *bool
	recursive *bool
	byId      *bool
}

func (cmd *copyCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.recursive = fs.Bool("r", false, "recursive copying")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "copy by id instead of path")
	return fs
}

func (cmd *copyCmd) Run(args []string) {
	if len(args) < 2 {
		args = append(args, ".")
	}

	end := len(args) - 1
	if end < 1 {
		exitWithError(fmt.Errorf("copy: expected more than one path"))
	}

	dest := args[end]

	sources, context, path := preprocessArgsByToggle(args, *cmd.byId)
	// Unshift by the end path
	sources = sources[:len(sources)-1]
	destRels, err := relativePaths(context.AbsPathOf(""), dest)
	exitWithError(err)

	dest = destRels[0]
	sources = append(sources, dest)

	exitWithError(drive.New(context, &drive.Options{
		Path:      path,
		Sources:   sources,
		Recursive: *cmd.recursive,
		Quiet:     *cmd.quiet,
	}).Copy(*cmd.byId))
}

type untrashCmd struct {
	hidden  *bool
	matches *bool
	quiet   *bool
	byId    *bool
}

func (cmd *untrashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows untrashing hidden paths")
	cmd.matches = fs.Bool(drive.MatchesKey, false, "search by prefix and untrash")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "untrash by id instead of path")
	return fs
}

func (cmd *untrashCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.byId || *cmd.matches)

	opts := drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.quiet,
	}

	if !*cmd.matches {
		exitWithError(drive.New(context, &opts).Untrash(*cmd.byId))
	} else {
		exitWithError(drive.New(context, &opts).UntrashByMatch())
	}
}

func (cmd *publishCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool(drive.HiddenKey, false, "allows publishing of hidden paths")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "publish by id instead of path")
	return fs
}

func (cmd *publishCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.byId)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.quiet,
	}).Publish(*cmd.byId))
}

type unshareCmd struct {
	noPrompt    *bool
	accountType *string
	quiet       *bool
	byId        *bool
}

func (cmd *unshareCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.accountType = fs.String("type", "", "scope of account to revoke access to")
	cmd.noPrompt = fs.Bool(drive.NoPromptKey, false, "disables the prompt")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "unshare by id instead of path")
	return fs
}

func (cmd *unshareCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.byId)

	meta := map[string][]string{
		"accountType": uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.accountType, ",")...)),
	}

	exitWithError(drive.New(context, &drive.Options{
		Meta:     &meta,
		Path:     path,
		Sources:  sources,
		NoPrompt: *cmd.noPrompt,
		Quiet:    *cmd.quiet,
	}).Unshare(*cmd.byId))
}

type moveCmd struct {
	quiet *bool
	byId  *bool
}

func (cmd *moveCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "unshare by id instead of path")
	return fs
}

func (cmd *moveCmd) Run(args []string) {
	argc := len(args)
	if argc < 1 {
		exitWithError(fmt.Errorf("move: expecting a path or more"))
	}
	sources, context, path := preprocessArgsByToggle(args, *cmd.byId)
	// Unshift by the end path
	sources = sources[:len(sources)-1]

	dest := args[argc-1]
	destRels, err := relativePaths(context.AbsPathOf(""), dest)
	exitWithError(err)

	sources = append(sources, destRels[0])

	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.quiet,
	}).Move(*cmd.byId))
}

type renameCmd struct {
	force *bool
	quiet *bool
	byId  *bool
}

func (cmd *renameCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.force = fs.Bool("force", false, "coerce rename even if remote already exists")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "unshare by id instead of path")
	return fs
}

func (cmd *renameCmd) Run(args []string) {
	argc := len(args)
	if argc < 2 {
		exitWithError(fmt.Errorf("rename: expecting <src> <dest>"))
	}
	rest, last := args[:argc-1], args[argc-1]
	sources, context, path := preprocessArgsByToggle(rest, *cmd.byId)

	sources = append(sources, last)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
		Force:   *cmd.force,
		Quiet:   *cmd.quiet,
	}).Rename(*cmd.byId))
}

type shareCmd struct {
	byId        *bool
	emails      *string
	message     *string
	role        *string
	accountType *string
	noPrompt    *bool
	notify      *bool
	quiet       *bool
}

func (cmd *shareCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.emails = fs.String("emails", "", "emails to share the file to")
	cmd.message = fs.String("message", "", "message to send receipients")
	cmd.role = fs.String("role", "", "role to set to receipients of share. Possible values: "+drive.DescRoles)
	cmd.accountType = fs.String("type", "", "scope of accounts to share files with. Possible values: "+drive.DescAccountTypes)
	cmd.notify = fs.Bool("notify", true, "toggle whether to notify receipients about share")
	cmd.noPrompt = fs.Bool(drive.NoPromptKey, false, "disables the prompt")
	cmd.quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.byId = fs.Bool(drive.CLIOptionId, false, "share by id instead of path")
	return fs
}

func (cmd *shareCmd) Run(args []string) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.byId)

	meta := map[string][]string{
		"emailMessage": []string{*cmd.message},
		"emails":       uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.emails, ",")...)),
		"role":         uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.role, ",")...)),
		"accountType":  uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.accountType, ",")...)),
	}

	mask := drive.NoopOnShare
	if *cmd.notify {
		mask = drive.Notify
	}

	exitWithError(drive.New(context, &drive.Options{
		Meta:     &meta,
		Path:     path,
		Sources:  sources,
		TypeMask: mask,
		NoPrompt: *cmd.noPrompt,
		Quiet:    *cmd.quiet,
	}).Share(*cmd.byId))
}

func initContext(args []string) *config.Context {
	var err error
	var gdPath string
	var firstInit bool

	gdPath, firstInit, context, err = config.Initialize(getContextPath(args))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// The signal handler should clean up the .gd path if this is the first time
	go func() {
		_ = <-c
		if firstInit {
			os.RemoveAll(gdPath)
		}
		os.Exit(1)
	}()

	exitWithError(err)
	return context
}

func discoverContext(args []string) (*config.Context, string) {
	var err error
	context, err = config.Discover(getContextPath(args))
	exitWithError(err)
	relPath := ""
	if len(args) > 0 {
		var headAbsArg string
		headAbsArg, err = filepath.Abs(args[0])
		if err == nil {
			relPath, err = filepath.Rel(context.AbsPath, headAbsArg)
		}
	}

	exitWithError(err)

	// relPath = strings.Join([]string{"", relPath}, "/")
	return context, relPath
}

func getContextPath(args []string) (contextPath string) {
	if len(args) > 0 {
		contextPath, _ = filepath.Abs(args[0])
	}
	if contextPath == "" {
		contextPath, _ = os.Getwd()
	}
	return
}

func uniqOrderedStr(sources []string) []string {
	cache := map[string]bool{}
	var uniqPaths []string
	for _, p := range sources {
		ok := cache[p]
		if ok {
			continue
		}
		uniqPaths = append(uniqPaths, p)
		cache[p] = true
	}
	return uniqPaths
}

func exitWithError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func relativePaths(root string, args ...string) ([]string, error) {
	return relativePathsOpt(root, args, false)
}

func relativePathsOpt(root string, args []string, leastNonExistant bool) ([]string, error) {
	var err error
	var relPath string
	var relPaths []string

	for _, p := range args {
		p, err = filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", p, err)
			continue
		}

		if leastNonExistant {
			sRoot := config.LeastNonExistantRoot(p)
			if sRoot != "" {
				p = sRoot
			}
		}

		relPath, err = filepath.Rel(root, p)
		if err != nil {
			break
		}

		if relPath == "." {
			relPath = ""
		}

		relPath = "/" + relPath
		relPaths = append(relPaths, relPath)
	}

	return relPaths, err
}

func preprocessArgs(args []string) ([]string, *config.Context, string) {
	context, path := discoverContext(args)
	root := context.AbsPathOf("")

	if len(args) < 1 {
		args = []string{"."}
	}

	relPaths, err := relativePaths(root, args...)
	exitWithError(err)

	return uniqOrderedStr(relPaths), context, path
}

func preprocessArgsByToggle(args []string, skipArgPreprocess bool) (sources []string, context *config.Context, path string) {
	if !skipArgPreprocess {
		return preprocessArgs(args)
	}

	cwd, err := os.Getwd()
	exitWithError(err)
	_, context, path = preprocessArgs([]string{cwd})
	sources = uniqOrderedStr(args)
	return sources, context, path
}
