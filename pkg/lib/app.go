package lib

import (
	"os"
	"path/filepath"
	"strings"
)

func AppName() string {
	path, _ := os.Executable()
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func AppDir() string {
	path, _ := os.Executable()
	return filepath.Dir(path)
}
