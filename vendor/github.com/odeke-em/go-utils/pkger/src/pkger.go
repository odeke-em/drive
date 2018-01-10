package pkger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var prettyGitHashGetArgs = []string{"log", "-n", "1", "--pretty=format:'%H'"}

type PkgInfo struct {
	CommitHash string
	GoVersion  string
	OsInfo     string
	BuildTime  string
}

func (p *PkgInfo) String() string {
	return fmt.Sprintf("Commit Hash: %s\nGo Version: %s\nOS: %s\nBuildTime: %v",
		p.CommitHash, p.GoVersion, p.OsInfo, p.BuildTime)
}

func goSrcify(p string) string {
	goPath := os.Getenv("GOPATH")
	return filepath.Join(goPath, "src", p)
}

var GoSrcify = goSrcify

func Recon(pkgPath string) (pkgInfo *PkgInfo, err error) {
	var gitProgPath, pkgPathAbs string

	// Firstly look up if they've got Git
	gitProgPath, err = exec.LookPath("git")
	if err != nil {
		return
	}

	pkgPathAbs = goSrcify(pkgPath)

	args := []string{gitProgPath}
	args = append(args, prettyGitHashGetArgs...)

	cmd := exec.Cmd{
		Args: args,
		Dir:  pkgPathAbs,
		Path: gitProgPath,
	}

	var output []byte
	if output, err = cmd.Output(); err != nil {
		return
	}

	pkgInfo = &PkgInfo{
		CommitHash: string(output),
		OsInfo:     strings.Join([]string{runtime.GOOS, runtime.GOARCH}, "/"),
		GoVersion:  runtime.Version(),
		BuildTime:  time.Now().String(),
	}

	return
}
