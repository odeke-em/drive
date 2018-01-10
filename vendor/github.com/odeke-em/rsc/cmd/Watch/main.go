package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"code.google.com/p/goplan9/plan9/acme"
)

var args []string
var win *acme.Win
var needrun = make(chan bool, 1)

var kq struct {
	fd   int
	dir  *os.File
	m    map[string]*os.File
	name map[int]string
}

func kadd(fd int) {
	kbuf := make([]syscall.Kevent_t, 1)
	kbuf[0] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_VNODE,
		Flags:  syscall.EV_ADD | syscall.EV_RECEIPT | syscall.EV_ONESHOT,
		Fflags: syscall.NOTE_DELETE | syscall.NOTE_EXTEND | syscall.NOTE_WRITE,
	}
	n, err := syscall.Kevent(kq.fd, kbuf[:1], kbuf[:1], nil)
	if err != nil {
		log.Fatalf("kevent: %v", err)
	}
	ev := &kbuf[0]
	if n != 1 || (ev.Flags&syscall.EV_ERROR) == 0 || int(ev.Ident) != int(fd) || int(ev.Filter) != syscall.EVFILT_VNODE {
		log.Fatal("kqueue phase error")
	}
	if ev.Data != 0 {
		log.Fatalf("kevent: kqueue error %s", syscall.Errno(ev.Data))
	}
}

func main() {
	flag.Parse()
	args = flag.Args()

	var err error
	win, err = acme.New()
	if err != nil {
		log.Fatal(err)
	}
	pwd, _ := os.Getwd()
	win.Name(pwd + "/+watch")
	win.Ctl("clean")
	win.Fprintf("tag", "Get ")
	needrun <- true
	go events()
	go runner()

	kq.fd, err = syscall.Kqueue()
	if err != nil {
		log.Fatal(err)
	}
	kq.m = make(map[string]*os.File)
	kq.name = make(map[int]string)

	dir, err := os.Open(".")
	if err != nil {
		log.Fatal(err)
	}
	kq.dir = dir
	kadd(int(dir.Fd()))
	readdir := true

	for {
		if readdir {
			kq.dir.Seek(0, 0)
			names, err := kq.dir.Readdirnames(-1)
			if err != nil {
				log.Fatalf("readdir: %v", err)
			}
			for _, name := range names {
				if kq.m[name] != nil {
					continue
				}
				f, err := os.Open(name)
				if err != nil {
					continue
				}
				kq.m[name] = f
				fd := int(f.Fd())
				kq.name[fd] = name
				kadd(fd)
			}
		}

		kbuf := make([]syscall.Kevent_t, 1)
		var n int
		for {
			n, err = syscall.Kevent(kq.fd, nil, kbuf[:1], nil)
			if err == syscall.EINTR {
				continue
			}
			break
		}
		if err != nil {
			log.Fatalf("kevent wait: %v", err)
		}
		ev := &kbuf[0]
		if n != 1 || int(ev.Filter) != syscall.EVFILT_VNODE {
			log.Fatal("kqueue phase error")
		}

		select {
		case needrun <- true:
		default:
		}

		fd := int(ev.Ident)
		readdir = fd == int(kq.dir.Fd())
		time.Sleep(100 * time.Millisecond)
		kadd(fd)
	}
}

func events() {
	for e := range win.EventChan() {
		switch e.C2 {
		case 'x', 'X': // execute
			if string(e.Text) == "Get" {
				select {
				case needrun <- true:
				default:
				}
				continue
			}
			if string(e.Text) == "Del" {
				win.Ctl("delete")
			}
		}
		win.WriteEvent(e)
	}
	os.Exit(0)
}

var run struct {
	sync.Mutex
	id int
}

func runner() {
	var lastcmd *exec.Cmd
	for _ = range needrun {
		run.Lock()
		run.id++
		id := run.id
		run.Unlock()
		if lastcmd != nil {
			lastcmd.Process.Kill()
		}
		lastcmd = nil
		cmd := exec.Command(args[0], args[1:]...)
		r, w, err := os.Pipe()
		if err != nil {
			log.Fatal(err)
		}
		win.Addr(",")
		win.Write("data", nil)
		win.Ctl("clean")
		win.Fprintf("body", "$ %s\n", strings.Join(args, " "))
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Start(); err != nil {
			r.Close()
			w.Close()
			win.Fprintf("body", "%s: %s\n", strings.Join(args, " "), err)
			continue
		}
		lastcmd = cmd
		w.Close()
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := r.Read(buf)
				if err != nil {
					break
				}
				run.Lock()
				if id == run.id {
					win.Write("body", buf[:n])
				}
				run.Unlock()
			}
			if err := cmd.Wait(); err != nil {
				run.Lock()
				if id == run.id {
					win.Fprintf("body", "%s: %s\n", strings.Join(args, " "), err)
				}
				run.Unlock()
			}
			win.Fprintf("body", "$\n")
			win.Fprintf("addr", "#0")
			win.Ctl("dot=addr")
			win.Ctl("show")
			win.Ctl("clean")
		}()
	}
}
