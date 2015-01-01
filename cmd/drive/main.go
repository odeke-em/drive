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
	"strings"

	"github.com/odeke-em/drive"
	"github.com/odeke-em/drive/config"
	"github.com/rakyll/command"
)

var context *config.Context

const (
	descInit      = "inits a directory and authenticates user"
	descPull      = "pulls remote changes from google drive"
	descPush      = "push local changes to google drive"
	descDiff      = "compares a local file with remote"
	descPublish   = "publishes a file and prints its publicly available url"
	descUnpublish = "revokes public access to a file"
)

func main() {
	command.On("init", descInit, &initCmd{}, []string{})
	command.On("pull", descPull, &pullCmd{}, []string{})
	command.On("push", descPush, &pushCmd{}, []string{})
	command.On("diff", descDiff, &diffCmd{}, []string{})
	command.On("pub", descPublish, &publishCmd{}, []string{})
	command.On("unpub", descUnpublish, &unpublishCmd{}, []string{})
	command.ParseAndRun()
}

type initCmd struct{}

func (cmd *initCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *initCmd) Run(args []string) {
	exitWithError(drive.New(initContext(args), nil).Init())
}

type pullCmd struct {
	export      *string
	isRecursive *bool
	isNoPrompt  *bool
	noClobber   *bool
}

func (cmd *pullCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noClobber = fs.Bool("no-clobber", false, "prevents overwriting of old content")
	cmd.export = fs.String(
		"export", "", "comma separated list of formats to export your docs + sheets files")
	cmd.isRecursive = fs.Bool("r", true, "performs the pull action recursively")
	cmd.isNoPrompt = fs.Bool("no-prompt", false, "shows no prompt before applying the pull action")
	return fs
}

func nonEmptyStrings(v []string) (splits []string) {
	for _, elem := range v {
		if elem != "" {
			splits = append(splits, elem)
		}
	}
	return
}

func (cmd *pullCmd) Run(args []string) {
	context, path := discoverContext(args)

	// Filter out empty strings.
	exports := nonEmptyStrings(strings.Split(*cmd.export, ","))

	exitWithError(drive.New(context, &drive.Options{
		Exports:     exports,
		IsNoPrompt:  *cmd.isNoPrompt,
		IsRecursive: *cmd.isRecursive,
		NoClobber:   *cmd.noClobber,
		Path:        path,
	}).Pull())
}

type pushCmd struct {
	noClobber   *bool
	hidden      *bool
	isNoPrompt  *bool
	isRecursive *bool
	mountedPush *bool
}

func (cmd *pushCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noClobber = fs.Bool("no-clobber", false, "allows overwriting of old content")
	cmd.hidden = fs.Bool("hidden", false, "allows syncing of hidden paths")
	cmd.isRecursive = fs.Bool("r", true, "performs the push action recursively")
	cmd.isNoPrompt = fs.Bool("no-prompt", false, "shows no prompt before applying the push action")
	cmd.mountedPush = fs.Bool("m", false, "allows pushing of mounted paths")
	return fs
}

func (cmd *pushCmd) Run(args []string) {
	if *cmd.mountedPush {
		pushMounted(cmd, args)
	} else {
		context, path := discoverContext(args)
		exitWithError(drive.New(context, &drive.Options{
			NoClobber:   *cmd.noClobber,
			Hidden:      *cmd.hidden,
			IsNoPrompt:  *cmd.isNoPrompt,
			IsRecursive: *cmd.isRecursive,
			Path:        path,
		}).Push())
	}
}

func pushMounted(cmd *pushCmd, args []string) {
	argc := len(args)

	var contextArgs, rest, sources []string

	if !*cmd.mountedPush {
		contextArgs = args
	} else {
		// Expectation is that at least one path has to be passed
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

	rest = nonEmptyStrings(rest)
	context, path := discoverContext(contextArgs)
	contextAbsPath, err := filepath.Abs(path)
	exitWithError(err)

	mountPoints, auxSrcs := config.MountPoints(path, contextAbsPath, rest, *cmd.hidden)
	sources = append(sources, auxSrcs...)

	exitWithError(drive.New(context, &drive.Options{
		Hidden:      *cmd.hidden,
		IsNoPrompt:  *cmd.isNoPrompt,
		IsRecursive: *cmd.isRecursive,
		Mounts:      mountPoints,
		NoClobber:   *cmd.noClobber,
		Path:        path,
		Sources:     sources,
	}).Push())
}

type diffCmd struct{}

func (cmd *diffCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *diffCmd) Run(args []string) {
	context, path := discoverContext(args)
	exitWithError(drive.New(context, &drive.Options{
		IsRecursive: true,
		Path:        path,
	}).Diff())
}

type publishCmd struct{}
type unpublishCmd struct{}

func (cmd *unpublishCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *unpublishCmd) Run(args []string) {
	context, path := discoverContext(args)
	exitWithError(drive.New(context, &drive.Options{
		Path: path,
	}).Unpublish())
}

func (cmd *publishCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *publishCmd) Run(args []string) {
	context, path := discoverContext(args)
	exitWithError(drive.New(context, &drive.Options{
		Path: path,
	}).Publish())
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

func exitWithError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
