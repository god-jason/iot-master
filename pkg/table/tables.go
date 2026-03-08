package table

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/busy-cloud/boat/lib"
)

var tables lib.Map[Table]

//var tables []*Table

func Register(table *Table) {
	//tables = append(tables, table)
	tables.Store(table.Name, table)
}

func Get(name string) (*Table, error) {
	tb := tables.Load(name)
	if tb == nil {
		return nil, errors.New("table not found")
	}
	return tb, nil
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
