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

func (g *Commands) Url(byId bool) error {
	kvChan := g.urler(byId, g.opts.Sources)

	for kv := range kvChan {
		switch kv.value.(type) {
		case error:
			g.log.LogErrf("%s: %s\n", kv.key, kv.value)
		default:
			g.log.Logf("%s: %v\n", kv.key, kv.value)
		}
	}

	return nil
}

func (g *Commands) urler(byId bool, sources []string) (kvChan chan *keyValue) {
	fileUrler := func(f *File) interface{} {
		if f == nil {
			return ""
		}
		return f.Url()
	}

	return resolver(g, byId, g.opts.Sources, fileUrler)
}
