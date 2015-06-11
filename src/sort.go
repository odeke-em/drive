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
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"sort"
	"strings"
)

const (
	AttrUnknown = iota
	AttrSize
	AttrModTime
	AttrLastViewedByMeTime
	AttrVersion
	AttrIsDir
	AttrMd5Checksum
	AttrMimeType
	AttrName
)

type attr int

type fileList []*File
type lastViewedTimeFlist fileList
type md5Flist fileList
type mimeTypeFlist fileList
type modTimeFlist fileList
type nameFlist fileList
type sizeFlist fileList
type typeFlist fileList
type versionFlist fileList

var (
	lastViewedTimeCmpLess = _lessCmper(AttrLastViewedByMeTime)
	md5CmpLess            = _lessCmper(AttrMd5Checksum)
	mimeTypeCmpLess       = _lessCmper(AttrMimeType)
	modTimeCmpLess        = _lessCmper(AttrModTime)
	nameCmpLess           = _lessCmper(AttrName)
	sizeCmpLess           = _lessCmper(AttrSize)
	versionCmpLess        = _lessCmper(AttrVersion)
	typeCmpLess           = _lessCmper(AttrIsDir)
)

var (
	// TODO get collation order from system's locale
	// language.Und seems to work well for common western locales

	collator *collate.Collator = collate.New(language.Und)
)

func (fl modTimeFlist) Less(i, j int) bool {
	return modTimeCmpLess(fl[i], fl[j])
}

func (fl modTimeFlist) Len() int {
	return len(fl)
}

func (fl modTimeFlist) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl typeFlist) Less(i, j int) bool {
	return typeCmpLess(fl[i], fl[j])
}

func (fl typeFlist) Len() int {
	return len(fl)
}

func (fl typeFlist) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl mimeTypeFlist) Less(i, j int) bool {
	return mimeTypeCmpLess(fl[i], fl[j])
}

func (fl mimeTypeFlist) Len() int {
	return len(fl)
}

func (fl mimeTypeFlist) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl lastViewedTimeFlist) Less(i, j int) bool {
	return lastViewedTimeCmpLess(fl[i], fl[j])
}

func (fl lastViewedTimeFlist) Len() int {
	return len(fl)
}

func (fl lastViewedTimeFlist) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl versionFlist) Less(i, j int) bool {
	return versionCmpLess(fl[i], fl[j])
}

func (fl versionFlist) Len() int {
	return len(fl)
}

func (fl versionFlist) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl nameFlist) Less(i, j int) bool {
	return nameCmpLess(fl[i], fl[j])
}

func (fl nameFlist) Len() int {
	return len(fl)
}

func (fl nameFlist) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl md5Flist) Less(i, j int) bool {
	return md5CmpLess(fl[i], fl[j])
}

func (fl md5Flist) Len() int {
	return len(fl)
}

func (fl md5Flist) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl sizeFlist) Less(i, j int) bool {
	return sizeCmpLess(fl[i], fl[j])
}

func (fl sizeFlist) Len() int {
	return len(fl)
}

func (fl sizeFlist) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func attrAtoiSorter(a string, fl []*File) (attr, sort.Interface, bool) {
	aLower := strings.ToLower(a)
	if len(aLower) < 1 {
		return AttrUnknown, nil, false
	}

	reverse := hasAnySuffix(aLower, "_r", "-")

	if hasAnyPrefix(aLower, Md5Key) {
		return AttrMd5Checksum, md5Flist(fl), reverse
	}
	if hasAnyPrefix(aLower, NameKey) {
		return AttrName, nameFlist(fl), reverse
	}
	if hasAnyPrefix(aLower, SizeKey) {
		return AttrSize, sizeFlist(fl), reverse
	}
	if hasAnyPrefix(aLower, TypeKey) {
		return AttrIsDir, typeFlist(fl), reverse
	}
	if hasAnyPrefix(aLower, ModTimeKey) {
		return AttrModTime, modTimeFlist(fl), reverse
	}
	if hasAnyPrefix(aLower, LastViewedByMeTimeKey) {
		return AttrLastViewedByMeTime, lastViewedTimeFlist(fl), reverse
	}
	if hasAnyPrefix(aLower, VersionKey) {
		return AttrVersion, versionFlist(fl), reverse
	}

	return AttrUnknown, nil, false
}

func nilCmpOrProceed(fallback func(*File, *File) bool) func(*File, *File) bool {
	return func(l, r *File) bool {
		if l == nil {
			return false
		}
		if r == nil {
			return true
		}
		return fallback(l, r)
	}
}

func _lessCmper(_attr attr) func(*File, *File) bool {
	switch _attr {
	case AttrSize:
		return nilCmpOrProceed(func(l, r *File) bool { return l.Size < r.Size })
	case AttrVersion:
		return nilCmpOrProceed(func(l, r *File) bool { return l.Version < r.Version })
	case AttrIsDir:
		return nilCmpOrProceed(func(l, r *File) bool { return !l.IsDir })
	case AttrMd5Checksum:
		return nilCmpOrProceed(func(l, r *File) bool { return l.Md5Checksum < r.Md5Checksum })
	case AttrName:
		return nilCmpOrProceed(func(l, r *File) bool { return collator.CompareString(l.Name, r.Name) < 0 })
	case AttrModTime:
		return nilCmpOrProceed(func(l, r *File) bool { return l.ModTime.Before(r.ModTime) })
	case AttrLastViewedByMeTime:
		return nilCmpOrProceed(func(l, r *File) bool { return l.LastViewedByMeTime.Before(r.LastViewedByMeTime) })
	}

	return nilCmpOrProceed(func(l, r *File) bool { return true })
}

func (g *Commands) sort(fl []*File, attrStrValues ...string) []*File {
	for _, attrStr := range attrStrValues {
		attrEnum, sortInterface, reverse := attrAtoiSorter(attrStr, fl)

		if attrEnum == AttrUnknown {
			g.log.LogErrf("%s is an unknown sort attribute\n", attrStr)
			continue
		}

		if reverse {
			sortInterface = sort.Reverse(sortInterface)
		} else {
			// Stable is needed if more than one sort keyword is used
			sort.Stable(sortInterface)
		}
	}

	return fl
}
