package tmpfile

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/odeke-em/go-uuid"
)

type readSeekerAt interface {
	io.ReadSeeker
	ReadAt(b []byte, off int64) (int, error)
}

type readSeekerAtWriteCloser interface {
	readSeekerAt
	io.WriteCloser
}

type statInfoer interface {
	Stat() (os.FileInfo, error)
}

type readSeekerCloseWriterStatInfoer interface {
	readSeekerAtWriteCloser
	statInfoer
}

var _ readSeekerCloseWriterStatInfoer = &TmpFile{}

type TmpFile struct {
	tmpf          *os.File
	dir           string
	path          string
	doneOnce      sync.Once
	ownCreatedDir bool
}

func (tf TmpFile) Done() error {
	var err error
	tf.doneOnce.Do(func() {
		_ = tf.tmpf.Close()
		targetPath := tf.tmpf.Name()
		if tf.ownCreatedDir {
			// Since we are removing the entire dir,
			// any file in the dir will be cleaned out.
			targetPath = tf.dir
		}
		err = os.RemoveAll(targetPath)
	})
	return err
}

type Context struct {
	Suffix            string
	Dir               string
	CreateIsolatedDir bool
	// NoOverrideIfSuffixEmpty if set signifies that the caller does not want
	// any suffix to be created for them if no suffix was previously provided,
	// that is they want to main the empty suffix.
	NoOverrideIfSuffixEmpty bool
}

func New(ctx *Context) (*TmpFile, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context passed in")
	}

	dir := ctx.Dir
	var abortFn func() error
	ownCreatedDir := false
	if ctx.CreateIsolatedDir {
		dirSuffix := dir
		if dirSuffix == "" {
			dirSuffix = uuid.UUID4().String()
		}
		dir = filepath.Join(os.Getenv("TMPDIR"), dirSuffix)
	}

	if _, err := os.Stat(dir); err != nil {
		if dir != "" {
			if !os.IsNotExist(err) {
				return nil, err
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
			abortFn = func() error { return os.RemoveAll(dir) }
			ownCreatedDir = true
		}
	}

	suffix := ctx.Suffix
	if ctx.Suffix == "" && !ctx.NoOverrideIfSuffixEmpty {
		suffix = uuid.UUID4().String()
	}

	tmpf, err := ioutil.TempFile(dir, suffix)
	if err != nil {
		if abortFn != nil {
			_ = abortFn()
		}
		return nil, err
	}
	ttmpf := new(TmpFile)
	ttmpf.tmpf = tmpf
	ttmpf.ownCreatedDir = ownCreatedDir
	ttmpf.path = filepath.Join(dir, tmpf.Name())
	ttmpf.dir = dir
	return ttmpf, nil
}

// Invoke this function to create a temp file in an isolated directory
func NewInIsolatedDir() (*TmpFile, error) {
	return New(&Context{CreateIsolatedDir: true})
}

func (tf *TmpFile) Write(b []byte) (int, error) {
	return tf.tmpf.Write(b)
}

func (tf *TmpFile) Close() error {
	return tf.tmpf.Close()
}

func (tf *TmpFile) Name() string {
	return tf.tmpf.Name()
}

func (tf *TmpFile) Path() string {
	return tf.path
}

func (tf *TmpFile) ReadAt(b []byte, off int64) (int, error) {
	return tf.tmpf.ReadAt(b, off)
}

func (tf *TmpFile) Read(b []byte) (int, error) {
	return tf.tmpf.Read(b)
}

func (tf *TmpFile) Seek(offset int64, whence int) (ret int64, err error) {
	return tf.tmpf.Seek(offset, whence)
}

func (tf *TmpFile) Stat() (os.FileInfo, error) {
	return tf.tmpf.Stat()
}

func (tf *TmpFile) Dir() string {
	return tf.dir
}
