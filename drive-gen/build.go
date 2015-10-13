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
	"os/exec"
	"path/filepath"
	"time"

	"github.com/odeke-em/drive/src"
	"github.com/odeke-em/ripper/src"
	"github.com/odeke-em/xon/pkger/src"

	_ "github.com/odeke-em/command"
	_ "github.com/odeke-em/drive/config"
)

var AliasBinaryDir = "drive-google"

func logErr(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
}

func exitIfError(err error) {
	if err != nil {
		logErr(err)
		os.Exit(-1)
	}
}

func main() {
	pkgInfo, err := pkger.Recon(drive.DriveRepoRelPath)
	exitIfError(err)

	absDrivePath := pkger.GoSrcify(drive.DriveRepoRelPath)
	sampleFullpath := filepath.Join(absDrivePath, "src", "about.go")

	var license string
	license, err = ripper.ApacheTopLicenseRip(sampleFullpath)
	exitIfError(err)

	rubric := []struct {
		field string
		value string
	}{
		{"CommitHash", pkgInfo.CommitHash},
		{"GoVersion", pkgInfo.GoVersion},
		{"OsInfo", pkgInfo.OsInfo},
		{"BuildTime", pkgInfo.BuildTime},
	}

	importsClause := "import \"github.com/odeke-em/xon/pkger/src\"\n\n"
	formatted := "var PkgInfo = &pkger.PkgInfo {\n"

	for _, v := range rubric {
		formatted += fmt.Sprintf("\t%s: \"%s\",\n", v.field, v.value)
	}
	formatted += "}\n"

	generatedDir := filepath.Join(absDrivePath, "gen")
	err = os.MkdirAll(generatedDir, 0755)
	exitIfError(err)

	generatedInfoPath := filepath.Join(generatedDir, "generated.go")

	originalContent, origReadErr := drive.ReadFullFile(generatedInfoPath)
	exitIfError(origReadErr)

	f, fErr := os.Create(generatedInfoPath)
	exitIfError(fErr)

	defer f.Close()

	packageInfo := "\n\npackage generated\n\n"

	autoGenerationInfo := fmt.Sprintf("\n\n// This file was auto-generated at %s\n// Edits will be overwritten!\n", time.Now().Round(time.Millisecond))

	clauses := []string{
		license,
		autoGenerationInfo,
		packageInfo,
		importsClause,
		formatted,
	}

	for _, clause := range clauses {
		_, wErr := f.Write([]byte(clause))
		if wErr != nil {
			logErr(wErr)
		}
	}

	// Next step will be to go get
	goBinaryPath, lookUpErr := exec.LookPath("go")
	exitIfError(lookUpErr)

	argc := len(os.Args)

	srcDirSegments := []string{"cmd", "drive"}

	if argc >= 2 {
		if os.Args[1] == AliasBinaryDir {
			srcDirSegments = []string{AliasBinaryDir}
		}
	}

	allCombined := append([]string{drive.DriveRepoRelPath}, srcDirSegments...)
	driveMainPath := filepath.Join(allCombined...)
	generateCmd := exec.Cmd{
		Args:   []string{goBinaryPath, "get", driveMainPath},
		Dir:    ".",
		Path:   goBinaryPath,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err = generateCmd.Run()

	// Now revert the original content
	revertHandle, revertOpenErr := os.Create(generatedInfoPath)
	exitIfError(revertOpenErr)

	for _, revertedClause := range originalContent {
		fullClause := fmt.Sprintf("%s\n", revertedClause)
		_, wErr := revertHandle.Write([]byte(fullClause))
		if wErr != nil {
			logErr(wErr)
		}
	}
	revertHandle.Close()

	exitIfError(err)
}
