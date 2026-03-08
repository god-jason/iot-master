package store

import (
	"archive/zip"
	"io"
	"io/fs"
)

type zipEntry struct {
	*zip.File
}

func (z *zipEntry) Name() string {
	return z.File.Name
}

func (z *zipEntry) IsDir() bool {
	return z.FileHeader.FileInfo().IsDir()
}

func (z *zipEntry) Type() fs.FileMode {
	return z.FileHeader.FileInfo().Mode()
}

func (z *zipEntry) Info() (fs.FileInfo, error) {
	return z.FileHeader.FileInfo(), nil
}

type ZipFS struct {
	Filename string
	r        *zip.ReadCloser
}

func (z *ZipFS) Open(name string) (file fs.File, err error) {
	if z.r == nil {
		z.r, err = zip.OpenReader(z.Filename)
		if err != nil {
			return
		}
	}

	//打开压缩包内的文件
	return z.r.Open(name)
}

func (z *ZipFS) ReadDir(name string) (entries []fs.DirEntry, err error) {
	for _, f := range z.r.File {
		entries = append(entries, &zipEntry{File: f})
	}
	return
}

func (z *ZipFS) ReadFile(name string) ([]byte, error) {
	file, err := z.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
