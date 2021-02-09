// Copyright 2017 Google Inc. All Rights Reserved.
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
	uuid "github.com/odeke-em/go-uuid"

	drive "google.golang.org/api/drive/v2"
)

type Watch struct {
	Cancel       chan<- struct{}
	ResponseChan chan *WatchResponse
	Filepath     string
	File         *File
}

type WatchChannel drive.Channel

func (c *Commands) FileWatch(wreq *WatchChannel) (map[string]*Watch, error) {
	watchesMapping := make(map[string]*Watch)

	// cleanUp is invoked when we encounter errors.
	cleanUp := func() error {
		for _, watch := range watchesMapping {
			watch.Cancel <- struct{}{}
		}

		// Now explicitly set it to nil to discard any content.
		watchesMapping = nil
		return nil
	}

	for _, relToRootPath := range c.opts.Sources {
		f, err := c.rem.FindByPath(relToRootPath)
		if err != nil {
			if err == ErrPathNotExists {
				// non-existent files are excusible
				continue
			}
			cleanUp()
			return nil, err
		}

		castReq := drive.Channel(*wreq)
		if castReq.Id == "" {
			castReq.Id = uuid.NewRandom().String()
		}

		// Otherwise now let's set up this fileId for watching
		cancel := make(chan struct{})
		wreq := &WatchRequest{
			FileId:  f.Id,
			Cancel:  cancel,
			Request: &castReq,
		}

		reschan, err := c.rem.watchForChanges(wreq)
		if err != nil {
			cleanUp()
			return nil, err
		}

		watchesMapping[f.Id] = &Watch{
			Cancel:   cancel,
			File:     f,
			Filepath: relToRootPath,

			ResponseChan: reschan,
		}
	}

	return watchesMapping, nil
}
