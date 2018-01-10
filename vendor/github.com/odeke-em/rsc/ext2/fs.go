// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ext2 implements read-only access to EXT2 file systems.
package ext2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	bytesPerSector = 512

	superOff   = 1024   // superblock offset
	superSize  = 1024   // superblock size
	superMagic = 0xEF53 // superblock magic

	minBlockSize = 1024
	maxBlockSize = 4096

	rootInode  = 2
	firstInode = 11

	validFS = 0x0001
	errorFS = 0x0002

	numDirBlocks = 12
	indBlock     = numDirBlocks  // indirect block
	ind2Block    = indBlock + 1  // double-indirect block
	ind3Block    = ind2Block + 1 // triple-indirect block
	numBlocks    = ind3Block + 1

	nameLen = 255

	// permissions in Inode.mode
	iexec  = 00100
	iwrite = 00200
	iread  = 00400
	isvtx  = 01000
	isgid  = 02000
	isuid  = 04000

	// type in Inode.mode
	ifmt   = 0170000
	ififo  = 0010000
	ifchr  = 0020000
	ifdir  = 0040000
	ifblk  = 0060000
	ifreg  = 0100000
	iflnk  = 0120000
	ifsock = 0140000
	ifwht  = 0160000
)

func dirlen(nameLen int) int {
	return (nameLen + 8 + 3) &^ 3
}

type diskSuper struct {
	Ninode         uint32 /* Inodes count */
	Nblock         uint32 /* Blocks count */
	Rblockcount    uint32 /* Reserved blocks count */
	Freeblockcount uint32 /* Free blocks count */
	Freeinodecount uint32 /* Free inodes count */
	Firstdatablock uint32 /* First Data Block */
	Logblocksize   uint32 /* Block size */
	Logfragsize    uint32 /* Fragment size */
	Blockspergroup uint32 /* # Blocks per group */
	Fragpergroup   uint32 /* # Fragments per group */
	Inospergroup   uint32 /* # Inodes per group */
	Mtime          uint32 /* Mount time */
	Wtime          uint32 /* Write time */
	Mntcount       uint16 /* Mount count */
	Maxmntcount    uint16 /* Maximal mount count */
	Magic          uint16 /* Magic signature */
	State          uint16 /* File system state */
	Errors         uint16 /* Behaviour when detecting errors */
	Pad            uint16
	Lastcheck      uint32 /* time of last check */
	Checkinterval  uint32 /* max. time between checks */
	Creatoros      uint32 /* OS */
	Revlevel       uint32 /* Revision level */
	Defresuid      uint16 /* Default uid for reserved blocks */
	Defresgid      uint16 /* Default gid for reserved blocks */

	/* the following are only available with revlevel = 1 */
	Firstino     uint32 /* First non-reserved inode */
	Inosize      uint16 /* size of inode structure */
	Blockgroupnr uint16 /* block group # of this super block */
}

type diskGroup struct {
	Bitblock        uint32 /* Blocks bitmap block */
	Inodebitblock   uint32 /* Inodes bitmap block */
	Inodeaddr       uint32 /* Inodes table block */
	Freeblockscount uint16 /* Free blocks count */
	Freeinodescount uint16 /* Free inodes count */
	Useddirscount   uint16 /* Directories count */
}

const diskGroupSize = 32

type diskInode struct {
	Mode    uint16 /* File mode */
	Uid     uint16 /* Owner Uid */
	Size    uint32 /* Size in bytes */
	Atime   uint32 /* Access time */
	Ctime   uint32 /* Creation time */
	Mtime   uint32 /* Modification time */
	Dtime   uint32 /* Deletion Time */
	Gid     uint16 /* Group Id */
	Nlink   uint16 /* Links count */
	Nblock  uint32 /* Blocks count */
	Flags   uint32 /* File flags */
	Osd1    uint32
	Block   [numBlocks]uint32 /* Pointers to blocks */
	Version uint32            /* File version (for NFS) */
	Fileacl uint32            /* File ACL */
	Diracl  uint32            /* Directory ACL or high size bits */
	Faddr   uint32            /* Fragment address */
}

