// Copyright 2015 Google Inc. All Rights Reserved.
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
	"strings"
)

func noopOnFile(f *File) interface{} {
	return f
}

func noopOnIgnorer(s string) bool {
	return false
}

func (g *Commands) EditDescription(byId bool) (composedErr error) {
	metaPtr := g.opts.Meta

	if metaPtr == nil {
		return fmt.Errorf("edit: no descriptions passed in")
	}

	meta := *metaPtr

	description := strings.Join(meta[EditDescriptionKey], "\n")

	if _, ok := meta[PipedKey]; ok {
		clauses, err := readFileFromStdin(noopOnIgnorer)
		if err != nil {
			return err
		}

		description = strings.Join(clauses, "\n")
	}

	if description == "" && g.opts.canPrompt() {
		g.log.Logln("Using an empty description will clear out the previous one")
		if !promptForChanges() {
			return
		}
	}

	kvChan := resolver(g, byId, g.opts.Sources, noopOnFile)

	for kv := range kvChan {
		file, ok := kv.value.(*File)
		if !ok {
			g.log.LogErrf("%s: %s\n", kv.key, kv.value)
			continue
		}

		if file == nil {
			continue
		}

		updatedFile, err := g.rem.updateDescription(file.Id, description)
		if err != nil {
			composedErr = reComposeError(composedErr, fmt.Sprintf("%q %v", kv.key, err))
		} else if updatedFile != nil {
			name := fmt.Sprintf("%q", kv.key)
			if kv.key != updatedFile.Id {
				name = fmt.Sprintf("%s aka %q", name, updatedFile.Id)
			}
			g.log.LogErrf("Description updated for %s\n", name)
		}
	}

	return
}
