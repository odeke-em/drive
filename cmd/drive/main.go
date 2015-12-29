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
	"encoding/json"
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

type errorer func() error

func bindCommandWithAliases(key, description string, cmd command.Cmd, requiredFlags []string) {
	command.On(key, description, cmd, requiredFlags)
	aliases, ok := drive.Aliases[key]
	if ok {
		for _, alias := range aliases {
			command.On(alias, description, cmd, requiredFlags)
		}
	}
}

func translateKeyChecks(definedFlags map[string]*flag.Flag) map[string]bool {
	keysOnly := map[string]bool{}

	for k, _ := range definedFlags {
		keysOnly[k] = true
	}

	return keysOnly
}

type defaultsFiller struct {
	from, to     interface{}
	rcSourcePath string
	definedFlags map[string]*flag.Flag
}

func fillWithDefaults(df defaultsFiller) error {
	alreadyDefined := translateKeyChecks(df.definedFlags)
	jsonStringified, err := drive.JSONStringifySiftedCLITags(df.from, df.rcSourcePath, alreadyDefined)

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(jsonStringified), df.to)
}

func main() {
	maxProcs, err := strconv.ParseInt(os.Getenv(drive.GoMaxProcsKey), 10, 0)
	if err != nil || maxProcs < 1 {
		maxProcs = int64(drive.DefaultMaxProcs)
	}
	runtime.GOMAXPROCS(int(maxProcs))

	bindCommandWithAliases(drive.AboutKey, drive.DescAbout, &aboutCmd{}, []string{})
	bindCommandWithAliases(drive.CopyKey, drive.DescCopy, &copyCmd{}, []string{})
	bindCommandWithAliases(drive.DiffKey, drive.DescDiff, &diffCmd{}, []string{})
	bindCommandWithAliases(drive.EmptyTrashKey, drive.DescEmptyTrash, &emptyTrashCmd{}, []string{})
	bindCommandWithAliases(drive.FeaturesKey, drive.DescFeatures, &featuresCmd{}, []string{})
	bindCommandWithAliases(drive.InitKey, drive.DescInit, &initCmd{}, []string{})
	bindCommandWithAliases(drive.DeInitKey, drive.DescDeInit, &deInitCmd{}, []string{})
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
	bindCommandWithAliases(drive.NewKey, drive.DescNew, &newCmd{}, []string{})
	bindCommandWithAliases(drive.IndexKey, drive.DescIndex, &indexCmd{}, []string{})
	bindCommandWithAliases(drive.UrlKey, drive.DescUrl, &urlCmd{}, []string{})
	bindCommandWithAliases(drive.OpenKey, drive.DescOpen, &openCmd{}, []string{})
	bindCommandWithAliases(drive.EditDescriptionKey, drive.DescEdit, &editDescriptionCmd{}, []string{})
	bindCommandWithAliases(drive.QRLinkKey, drive.DescQR, &qrLinkCmd{}, []string{})
	bindCommandWithAliases(drive.DuKey, drive.DescDu, &duCmd{}, []string{})
	bindCommandWithAliases(drive.StarKey, drive.DescStar, &starCmd{}, []string{})
	bindCommandWithAliases(drive.UnStarKey, drive.DescUnStar, &unstarCmd{}, []string{})
	bindCommandWithAliases(drive.ClashesKey, drive.DescFixClashes, &clashesCmd{}, []string{})
	bindCommandWithAliases(drive.IdKey, drive.DescId, &idCmd{}, []string{})
	bindCommandWithAliases(drive.ReportIssueKey, drive.DescReportIssue, &issueCmd{}, []string{})

	command.DefineHelp(&helpCmd{})
	command.ParseAndRun()
}

type helpCmd struct {
	args []string
}

func (cmd *helpCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *helpCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	drive.ShowDescriptions(args...)
	exitWithError(nil)
}

type featuresCmd struct{}

func (cmd *featuresCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *featuresCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	context, path := discoverContext(args)

	exitWithError(drive.New(context, &drive.Options{
		Path: path,
	}).About(drive.AboutFeatures))
}

type versionCmd struct{}

func (cmd *versionCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *versionCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	drive.StdoutPrintf("drive version: %s\n%s\n", drive.Version, generated.PkgInfo)
	exitWithError(nil)
}

type initCmd struct{}

func (cmd *initCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *initCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	exitWithError(drive.New(initContext(args), nil).Init())
}

type deInitCmd struct {
	noPrompt *bool
}

func (cmd *deInitCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noPrompt = fs.Bool(drive.NoPromptKey, false, "disables the prompt")
	return fs
}

func (cmd *deInitCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	_, context, path := preprocessArgsByToggle(args, true)
	opts := &drive.Options{
		NoPrompt: *cmd.noPrompt,
		Path:     path,
	}

	exitWithError(drive.New(context, opts).DeInit())
}

type quotaCmd struct{}

func (cmd *quotaCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *quotaCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	context, path := discoverContext(args)

	exitWithError(drive.New(context, &drive.Options{
		Path: path,
	}).About(drive.AboutQuota))
}

type openCmd struct {
	ById        *bool `json:"by-id"`
	FileBrowser *bool `json:"file-browser"`
	WebBrowser  *bool `json:"web-browser"`
}

func (cmd *openCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "open by id instead of path")
	cmd.FileBrowser = fs.Bool(drive.CLIOptionFileBrowser, true, "open file with the local file manager")
	cmd.WebBrowser = fs.Bool(drive.CLIOptionWebBrowser, true, "open file in default browser")
	return fs
}

func (cmd *openCmd) Run(args []string, definedArgs map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	opts := drive.Options{
		Path:    path,
		Sources: sources,
	}

	openType := drive.OpenNone

	if *cmd.ById {
		openType |= drive.IdOpen
	}
	if *cmd.WebBrowser {
		openType |= drive.BrowserOpen
	}
	if *cmd.FileBrowser {
		openType |= drive.FileManagerOpen
	}

	exitWithError(drive.New(context, &opts).Open(openType))
}

type editDescriptionCmd struct {
	ById        *bool   `json:"by-id"`
	Description *string `json:"description"`
	Piped       *bool   `json:"piped"`
}

func (cmd *editDescriptionCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "open by id instead of path")
	cmd.Description = fs.String(drive.CLIOptionDescription, "", drive.DescDescription)
	cmd.Piped = fs.Bool(drive.CLIOptionPiped, false, drive.DescPiped)

	return fs
}

func (cmd *editDescriptionCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	meta := map[string][]string{
		drive.EditDescriptionKey: []string{*cmd.Description},
	}

	if *cmd.Piped {
		meta[drive.PipedKey] = []string{fmt.Sprintf("%v", *cmd.Piped)}
	}

	opts := drive.Options{
		Meta:    &meta,
		Path:    path,
		Sources: sources,
	}

	exitWithError(drive.New(context, &opts).EditDescription(*cmd.ById))
}

