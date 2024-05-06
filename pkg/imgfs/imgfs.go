package imgfs

import (
	"context"
	"io"
	"os"

	"github.com/athoune/stream-my-root/pkg/backend"
	"github.com/athoune/stream-my-root/pkg/backend/ro"
	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

const (
	rootInode fuseops.InodeID = fuseops.RootInodeID + iota
	imgInode
)

type inodeInfo struct {
	attributes fuseops.InodeAttributes

	// File or directory?
	dir bool

	// For directories, children.
	children []fuseutil.Dirent
}

var gInodeInfo = map[fuseops.InodeID]inodeInfo{
	// root
	rootInode: inodeInfo{
		attributes: fuseops.InodeAttributes{
			Nlink: 1,
			Mode:  0555 | os.ModeDir,
		},
		dir: true,
		children: []fuseutil.Dirent{
			fuseutil.Dirent{
				Offset: 1,
				Inode:  imgInode,
				Name:   "root.img",
				Type:   fuseutil.DT_File,
			},
		},
	},

	// img
	imgInode: inodeInfo{
		attributes: fuseops.InodeAttributes{
			Nlink: 1,
			Mode:  0444,
			Size:  uint64(len("Hello, world!")),
		},
	},
}

type LazyBlockFS struct {
	fuseutil.NotImplementedFileSystem
	backend backend.Backend
}

func New(recipe *blocks.Recipe, reader *blocks.Reader) (fuse.Server, error) {
	backend := ro.NewROBackend(recipe, reader)
	fs := &LazyBlockFS{
		backend: backend,
	}
	return fuseutil.NewFileSystemServer(fs), nil
}

func (l *LazyBlockFS) ReadDir(
	ctx context.Context,
	op *fuseops.ReadDirOp) error {
	// Find the info for this inode.
	info, ok := gInodeInfo[op.Inode]
	if !ok {
		return fuse.ENOENT
	}

	if !info.dir {
		return fuse.EIO
	}

	entries := info.children

	// Grab the range of interest.
	if op.Offset > fuseops.DirOffset(len(entries)) {
		return nil
	}

	entries = entries[op.Offset:]

	// Resume at the specified offset into the array.
	for _, e := range entries {
		n := fuseutil.WriteDirent(op.Dst[op.BytesRead:], e)
		if n == 0 {
			break
		}

		op.BytesRead += n
	}

	return nil
}

func (f *LazyBlockFS) OpenFile(
	ctx context.Context,
	op *fuseops.OpenFileOp) error {
	// Allow opening any file.
	return nil
}

func (f *LazyBlockFS) ReadFile(
	ctx context.Context,
	op *fuseops.ReadFileOp) error {
	// Let io.ReaderAt deal with the semantics.

	var err error
	op.BytesRead, err = f.backend.ReadAt(op.Dst, op.Offset)

	// Special case: FUSE doesn't expect us to return io.EOF.
	if err == io.EOF {
		return nil
	}

	return err
}

func (l *LazyBlockFS) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {
	return nil
}

func (l *LazyBlockFS) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) error {
	// Find the info for this inode.
	info, ok := gInodeInfo[op.Inode]
	if !ok {
		return fuse.ENOENT
	}

	// Copy over its attributes.
	op.Attributes = info.attributes
	if op.Inode == imgInode { // What an ugly hack
		s, err := l.backend.Size()
		if err != nil {
			panic(err)
		}
		op.Attributes.Size = uint64(s)
	}

	return nil
}

func (l *LazyBlockFS) OpenDir(
	ctx context.Context,
	op *fuseops.OpenDirOp) error {
	// Allow opening any directory.
	return nil
}

func findChildInode(
	name string,
	children []fuseutil.Dirent) (fuseops.InodeID, error) {
	for _, e := range children {
		if e.Name == name {
			return e.Inode, nil
		}
	}

	return 0, fuse.ENOENT
}

func (l *LazyBlockFS) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {
	// Find the info for the parent.
	parentInfo, ok := gInodeInfo[op.Parent]
	if !ok {
		return fuse.ENOENT
	}

	// Find the child within the parent.
	childInode, err := findChildInode(op.Name, parentInfo.children)
	if err != nil {
		return err
	}

	// Copy over information.
	op.Entry.Child = childInode
	op.Entry.Attributes = gInodeInfo[childInode].attributes

	return nil
}
