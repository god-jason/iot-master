package app

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var priKey, _ = hex.DecodeString("fdf6dce03f6d7b9e8c6e51d99c2ab160aa9dc46ff839edaba794536aeb9335454f851cec1f93a757037fbb7771aead9a346df9cdd1cf623a8c00b691ac369ed5")

func Pack(dir string, out string) error {
	dir = strings.ReplaceAll(dir, "\\", "/") //全部使用unix分隔符

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	zipWriter := zip.NewWriter(f)

	list := bytes.NewBuffer(nil)

	buf := make([]byte, 32*1024)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		path = strings.ReplaceAll(path, "\\", "/") //全部使用unix分隔符

		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(path, dir)
		header.Name = strings.TrimPrefix(header.Name, "/") //去掉第一个
		header.Method = zip.Deflate
		//header.Modified = time.Unix(1459468800, 0) //清空时间

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		//sum := sha256.New()
		sum := crc32.NewIEEE() //与zip CRC32保持一致，解压时不需要再重复计算

		//复制内容
		for {
			n, e := file.Read(buf)
			if e != nil {
				if e == io.EOF {
					break
				}
				return e
			}
			if n > 0 {
				n, e = writer.Write(buf[:n])
				if e != nil {
					return e
				}
				//计算hash
				_, _ = sum.Write(buf[:n])
			}
		}

		list.WriteString(hex.EncodeToString(sum.Sum(nil))) //sum.Sum32()
		list.WriteByte(':')
		list.WriteString(header.Name)
		list.WriteByte('\n')

		//_, err = io.Copy(writer, file)
		return nil
	})
	if err != nil {
		return err
	}

	//写入hash文件
	w, err := zipWriter.Create(ListName)
	if err != nil {
		return err
	}
	_, err = w.Write(list.Bytes())
	if err != nil {
		return err
	}

	sign := ed25519.Sign(priKey, list.Bytes())
	//写入签名文件
	w, err = zipWriter.Create(SignName)
	if err != nil {
		return err
	}
	_, err = w.Write(sign)
	if err != nil {
		return err
	}

	return zipWriter.Close()
}
