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

package ripper

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
)

type diction struct {
	skip  *regexp.Regexp
	end   *regexp.Regexp
	begin *regexp.Regexp
}

func readFile(absPath string, dctn *diction) (content string, err error) {
	var f *os.File
	f, err = os.Open(absPath)
	if err != nil || f == nil {
		return
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)

	results := [][]byte{}

	canSkip := dctn.skip != nil
	canEnd := dctn.end != nil
	canBegin := dctn.begin != nil

	if canBegin {
		for scanner.Scan() {
			lineBytes := scanner.Bytes()
			if dctn.begin.Match(lineBytes) {
				results = append(results, lineBytes)
				break
			}
		}
	}

	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		if canEnd && dctn.end.Match(lineBytes) {
			break
		}
		if canSkip && dctn.skip.Match(lineBytes) {
			continue
		}

		results = append(results, lineBytes)
	}

	joined := bytes.Join(results, []byte{'\n'})
	content = string(joined)
	return
}

func ApacheTopLicenseRip(absPath string) (content string, err error) {
	var begin, end *regexp.Regexp

	begin, err = regexp.Compile("^//\\s*Copyright")
	if err != nil {
		return
	}

	end, err = regexp.Compile("^package")
	if err != nil {
		return
	}

	dctn := diction{
		skip:  nil,
		end:   end,
		begin: begin,
	}

	return readFile(absPath, &dctn)
}