type diskDirent struct {
	Ino    uint32 /* Inode number */
	Reclen uint16 /* Directory entry length */
	Namlen uint8  /* Name length */
}

const minDirentSize = 4 + 2 + 1 + 1

// An FS represents a file system.
type FS struct {
	BlockSize      int
	NumBlock       int64
	numGroup       uint32
	inodesPerGroup uint32
	blocksPerGroup uint32
	inodesPerBlock uint32
	inodeSize      uint32
	groupAddr      uint32
	descPerBlock   uint32
	firstBlock     uint32

	g   []*diskGroup
	r   io.ReaderAt
	c   io.Closer
	buf []byte

	cache    [16]block
	cacheAge int64
	cacheHit int64
}

type block struct {
	off     int64
	lastUse int64
	buf     []byte
}

// A readerAtOffset wraps a ReaderAt but translates all the I/O by the given offset.
type readerAtOffset struct {
	r   io.ReaderAt
	off int64
}

func (r *readerAtOffset) ReadAt(p []byte, off int64) (n int, err error) {
	return r.r.ReadAt(p, off+r.off)
}

// Open opens the file system in the named file.
//
// If the name contains an @ sign, it is taken to be
// of the form file@offset, where offset is a decimal,
// hexadecimal, or octal number according to its prefix,
// and the file system is assumed to start at the given
// offset in the file instead of at the beginning of the file.
func Open(name string) (*FS, error) {
	var off int64
	if i := strings.Index(name, "@"); i >= 0 {
		v, err := strconv.ParseInt(name[i+1:], 0, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid offset in name %q", name)
		}
		off = v
		name = name[:i]
	}

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	var r io.ReaderAt = f
	if off != 0 {
		r = &readerAtOffset{r, off}
	}

	fs, err := Init(r)
	if err != nil {
		f.Close()
		return nil, err
	}
	fs.c = f
	return fs, nil
}

// Init returns an FS reading from the file system stored in r.
func Init(r io.ReaderAt) (*FS, error) {
	fs := &FS{
		r:         r,
		buf:       make([]byte, 1024),
		BlockSize: superSize,
	}

	var super diskSuper
	if _, err := fs.read(superOff, 0, &super); err != nil {
		return nil, err
	}

	if super.Magic != superMagic {
		return nil, fmt.Errorf("bad magic %#x wanted %#x", super.Magic, superMagic)
	}

	bsize := uint32(minBlockSize << super.Logblocksize)
	fs.BlockSize = int(bsize)
	fs.NumBlock = int64(super.Nblock)
	fs.numGroup = (super.Nblock + super.Blockspergroup - 1) / super.Blockspergroup
	fs.g = make([]*diskGroup, fs.numGroup)
	fs.inodesPerGroup = super.Inospergroup
	fs.blocksPerGroup = super.Blockspergroup
	if super.Revlevel >= 1 {
		fs.inodeSize = uint32(super.Inosize)
	} else {
		fs.inodeSize = 128
	}
	fs.inodesPerBlock = bsize / fs.inodeSize
	if bsize == superOff {
		fs.groupAddr = 2
	} else {
		fs.groupAddr = 1
	}
	fs.descPerBlock = bsize / diskGroupSize
	fs.firstBlock = super.Firstdatablock

	return fs, nil
}

// A File represents a file or directory in a file system.
type File struct {
	fs   *FS
	inum uint32
	ino  diskInode
}

// File returns the file with the given inode number.
func (fs *FS) File(inode uint32) (*File, error) {
	g, ioff, err := fs.igroup(inode)
	if err != nil {
		return nil, err
	}

	addr := int64(fs.BlockSize) * int64(g.Inodeaddr+ioff/fs.inodesPerBlock)
	ivoff := (ioff % fs.inodesPerBlock) * fs.inodeSize

	file := &File{fs: fs, inum: inode}
	if _, err := fs.read(addr, int(ivoff), &file.ino); err != nil {
		return nil, err
	}

	switch file.ino.Mode & ifmt {
	case ififo, ifchr, ifdir, ifblk, ifreg, iflnk, ifsock:
		// okay
	default:
		return nil, fmt.Errorf("invalid inode mode %#x", file.ino.Mode)
	}

	return file, nil
}