type urlCmd struct {
	ById *bool `json:"by-id"`
}

func (cmd *urlCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "resolve url by id instead of path")
	return fs
}

func (cmd *urlCmd) Run(args []string, definedArgs map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	opts := drive.Options{
		Path:    path,
		Sources: sources,
	}

	exitWithError(drive.New(context, &opts).Url(*cmd.ById))
}

type listCmd struct {
	ById         *bool   `json:"by-id"`
	Hidden       *bool   `json:"hidden"`
	Recursive    *bool   `json:"recursive"`
	Files        *bool   `json:"files"`
	Directories  *bool   `json:"directories"`
	Depth        *int    `json:"depth"`
	PageSize     *int64  `json:"page-size"`
	LongFmt      *bool   `json:"long"`
	NoPrompt     *bool   `json:"no-prompt"`
	Shared       *bool   `json:"shared"`
	InTrash      *bool   `json:"trashed"`
	Version      *bool   `json:"version"`
	Matches      *bool   `json:"matches"`
	Owners       *bool   `json:"owners"`
	Quiet        *bool   `json:"quiet"`
	SkipMimeKey  *string `json:"skip-mime"`
	MatchMimeKey *string `json:"match-mime"`
	ExactTitle   *string `json:"exact-title"`
	MatchOwner   *string `json:"match-owner"`
	ExactOwner   *string `json:"exact-owner"`
	NotOwner     *string `json:"not-owner"`
	Sort         *string `json:"sort"`
}

func (cmd *listCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Depth = fs.Int(drive.DepthKey, 1, "maximum recursion depth")
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "list all paths even hidden ones")
	cmd.Files = fs.Bool(drive.CLIOptionFiles, false, "list only files")
	cmd.Directories = fs.Bool(drive.CLIOptionDirectories, false, "list all directories")
	cmd.LongFmt = fs.Bool(drive.CLIOptionLongFmt, false, "long listing of contents")
	cmd.PageSize = fs.Int64(drive.PageSizeKey, 100, "number of results per pagination")
	cmd.Shared = fs.Bool("shared", false, "show files that are shared with me")
	cmd.InTrash = fs.Bool(drive.TrashedKey, false, "list content in the trash")
	cmd.Version = fs.Bool("version", false, "show the number of times that the file has been modified on \n\t\tthe server even with changes not visible to the user")
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before pagination")
	cmd.Owners = fs.Bool("owners", false, "shows the owner names per file")
	cmd.Recursive = fs.Bool(drive.RecursiveKey, false, "recursively list subdirectories")
	cmd.Sort = fs.String(drive.SortKey, "", drive.DescSort)
	cmd.Matches = fs.Bool(drive.MatchesKey, false, "list by prefix")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.SkipMimeKey = fs.String(drive.CLIOptionSkipMime, "", drive.DescSkipMime)
	cmd.MatchMimeKey = fs.String(drive.CLIOptionMatchMime, "", drive.DescMatchMime)
	cmd.ExactTitle = fs.String(drive.CLIOptionExactTitle, "", drive.DescExactTitle)
	cmd.MatchOwner = fs.String(drive.CLIOptionMatchOwner, "", drive.DescMatchOwner)
	cmd.ExactOwner = fs.String(drive.CLIOptionExactOwner, "", drive.DescExactOwner)
	cmd.NotOwner = fs.String(drive.CLIOptionNotOwner, "", drive.DescNotOwner)
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "list by id instead of path")

	return fs
}

func (lCmd *listCmd) _run(args []string, definedFlags map[string]*flag.Flag, diskUsageSubset bool) error {
	sources, context, path := preprocessArgsByToggle(args, (*lCmd.ById || *lCmd.Matches))
	cmd := listCmd{}
	df := defaultsFiller{
		from: *lCmd, to: &cmd,
		rcSourcePath: context.AbsPathOf(path),
		definedFlags: definedFlags,
	}

	if err := fillWithDefaults(df); err != nil {
		exitWithError(err)
	}

	typeMask := 0
	if *cmd.Directories {
		typeMask |= drive.Folder
	}
	if *cmd.Shared {
		typeMask |= drive.Shared
	}
	if *cmd.Owners {
		typeMask |= drive.Owners
	}
	if *cmd.Version {
		typeMask |= drive.CurrentVersion
	}
	if *cmd.Files {
		typeMask |= drive.NonFolder
	}
	if *cmd.InTrash {
		typeMask |= drive.InTrash
	}

	if diskUsageSubset {
		typeMask |= drive.DiskUsageOnly
	}

	if !*cmd.LongFmt {
		typeMask |= drive.Minimal
	}

	depth := *cmd.Depth
	if *cmd.Recursive {
		depth = drive.InfiniteDepth
	}

	meta := map[string][]string{
		drive.SortKey:         drive.NonEmptyTrimmedStrings(*cmd.Sort),
		drive.SkipMimeKeyKey:  drive.NonEmptyTrimmedStrings(strings.Split(*cmd.SkipMimeKey, ",")...),
		drive.MatchMimeKeyKey: drive.NonEmptyTrimmedStrings(strings.Split(*cmd.MatchMimeKey, ",")...),
		drive.ExactTitleKey:   drive.NonEmptyTrimmedStrings(strings.Split(*cmd.ExactTitle, ",")...),
		drive.MatchOwnerKey:   drive.NonEmptyTrimmedStrings(strings.Split(*cmd.MatchOwner, ",")...),
		drive.ExactOwnerKey:   drive.NonEmptyTrimmedStrings(strings.Split(*cmd.ExactOwner, ",")...),
		drive.NotOwnerKey:     drive.NonEmptyTrimmedStrings(strings.Split(*cmd.NotOwner, ",")...),
	}

	opts := &drive.Options{
		Path:      path,
		Sources:   sources,
		Depth:     depth,
		Hidden:    *cmd.Hidden,
		InTrash:   *cmd.InTrash,
		PageSize:  *cmd.PageSize,
		NoPrompt:  *cmd.NoPrompt,
		Recursive: *cmd.Recursive,
		TypeMask:  typeMask,
		Quiet:     *cmd.Quiet,
		Meta:      &meta,
		Match:     *cmd.Matches,
	}

	if *cmd.Shared {
		return drive.New(context, opts).ListShared()
	} else if *cmd.Matches {
		return drive.New(context, opts).ListMatches()
	} else {
		return drive.New(context, opts).List(*cmd.ById)
	}

	return nil
}

type duCmd struct {
	listCmd
}

func (lCmd *listCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	exitWithError(lCmd._run(args, definedFlags, false))
}

func (dCmd *duCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	exitWithError(dCmd._run(args, definedFlags, true))
}

