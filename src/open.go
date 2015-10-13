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
	"github.com/skratchdot/open-golang/open"
)

type OpenType uint

const (
	OpenNone OpenType = 1 << iota
	FileManagerOpen
	BrowserOpen
	IdOpen
)

type opener func(string) error

func (g *Commands) Open(ot OpenType) error {
	byId := (ot & IdOpen) != 0
	kvChan := g.urler(byId, g.opts.Sources)

	for kv := range kvChan {
		switch kv.value.(type) {
		case error:
			g.log.LogErrf("%s: %s\n", kv.key, kv.value)
			continue
		}

		openArgs := []string{}
		canAddUrl := (ot & BrowserOpen) != 0

		if byId {
			canAddUrl = true
		} else if (ot & FileManagerOpen) != 0 {
			openArgs = append(openArgs, g.context.AbsPathOf(kv.key))
		}

		if canAddUrl {
			if castKey, ok := kv.value.(string); ok {
				openArgs = append(openArgs, castKey)
			}
		}

		for _, arg := range openArgs {
			open.Start(arg)
		}
	}

	return nil
}
