package store

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"slices"
	"sync"
)

type Item struct {
	fs   FS
	base string
}

func (s *Item) Open(name string) (fs.File, error) {
	return s.fs.Open(path.Join(s.base, name))
}

func (s *Item) ReadDir(name string) ([]fs.DirEntry, error) {
	return s.fs.ReadDir(path.Join(s.base, name))
}

func (s *Item) ReadFile(name string) ([]byte, error) {
	return s.fs.ReadFile(path.Join(s.base, name))
}

type Store struct {
	Items []FS
	lock  sync.RWMutex
}

func (s *Store) Remove(fs FS) {
	for i, item := range s.Items {
		if item == fs {
			slices.Delete(s.Items, i, 1)
			return
		}
	}
}

func (s *Store) Add(fs FS) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Items = append(s.Items, fs)
}

func (s *Store) AddDir(root string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Items = append(s.Items, Dir(root))
}

func (s *Store) AddZip(zip string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Items = append(s.Items, &ZipFS{Filename: zip})
}

func (s *Store) AddEmbedFS(fs *embed.FS) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Items = append(s.Items, fs)
}

func (s *Store) Open(name string) (http.File, error) {
	file, err := s.OpenFile(name)
	if err != nil {
		return nil, err
	}
	return HttpFile(file), err
}

func (s *Store) OpenFile(name string) (file fs.File, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, item := range s.Items {
		//查找文件
		file, err = item.Open(name)
		if err == nil {
			fi, e := file.Stat()
			if e != nil {
				return nil, e
			}
			if fi != nil && !fi.IsDir() {
				return
			}
		}
	}
	return nil, os.ErrNotExist
}
