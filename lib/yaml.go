package lib

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func LoadYaml(filename string, cfg any) error {
	y, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer y.Close()

	d := yaml.NewDecoder(y)
	err = d.Decode(cfg)
	if err != nil {
		return err
	}
	return nil
}

func StoreYaml(filename string, cfg any) error {
	_ = os.MkdirAll(filepath.Dir(filename), os.ModePerm)         //创建目录
	y, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755) //os.Create(filename)
	if err != nil {
		return err
	}
	defer y.Close()

	e := yaml.NewEncoder(y)
	defer e.Close()

	err = e.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}
