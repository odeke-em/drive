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

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/jwt"

	"github.com/boltdb/bolt"
)

var (
	GDDirSuffix   = ".gd"
	PathSeparator = fmt.Sprintf("%c", os.PathSeparator)

	ErrNoDriveContext      = errors.New("no drive context found; run `drive init` or go into one of the directories (sub directories) that you performed `drive init`")
	ErrDerefNilIndex       = errors.New("cannot dereference a nil index")
	ErrDerefNilDB          = errors.New("cannot dereference a nil db")
	ErrEmptyFileIdForIndex = errors.New("fileId for index must be non-empty")
	ErrNoSuchDbKey         = errors.New("no such db key exists")
	ErrNoSuchDbBucket      = errors.New("no such bucket exists")
)

const (
	IndicesKey = "indices"
	DriveDb    = "drivedb"
)

const (
	O_RWForAll = 0666
)

type Context struct {
	GSAJWTConfig *jwt.Config `json:"gsa_jwt_config,omitempty"`

	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	AbsPath      string `json:"-"`
}

type Index struct {
	FileId      string `json:"id"`
	Etag        string `json:"etag"`
	Md5Checksum string `json:"md5"`
	MimeType    string `json:"mtype"`
	ModTime     int64  `json:"mtime"`
	Version     int64  `json:"version"`
	IndexTime   int64  `json:"itime"`
}

type MountPoint struct {
	CanClean  bool
	Name      string
	AbsPath   string
	MountPath string
}

type Mount struct {
	CreatedMountDir   string
	ShortestMountRoot string
	Points            []*MountPoint
}

func byteify(s string) []byte {
	return []byte(s)
}

func (mpt *MountPoint) mounted() bool {
	// TODO: Find proper scheme for resolving symlinks
	return mpt.CanClean
}

func (mpt *MountPoint) Unmount() error {
	if mpt.mounted() {
		return os.RemoveAll(mpt.MountPath)
	}
	return nil
}

func (c *Context) AbsPathOf(fileOrDirPath string) string {
	return path.Join(c.AbsPath, fileOrDirPath)
}

func (c *Context) Cwd() string {
	cwd, _ := os.Getwd()
	return cwd
}

func (c *Context) Read() error {
	data, err := ioutil.ReadFile(credentialsPath(c.AbsPath))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

func (c *Context) DeserializeIndex(key string) (*Index, error) {
	if creationErr := c.CreateIndicesBucket(); creationErr != nil {
		return nil, creationErr
	}

	db, err := c.OpenDB()
	if err != nil {
		return nil, err
	}

	defer db.Close()

	var data []byte

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(byteify(IndicesKey))
		if bucket == nil {
			return ErrNoSuchDbBucket
		}

		retr := bucket.Get(byteify(key))
		if len(retr) < 1 {
			return ErrNoSuchDbKey
		}
		data = retr
		return nil
	})

	if err != nil {
		return nil, err
	}

	index := Index{}
	err = json.Unmarshal(data, &index)
	return &index, err
}

func (c *Context) ListKeys(dir, bucketName string) (chan string, error) {
	keysChan := make(chan string)
	if err := c.CreateIndicesBucket(); err != nil {
		close(keysChan)
		return keysChan, err
	}

	db, err := c.OpenDB()
	if err != nil {
		close(keysChan)
		return keysChan, err
	}

	go func() {
		defer func() {
			db.Close()
			close(keysChan)
		}()

		db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(byteify(bucketName))
			if bucket == nil {
				return ErrNoSuchDbBucket
			}

			cur := bucket.Cursor()

			for key, _ := cur.First(); key != nil; key, _ = cur.Next() {
				keysChan <- string(key)
			}

			return nil
		})
	}()

	return keysChan, nil
}

func (c *Context) PopIndicesKey(key string) error {
	return c.popDbKey(IndicesKey, key)
}

func (c *Context) popDbKey(bucketName, key string) error {
	db, err := c.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(byteify(IndicesKey))
		if err != nil {
			return err
		}
		if bucket == nil {
			return ErrNoSuchDbBucket
		}

		return bucket.Delete(byteify(key))
	})
}

func (c *Context) RemoveIndex(index *Index, p string) error {
	if index == nil {
		return ErrDerefNilIndex
	}
	if empty(index.FileId) {
		return ErrEmptyFileIdForIndex
	}

	db, err := c.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(byteify(IndicesKey))
		if err != nil {
			return err
		}
		if bucket == nil {
			return ErrNoSuchDbBucket
		}
		return bucket.Delete(byteify(index.FileId))
	})
}

func (c *Context) CreateIndicesBucket() error {
	db, err := c.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(byteify(IndicesKey))
		if err != nil {
			return err
		}
		if bucket == nil {
			return ErrNoSuchDbBucket
		}
		return nil
	})
}

func (c *Context) SerializeIndex(index *Index) error {
	data, err := json.Marshal(index)
	if err != nil {
		return err
	}

	db, err := c.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(byteify(IndicesKey))
		if err != nil {
			return err
		}
		if bucket == nil {
			return ErrNoSuchDbBucket
		}
		return bucket.Put(byteify(index.FileId), data)
	})
}

