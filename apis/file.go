package apis

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
)

var UploadPath = "." //TODO 添加到参数

func init() {
	api.RegisterUnAuthorized("POST", "upload", fileUpload)
	api.RegisterUnAuthorized("GET", "download/*filepath", fileDownload)
}

func filename(raw string, num int) string {
	now := time.Now()
	dir := fmt.Sprintf("%d/%d", now.Year(), now.Month())
	_ = os.MkdirAll(dir, os.ModePerm)
	return fmt.Sprintf("%d/%d/%d-%d", now.Year(), now.Month(), now.UnixMilli(), num) + filepath.Ext(raw)
}

func handleUpload(ff *multipart.FileHeader, num int) (string, error) {
	file, err := ff.Open()
	if err != nil {
		return "", err
	}

	now := time.Now()
	dir := fmt.Sprintf("%d/%d", now.Year(), now.Month())
	_ = os.MkdirAll(UploadPath+"/"+dir, os.ModePerm)
	fn := fmt.Sprintf("%d/%d/%d-%d", now.Year(), now.Month(), now.UnixMilli(), num) + filepath.Ext(ff.Filename)

	file2, err := os.Create(UploadPath + "/" + fn)
	if err != nil {
		return "", err
	}
	defer file2.Close()

	_, err = io.Copy(file2, file)
	return fn, err
}

func fileUpload(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var files []string

	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	url := scheme + "://" + ctx.Request.Host + "/api/download/"

	i := 1
	for _, f := range form.File {
		for _, ff := range f {
			i++
			fn, err := handleUpload(ff, i)
			if err != nil {
				_ = ctx.Error(err)
				return
			}

			files = append(files, url+fn)
		}
	}
	api.OK(ctx, files)
}

func fileDownload(ctx *gin.Context) {
	fn := ctx.Param("filepath")
	ctx.File(UploadPath + fn)
}