type md5SumCmd struct {
	ById      *bool `json:"by-id"`
	Depth     *int  `json:"depth"`
	Hidden    *bool `json:"hidden"`
	Recursive *bool `json:"recursive"`
	Quiet     *bool `json:"quiet"`
}

func (cmd *md5SumCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Depth = fs.Int(drive.DepthKey, 1, "max traversal depth")
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "discover hidden paths")
	cmd.Recursive = fs.Bool(drive.RecursiveKey, false, "recursively discover folders")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "stat by id instead of path")
	return fs
}

func (cmd *md5SumCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	depth := *cmd.Depth
	if *cmd.Recursive {
		depth = drive.InfiniteDepth
	}

	opts := drive.Options{
		Path:      path,
		Sources:   sources,
		Depth:     depth,
		Hidden:    *cmd.Hidden,
		Recursive: *cmd.Recursive,
		Quiet:     *cmd.Quiet,
		Md5sum:    true,
	}

	if *cmd.ById {
		exitWithError(drive.New(context, &opts).StatById())
	} else {
		exitWithError(drive.New(context, &opts).Stat())
	}
}

type statCmd struct {
	ById      *bool `json:"by-id"`
	Depth     *int  `json:"depth"`
	Hidden    *bool `json:"hidden"`
	Recursive *bool `json:"recursive"`
	Quiet     *bool `json:"quiet"`
	Md5sum    *bool `json:"md5sum"`
}

func (cmd *statCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Depth = fs.Int(drive.DepthKey, 1, "max traversal depth")
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "discover hidden paths")
	cmd.Recursive = fs.Bool(drive.RecursiveKey, false, "recursively discover folders")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "stat by id instead of path")
	cmd.Md5sum = fs.Bool(drive.Md5sumKey, false, "produce output compatible with md5sum(1)")
	return fs
}

func (cmd *statCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	depth := *cmd.Depth
	if *cmd.Recursive {
		depth = drive.InfiniteDepth
	}

	opts := drive.Options{
		Depth:     depth,
		Path:      path,
		Sources:   sources,
		Hidden:    *cmd.Hidden,
		Recursive: *cmd.Recursive,
		Quiet:     *cmd.Quiet,
		Md5sum:    *cmd.Md5sum,
	}

	if *cmd.ById {
		exitWithError(drive.New(context, &opts).StatById())
	} else {
		exitWithError(drive.New(context, &opts).Stat())
	}
}

type indexCmd struct {
	ById              *bool   `json:"by-id"`
	IgnoreConflict    *bool   `json:"ignore-conflict"`
	Recursive         *bool   `json:"recursive"`
	NoPrompt          *bool   `json:"no-prompt"`
	Hidden            *bool   `json:"hidden"`
	Force             *bool   `json:"force"`
	IgnoreNameClashes *bool   `json:"ignore-name-clashes"`
	Quiet             *bool   `json:"quiet"`
	ExcludeOps        *string `json:"exclude-ops"`
	SkipMimeKey       *string `json:"skip-mime"`
	IgnoreChecksum    *bool   `json:"ignore-checksum"`
	NoClobber         *bool   `json:"no-clobber"`
	Prune             *bool   `json:"prune"`
	AllOps            *bool   `json:"all-ops"`
	Matches           *bool   `json:"matches"`
}

func (cmd *indexCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "fetch by id instead of path")
	cmd.IgnoreConflict = fs.Bool(drive.CLIOptionIgnoreConflict, true, drive.DescIgnoreConflict)
	cmd.Recursive = fs.Bool(drive.RecursiveKey, true, "fetch recursively for children")
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before applying the fetch action")
	cmd.Hidden = fs.Bool(drive.HiddenKey, true, "allows fetching of hidden paths")
	cmd.Force = fs.Bool(drive.ForceKey, false, "forces a fetch even if no changes present")
	cmd.IgnoreNameClashes = fs.Bool(drive.CLIOptionIgnoreNameClashes, true, drive.DescIgnoreNameClashes)
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ExcludeOps = fs.String(drive.CLIOptionExcludeOperations, "", drive.DescExcludeOps)
	cmd.SkipMimeKey = fs.String(drive.CLIOptionSkipMime, "", drive.DescSkipMime)
	cmd.IgnoreChecksum = fs.Bool(drive.CLIOptionIgnoreChecksum, true, drive.DescIgnoreChecksum)
	cmd.NoClobber = fs.Bool(drive.CLIOptionNoClobber, false, "prevents overwriting of old content")
	cmd.Prune = fs.Bool(drive.CLIOptionPruneIndices, false, drive.DescPruneIndices)
	cmd.AllOps = fs.Bool(drive.CLIOptionAllIndexOperations, false, drive.DescAllIndexOperations)
	cmd.Matches = fs.Bool(drive.MatchesKey, false, "search by prefix")

	return fs
}

func (cmd *indexCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	byId := *cmd.ById
	byMatches := *cmd.Matches
	sources, context, path := preprocessArgsByToggle(args, byMatches || byId)

	options := &drive.Options{
		Path:              path,
		Sources:           sources,
		Hidden:            *cmd.Hidden,
		IgnoreChecksum:    *cmd.IgnoreChecksum,
		IgnoreConflict:    *cmd.IgnoreConflict,
		NoPrompt:          *cmd.NoPrompt,
		NoClobber:         *cmd.NoClobber,
		Recursive:         *cmd.Recursive,
		Quiet:             *cmd.Quiet,
		Force:             *cmd.Force,
		IgnoreNameClashes: *cmd.IgnoreNameClashes,
		Match:             *cmd.Matches,
	}

	dr := drive.New(context, options)

	fetchFn := dr.Fetch
	if byId {
		fetchFn = dr.FetchById
	} else if *cmd.Matches {
		fetchFn = dr.FetchMatches
	}

	scheduling := []errorer{}
	if *cmd.AllOps {
		scheduling = append(scheduling, dr.Prune, fetchFn)
	} else if *cmd.Prune {
		scheduling = append(scheduling, dr.Prune)
	} else {
		scheduling = append(scheduling, fetchFn)
	}

	for _, fn := range scheduling {
		exitWithError(fn())
	}
}

type pullCmd struct {
	ById              *bool   `json:"by-id"`
	ExportsDir        *string `json:"exports-dir"`
	Export            *string `json:"export"`
	ExcludeOps        *string `json:"exclude-ops"`
	Force             *bool   `json:"force"`
	Hidden            *bool   `json:"hidden"`
	Matches           *bool   `json:"matches"`
	NoPrompt          *bool   `json:"no-prompt"`
	NoClobber         *bool   `json:"no-clobber"`
	Recursive         *bool   `json:"recursive"`
	IgnoreChecksum    *bool   `json:"ignore-checksum"`
	IgnoreConflict    *bool   `json:"ignore-conflict"`
	Piped             *bool   `json:"piped"`
	Quiet             *bool   `json:"quiet"`
	IgnoreNameClashes *bool   `json:"ignore-name-clashes"`
	SkipMimeKey       *string `json:"skip-mime"`
	ExplicitlyExport  *bool   `json:"explicitly-export"`
	FixClashes        *bool   `json:"fix-clashes"`

	Verbose    *bool `json:"verbose"`
	Depth      *int  `json:"depth"`
	Starred    *bool `json:"starred"`
	AllStarred *bool `json:"all-starred"`
	InTrash    *bool `json:"trashed"`
}

