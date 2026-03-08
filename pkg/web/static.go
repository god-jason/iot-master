package web

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type item struct {
	fs     http.FileSystem
	path   string
	prefix string
	index  string
}

var items []*item

// Static 静态目录
// path url路径。
// prefix 基本路径（主要用于zip文件）。
// index 默认首页，index.html 代表SPA应用。
func Static(fs http.FileSystem, path, prefix, index string) {
	items = append(items, &item{fs: fs, path: path, prefix: prefix, index: index})
}

func StaticFS(fs fs.FS, path, prefix, index string) {
	items = append(items, &item{fs: http.FS(fs), path: path, prefix: prefix, index: index})
}

func StaticDir(dir string, path, prefix, index string) {
	items = append(items, &item{fs: http.Dir(dir), path: path, prefix: prefix, index: index})
}

func StaticZip(zip string, path, prefix, index string) {
	items = append(items, &item{fs: &ZipFS{Filename: zip}, path: path, prefix: prefix, index: index})
}

func StaticEmbedFS(fs embed.FS, path, prefix, index string) {
	items = append(items, &item{fs: http.FS(fs), path: path, prefix: prefix, index: index})
}

func OpenStaticFile(name string) (file http.File, err error) {
	//低效
	//for _, f := range items {
	//逆序，优先用后来者
	for i := len(items) - 1; i >= 0; i-- {
		f := items[i]
		//fn := path.Join(fbase, name)
		// && !strings.HasPrefix(name, "/$")
		if f.path == "" || f.path != "" && strings.HasPrefix(name, f.path) {
			//去除前缀
			fn := path.Join(f.prefix, strings.TrimPrefix(name, f.path))

			//查找文件
			file, err = f.fs.Open(fn)
			if err == nil {
				fi, _ := file.Stat()
				if fi != nil && !fi.IsDir() {
					return
				}
			}

			//尝试默认页
			if f.index != "" {
				file, err = f.fs.Open(path.Join(f.prefix, f.index))
				if err == nil {
					fi, _ := file.Stat()
					if fi != nil && !fi.IsDir() {
						return
					}
				}
			}

			//return nil, errors.New("not found")
		}
	}
	return nil, errors.New("not found")
}
