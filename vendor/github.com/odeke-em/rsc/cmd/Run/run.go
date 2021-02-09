// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"os/exec"
	"fmt"
	"time"
	"path"
	"strings"
	"strconv"
	"syscall"

	"code.google.com/p/goplan9/plan9/acme"
)

var _ = fmt.Printf

func main() {
	log.SetFlags(0)
	log.SetPrefix("Run: ")

	file := os.Getenv("samfile")
	if file == "" {
		log.Fatal("not running in acme")
	}
	id, _ := strconv.Atoi(os.Getenv("winid"))
	wfile, err := acme.Open(id, nil)
	if err != nil {
		log.Fatal(err)
	}
	wfile.Ctl("put")
	wfile.CloseFiles()

	wname := "/go/run/" + strings.TrimSuffix(path.Base(file), ".go")
	windows, _ := acme.Windows()
	var w *acme.Win
	for _, info := range windows {
		if info.Name == wname {
			ww, err := acme.Open(info.ID, nil)
			if err != nil {
				log.Fatal(err)
			}
			ww.Addr(",")
			ww.Write("data", nil)
			w = ww
			break
		}
	}
	if w == nil {
		ww, err := acme.New()
		if err != nil {
			log.Fatal(err)
		}
		ww.Name(wname)
		w = ww
	}
	w.Ctl("clean")
	defer w.Ctl("clean")

	cmd := exec.Command("go", append([]string{"run", os.Getenv("samfile")}, os.Args[1:]...)...)
	cmd.Stdout = bodyWriter{w}
	cmd.Stderr = cmd.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		w.Fprintf("body", "error starting command: %v\n", err)
		return
	}
	
	//stop := blinker(w)
	w.Ctl("cleartag")
	w.Fprintf("tag", " Kill Stack")

	done := make(chan bool)
	go func() {
		err := cmd.Wait()
		if err != nil {
			w.Fprintf("body", "\nerror running command: %v\n", err)
		}
		//stop <- true
		done <- true
	}()
	
	deleted := make(chan bool, 1)
	go func() {
		for e := range w.EventChan() {
			if e.C2 == 'x' || e.C2 == 'X' {
				switch string(e.Text) {
				case "Del":
					select {
					case deleted <- true:
					default:
					}
					syscall.Kill(-cmd.Process.Pid, 2)
					continue
				case "Kill":
					syscall.Kill(-cmd.Process.Pid, 2)
					continue
				case "Stack":
					syscall.Kill(-cmd.Process.Pid, 3)
					continue
				}
				w.WriteEvent(e)
			}
		}
	}()
	
	<-done
	w.Ctl("cleartag")
	
	select {
	case <-deleted:
		w.Ctl("delete")
	default:
	}
}

type bodyWriter struct {
	w *acme.Win
}

func (w bodyWriter) Write(b []byte) (int, error) {
	return w.w.Write("body", b)
}

func blinker(w *acme.Win) chan bool {
	c := make(chan bool)
	go func() {
		t := time.NewTicker(300 * time.Millisecond)
		defer t.Stop()
		dirty := false
		for {
			select {
			case <-t.C:
				dirty = !dirty
				if dirty {
					w.Ctl("dirty")
				} else {
					w.Ctl("clean")
				}
			case <-c:
				if dirty {
					w.Ctl("clean")
				}
				return
			}
		}
	}()
	return c
}