func (cmd *pullCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.AllStarred = fs.Bool(drive.CLIOptionAllStarred, false, drive.DescAllStarred)
	cmd.NoClobber = fs.Bool(drive.CLIOptionNoClobber, false, "prevents overwriting of old content")
	cmd.Export = fs.String(
		drive.ExportsKey, "", "comma separated list of formats to export your docs + sheets files")
	cmd.Recursive = fs.Bool(drive.RecursiveKey, true, "performs the pull action recursively")
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before applying the pull action")
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows pulling of hidden paths")
	cmd.Force = fs.Bool(drive.ForceKey, false, "forces a pull even if no changes present")
	cmd.IgnoreChecksum = fs.Bool(drive.CLIOptionIgnoreChecksum, true, drive.DescIgnoreChecksum)
	cmd.IgnoreConflict = fs.Bool(drive.CLIOptionIgnoreConflict, false, drive.DescIgnoreConflict)
	cmd.IgnoreNameClashes = fs.Bool(drive.CLIOptionIgnoreNameClashes, false, drive.DescIgnoreNameClashes)
	cmd.ExportsDir = fs.String(drive.ExportsDirKey, "", "directory to place exports")
	cmd.Matches = fs.Bool(drive.MatchesKey, false, "search by prefix")
	cmd.Piped = fs.Bool(drive.CLIOptionPiped, false, drive.DescPiped)
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ExcludeOps = fs.String(drive.CLIOptionExcludeOperations, "", drive.DescExcludeOps)
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "pull by id instead of path")
	cmd.SkipMimeKey = fs.String(drive.CLIOptionSkipMime, "", drive.DescSkipMime)
	cmd.ExplicitlyExport = fs.Bool(drive.CLIOptionExplicitlyExport, false, drive.DescExplicitylPullExports)
	cmd.Verbose = fs.Bool(drive.CLIOptionVerboseKey, false, drive.DescVerbose)
	cmd.Depth = fs.Int(drive.DepthKey, drive.DefaultMaxTraversalDepth, "max traversal depth")
	cmd.FixClashes = fs.Bool(drive.CLIOptionFixClashesKey, false, drive.DescFixClashes)
	cmd.Starred = fs.Bool(drive.CLIOptionStarred, false, drive.DescStarred)
	cmd.InTrash = fs.Bool(drive.TrashedKey, false, "pull content in the trash")

	return fs
}

func (pCmd *pullCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, (*pCmd.ById || *pCmd.Matches || *pCmd.Starred))
	cmd := pullCmd{}
	df := defaultsFiller{
		from: *pCmd, to: &cmd,
		rcSourcePath: context.AbsPathOf(path),
		definedFlags: definedFlags,
	}

	if err := fillWithDefaults(df); err != nil {
		exitWithError(err)
	}

	excludes := drive.NonEmptyTrimmedStrings(strings.Split(*cmd.ExcludeOps, ",")...)
	excludeCrudMask := drive.CrudAtoi(excludes...)
	if excludeCrudMask == drive.AllCrudOperations {
		exitWithError(fmt.Errorf("all CRUD operations forbidden"))
	}

	meta := map[string][]string{
		drive.SkipMimeKeyKey: drive.NonEmptyTrimmedStrings(strings.Split(*cmd.SkipMimeKey, ",")...),
	}

	// Filter out empty strings.
	exports := drive.NonEmptyTrimmedStrings(strings.Split(*cmd.Export, ",")...)

	options := &drive.Options{
		Path:              path,
		Sources:           sources,
		Exports:           uniqOrderedStr(exports),
		ExportsDir:        strings.Trim(*cmd.ExportsDir, " "),
		Force:             *cmd.Force,
		Hidden:            *cmd.Hidden,
		IgnoreChecksum:    *cmd.IgnoreChecksum,
		IgnoreConflict:    *cmd.IgnoreConflict,
		NoPrompt:          *cmd.NoPrompt,
		NoClobber:         *cmd.NoClobber,
		Recursive:         *cmd.Recursive,
		Piped:             *cmd.Piped,
		Quiet:             *cmd.Quiet,
		IgnoreNameClashes: *cmd.IgnoreNameClashes,
		ExcludeCrudMask:   excludeCrudMask,
		ExplicitlyExport:  *cmd.ExplicitlyExport,
		Meta:              &meta,
		Verbose:           *cmd.Verbose,
		Depth:             *cmd.Depth,
		FixClashes:        *cmd.FixClashes,
		Starred:           *cmd.Starred,
		Match:             *cmd.Matches,
		InTrash:           *cmd.InTrash,
	}

	if *cmd.Matches || *cmd.Starred {
		if *cmd.AllStarred {
			exitWithError(drive.New(context, options).PullAllStarred())
		} else {
			exitWithError(drive.New(context, options).PullMatchLike())
		}
	} else if *cmd.Piped {
		exitWithError(drive.New(context, options).PullPiped(*cmd.ById))
	} else if *cmd.ById {
		exitWithError(drive.New(context, options).PullById())
	} else {
		exitWithError(drive.New(context, options).Pull())
	}
}

type pushCmd struct {
	NoClobber   *bool `json:"no-clobber"`
	Hidden      *bool `json:"hidden"`
	Force       *bool `json:"force"`
	NoPrompt    *bool `json:"no-prompt"`
	Recursive   *bool `json:"recursive"`
	Piped       *bool `json:"piped"`
	MountedPush *bool `json:"m"`
	// convert when set tells Google drive to convert the document into
	// its appropriate Google Docs format
	Convert *bool `json:"convert"`
	// ocr when set indicates that Optical Character Recognition should be
	// attempted on .[gif, jpg, pdf, png] uploads
	Ocr               *bool   `json:"ocr"`
	IgnoreChecksum    *bool   `json:"ignore-checksum"`
	IgnoreConflict    *bool   `json:"ignore-conflict"`
	IgnoreNameClashes *bool   `json:"ignore-name-clashes"`
	Quiet             *bool   `json:"quiet"`
	CoercedMimeKey    *string `json:"coerced-mime"`
	ExcludeOps        *string `json:"exclude-ops"`
	SkipMimeKey       *string `json:"skip-mime"`
	Verbose           *bool   `json:"verbose"`
	Depth             *int    `json:"depth"`
	FixClashes        *bool   `json:"fix-clashes"`
}

