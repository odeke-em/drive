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
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
)

// MaxFileSize is the max number of bytes we
// can accept for diffing (Arbitrary value)
const MaxFileSize = 50 * 1024 * 1024

var Ruler = strings.Repeat("*", 80)

func (g *Commands) Diff() (err error) {
	var cl []*Change
	cl, err = g.changeListResolve(true)
	if err != nil {
		return
	}

	var diffUtilPath string
	diffUtilPath, err = exec.LookPath("diff")
	if err != nil {
		return
	}

	for _, c := range cl {
		dErr := g.perDiff(c, diffUtilPath, ".")
		if dErr != nil {
			fmt.Println(dErr)
		}
	}
	return
}

func sysHasDiff() bool {
	_, err := exec.LookPath("diff")
	return err == nil
}

func (g *Commands) perDiff(change *Change, diffProgPath, cwd string) (err error) {
	defer func() {
		fmt.Println(Ruler)
	}()

	l, r := change.Src, change.Dest
	if l == nil && r == nil {
		return fmt.Errorf("Neither remote nor local exists")
	}
	if r == nil && l != nil {
		return fmt.Errorf("%s only on local", change.Path)
	}
	if l == nil && r != nil {
		return fmt.Errorf("%s only on remote", change.Path)
	}
	// Pre-screening phase
	if r.IsDir && l.IsDir {
		return fmt.Errorf("Both local and remote are directories")
	}
	if r.IsDir && !l.IsDir {
		return fmt.Errorf("Remote is a directory while local is an ordinary file")
	}

	if l.IsDir && !r.IsDir {
		return fmt.Errorf("Local is a directory while remote is an ordinary file")
	}

	if r.BlobAt == "" {
		return fmt.Errorf("Cannot access download link for '%v'", r.Name)
	}

	if r.Size > MaxFileSize {
		return fmt.Errorf("%s Remote too large for display \033[94m[%v bytes]\033[00m",
			change.Path, r.Size)
	}
	if l.Size > MaxFileSize {
		return fmt.Errorf("%s Local too large for display \033[92m[%v bytes]\033[00m",
			change.Path, l.Size)
	}

	if isSameFile(r, l) {
		// No output when "no changes found"
		return nil
	}

	var frTmp, fl *os.File
	var blob io.ReadCloser

	// Clean-up
	defer func() {
		if frTmp != nil {
			os.RemoveAll(frTmp.Name())
		}
		if fl != nil {
			fl.Close()
		}
		if blob != nil {
			blob.Close()
		}
	}()

	blob, err = g.rem.Download(r.Id, "")
	if err != nil {
		return err
	}

	// Next step: Create a temp file with an obscure name unlikely to clash.
	tmpName := strings.Join([]string{
		".",
		fmt.Sprintf("tmp%v.tmp", rand.Int()),
	}, "x")

	frTmp, err = ioutil.TempFile(".", tmpName)
	if err != nil {
		return
	}
	_, err = io.Copy(frTmp, blob)
	if err != nil {
		return
	}

	fmt.Printf("%s\n%s %s\n", Ruler, l.Name, r.Name)

	diffCmd := exec.Cmd{
		Args:   []string{diffProgPath, l.BlobAt, frTmp.Name()},
		Dir:    cwd,
		Path:   diffProgPath,
		Stdin:  nil,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	// Normally when elements differ diff returns a non-zero code
	_ = diffCmd.Run()
	return
}
