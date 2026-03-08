package app

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func parseUint32(buf []byte) uint32 {
	return uint32(buf[0])<<24 +
		uint32(buf[1])<<16 +
		uint32(buf[2])<<8 +
		uint32(buf[3])
}

func Unpack(key []byte, filename string, dir string) error {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer reader.Close()

	//读取签名单
	sign, err := reader.Open(SignName)
	if err != nil {
		return err
	}
	signs, err := io.ReadAll(sign)
	if err != nil {
		return err
	}

	//读取校验单
	list, err := reader.Open(ListName)
	if err != nil {
		return err
	}
	lists, err := io.ReadAll(list)
	if err != nil {
		return err
	}

	//验证签名
	ret := ed25519.Verify(key, lists, signs)
	if !ret {
		return errors.New("invalid signature")
	}

	//逐行验证文件校验
	rdr := bufio.NewReader(bytes.NewReader(lists))
	for {
		//line, err := rdr.ReadString('\n')
		line, _, err := rdr.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		ss := strings.SplitN(string(line), ":", 2)
		if len(ss) != 2 {
			return errors.New("invalid list file: " + string(line))
		}

		found := false
		//效率有点低，但是也没办法，要不先搞个map索引？？？
		for _, f := range reader.File {
			if f.Name == ss[1] {
				found = true
				b, e := hex.DecodeString(ss[0])
				if e != nil {
					return e
				}

				if parseUint32(b) != f.CRC32 {
					return errors.New("invalid file:" + ss[1])
				}
				break
			}
		}

		if !found {
			return errors.New("not found file:" + ss[1])
		}
	}

	//正式解压
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		fn := filepath.Join(dir, f.Name)
		_ = os.MkdirAll(filepath.Dir(fn), os.ModePerm)
		file, err := f.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		out, err := os.Create(fn)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			return err
		}
	}

	return nil
}
