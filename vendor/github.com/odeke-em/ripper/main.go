// Copyright 2015 Emmanuel Odeke. All Rights Reserved.
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

package main

import (
	"fmt"
	"os"

	"github.com/odeke-em/ripper/src"
)

func main() {
	argv := os.Args[1:]

	if len(argv) < 1 {
		argv = append(argv, ripper.CallerFilename())
	}

	for _, arg := range argv {
		content, err := ripper.ApacheTopLicenseRip(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else if len(content) >= 1 {
			fmt.Printf("* %s\n%s\n", arg, content)
		}
	}
}
