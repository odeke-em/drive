// Asmlint looks for errors in and sometimes rewrites assembly source files.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	warnCall  = flag.Bool("warncall", false, "warn about CALLs")
	fixOffset = flag.Bool("fixoffset", false, "fix offsets")
	fixNames  = flag.Bool("fixnames", false, "fix variable names")
	useInt64  = flag.Bool("int64", false, "assume 64-bit int on amd64")
	fixInt64  = flag.Bool("fixint64", false, "assume converting to 64-bit int on amd64")

	diffMode = flag.Bool("d", false, "diff")
	verbose  = flag.Bool("v", false, "verbose")

	fset = token.NewFileSet()

	intKind = "L"
	intSize = 4
	ptrKind = "L"
	ptrSize = 4

	newIntSize = intSize
)

func init() {
	flag.StringVar(&build.Default.GOARCH, "goarch", build.Default.GOARCH, "goarch")
}

func main() {
	flag.Parse()
	if build.Default.GOARCH == "amd64" {
		ptrKind = "Q"
		ptrSize = 8
		if *useInt64 {
			intSize = 8
			newIntSize = intSize
			intKind = "Q"
		}
		if *fixInt64 {
			newIntSize = 8
			*fixOffset = true
		}
	}
	args := flag.Args()
	if len(args) == 0 {
		p, err := build.ImportDir(".", 0)
		if err != nil {
			log.Fatal(err)
		}
		check(p)
	} else {
		for _, arg := range args {
			p, err := build.Import(arg, "", 0)
			if err != nil {
				log.Print(err)
				continue
			}
			check(p)
		}
	}
}

type varInfo struct {
	name   string
	kind   string
	typ    string
	off    int
	newOff int
}

type funcInfo struct {
	vars        map[string]*varInfo
	varByOffset map[int]*varInfo
	size        int
}

var funcByName map[string]*funcInfo

var (
	call  = regexp.MustCompile(`\b(CALL|BL)\b`)
	text  = regexp.MustCompile(`\bTEXT\b.*·([^\(]+)\(SB\)(?:\s*,\s*([0-9]+))?(?:\s*,\s*\$([0-9]+)(?:-([0-9]+))?)?`)
	globl = regexp.MustCompile(`\b(DATA|GLOBL)\b`)
	fp    = regexp.MustCompile(`([a-zA-Z0-9_\xFF-\x{10FFFF}]+)(?:\+([0-9]+))\(FP\)`)
	fp2   = regexp.MustCompile(`[^+\-0-9]](([0-9]+)\(FP\))`)
	inst  = regexp.MustCompile(`^\s*(?:[A-Z0-9a-z_]+:)?\s*([A-Z]+)\s*([^,]*)(?:,\s*(.*))?`)
)