func (cmd *pushCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.NoClobber = fs.Bool(drive.CLIOptionNoClobber, false, "allows overwriting of old content")
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows pushing of hidden paths")
	cmd.Recursive = fs.Bool(drive.RecursiveKey, true, "performs the push action recursively")
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before applying the push action")
	cmd.Force = fs.Bool(drive.ForceKey, false, "forces a push even if no changes present")
	cmd.MountedPush = fs.Bool("m", false, "allows pushing of mounted paths")
	cmd.Convert = fs.Bool(drive.ConvertKey, false, "toggles conversion of the file to its appropriate Google Doc format")
	cmd.Ocr = fs.Bool(drive.OcrKey, false, "if true, attempt OCR on gif, jpg, pdf and png uploads")
	cmd.Piped = fs.Bool(drive.CLIOptionPiped, false, drive.DescPiped)
	cmd.IgnoreChecksum = fs.Bool(drive.CLIOptionIgnoreChecksum, true, drive.DescIgnoreChecksum)
	cmd.IgnoreConflict = fs.Bool(drive.CLIOptionIgnoreConflict, false, drive.DescIgnoreConflict)
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.CoercedMimeKey = fs.String(drive.CoercedMimeKeyKey, "", "the mimeType you are trying to coerce this file to be")
	cmd.IgnoreNameClashes = fs.Bool(drive.CLIOptionIgnoreNameClashes, false, drive.DescIgnoreNameClashes)
	cmd.ExcludeOps = fs.String(drive.CLIOptionExcludeOperations, "", drive.DescExcludeOps)
	cmd.SkipMimeKey = fs.String(drive.CLIOptionSkipMime, "", drive.DescSkipMime)
	cmd.Verbose = fs.Bool(drive.CLIOptionVerboseKey, false, drive.DescVerbose)
	cmd.Depth = fs.Int(drive.DepthKey, drive.DefaultMaxTraversalDepth, "max traversal depth")
	cmd.FixClashes = fs.Bool(drive.CLIOptionFixClashesKey, false, drive.DescFixClashes)
	return fs
}

func (cmd *pushCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	if *cmd.MountedPush {
		cmd.pushMounted(args, definedFlags)
	} else {
		sources, context, path := preprocessArgs(args)

		options, err := cmd.createPushOptions(context.AbsPathOf(path), definedFlags)
		if err != nil {
			exitWithError(err)
		}

		options.Path = path
		options.Sources = sources

		if *cmd.Piped {
			exitWithError(drive.New(context, options).PushPiped())
		} else {
			exitWithError(drive.New(context, options).Push())
		}
	}
}

type qrLinkCmd struct {
	Address *string `json:"address"`
	ById    *bool   `json:"by-id"`
	Verbose *bool   `json:"verbose"`
}

func (cmd *qrLinkCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Address = fs.String(drive.AddressKey, drive.DefaultQRShareServer, "address on which the QR code generator is running")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "share by id instead of path")
	cmd.Verbose = fs.Bool(drive.CLIOptionVerboseKey, true, drive.DescVerbose)
	return fs
}

func (qCmd *qrLinkCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *qCmd.ById)

	cmd := qrLinkCmd{}
	df := defaultsFiller{
		from: *qCmd, to: &cmd,
		rcSourcePath: path,
		definedFlags: definedFlags,
	}

	if err := fillWithDefaults(df); err != nil {
		exitWithError(err)
	}

	meta := map[string][]string{
		drive.AddressKey: []string{*cmd.Address},
	}

	opts := drive.Options{
		Path:    path,
		Sources: sources,
		Meta:    &meta,
		Verbose: *cmd.Verbose,
	}

	exitWithError(drive.New(context, &opts).QR(*cmd.ById))
}

type touchCmd struct {
	ById      *bool `json:"by-id"`
	Depth     *int  `json:"depth"`
	Hidden    *bool `json:"hidden"`
	Recursive *bool `json:"recursive"`
	Matches   *bool `json:"matches"`
	Quiet     *bool `json:"quiet"`
	Verbose   *bool `json:"verbose"`
}

func (cmd *touchCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows pushing of hidden paths")
	cmd.Recursive = fs.Bool(drive.RecursiveKey, false, "toggles recursive touching")
	cmd.Matches = fs.Bool(drive.MatchesKey, false, "search by prefix and touch")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "share by id instead of path")
	cmd.Depth = fs.Int(drive.DepthKey, drive.DefaultMaxTraversalDepth, "max traversal depth")
	cmd.Verbose = fs.Bool(drive.CLIOptionVerboseKey, true, drive.DescVerbose)

	return fs
}

func (cmd *touchCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.Matches || *cmd.ById)

	depth := *cmd.Depth
	if *cmd.Recursive {
		depth = drive.InfiniteDepth
	}

	opts := drive.Options{
		Path:      path,
		Sources:   sources,
		Hidden:    *cmd.Hidden,
		Recursive: *cmd.Recursive,
		Depth:     depth,
		Quiet:     *cmd.Quiet,
		Match:     *cmd.Matches,
		Verbose:   *cmd.Verbose,
	}

	if *cmd.Matches {
		exitWithError(drive.New(context, &opts).TouchByMatch())
	} else {
		exitWithError(drive.New(context, &opts).Touch(*cmd.ById))
	}
}

func (pCmd *pushCmd) createPushOptions(absEntryPath string, definedFlags map[string]*flag.Flag) (*drive.Options, error) {
	cmd := pushCmd{}
	df := defaultsFiller{
		from: *pCmd, to: &cmd,
		rcSourcePath: absEntryPath,
		definedFlags: definedFlags,
	}

	if err := fillWithDefaults(df); err != nil {
		exitWithError(err)
	}

	mask := drive.OptNone
	if *cmd.Convert {
		mask |= drive.OptConvert
	}
	if *cmd.Ocr {
		mask |= drive.OptOCR
	}

	meta := map[string][]string{
		drive.CoercedMimeKeyKey: drive.NonEmptyTrimmedStrings(*cmd.CoercedMimeKey),
		drive.SkipMimeKeyKey:    drive.NonEmptyTrimmedStrings(strings.Split(*cmd.SkipMimeKey, ",")...),
	}

	excludes := drive.NonEmptyTrimmedStrings(strings.Split(*cmd.ExcludeOps, ",")...)
	excludeCrudMask := drive.CrudAtoi(excludes...)
	if excludeCrudMask == drive.AllCrudOperations {
		exitWithError(fmt.Errorf("all CRUD operations forbidden yet asking to push"))
	}

	opts := &drive.Options{
		Force:             *cmd.Force,
		Hidden:            *cmd.Hidden,
		IgnoreChecksum:    *cmd.IgnoreChecksum,
		IgnoreConflict:    *cmd.IgnoreConflict,
		NoClobber:         *cmd.NoClobber,
		NoPrompt:          *cmd.NoPrompt,
		Recursive:         *cmd.Recursive,
		Piped:             *cmd.Piped,
		Quiet:             *cmd.Quiet,
		Meta:              &meta,
		TypeMask:          mask,
		ExcludeCrudMask:   excludeCrudMask,
		IgnoreNameClashes: *cmd.IgnoreNameClashes,
		Verbose:           *cmd.Verbose,
		Depth:             *cmd.Depth,
		FixClashes:        *cmd.FixClashes,
	}

	return opts, nil
}

