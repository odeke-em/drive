// +build ignore

package main

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"code.google.com/p/rsc/c2go"
	"code.google.com/p/rsc/cc"
)

// print an error; fprintf is a bad name but helps go vet.
func fprintf(span cc.Span, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s:%d: %s\n", span.Start.File, span.Start.Line, msg)
}

type outFile struct {
	p c2go.Printer
}

// write actual output
func write(prog *cc.Prog, filenames []string, cfg *Config) {
	files := map[string]*outFile{}
	havePkg := map[string]bool{}
	for _, decl := range prog.Decls {
		if decl.GoPackage == "" {
			pkg := findPkg(cfg, decl.Span.Start.File)
			if pkg == "" {
				continue
			}
			decl.GoPackage = pkg
		}
		file := decl.Span.Start.File
		file = strings.TrimSuffix(strings.TrimSuffix(file, ".c"), ".h") + ".go"
		file = filepath.Base(file)
		pkg := decl.GoPackage
		file = pkg + "/" + file
		f := files[file]
		if f == nil {
			f = new(outFile)
			f.p.Package = pkg
			files[file] = f
			f.p.Print("package ", path.Base(pkg), "\n\n")
			if !havePkg[pkg] {
				havePkg[pkg] = true
				f.p.Print(bool2int)
			}
		}

		if cfg.Delete[decl.Name] {
			f.p.Print(decl.Comments.Before)
			f.p.Print(decl.Comments.Suffix, decl.Comments.After)
			continue
		}
		off := len(f.p.Bytes())
		repl, ok := cfg.Replace[decl.Name]
		if !ok {
			repl, ok = cfg.Replace[strings.ToLower(decl.Name)]
		}
		if ok {
			f.p.Print(decl.Comments.Before)
			f.p.Print(repl)
			f.p.Print(decl.Comments.Suffix, decl.Comments.After)
		} else {
			f.p.Print(decl)
		}
		if len(f.p.Bytes()) > off {
			f.p.Print(c2go.Newline)
			f.p.Print(c2go.Newline)
		}
	}

	for pkgfile, f := range files {
		dstfile := filepath.Join(*out, pkgfile)
		buf := f.p.Bytes()
		buf1, err := format.Source(f.p.Bytes())
		if err == nil {
			buf = buf1
		}
		out := string(buf)
		for i, d := range cfg.Diffs {
			if strings.Contains(out, d.Before) {
				out = strings.Replace(out, d.Before, d.After, -1)
				cfg.Diffs[i].Used++
			}
		}
		if err := os.MkdirAll(filepath.Dir(dstfile), 0777); err != nil {
			log.Print(err)
		}
		if err := ioutil.WriteFile(dstfile, []byte(out), 0666); err != nil {
			log.Print(err)
		}
	}
}

var bool2int = `
func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

const (
	fmtLong = 1 << iota
)

`
