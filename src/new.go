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
	"path/filepath"
	"time"
)

func (g *Commands) NewFolder() (err error) {
	return _newFile(g, true)
}

func (g *Commands) NewFile() (err error) {
	return _newFile(g, false)
}

func _newFile(g *Commands, folder bool) (err error) {
	var coercedMimeKey string
	if g.opts.Meta != nil {
		meta := *(g.opts.Meta)
		mimeKeys, ok := meta[MimeKey]
		if ok {
			coercedMimeKey = sepJoin("", mimeKeys...)
		}
	}

	spin := g.playabler()
	spin.play()
	defer spin.stop()

	for _, relToRootPath := range g.opts.Sources {
		parentPath, basename := g.pathSplitter(relToRootPath)

		parent, parErr := g.remoteMkdirAll(parentPath)
		if parErr != nil {
			g.log.LogErrf("newFile: %s %v\n", relToRootPath, parErr)
			continue
		}

		f := &File{
			ModTime: time.Now(),
			Name:    urlToPath(basename, false),
		}

		if folder {
			f.IsDir = true
		} else {
			mimeKey := coercedMimeKey
			if coercedMimeKey == "" {
				mimeKey = filepath.Ext(relToRootPath)
			}
			f.MimeType = mimeTypeFromQuery(mimeKey)
		}

		upArg := upsertOpt{
			parentId: parent.Id,
			src:      f,
		}

		freshFile, _, fErr := g.rem.upsertByComparison(nil, &upArg)
		if fErr != nil {
			g.log.LogErrf("newFile: %s creation failed %v\n", relToRootPath, fErr)
			continue
		}

		g.log.Logf("%s %s\n", relToRootPath, freshFile.Id)
	}

	return err
}