func (fs *FS) igroup(inum uint32) (g *diskGroup, ioff uint32, err error) {
	gnum := (inum - 1) / fs.inodesPerGroup
	if gnum >= fs.numGroup {
		return nil, 0, fmt.Errorf("inode number %#x out of range", inum)
	}
	ioff = (inum - 1) % fs.inodesPerGroup
	g, err = fs.group(gnum)
	return
}

func (fs *FS) group(gnum uint32) (g *diskGroup, err error) {
	// cache to avoid repeated loads from disk
	if g := fs.g[gnum]; g != nil {
		return g, nil
	}

	g = new(diskGroup)
	addr := int64(fs.BlockSize) * int64(fs.groupAddr+gnum/fs.descPerBlock)
	voff := gnum % fs.descPerBlock * diskGroupSize
	if _, err := fs.read(addr, int(voff), g); err != nil {
		return nil, err
	}

	if g.Inodeaddr < fs.groupAddr+fs.numGroup/fs.descPerBlock {
		return nil, fmt.Errorf("implausible inode group descriptor at %#x[%d:]: %+v", addr, voff, *g)
	}

	fs.g[gnum] = g
	return g, nil
}

// Size returns the file's size in bytes.
func (f *File) Size() int64 {
	size := int64(f.ino.Size)
	if f.ino.Mode&ifmt == ifreg {
		size |= int64(f.ino.Diracl) << 32
	}
	return size
}

// Mode returns the file's mode.
func (f *File) Mode() os.FileMode {
	mode := os.FileMode(f.ino.Mode & 0777)
	switch f.ino.Mode & ifmt {
	case ififo:
		mode |= os.ModeNamedPipe
	default: // ifchr, ifblk, unknown
		mode |= os.ModeDevice
	case ifdir:
		mode |= os.ModeDir
	case ifreg:
		// no bits
	case iflnk:
		mode |= os.ModeSymlink
	case ifsock:
		mode |= os.ModeSocket
	}
	return mode
}

func (f *File) IsDir() bool {
	return f.ino.Mode&ifmt == ifdir
}

// ModTime returns the file's modification time.
func (f *File) ModTime() time.Time {
	return time.Unix(int64(f.ino.Mtime), 0)
}

// ReadAt implements the io.ReaderAt interface.
func (f *File) ReadAt(buf []byte, off int64) (n int, err error) {
	size := f.Size()
	if off >= size {
		return 0, io.EOF
	}
	if n = len(buf); int64(n) > size-off {
		n = int(size - off)
	}

	lfrag := int(uint32(off) % uint32(f.fs.BlockSize))
	off -= int64(lfrag)
	want := lfrag + n
	rfrag := -want & (f.fs.BlockSize - 1)
	want += rfrag

	offb := uint32(off / int64(f.fs.BlockSize))

	nblock := want / f.fs.BlockSize
	for i := 0; i < nblock; i++ {
		b, err := f.dblock(offb + uint32(i))
		if err != nil {
			return 0, err
		}
		dbuf, err := f.fs.readData(b, 0, nil)
		if err != nil {
			return 0, err
		}
		m := copy(buf, dbuf[lfrag:])
		buf = buf[m:]
		lfrag = 0
	}
	return n, nil
}

// ReadLink returns the symbolic link content of f.
func (f *File) ReadLink() (string, error) {
	if f.ino.Mode&ifmt != iflnk {
		return "", fmt.Errorf("not a symbolic link")
	}

	size := f.ino.Size

	if f.ino.Nblock != 0 {
		if size > uint32(f.fs.BlockSize) {
			return "", fmt.Errorf("invalid symlink size")
		}
		// Symlink fits in one block.
		b, err := f.dblock(0)
		if err != nil {
			return "", err
		}
		buf, err := f.fs.readData(b, 0, nil)
		if err != nil {
			return "", err
		}
		return string(buf[:size]), nil
	}

	if size > 4*numBlocks {
		return "", fmt.Errorf("invalid symlink size")
	}
	var buf [4 * numBlocks]byte
	for i := 0; i < numBlocks; i++ {
		binary.LittleEndian.PutUint32(buf[4*i:], f.ino.Block[i])
	}
	return string(buf[:size]), nil
}