func (cmd *pushCmd) pushMounted(args []string, definedFlags map[string]*flag.Flag) {
	argc := len(args)

	var err error
	var contextArgs, rest, sources []string

	if !*cmd.MountedPush {
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

	if path == "." {
		path = ""
	}

	root := context.AbsPathOf("")
	contextAbsPath := filepath.Join(root, path)

	mount, auxSrcs := config.MountPoints(path, contextAbsPath, rest, *cmd.Hidden)

	sources, err = relativePathsOpt(root, auxSrcs, true)
	exitWithError(err)

	options, err := cmd.createPushOptions(path, definedFlags)
	if err != nil {
		exitWithError(err)
	}

	options.Path = path
	options.Mount = mount
	options.Sources = sources

	exitWithError(drive.New(context, options).Push())
}

type aboutCmd struct {
	Features *bool `json:"features"`
	Filesize *bool `json:"filesize"`
	Quiet    *bool `json:"quiet"`
	Quota    *bool `json:"quota"`
}

func (cmd *aboutCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Features = fs.Bool("features", false, "gives information on features present on this drive")
	cmd.Filesize = fs.Bool("filesize", false, "prints out information about file sizes e.g the max upload size for a specific file size")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.Quota = fs.Bool("quota", false, "prints out quota information for this drive")
	return fs
}

func (cmd *aboutCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	_, context, _ := preprocessArgs(args)

	mask := drive.AboutNone
	if *cmd.Features {
		mask |= drive.AboutFeatures
	}
	if *cmd.Quota {
		mask |= drive.AboutQuota
	}
	if *cmd.Filesize {
		mask |= drive.AboutFileSizes
	}

	if mask == drive.AboutNone { // No option set
		mask = drive.AboutQuota | drive.AboutFeatures | drive.AboutFileSizes
	}

	exitWithError(drive.New(context, &drive.Options{
		Quiet: *cmd.Quiet,
	}).About(mask))
}

type diffCmd struct {
	Hidden            *bool `json:"hidden"`
	IgnoreConflict    *bool `json:"ignore-conflict"`
	IgnoreChecksum    *bool `json:"ignore-checksum"`
	IgnoreNameClashes *bool `json:"ignore-name-clashes"`
	Quiet             *bool `json:"quiet"`
	Depth             *int  `json:"depth"`
	Recursive         *bool `json:"recursive"`
	Unified           *bool `json:"unified"`
	BaseLocal         *bool `json:"base-local"`
}

func (cmd *diffCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows pulling of hidden paths")
	cmd.IgnoreChecksum = fs.Bool(drive.CLIOptionIgnoreChecksum, true, drive.DescIgnoreChecksum)
	cmd.IgnoreConflict = fs.Bool(drive.CLIOptionIgnoreConflict, false, drive.DescIgnoreConflict)
	cmd.IgnoreNameClashes = fs.Bool(drive.CLIOptionIgnoreNameClashes, false, drive.DescIgnoreNameClashes)
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.Depth = fs.Int(drive.DepthKey, drive.DefaultMaxTraversalDepth, "max traversal depth")
	cmd.Recursive = fs.Bool(drive.RecursiveKey, true, "recursively diff")
	cmd.Unified = fs.Bool(drive.CLIOptionUnifiedShortKey, true, drive.DescUnifiedDiff)
	cmd.BaseLocal = fs.Bool(drive.CLIOptionDiffBaseLocal, true, drive.DescDiffBaseLocal)

	return fs
}

func (cmd *diffCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgs(args)

	mask := drive.DiffNone
	if *cmd.Unified {
		mask |= drive.DiffUnified
	}

	exitWithError(drive.New(context, &drive.Options{
		Path:              path,
		Sources:           sources,
		Hidden:            *cmd.Hidden,
		Recursive:         *cmd.Recursive,
		IgnoreChecksum:    *cmd.IgnoreChecksum,
		IgnoreNameClashes: *cmd.IgnoreNameClashes,
		IgnoreConflict:    *cmd.IgnoreConflict,
		Quiet:             *cmd.Quiet,
		Depth:             *cmd.Depth,
		BaseLocal:         *cmd.BaseLocal,
		TypeMask:          mask,
	}).Diff())
}

type unpublishCmd struct {
	Hidden *bool `json:"hidden"`
	Quiet  *bool `json:"quiet"`
	ById   *bool `json:"by-id"`
}

func (cmd *unpublishCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows pulling of hidden paths")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "unpublish by id instead of path")
	return fs
}

func (uCmd *unpublishCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *uCmd.ById)

	cmd := unpublishCmd{}
	df := defaultsFiller{
		from: *uCmd, to: &cmd,
		rcSourcePath: context.AbsPathOf(path),
		definedFlags: definedFlags,
	}

	if err := fillWithDefaults(df); err != nil {
		exitWithError(err)
	}

	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.Quiet,
	}).Unpublish(*cmd.ById))
}

type emptyTrashCmd struct {
	NoPrompt *bool `json:"no-prompt"`
	Quiet    *bool `json:"quiet"`
}

func (cmd *emptyTrashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "shows no prompt before emptying the trash")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	return fs
}

func (cmd *emptyTrashCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	_, context, _ := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		NoPrompt: *cmd.NoPrompt,
		Quiet:    *cmd.Quiet,
	}).EmptyTrash())
}

type deleteCmd struct {
	Hidden   *bool `json:"hidden"`
	Matches  *bool `json:"matches"`
	Quiet    *bool `json:"quiet"`
	ById     *bool `json:"by-id"`
	NoPrompt *bool `json:"no-prompt"`
}

func (cmd *deleteCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows trashing hidden paths")
	cmd.Matches = fs.Bool(drive.MatchesKey, false, "search by prefix and delete")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "delete by id instead of path")
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "disables the prompt")

	return fs
}