func check(p *build.Package) {
	if len(p.SFiles) == 0 {
		return
	}

	useFile := func(fi os.FileInfo) bool {
		name := fi.Name()
		for _, f := range p.GoFiles {
			if name == f {
				return true
			}
		}
		for _, f := range p.CgoFiles {
			if name == f {
				return true
			}
		}
		return false
	}
	pkgs, err := parser.ParseDir(fset, p.Dir, useFile, 0)
	if err != nil {
		log.Fatalf("parsing Go code: %v", err)
	}

	funcByName := make(map[string]*funcInfo)
	pkg := pkgs[p.Name]
	if pkg == nil {
		log.Printf("no package %s in %s", p.Name, p.Dir)
		return
	}

	for filename, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Body == nil {
					funcByName[decl.Name.Name] = newFuncInfo(filename, decl)
				}
			}
		}
	}

	for _, sfile := range p.SFiles {
		var buf bytes.Buffer
		var curFunc *funcInfo
		path := filepath.Join(p.Dir, sfile)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		lines := strings.SplitAfter(string(data), "\n")
		for i, line := range lines {
			if *warnCall && call.FindString(line) != "" {
				fmt.Printf("%s:%d: CALL instruction\n", path, i+1)
			}
			if m := text.FindStringSubmatch(line); m != nil {
				curFunc = funcByName[m[1]]
				if curFunc != nil && m[2] != "7" {
					size, _ := strconv.Atoi(m[4])
					if size != curFunc.size {
						fmt.Printf("%s:%d: wrong argument size %d - should be %d\n", path, i+1, size, curFunc.size)
					}
				}
				buf.WriteString(line)
				continue
			}
			if globl.FindStringSubmatch(line) != nil {
				buf.WriteString(line)
				curFunc = nil
				continue
			}
			if curFunc == nil {
				buf.WriteString(line)
				continue
			}
			origLine := line
			for _, m := range fp2.FindAllStringSubmatch(line, -1) {
				fmt.Printf("%s:%d: use of unnamed argument %s\n", path, i+1, m[1])
				continue
			}
			for _, m := range fp.FindAllStringSubmatch(line, -1) {
				name := m[1]
				off := 0
				if m[2] != "" {
					off, _ = strconv.Atoi(m[2])
				}
				v := curFunc.vars[name]
				if v == nil {
					v = curFunc.varByOffset[off]
					if *fixNames {
						if v != nil {
							if *verbose {
								fmt.Printf("%s:%d: rename %s to %s\n", path, i+1, name, v.name)
							}
							line = regexp.MustCompile(`\b`+regexp.QuoteMeta(name)+`\b`).ReplaceAllString(line, v.name)
							goto foundVar
						}
					}
					if v != nil {
						fmt.Printf("%s:%d: unknown variable %s - maybe %s?\n", path, i+1, m[0], v.name)
					} else {
						fmt.Printf("%s:%d: unknown variable %s\n", path, i+1, m[0])
					}
					continue
				}
			foundVar:
				instm := inst.FindStringSubmatch(line)
				if instm == nil {
					fmt.Printf("%s:%d: cannot find opcode\n", path, i+1)
					continue
				}
				var instSrc, instDst, instKind string
				switch as := instm[1]; build.Default.GOARCH + "." + as {
				case "386.FMOVDP":
					instSrc = "Q"
				case "386.FMOVLP":
					instSrc = "Q"
					instDst = "L"
				case "arm.MOVD":
					instSrc = "Q"
				case "arm.MOVW":
					instSrc = "L"
				case "arm.MOVH", "arm.MOVHU":
					instSrc = "W"
				case "arm.MOVB", "arm.MOVBU":
					instSrc = "B"
				default:
					if build.Default.GOARCH != "arm" {
						if strings.HasSuffix(as, "SD") {
							instSrc = "Q"
							break
						}
						if strings.HasPrefix(as, "SET") {
							instSrc = "B"
							break
						}
						switch as[len(as)-1] {
						case 'B', 'W', 'L', 'Q':
							instSrc = as[len(as)-1:]
						case 'D':
							instSrc = "Q"
						}
					}
				}
				if instDst == "" {
					instDst = instSrc
				}
				if strings.Index(origLine, m[0]) > strings.Index(origLine, ",") {
					instKind = instDst
				} else {
					instKind = instSrc
				}
				newExtra := 0
				switch v.kind {
				case "B", "W", "L", "Q":
					halfMove := build.Default.GOARCH != "amd64" && instKind == "L" && v.kind == "Q"
					if off != v.off && (!halfMove || off != v.off+4) {
						fmt.Printf("%s:%d: invalid offset %s - expected +%d\n", path, i+1, m[0], v.off)
					}
					if instKind != v.kind && !halfMove {
						fmt.Printf("%s:%d: invalid %s of %s (%s)\n", path, i+1, instm[1], m[0], v.typ)
					}
				case "string":
					switch off - v.off {
					default:
						fmt.Printf("%s:%d: invalid offset %s (%d into string)\n", path, i+1, m[0], off-v.off)
					case 0:
						if instKind != ptrKind {
							fmt.Printf("%s:%d: invalid %s of %s (string base ptr)\n", path, i+1, instm[1], m[0])
						}
					case ptrSize:
						if instKind != intKind {
							fmt.Printf("%s:%d: invalid %s of %s (string len)\n", path, i+1, instm[1], m[0])
						}
					}
				case "slice":
					switch off - v.off {
					default:
						fmt.Printf("%s:%d: invalid offset %s (%d into slice)\n", path, i+1, m[0], off-v.off)
					case 0:
						if instKind != ptrKind {
							fmt.Printf("%s:%d: invalid %s of %s (slice base ptr)\n", path, i+1, instm[1], m[0])
						}
					case ptrSize:
						if instKind != intKind {
							fmt.Printf("%s:%d: invalid %s of %s (slice len)\n", path, i+1, instm[1], m[0])
						}
					case ptrSize + intSize:
						if instKind != intKind {
							fmt.Printf("%s:%d: invalid %s of %s (slice cap)\n", path, i+1, instm[1], m[0])
						}
						newExtra += newIntSize - intSize
					}
				}
				if *verbose {
					fmt.Printf("%s:%d: checked %s; newoff=%d\n", path, i+1, m[0], v.newOff+newExtra)
				}
				if *fixOffset && v.newOff+newExtra != v.off {
					off += v.newOff + newExtra - v.off
					newArg := fmt.Sprintf("%s+%d(FP)", v.name, off)
					re := `\b` + regexp.QuoteMeta(m[0])
					line = regexp.MustCompile(re).ReplaceAllString(line, newArg)
					if *verbose {
						fmt.Printf("%s:%d: %s -> %s\n", path, i+1, m[0], newArg)
					}
				}
			}
			buf.WriteString(line)
		}

		ndata := buf.Bytes()

		if !bytes.Equal(data, ndata) {
			if *diffMode {
				out, err := diff(data, ndata)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%s\n%s", path, out)
				continue
			}
			fmt.Printf("%s\n", path)
			err := ioutil.WriteFile(path, ndata, 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func newFuncInfo(file string, decl *ast.FuncDecl) *funcInfo {
	if decl.Recv != nil {
		log.Fatalf("%s: %s: assembly receivers not supported", file, decl.Name.Name)
	}

	funci := &funcInfo{
		vars:        map[string]*varInfo{},
		varByOffset: map[int]*varInfo{},
	}

	off := 0
	newOff := 0
	doparams := func(list []*ast.Field) {
		for i, f := range list {
			names := f.Names
			var align, size, newAlign, newSize int
			var kind string
			typ := gofmt(f.Type)
			switch t := f.Type.(type) {
			case *ast.ChanType, *ast.FuncType, *ast.MapType, *ast.StarExpr:
				kind = ptrKind
				align = ptrSize
				size = ptrSize
			default:
				switch typ {
				default:
					log.Fatalf("%s: %s: unknown type %s", file, decl.Name.Name, typ)
				case "int8", "uint8", "byte", "bool":
					kind = "B"
					align = 1
					size = 1
				case "int16", "uint16":
					kind = "W"
					align = 2
					size = 2
				case "int32", "uint32", "float32":
					kind = "L"
					align = 4
					size = 4
				case "int64", "uint64", "float64":
					kind = "Q"
					align = ptrSize
					size = 8
				case "int", "uint":
					kind = intKind
					size = intSize
					align = intSize
					newSize = newIntSize
					newAlign = newIntSize
				case "uintptr", "Word", "Errno", "unsafe.Pointer":
					kind = ptrKind
					size = ptrSize
					align = ptrSize
				case "string":
					kind = "string"
					size = ptrSize * 2
					align = ptrSize
				}
			case *ast.InterfaceType:
				kind = "interface"
				align = ptrSize
				size = 2 * ptrSize
			case *ast.ArrayType:
				if t.Len == nil {
					kind = "slice"
					align = ptrSize
					size = ptrSize + 2*intSize
					newSize = ptrSize + 2*newIntSize
					break
				}
				log.Fatalf("%s: %s: array type not supported", file, decl.Name.Name)
			case *ast.StructType:
				log.Fatalf("%s: %s: struct type not supported", file, decl.Name.Name)
			}

			if align == 0 {
				log.Fatalf("%s: %s: 0 alignment for %s", file, decl.Name.Name, typ)
			}
			if newAlign == 0 {
				newAlign = align
			}
			if newSize == 0 {
				newSize = size
			}

			off += -off & (align - 1)
			newOff += -newOff & (newAlign - 1)
			if len(names) == 0 {
				name := "_"
				if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 && &list[0] == &decl.Type.Results.List[0] && i == 0 {
					name = "r"
				}
				names = []*ast.Ident{{Name: name}}
			}
			for _, id := range names {
				name := id.Name
				v := &varInfo{
					name:   name,
					kind:   kind,
					typ:    typ,
					off:    off,
					newOff: newOff,
				}
				funci.vars[name] = v
				for j := 0; j < size; j++ {
					funci.varByOffset[v.off+j] = v
				}
				off += size
				newOff += newSize
			}
		}
	}

	doparams(decl.Type.Params.List)
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		off += -off & (ptrSize - 1)
		newOff += -newOff & (ptrSize - 1)
		doparams(decl.Type.Results.List)
	}
	funci.size = off

	return funci
}

const (
	tabWidth    = 4
	parserMode  = parser.ParseComments
	printerMode = printer.TabIndent | printer.UseSpaces
)

var printConfig = &printer.Config{
	Mode:     printerMode,
	Tabwidth: tabWidth,
}

var gofmtBuf bytes.Buffer

func gofmt(n interface{}) string {
	gofmtBuf.Reset()
	err := printConfig.Fprint(&gofmtBuf, fset, n)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return gofmtBuf.String()
}

func diff(b1, b2 []byte) (data []byte, err error) {
	f1, err := ioutil.TempFile("", "asmlint")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "asmlint")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.Write(b1)
	f2.Write(b2)

	data, err = exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return
}
