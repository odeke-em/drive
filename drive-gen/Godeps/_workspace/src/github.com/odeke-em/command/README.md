# command

[![Build Status](https://travis-ci.org/odeke-em/command.png?branch=master)](https://travis-ci.org/odeke-em/command)

command is a tiny package that helps you to add cli subcommands to your Go program with no effort, and prints a pretty guide if needed.

~~~
Usage: program <command>

where <command> is one of:
  version   prints the version
  command1  some description about command1
  command2  some description about command2

available flags:
  -exec-path="": a custom path to executable

program <command> -h for subcommand help
~~~

## Usage

In order to start, go get this repository:

~~~ sh
go get github.com/odeke-em/command
~~~

This package allows you to use flags package as you used to do, and provides additional parsing for subcommands and subcommand flags.

~~~ go
import "github.com/odeke-em/command"

// register any global flags
var flagExecPath = flag.String("exec-path", "", "a custom path to executable")

type VersionCommand struct{
	flagVerbose *bool
}

func (cmd *VersionCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	// define subcommand's flags
	cmd.flagVerbose = fs.Bool("v", false, "provides verbose output")
	return fs
}

func (cmd *VersionCommand) Run(args []string) {
	// implement the main body of the subcommand here
  // required and optional arguments are found in args
}

// register version as a subcommand
command.On("version", "prints the version", &VersionCommand{}, []string{"<required-arg>"})
command.On("command1", "some description about command1", ..., []string{})
command.On("command2", "some description about command2", ..., []string{})
command.Parse()
// ...
command.Run()
~~~

The program above will handle the registered commands and invoke the matching command's `Run` or print subcommand help if `-h` is set.

~~~
$ program -exec-path=/home/user/bin/someexec version -v=true history
~~~

will output the version of the program in a verbose way requring an argument (history), and will set the exec path to the provided path. If arguments doesn't match any subcommand or illegal arguments are provided, it will print the usage guide.


## License

Copyright 2013 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
