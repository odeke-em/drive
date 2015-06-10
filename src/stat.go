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
	"path/filepath"
	"time"

	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	drive "github.com/odeke-em/google-api-go-client/drive/v2"

	"github.com/odeke-em/log"
)

type keyValue struct {
	key   string
	value interface{}
}

func (g *Commands) StatById() error {
	return g.statfn("statById", g.rem.FindById)
}

func (g *Commands) Stat() error {
	return g.statfn("stat", g.rem.FindByPath)
}

func (g *Commands) statfn(fname string, fn func(string) (*File, error)) error {
	for _, src := range g.opts.Sources {
		f, err := fn(src)
		if err != nil {
			g.log.LogErrf("%s: %s err: %v\n", fname, src, err)
			continue
		}

		if g.opts.Md5sum {

			depth := g.opts.Depth
			src = f.Name // forces filename if -id is used

			// md5sum with no arguments should do md5sum *
			if f.IsDir && g.opts.Path == "/" {
				src = ""
				if depth == 1 {
					depth = 2
				}
			}

			err = g.stat(src, f, depth)
		} else {
			err = g.stat(src, f, g.opts.Depth)
		}

		if err != nil {
			g.log.LogErrf("%s: %s err: %v\n", fname, src, err)
			continue
		}
	}

	return nil
}

func prettyPermission(logf log.Loggerf, perm *drive.Permission) {
	logf("\n*\nName: %v <%s>\n", perm.Name, perm.EmailAddress)
	kvList := []*keyValue{
		&keyValue{"Role", perm.Role},
		&keyValue{"AccountType", perm.Type},
	}
	for _, kv := range kvList {
		logf("%-20s %-30v\n", kv.key, kv.value.(string))
	}
	logf("*\n")
}

func prettyFileStat(logf log.Loggerf, relToRootPath string, file *File) {
	dirType := "file"
	if file.IsDir {
		dirType = "folder"
	}

	logf("\n\033[92m%s\033[00m\n", relToRootPath)

	kvList := []*keyValue{
		&keyValue{"Filename", file.Name},
		&keyValue{"FileId", file.Id},
		&keyValue{"Bytes", fmt.Sprintf("%v", file.Size)},
		&keyValue{"Size", prettyBytes(file.Size)},
		&keyValue{"DirType", dirType},
		&keyValue{"VersionNumber", fmt.Sprintf("%v", file.Version)},
		&keyValue{"MimeType", file.MimeType},
		&keyValue{"Etag", file.Etag},
		&keyValue{"ModTime", fmt.Sprintf("%v", file.ModTime)},
		&keyValue{"LastViewedByMe", fmt.Sprintf("%v", file.LastViewedByMeTime)},
		&keyValue{"Shared", fmt.Sprintf("%v", file.Shared)},
		&keyValue{"Owners", sepJoin(" & ", file.OwnerNames...)},
		&keyValue{"LastModifyingUsername", file.LastModifyingUsername},
	}

	if file.Name != file.OriginalFilename {
		kvList = append(kvList, &keyValue{"OriginalFilename", file.OriginalFilename})
	}

	if !file.IsDir {
		kvList = append(kvList, &keyValue{"Md5Checksum", file.Md5Checksum})

		// By default, folders are non-copyable, but drive implements recursively copying folders
		kvList = append(kvList, &keyValue{"Copyable", fmt.Sprintf("%v", file.Copyable)})
	}

	if file.Labels != nil {
		kvList = append(kvList,
			&keyValue{"Starred", fmt.Sprintf("%v", file.Labels.Starred)},
			&keyValue{"Viewed", fmt.Sprintf("%v", file.Labels.Viewed)},
			&keyValue{"Trashed", fmt.Sprintf("%v", file.Labels.Trashed)},
			&keyValue{"ViewersCanDownload", fmt.Sprintf("%v", file.Labels.Restricted)},
		)
	}

	for _, kv := range kvList {
		logf("%-25s %-30v\n", kv.key, kv.value.(string))
	}
}

func (g *Commands) stat(relToRootPath string, file *File, depth int) error {

	if depth == 0 {
		return nil
	}

	if depth >= 1 {
		depth -= 1
	}

	// Arbitrary value for throttle pause duration
	// TODO Is this really needed now that everything is serialized?

	throttle := time.Tick(1e9 / 5)

	if !g.opts.Md5sum {
		prettyFileStat(g.log.Logf, relToRootPath, file)
		perms, permErr := g.rem.listPermissions(file.Id)
		if permErr != nil {
			return permErr
		}

		for _, perm := range perms {
			prettyPermission(g.log.Logf, perm)
		}
	} else if file.Md5Checksum != "" {
		g.log.Logf("%32s  %s\n", file.Md5Checksum, strings.TrimPrefix(relToRootPath, "/"))
	}

	if file.IsDir {
		//remoteChildren := FileArray{}
		var remoteChildren FileArray

		for child := range g.rem.FindByParentId(file.Id, g.opts.Hidden) {
			remoteChildren = append(remoteChildren, child)
			<-throttle
		}

		if g.opts.Md5sum {
			// TODO use g.sort instead of sort.stable
			// i.e g.sort(remoteChildren,"name")
			// The reason this is not done here is because g.sort does not sort in natural order
			
			sort.Stable(remoteChildren)
		}

		for _, child := range remoteChildren {
			g.stat(filepath.Clean(relToRootPath+"/"+child.Name), child, depth)
		}
	}

	return nil
}

// FileArray sorts by File.Name

type FileArray []*File

func (this FileArray) Len() int {
	return len(this)
}

// TODO get collation order from system's locale
// language.Und seems to work well for common western locales

var collator *collate.Collator = collate.New(language.Und)

func (this FileArray) Less(i, j int) bool {

	cmp := collator.CompareString(this[i].Name, this[j].Name)

	// This should ensure stable results when two files have the same name, I think

	if cmp == 0 {
		if this[i].Id < this[j].Id {
			cmp = -1
		} else {
			cmp = +1
		}
	}

	return cmp < 0
}

func (this FileArray) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