func (cmd *deleteCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	if *cmd.NoPrompt {
		exitWithError(drive.PermanentDeletionNoPromptError)
	}

	sources, context, path := preprocessArgsByToggle(args, *cmd.Matches || *cmd.ById)

	opts := drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.Quiet,
		Match:   *cmd.Matches,
	}

	if !*cmd.Matches {
		exitWithError(drive.New(context, &opts).Delete(*cmd.ById))
	} else {
		exitWithError(drive.New(context, &opts).DeleteByMatch())
	}
}

type trashCmd struct {
	Hidden  *bool `json:"hidden"`
	Matches *bool `json:"matches"`
	Quiet   *bool `json:"quiet"`
	ById    *bool `json:"by-id"`
}

func (cmd *trashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows trashing hidden paths")
	cmd.Matches = fs.Bool(drive.MatchesKey, false, "search by prefix and trash")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "trash by id instead of path")

	return fs
}

func (cmd *trashCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.Matches || *cmd.ById)

	opts := drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.Quiet,
		Match:   *cmd.Matches,
	}

	if !*cmd.Matches {
		exitWithError(drive.New(context, &opts).Trash(*cmd.ById))
	} else {
		exitWithError(drive.New(context, &opts).TrashByMatch())
	}
}

type newCmd struct {
	Folder  *bool   `json:"folder"`
	MimeKey *string `json:"mime"`
}

func (cmd *newCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Folder = fs.Bool("folder", false, "create a folder if set otherwise create a regular file")
	cmd.MimeKey = fs.String(drive.MimeKey, "", "coerce the file to this mimeType")
	return fs
}

func (cmd *newCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgs(args)

	opts := drive.Options{
		Path:    path,
		Sources: sources,
	}

	meta := map[string][]string{
		drive.MimeKey: drive.NonEmptyTrimmedStrings(strings.Split(*cmd.MimeKey, ",")...),
	}

	opts.Meta = &meta

	if *cmd.Folder {
		exitWithError(drive.New(context, &opts).NewFolder())
	} else {
		exitWithError(drive.New(context, &opts).NewFile())
	}
}

type copyCmd struct {
	Quiet     *bool `json:"quiet"`
	Recursive *bool `json:"recursive"`
	ById      *bool `json:"by-id"`
}

func (cmd *copyCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Recursive = fs.Bool(drive.RecursiveKey, false, "recursive copying")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "copy by id instead of path")
	return fs
}

func (cmd *copyCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	if len(args) < 2 {
		args = append(args, ".")
	}

	end := len(args) - 1
	if end < 1 {
		exitWithError(fmt.Errorf("copy: expected more than one path"))
	}

	dest := args[end]

	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	// Unshift by the end path
	sources = sources[:len(sources)-1]
	destRels, err := relativePaths(context.AbsPathOf(""), dest)
	exitWithError(err)

	dest = destRels[0]
	sources = append(sources, dest)

	exitWithError(drive.New(context, &drive.Options{
		Path:      path,
		Sources:   sources,
		Recursive: *cmd.Recursive,
		Quiet:     *cmd.Quiet,
	}).Copy(*cmd.ById))
}

type untrashCmd struct {
	Hidden  *bool `json:"hidden"`
	Matches *bool `json:"matches"`
	Quiet   *bool `json:"quiet"`
	ById    *bool `json:"by-id"`
}

func (cmd *untrashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows untrashing hidden paths")
	cmd.Matches = fs.Bool(drive.MatchesKey, false, "search by prefix and untrash")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "untrash by id instead of path")

	return fs
}

func (cmd *untrashCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById || *cmd.Matches)

	opts := drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.Quiet,
		Match:   *cmd.Matches,
	}

	if !*cmd.Matches {
		exitWithError(drive.New(context, &opts).Untrash(*cmd.ById))
	} else {
		exitWithError(drive.New(context, &opts).UntrashByMatch())
	}
}

type publishCmd struct {
	Hidden *bool `json:"hidden"`
	Quiet  *bool `json:"quiet"`
	ById   *bool `json:"by-id"`
}

func (cmd *publishCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Hidden = fs.Bool(drive.HiddenKey, false, "allows publishing of hidden paths")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "publish by id instead of path")
	return fs
}

func (cmd *publishCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.Quiet,
	}).Publish(*cmd.ById))
}

type shareCmd struct {
	ById        *bool   `json:"by-id"`
	Emails      *string `json:"emails"`
	Message     *string `json:"message"`
	Role        *string `json:"role"`
	AccountType *string `json:"type"`
	NoPrompt    *bool   `json:"no-prompt"`
	Notify      *bool   `json:"notify"`
	Quiet       *bool   `json:"quiet"`
	Verbose     *bool   `json:"verbose"`
}

func (cmd *shareCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Emails = fs.String(drive.EmailsKey, "", "emails to share the file to")
	cmd.Message = fs.String("message", "", "message to send receipients")
	cmd.Role = fs.String(drive.RoleKey, "", "role to set to receipients of share. Possible values: "+drive.DescRoles)
	cmd.AccountType = fs.String(drive.TypeKey, "", "scope of accounts to share files with. Possible values: "+drive.DescAccountTypes)
	cmd.Notify = fs.Bool(drive.CLIOptionNotify, true, "toggle whether to notify receipients about share")
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "disables the prompt")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "share by id instead of path")
	cmd.Verbose = fs.Bool(drive.CLIOptionVerboseKey, true, drive.DescVerbose)

	return fs
}

func (cmd *shareCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	meta := map[string][]string{
		drive.EmailMessageKey: []string{*cmd.Message},
		drive.EmailsKey:       uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.Emails, ",")...)),
		drive.RoleKey:         uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.Role, ",")...)),
		drive.AccountTypeKey:  uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.AccountType, ",")...)),
	}

	mask := drive.NoopOnShare
	if *cmd.Notify {
		mask = drive.Notify
	}

	exitWithError(drive.New(context, &drive.Options{
		Path:     path,
		Sources:  sources,
		Meta:     &meta,
		TypeMask: mask,
		NoPrompt: *cmd.NoPrompt,
		Quiet:    *cmd.Quiet,
		Verbose:  *cmd.Verbose,
	}).Share(*cmd.ById))
}

type unshareCmd struct {
	Role        *string `json:"role"`
	AccountType *string `json:"type"`
	Quiet       *bool   `json:"quiet"`
	ById        *bool   `json:"by-id"`
	Emails      *string `json:"emails"`
	NoPrompt    *bool   `json:"no-prompt"`
	Verbose     *bool   `json:"verbose"`
}

func (cmd *unshareCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "unshare by id instead of path")
	cmd.Role = fs.String(drive.RoleKey, "", "role to set to receipients of share. Possible values: "+drive.DescRoles)
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.Verbose = fs.Bool(drive.CLIOptionVerboseKey, true, drive.DescVerbose)
	cmd.Emails = fs.String(drive.EmailsKey, "", "emails to share the file to")
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "disables the prompt")
	cmd.AccountType = fs.String(drive.TypeKey, "", "scope of account to revoke access to")

	return fs
}

