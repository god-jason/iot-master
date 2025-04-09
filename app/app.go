package app

import (
	"archive/zip"
	_ "embed"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const RootPath = "apps"
const Extension = ".ipk"
const ManifestName = "manifest.json"
const IconName = "icon.png"
const ListName = "__LIST__"
const SignName = "__SIGN__"

//go:embed icon.png
var defaultIcon []byte

var pubKey, _ = hex.DecodeString("4f851cec1f93a757037fbb7771aead9a346df9cdd1cf623a8c00b691ac369ed5")

type Manifest struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Author      string `json:"author,omitempty"`
	Copyright   string `json:"copyright,omitempty"`
	Url         string `json:"url,omitempty"`
}

type App struct {
	Manifest

	zipReader     *zip.ReadCloser
	zipReaderLock sync.Mutex
}

func (a *App) ServeFile(path string, ctx *gin.Context) error {
	//TODO 这里会导致前端加载缓慢。。。
	a.zipReaderLock.Lock()
	defer a.zipReaderLock.Unlock()

	if a.zipReader == nil {
		var err error
		filename := filepath.Join(RootPath, a.Id+Extension)
		a.zipReader, err = zip.OpenReader(filename)
		if err != nil {
			return err
		}
	}

	//打开文件
	file, err := a.zipReader.Open(path)
	if err != nil {
		//查找默认首页
		if errors.Is(err, os.ErrNotExist) && path != "index.html" {
			path = "index.html"
			file, err = a.zipReader.Open(path)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	//TODO 只处理 GET与OPTIONS请求

	//加入日期，方便缓存
	st, _ := file.Stat()
	ctx.Header("Last-Modified", st.ModTime().UTC().Format(gmtFormat))
	ctx.Header("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
	ctx.Writer.WriteHeader(http.StatusOK)

	_, err = io.Copy(ctx.Writer, file)
	return err
}
