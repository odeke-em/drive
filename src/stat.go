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
	drive "github.com/google/google-api-go-client/drive/v2"
	"github.com/odeke-em/log"
	"path/filepath"
	"strings"
)

type keyValue struct {
	key   string
	value interface{}
}

func (g *Commands) StatById() error {
	return g.statfn("statById", g.rem.FindByIdMulti)
}

func (g *Commands) Stat() error {
	return g.statfn("stat", g.rem.FindByPath)
}

func (g *Commands) statfn(fname string, fn func(string) ([]*File, error)) error {
	for _, src := range g.opts.Sources {
		files, err := fn(src)
		if err != nil {
			g.log.LogErrf("%s: %s err: %v\n", fname, src, err)
			continue
		}

		for _, f := range files {
			if g.opts.Md5sum {
				src = f.Name // forces filename if -id is used

				// md5sum with no arguments should do md5sum *
				if f.IsDir && rootLike(g.opts.Path) {
					src = ""
				}
			}

			err = g.stat(src, f, g.opts.Depth)

			if err != nil {
				g.log.LogErrf("%s: %s err: %v\n", fname, src, err)
				continue
			}
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

	if g.opts.Md5sum {
		if file.Md5Checksum != "" {
			g.log.Logf("%32s  %s\n", file.Md5Checksum, strings.TrimPrefix(relToRootPath, "/"))
		}
	} else {
		prettyFileStat(g.log.Logf, relToRootPath, file)
		perms, permErr := g.rem.listPermissions(file.Id)
		if permErr != nil {
			return permErr
		}

		for _, perm := range perms {
			prettyPermission(g.log.Logf, perm)
		}
	}

	if depth == 0 || !file.IsDir {
		return nil
	}

	if depth >= 1 {
		depth -= 1
	}

	var remoteChildren []*File

	for child := range g.rem.FindByParentId(file.Id, g.opts.Hidden) {
		remoteChildren = append(remoteChildren, child)
	}

	if g.opts.Md5sum {
		g.sort(remoteChildren, Md5Key, NameKey)
	}

	for _, child := range remoteChildren {
		g.stat(filepath.Clean(relToRootPath+"/"+child.Name), child, depth)
	}

	return nil
}
