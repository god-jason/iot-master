package store

import (
	"io/fs"
	"os"
	"path"
)

type Dir string

func (d Dir) Open(name string) (fs.File, error) {
	return os.Open(path.Join(string(d), name))
}

func (d Dir) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(path.Join(string(d), name))
}

func (d Dir) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(path.Join(string(d), name))
}