func (cmd *unshareCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	meta := map[string][]string{
		drive.EmailsKey:      uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.Emails, ",")...)),
		drive.RoleKey:        uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.Role, ",")...)),
		drive.AccountTypeKey: uniqOrderedStr(drive.NonEmptyTrimmedStrings(strings.Split(*cmd.AccountType, ",")...)),
	}

	exitWithError(drive.New(context, &drive.Options{
		Meta:     &meta,
		Path:     path,
		Sources:  sources,
		NoPrompt: *cmd.NoPrompt,
		Quiet:    *cmd.Quiet,
		Verbose:  *cmd.Verbose,
	}).Unshare(*cmd.ById))
}

type moveCmd struct {
	Quiet *bool `json:"quiet"`
	ById  *bool `json:"by-id"`
}

func (cmd *moveCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "move by id instead of path")
	return fs
}

func (cmd *moveCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	argc := len(args)
	if argc < 1 {
		exitWithError(fmt.Errorf("move: expecting a path or more"))
	}
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	// Unshift by the end path
	sources = sources[:len(sources)-1]

	dest := args[argc-1]
	destRels, err := relativePaths(context.AbsPathOf(""), dest)
	exitWithError(err)

	sources = append(sources, destRels[0])

	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.Quiet,
	}).Move(*cmd.ById))
}

type renameCmd struct {
	Force *bool `json:"force"`
	Quiet *bool `json:"quiet"`
	ById  *bool `json:"by-id"`
}

func (cmd *renameCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Force = fs.Bool(drive.ForceKey, false, "coerce rename even if remote already exists")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "unshare by id instead of path")
	return fs
}

func (cmd *renameCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	argc := len(args)
	if argc < 2 {
		exitWithError(fmt.Errorf("rename: expecting <src> <dest>"))
	}
	rest, last := args[:argc-1], args[argc-1]
	sources, context, path := preprocessArgsByToggle(rest, *cmd.ById)

	sources = append(sources, last)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
		Force:   *cmd.Force,
		Quiet:   *cmd.Quiet,
	}).Rename(*cmd.ById))
}

type starCmd struct {
	ById     *bool `json:"by-id"`
	Quiet    *bool `json:"quiet"`
	NoPrompt *bool `json:"no-prompt"`
}

func (cmd *starCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.ById = fs.Bool(drive.CLIOptionId, false, "open by id instead of path")
	cmd.NoPrompt = fs.Bool(drive.NoPromptKey, false, "disables the prompt")
	cmd.Quiet = fs.Bool(drive.QuietKey, false, "if set, do not log anything but errors")
	return fs
}

func (cmd *starCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	opts := &drive.Options{
		Path:    path,
		Sources: sources,
		Quiet:   *cmd.Quiet,
	}

	exitWithError(drive.New(context, opts).Star(*cmd.ById))
}

type unstarCmd struct {
	starCmd
}

func (cmd *unstarCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *cmd.ById)

	opts := &drive.Options{
		Path:    path,
		Sources: sources,
	}

	exitWithError(drive.New(context, opts).UnStar(*cmd.ById))
}

type idCmd struct {
	Depth  *int  `json:"depth"`
	Hidden *bool `json:"hidden"`
}

func (cmd *idCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Depth = fs.Int(drive.DepthKey, 1, "maximum recursion depth")
	cmd.Hidden = fs.Bool(drive.HiddenKey, true, "allows operation on hidden paths")
	return fs
}

func (icmd *idCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgs(args)
	cmd := idCmd{}
	df := defaultsFiller{
		from: *icmd, to: &cmd,
		rcSourcePath: context.AbsPathOf(path),
		definedFlags: definedFlags,
	}

	if err := fillWithDefaults(df); err != nil {
		exitWithError(err)
	}

	opts := &drive.Options{
		Path:    path,
		Sources: sources,
		Depth:   *cmd.Depth,
		Hidden:  *cmd.Hidden,
	}

	exitWithError(drive.New(context, opts).Id())
}

type clashesCmd struct {
	Fix    *bool `json:"fix"`
	List   *bool `json:"list"`
	ById   *bool `json:"by-id"`
	Depth  *int  `json:"depth"`
	Hidden *bool `json:"hidden"`
}

func (cmd *clashesCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.Fix = fs.Bool(drive.CLIOptionFixClashes, false, drive.DescFixClashes)
	cmd.List = fs.Bool(drive.CLIOptionListClashes, true, drive.DescListClashes)
	cmd.ById = fs.Bool(drive.CLIOptionId, false, drive.DescClashesOpById)
	cmd.Depth = fs.Int(drive.DepthKey, drive.InfiniteDepth, "maximum recursion depth")
	cmd.Hidden = fs.Bool(drive.HiddenKey, true, "allows operation on hidden paths")

	return fs
}

func (ccmd *clashesCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgsByToggle(args, *ccmd.ById)
	cmd := clashesCmd{}
	df := defaultsFiller{
		from: *ccmd, to: &cmd,
		rcSourcePath: context.AbsPathOf(path),
		definedFlags: definedFlags,
	}

	if err := fillWithDefaults(df); err != nil {
		exitWithError(err)
	}

	opts := &drive.Options{
		Path:    path,
		Sources: sources,
		Depth:   *cmd.Depth,
		Hidden:  *cmd.Hidden,
	}

	driveInstance := drive.New(context, opts)
	fn := driveInstance.ListClashes
	if *cmd.Fix {
		fn = driveInstance.FixClashes
	}

	exitWithError(fn(*cmd.ById))
}

type issueCmd struct {
	Piped *bool
	Title *string
	Body  *string
}

func (ic *issueCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	ic.Piped = fs.Bool(drive.CLIOptionPiped, false, drive.DescPiped)
	ic.Title = fs.String(drive.IssueTitleKey, "", drive.DescIssueTitle)
	ic.Body = fs.String(drive.IssueBodyKey, "", drive.DescIssueBody)

	return fs
}

func (cmd *issueCmd) Run(args []string, definedFlags map[string]*flag.Flag) {
	sources, context, path := preprocessArgs(args)

	meta := map[string][]string{
		drive.IssueTitleKey: []string{*cmd.Title},
		drive.IssueBodyKey:  []string{*cmd.Body},
	}

	if *cmd.Piped {
		meta[drive.PipedKey] = []string{fmt.Sprintf("%v", *cmd.Piped)}
	}

	opts := &drive.Options{
		Path:    path,
		Sources: sources,
		Meta:    &meta,
	}

	exitWithError(drive.New(context, opts).FileIssue())
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
		drive.FprintfShadow(os.Stderr, "%v\n", err)
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
			drive.FprintfShadow(os.Stderr, "%s %v\n", p, err)
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
