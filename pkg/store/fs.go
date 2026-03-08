package store

import (
	"io/fs"
	"path"
)

type FS interface {
	Open(name string) (fs.File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

type prefixedFs struct {
	fs     FS
	prefix string
}

func (t *prefixedFs) Open(name string) (fs.File, error) {
	return t.fs.Open(path.Join(t.prefix, name))
}

func (t *prefixedFs) ReadDir(name string) ([]fs.DirEntry, error) {
	return t.fs.ReadDir(path.Join(t.prefix, name))
}

func (t *prefixedFs) ReadFile(name string) ([]byte, error) {
	return t.fs.ReadFile(path.Join(t.prefix, name))
}

func PrefixFS(fs FS, prefix string) FS {
	return &prefixedFs{
		fs:     fs,
		prefix: prefix,
	}
}
