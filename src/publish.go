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
)

func (c *Commands) Publish(byId bool) (err error) {
	for _, relToRoot := range c.opts.Sources {
		if pubErr := c.pub(relToRoot, byId); pubErr != nil {
			c.log.LogErrf("\033[91mPub\033[00m %s:  %v\n", relToRoot, pubErr)
		}
	}
	return
}

func (c *Commands) remFileResolve(relToRoot string, byId bool) ([]*File, error) {
	resolver := c.rem.FindByPath
	if byId {
		resolver = c.rem.FindByIdMulti
	}

	return resolver(relToRoot)
}

func (c *Commands) pub(relToRoot string, byId bool) (errs []error) {
	files, err := c.remFileResolve(relToRoot, byId)
	if err != nil || len(files) < 1 {
		errs = append(errs, err)
		return
	}

	for _, file := range files {
		link, err := c.rem.Publish(file.Id)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		link = file.Url()

		if byId {
			relToRoot = fmt.Sprintf("%s aka %s", relToRoot, file.Name)
		}

		c.log.Logf("%s published on %s\n", relToRoot, link)
	}

	return
}

func (c *Commands) Unpublish(byId bool) error {
	for _, relToRoot := range c.opts.Sources {
		if unpubErr := c.unpub(relToRoot, byId); unpubErr != nil {
			c.log.LogErrf("\033[91mUnpub\033[00m %s:  %v\n", relToRoot, unpubErr)
		}
	}
	return nil
}

func (c *Commands) unpub(relToRoot string, byId bool) (errs []error) {
	files, err := c.remFileResolve(relToRoot, byId)
	if err != nil {
		errs = append(errs, err)
		return
	}

	for _, file := range files {
		if upErr := c.rem.Unpublish(file.Id); upErr != nil {
			errs = append(errs, upErr)
		}
	}

	return errs
}