// A Dir represents a directory entry.
type Dir struct {
	Name  string
	Inode uint32
}

func (f *File) walkDir(run func(name []byte, ino uint32) bool) error {
	if f.ino.Mode&ifmt != ifdir {
		return fmt.Errorf("file is not a directory")
	}

	nblock := (int(f.ino.Size) + f.fs.BlockSize - 1) / f.fs.BlockSize
	for i := 0; i < nblock; i++ {
		b, err := f.dblock(uint32(i))
		if err != nil {
			return err
		}
		buf, err := f.fs.readData(b, 0, nil)
		if err != nil {
			return err
		}

		for len(buf) > 0 {
			var de diskDirent
			if err := unpack(buf, &de); err != nil {
				return err
			}
			minLen := minDirentSize
			recLen := int(de.Reclen)
			nameLen := int(de.Namlen)
			if minLen+nameLen > recLen || recLen > len(buf) {
				return fmt.Errorf("corrupt directory entry")
			}
			name := buf[minLen : minLen+nameLen]
			buf = buf[recLen:]

			if de.Ino == 0 {
				continue
			}

			if !run(name, de.Ino) {
				break
			}
		}
	}

	return nil
}

// ReadDir returns all the directory entries in f.
func (f *File) ReadDir() ([]Dir, error) {
	var dirs []Dir
	err := f.walkDir(func(name []byte, ino uint32) bool {
		dirs = append(dirs, Dir{string(name), ino})
		return true
	})
	return dirs, err
}

