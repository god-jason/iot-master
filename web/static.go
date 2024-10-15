package web

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

var Static FileSystem

type fsItem struct {
	fs    http.FileSystem
	path  string
	base  string
	index string
}

type FileSystem struct {
	items []*fsItem
	//items map[string]*fsItem
}

func (f *FileSystem) Put(path string, fs http.FileSystem, base string, index string) {
	f.items = append(f.items, &fsItem{fs: fs, path: path, base: base, index: index})
}

func (f *FileSystem) PutFS(path string, fs fs.FS, base string, index string) {
	f.items = append(f.items, &fsItem{fs: http.FS(fs), path: path, base: base, index: index})
}

func (f *FileSystem) PutDir(path string, dir string, base string, index string) {
	f.items = append(f.items, &fsItem{fs: http.Dir(dir), path: path, base: base, index: index})
}

func (f *FileSystem) PutZip(path string, zip string, base string, index string) {
	f.items = append(f.items, &fsItem{fs: &zipFS{filename: zip}, path: path, base: base, index: index})
}

func (f *FileSystem) PutEmbedFS(path string, fs embed.FS, base string, index string) {
	f.items = append(f.items, &fsItem{fs: http.FS(fs), path: path, base: base, index: index})
}

func (f *FileSystem) Open(name string) (file http.File, err error) {
	//低效
	for _, ff := range f.items {
		//fn := path.Join(ff.base, name)
		// && !strings.HasPrefix(name, "/$")
		if ff.path == "" || ff.path != "" && strings.HasPrefix(name, ff.path) {
			//去除前缀
			fn := path.Join(ff.base, strings.TrimPrefix(name, ff.path))

			//查找文件
			file, err = ff.fs.Open(fn)
			if file != nil {
				fi, _ := file.Stat()
				if !fi.IsDir() {
					return
				}
			}

			//尝试默认页
			if ff.index != "" {
				file, err = ff.fs.Open(path.Join(ff.base, ff.index))
				if file != nil {
					fi, _ := file.Stat()
					if !fi.IsDir() {
						return
					}
				}
			}

			return nil, errors.New("not found")
		}
	}
	return nil, errors.New("not found")
}
