package store

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
)

type httpFile struct {
	fs.File
}

func (f *httpFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, errors.New("does not implement Readdir()")
}

func (f *httpFile) Seek(offset int64, whence int) (int64, error) {
	s, ok := f.File.(io.Seeker)
	if !ok {
		return 0, errors.New("does not implement Seek()")
	}
	return s.Seek(offset, whence)
}

func HttpFile(file fs.File) http.File {
	return &httpFile{File: file}
}
