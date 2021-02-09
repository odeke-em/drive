// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Simple demo of using package ext2.

// +build ignore

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"code.google.com/p/rsc/ext2"
)

func main() {
	fs, err := ext2.Open("/dev/disk1@0x8ca30600")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(int64(fs.BlockSize) * fs.NumBlock)

	root, err := fs.Root()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(root.Size())

	f, err := root.Lookup("netdisk")
	if err != nil {
		log.Fatal(err)
	}
	f, err = f.Lookup("rsc")
	if err != nil {
		log.Fatal(err)
	}

	dirs, err := f.ReadDir()
	if err != nil {
		log.Fatal(err)
	}
	for _, dir := range dirs {
		fmt.Printf("%#08x %s\n", dir.Inode, dir.Name)
	}

	if len(os.Args) > 1 {
		walk(f, os.Args[1])
	}

	/*
		r, err := root.Open()
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(os.Stdout, r)
		if err != nil {
			log.Fatal(err)
		}
	*/
}

func walk(f *ext2.File, targ string) {
	if f.IsDir() {
		dirs, err := f.ReadDir()
		if err != nil {
			log.Print(err)
			return
		}
		if err := os.MkdirAll(targ, 0777); err != nil {
			log.Print(err)
			return
		}
		os.Chmod(targ, 0777) // make writable for now
		for _, dir := range dirs {
			if dir.Name == "." || dir.Name == ".." {
				continue
			}
			ff, err := f.fs.File(dir.Inode)
			if err != nil {
				log.Print(err)
				continue
			}
			walk(ff, targ+"/"+dir.Name)
		}
		t := f.ModTime()
		os.Chtimes(targ, t, t)
		os.Chmod(targ, uint32(f.ino.Mode&0777))
		return
	}

	m := f.Mode()
	if m&os.ModeSymlink != 0 {
		link, err := f.ReadLink()
		if err != nil {
			log.Printf("%s: %s", targ, err)
			return
		}
		current, _ := os.Readlink(targ)
		if current == link {
			//	log.Printf("skip donelink %s %s %s", targ, m, link)
			return
		}
		os.Remove(targ)
		if err := os.Symlink(link, targ); err != nil {
			log.Printf("%s: %v", targ, err)
			return
		}
		log.Printf("%s link %s", targ, link)
		return
	}

	if m&os.ModeType != 0 {
		log.Printf("skipping special file %s %#x %s", targ, m, m)
		return
	}

	st, err := os.Stat(targ)
	size := f.Size()
	t := f.ModTime()
	if err == nil && st.Size() == size && st.ModTime().Equal(t) {
		//	log.Printf("skip done %s", targ)
		return
	}

	if strings.Contains(targ, ".unison.") {
		log.Printf("skip bad %s", targ)
		return
	}

	r, err := f.Open()
	if err != nil {
		log.Printf("%s: %s", targ, err)
		return
	}

	os.Remove(targ)
	ff, err := os.Create(targ)
	if err != nil {
		log.Print(err)
		return
	}

	n, err := io.Copy(ff, r)
	if err != nil {
		log.Printf("%s: %s", targ, err)
	}
	if n != size {
		log.Printf("%s: expected %d bytes but got %d", targ, size, n)
	}

	ff.Close()

	os.Chtimes(targ, t, t)
	os.Chmod(targ, uint32(f.ino.Mode&0777))

	log.Printf("%s %d %s", targ, size, t)
}
