// +build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"code.google.com/p/rsc/cc"
)

type Config struct {
	Replace   map[string]string
	Delete    map[string]bool
	ForceType map[string]*cc.Type
	Len       map[string]string // map T.n to T.p where t.n = len(t.p)
	StopFlow  map[string]bool
	Packages  []Package
	Diffs     []Diff
	Exports   []string

	// Derived during analysis
	TopDecls []*cc.Decl
}

type Diff struct {
	Line   string
	Before string
	After  string
	Used   int
}

type Package struct {
	Pattern    string
	ImportPath string
}

func readConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &Config{
		Replace:   make(map[string]string),
		Delete:    make(map[string]bool),
		ForceType: make(map[string]*cc.Type),
		Len:       make(map[string]string),
		StopFlow:  make(map[string]bool),
	}

	r := bufio.NewReader(f)
	lineno := 0
	for {
		s, err := r.ReadString('\n')
		lineno++
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		var buf, buf2 bytes.Buffer
		line := s
		s = strings.TrimSpace(s)
		switch {
		default:
			return nil, fmt.Errorf("unknown directive: %v", strings.TrimSuffix(line, "\n"))

		case s == "", strings.HasPrefix(s, "//"):
			continue

		case strings.HasPrefix(s, "func "), strings.HasPrefix(s, "type "):
			buf.Reset()
			buf.WriteString(line)
			if strings.HasSuffix(s, "{") {
				for {
					s, err := r.ReadString('\n')
					lineno++
					buf.WriteString(s)
					if s == "}\n" {
						break
					}
					if err != nil {
						if err == io.EOF {
							err = fmt.Errorf("unexpected EOF reading func body")
						}
						return nil, err
					}
				}
			}
			name := strings.Fields(s)[1]
			if i := strings.Index(name, "("); i >= 0 {
				name = name[:i]
			}
			cfg.Replace[name] = buf.String()

		case strings.HasPrefix(s, "delete "):
			for _, f := range strings.Fields(s)[1:] {
				cfg.Delete[f] = true
			}

		case strings.HasPrefix(s, "package "):
			fields := strings.Fields(s)[1:]
			pkg, fields := fields[0], fields[1:]
			if len(fields) == 0 {
				fields = append(fields, "")
			}
			for _, f := range fields {
				cfg.Packages = append(cfg.Packages, Package{Pattern: f, ImportPath: pkg})
			}

		case strings.HasPrefix(s, "export "):
			cfg.Exports = append(cfg.Exports, strings.Fields(s)[1:]...)

		case strings.HasPrefix(s, "stopflow "):
			for _, f := range strings.Fields(s)[1:] {
				cfg.StopFlow[f] = true
			}

		case strings.HasPrefix(s, "uselen "):
			fields := strings.Fields(s)
			if len(fields) != 3 {
				fmt.Fprintf(os.Stderr, "%s:%d: invalid uselen: %s\n", file, lineno, strings.TrimSuffix(line, "\n"))
				continue
			}
			cfg.Len[fields[1]] = fields[2]

		case s == "diff {":
			buf.Reset()
			buf2.Reset()
			fileline := fmt.Sprintf("%s:%d", file, lineno)
			for {
				s, err := r.ReadString('\n')
				lineno++
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}
				if s == "}\n" {
					break
				}
				switch {
				case strings.HasPrefix(s, "+"):
					s = strings.TrimPrefix(s[1:], " ")
					buf2.WriteString(s)

				case strings.HasPrefix(s, "-"):
					s = strings.TrimPrefix(s[1:], " ")
					buf.WriteString(s)

				case strings.HasPrefix(s, " "), strings.HasPrefix(s, "\t"), s == "\n":
					s = strings.TrimPrefix(strings.TrimPrefix(s, " "), " ")
					buf.WriteString(s)
					buf2.WriteString(s)

				default:
					return nil, fmt.Errorf("unexpected line in diff: %v\n", strings.TrimSuffix(line, "\n"))
				}
			}
			cfg.Diffs = append(cfg.Diffs, Diff{
				Line:   fileline,
				Before: buf.String(),
				After:  buf2.String(),
			})
		}
	}

	return cfg, nil
}

func findPkg(cfg *Config, file string) string {
	best := "main"
	for _, p := range cfg.Packages {
		if strings.HasSuffix(p.Pattern, "/") {
			if strings.Contains(file, p.Pattern) {
				best = p.ImportPath
			}
		} else {
			if strings.HasSuffix(file, p.Pattern) {
				best = p.ImportPath
			}
		}
	}
	if best == "none" {
		best = ""
	}
	return best
}
