package app

import (
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"os"
	"path/filepath"
)

const licenseRoot = "licenses"
const licenseExt = ".lic"

var licenses lib.Map[License]

func ParseLicense(raw string) (*License, error) {
	var lic License
	err := lic.Decode(raw)
	if err != nil {
		return nil, err
	}

	err = lic.Verify(pubKey)
	if err != nil {
		return nil, err
	}

	licenses.Store(lic.AppId, &lic)

	return &lic, nil
}

func LoadLicense(name string) (*License, error) {
	buf, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return ParseLicense(string(buf))
}

func LoadLicenses() error {
	_ = os.MkdirAll(licenseRoot, 0755)
	entries, err := os.ReadDir(licenseRoot)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != licenseExt {
			continue
		}
		name := filepath.Join(licenseRoot, entry.Name())
		_, err = LoadLicense(name)
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}