func (c *Context) Write() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(credentialsPath(c.AbsPath), data, 0600)
}

func (c *Context) DeInitialize(prompter func(...interface{}) bool, returnOnAnyError bool) error {
	rootDir := c.AbsPathOf("")
	pathsToRemove := []string{
		credentialsPath(rootDir),
		DbSuffixedPath(rootDir),
	}

	for _, p := range pathsToRemove {
		if !prompter("remove: ", p, ". This operation is permanent (Y/n) ") {
			continue
		}

		rmErr := os.RemoveAll(p)
		if rmErr != nil {
			if returnOnAnyError {
				return rmErr
			}
			fmt.Fprintf(os.Stderr, "deinit.removeAll: %s %v\n", p, rmErr)
		}
	}

	return nil
}

func (c *Context) OpenDB() (*bolt.DB, error) {
	dbPath := DbSuffixedPath(c.AbsPathOf(""))
	db, err := bolt.Open(dbPath, O_RWForAll, nil)
	if err != nil {
		return db, err
	}

	if db == nil {
		return db, ErrDerefNilDB
	}

	return db, nil
}

// Discovers the gd directory, if no gd directory or credentials
// could be found for the path, returns ErrNoContext.
func Discover(currentAbsPath string) (*Context, error) {
	p := currentAbsPath
	found := false
	for {
		info, e := os.Stat(gdPath(p))
		if e == nil && info.IsDir() {
			found = true
			break
		}
		newPath := filepath.Join(p, "..")
		if p == newPath {
			break
		}
		p = newPath
	}

	if !found {
		return nil, ErrNoDriveContext
	}
	context := &Context{AbsPath: p}
	if err := context.Read(); err != nil {
		return nil, err
	}
	return context, nil
}

func Initialize(absPath string) (pathGD string, firstInit bool, c *Context, err error) {
	pathGD = gdPath(absPath)
	sInfo, sErr := os.Stat(pathGD)
	if sErr != nil {
		if os.IsNotExist(sErr) {
			firstInit = true
		} else { // An err not related to path existence
			err = sErr
			return
		}
	}
	if sInfo != nil && !sInfo.IsDir() {
		err = fmt.Errorf("%s is not a directory", pathGD)
		return
	}
	if err = os.MkdirAll(pathGD, 0755); err != nil {
		return
	}
	c = &Context{AbsPath: absPath}
	err = c.Write()
	return
}

func gdPath(absPath string) string {
	return path.Join(absPath, GDDirSuffix)
}

func credentialsPath(absPath string) string {
	return path.Join(gdPath(absPath), "credentials.json")
}

func DbSuffixedPath(dir string) string {
	return path.Join(gdPath(dir), DriveDb)
}

func LeastNonExistantRoot(contextAbsPath string) string {
	last := ""
	p := contextAbsPath
	for p != "" {
		fInfo, _ := os.Stat(p)
		if fInfo != nil {
			break
		}
		last = p
		p, _ = filepath.Split(strings.TrimRight(p, PathSeparator))
	}
	return last
}

func empty(p string) bool {
	return p == ""
}

func MountPoints(contextPath, contextAbsPath string, paths []string, hidden bool) (
	mount *Mount, sources []string) {

	createdMountDir := false
	shortestMountRoot := ""

	_, err := os.Stat(contextAbsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}

		if sRoot := LeastNonExistantRoot(contextAbsPath); sRoot != "" {
			shortestMountRoot = sRoot
			// Link traversal is already implemented, no
			// need to append the shortestMountRoot
			if false {
				sources = append(sources, sRoot)
			}
		}

		if err := os.MkdirAll(contextAbsPath, os.ModeDir|0755); err != nil {
			fmt.Fprintf(os.Stderr, "mountpoint: %v\n", err)
			return
		}

		createdMountDir = true
	}

	var mtPoints []*MountPoint
	visitors := map[string]bool{}

	for _, path := range paths {
		_, visited := visitors[path]
		if visited {
			continue
		}
		visitors[path] = true

		localinfo, err := os.Stat(path)
		if err != nil || localinfo == nil {
			continue
		}

		base := filepath.Base(path)
		if !hidden && strings.HasPrefix(base, ".") {
			continue
		}

		canClean := true
		mountPath := filepath.Join(contextAbsPath, base)
		err = os.Symlink(path, mountPath)

		if err != nil {
			if !os.IsExist(err) {
				continue
			}
			// This is an old symlink probably due to a name clash.
			// TODO: Due to the name clash, find a good name for this symlink.
			canClean = false
		}

		var relPath = ""
		if contextPath == "" {
			relPath = strings.Join([]string{"", base}, "/")
		} else {
			relPath = strings.Join([]string{"", contextPath, base}, "/")
		}

		mtPoints = append(mtPoints, &MountPoint{
			AbsPath:   path,
			CanClean:  canClean,
			MountPath: mountPath,
			Name:      relPath,
		})
	}
	if len(mtPoints) >= 1 {
		mount = &Mount{
			Points: mtPoints,
		}
		if createdMountDir {
			mount.CreatedMountDir = contextAbsPath
			mount.ShortestMountRoot = shortestMountRoot
		}
	}
	return
}
