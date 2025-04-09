package app

import (
	"archive/zip"
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
	"io"
	"path/filepath"
	"time"
)

const gmtFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func init() {

	//启动时间作为默认图标的修改时间
	bootTime := time.Now()

	api.Register("GET", "iot/app/list", func(ctx *gin.Context) {
		var as []*Manifest
		apps.Range(func(name string, item *App) bool {
			as = append(as, &item.Manifest)
			return true
		})
		api.OK(ctx, as)
	})

	api.Register("GET", "iot/app/:app", func(ctx *gin.Context) {
		app, err := Load(ctx.Param("app"))
		if err != nil {
			api.Error(ctx, err)
			return
		}
		api.OK(ctx, app.Manifest)
	})

	api.Register("GET", "iot/app/:app/icon", func(ctx *gin.Context) {
		reader, err := zip.OpenReader(filepath.Join(RootPath, ctx.Param("app")+Extension))
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		defer reader.Close()

		//ctx.Writer.WriteHeader(http.StatusNotModified)
		//reader.File[0].Comment

		file, err := reader.Open(IconName)
		if err != nil {
			//return nil, err
			//return icon, nil //使用默认图片
			ctx.Header("Last-Modified", bootTime.UTC().Format(gmtFormat))
			ctx.Header("Content-Type", "image/png")
			_, _ = ctx.Writer.Write(defaultIcon)
			return
		}
		defer file.Close()

		st, _ := file.Stat()
		buf, err := io.ReadAll(file)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.Header("Last-Modified", st.ModTime().UTC().Format(gmtFormat))
		ctx.Header("Content-Type", "image/png")
		_, _ = ctx.Writer.Write(buf)
	})
}