// Lookup looks up the name in the directory f, returning the corresponding child file.
func (f *File) Lookup(name string) (*File, error) {
	var bino uint32
	bname := []byte(name)
	err := f.walkDir(func(name []byte, ino uint32) bool {
		if bytes.Equal(bname, name) {
			bino = ino
			return false
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	if bino == 0 {
		return nil, fmt.Errorf("file not found")
	}
	return f.fs.File(bino)
}

func (f *File) dblock(block uint32) (uint32, error) {
	b := block

	// direct?
	if b < numDirBlocks {
		return f.ino.Block[b], nil
	}
	b -= numDirBlocks

	// single indirect?
	p := uint32(f.fs.BlockSize) / 4
	if b < p {
		var b1 uint32
		if _, err := f.fs.readData(f.ino.Block[indBlock], int(4*b), &b1); err != nil {
			return 0, err
		}
		return b1, nil
	}
	b -= p

	// double indirect?
	if b < p*p {
		i1 := b / p
		i2 := b % p
		var b1, b2 uint32
		if _, err := f.fs.readData(f.ino.Block[ind2Block], int(4*i1), &b1); err != nil {
			return 0, err
		}
		if _, err := f.fs.readData(b1, int(4*i2), &b2); err != nil {
			return 0, err
		}
		return b2, nil
	}
	b -= p * p

	// triple indirect?
	if b < p*p*p {
		i1 := b / (p * p)
		i2 := b % (p * p) / p
		i3 := b % p
		var b1, b2, b3 uint32
		if _, err := f.fs.readData(f.ino.Block[ind3Block], int(4*i1), &b1); err != nil {
			return 0, err
		}
		if _, err := f.fs.readData(b1, int(4*i2), &b2); err != nil {
			return 0, err
		}
		if _, err := f.fs.readData(b1, int(4*i3), &b3); err != nil {
			return 0, err
		}
		return b3, nil
	}

	return 0, fmt.Errorf("block number %d out of range", block)
}

// Open returns an io.Reader for the file.
func (f *File) Open() (io.Reader, error) {
	if f.ino.Mode&ifmt != ifreg {
		return nil, fmt.Errorf("not a regular file")
	}
	return &fileReader{f, 0}, nil
}

type fileReader struct {
	f   *File
	off int64
}

func (r *fileReader) Read(buf []byte) (n int, err error) {
	n, err = r.f.ReadAt(buf, r.off)
	if n > 0 {
		err = nil
	}
	r.off += int64(n)
	return
}

// Root returns the root directory of the file system.
func (fs *FS) Root() (*File, error) {
	return fs.File(rootInode)
}

const maxCache = 32

func (fs *FS) read(off int64, voff int, val interface{}) ([]byte, error) {
	var buf []byte

	// look in cache
	fs.cacheAge++
	if fs.cacheAge%10000 == 0 {
		fmt.Printf("cache %d/%d\n", fs.cacheHit, fs.cacheAge)
	}

	var oldest *block
	for i := range fs.cache {
		b := &fs.cache[i]
		if b.off == off {
			b.lastUse = fs.cacheAge
			buf = b.buf
			fs.cacheHit++
			goto unpack
		}
		if oldest == nil || b.lastUse < oldest.lastUse {
			oldest = b
		}
	}

	// load into b
	{
		b := oldest
		if len(b.buf) < fs.BlockSize {
			b.buf = make([]byte, fs.BlockSize)
		}
		b.off = 0
		n, err := fs.r.ReadAt(b.buf, off)
		if n < fs.BlockSize {
			if err == nil {
				err = fmt.Errorf("short read")
			}
			return nil, fmt.Errorf("reading %d bytes at offset %#x: %v", len(b.buf), off, err)
		}
		b.off = off
		b.lastUse = fs.cacheAge
		buf = b.buf
	}

unpack:
	if val != nil {
		if err := unpack(buf[voff:], val); err != nil {
			return nil, fmt.Errorf("parsing offset %#x[%d:%d] as %T: %v", off, voff, fs.BlockSize, val, err)
		}
	}

	return buf, nil
}

func (fs *FS) readData(block uint32, voff int, val *uint32) ([]byte, error) {
	b := block

	if b < fs.firstBlock {
		return nil, fmt.Errorf("block number %d out of range", block)
	}
	b -= fs.firstBlock

	g, err := fs.group(b / fs.blocksPerGroup)
	if err != nil {
		return nil, err
	}

	buf, err := fs.read(int64(g.Bitblock)*int64(fs.BlockSize), 0, nil)
	if err != nil {
		return nil, err
	}

	boff := b % fs.blocksPerGroup
	if buf[boff>>3]&(1<<(boff&7)) == 0 {
		return nil, fmt.Errorf("block %d not allocated", block)
	}

	buf, err = fs.read(int64(block)*int64(fs.BlockSize), 0, nil)
	if err != nil {
		return nil, err
	}

	if val != nil {
		*val = binary.LittleEndian.Uint32(buf[voff:])
	}
	return buf, nil
}

func unpack(data []byte, val interface{}) error {
	v := reflect.ValueOf(val).Elem()
	if v.Kind() != reflect.Struct || !v.CanSet() {
		return fmt.Errorf("must unpack into ptr to struct")
	}

	for i := 0; i < v.NumField(); i++ {
		switch p := v.Field(i).Addr().Interface().(type) {
		case *uint8:
			if len(data) < 1 {
				return fmt.Errorf("buffer smaller than data structure")
			}
			*p = data[0]
			data = data[1:]

		case *uint16:
			if len(data) < 2 {
				return fmt.Errorf("buffer smaller than data structure")
			}
			*p = binary.LittleEndian.Uint16(data)
			data = data[2:]

		case *uint32:
			if len(data) < 4 {
				return fmt.Errorf("buffer smaller than data structure")
			}
			*p = binary.LittleEndian.Uint32(data)
			data = data[4:]

		case *[numBlocks]uint32:
			for i := 0; i < numBlocks; i++ {
				if len(data) < 4 {
					return fmt.Errorf("buffer smaller than data structure")
				}
				p[i] = binary.LittleEndian.Uint32(data)
				data = data[4:]
			}

		default:
			return fmt.Errorf("unexpected field type %T", p)
		}
	}

	return nil
}
