package web

import (
	"archive/zip"
	"io/fs"
	"net/http"
)

type zipFile struct {
	fs.File
}

func (f *zipFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f *zipFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

type zipFS struct {
	filename string
	r        *zip.ReadCloser
}

func (z *zipFS) Open(name string) (file http.File, err error) {
	if z.r == nil {
		z.r, err = zip.OpenReader(z.filename)
		if err != nil {
			return
		}
	}

	//打开压缩包内的文件
	f, err := z.r.Open(name)
	if err != nil {
		return nil, err
	}

	return &zipFile{File: f}, nil
}
