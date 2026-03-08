package web

import (
	"archive/zip"
	"io/fs"
	"net/http"
)

type ZipFile struct {
	fs.File
}

func (f *ZipFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f *ZipFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

type ZipFS struct {
	Filename string
	r        *zip.ReadCloser
}

func (z *ZipFS) Open(name string) (file http.File, err error) {
	if z.r == nil {
		z.r, err = zip.OpenReader(z.Filename)
		if err != nil {
			return
		}
	}

	// 打开压缩包内的文件
	f, err := z.r.Open(name)
	if err != nil {
		return nil, err
	}

	return &ZipFile{File: f}, nil
}
