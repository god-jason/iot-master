package table

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var tables sync.Map

func Register(table *Table) {
	tables.Store(table.Name, table)
}

func Get(name string) (*Table, error) {
	if v, ok := tables.Load(name); ok {
		return v.(*Table), nil
	}
	return nil, errors.New("表不存在")
}

func Load(path string) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var table Table
	err = json.Unmarshal(buf, &table)
	if err != nil {
		return err
	}

	Register(&table)

	return nil
}

func Scan(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".json" {
			err = Load(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func Range(iterator func(name string, tb *Table) bool) {
	tables.Range(func(key, value any) bool {
		return iterator(key.(string), value.(*Table))
	})
}
