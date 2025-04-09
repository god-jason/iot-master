package app

import (
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"path/filepath"
)

func init() {

	api.Register("GET", "iot/license/list", func(ctx *gin.Context) {
		var ps []*License
		licenses.Range(func(name string, item *License) bool {
			ps = append(ps, item)
			return true
		})
		api.OK(ctx, ps)
	})

	api.Register("POST", "iot/license/upload", func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			api.Error(ctx, err)
			return
		}

		f, err := file.Open()
		if err != nil {
			api.Error(ctx, err)
			return
		}
		defer f.Close()

		buf, err := io.ReadAll(f)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		lic, err := ParseLicense(string(buf))
		if err != nil {
			api.Error(ctx, err)
			return
		}

		//写入文件
		name := filepath.Join(licenseRoot, lic.AppId+licenseExt)
		err = os.WriteFile(name, buf, os.ModePerm)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, lic)
	})

	api.Register("GET", "iot/license/:app/download", func(ctx *gin.Context) {
		app := ctx.Param("app")
		name := app + licenseExt
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Disposition", "attachment; filename="+name)
		ctx.Header("Content-Transfer-Encoding", "binary")
		ctx.File(filepath.Join(licenseRoot, name))
	})

	api.Register("GET", "iot/license/:app", func(ctx *gin.Context) {
		app := ctx.Param("app")
		lic := licenses.Load(app)
		var err error
		if lic != nil {
			err = lic.Verify(pubKey)
		} else {
			//二次加载，适用于主动放入
			lic, err = LoadLicense(app + licenseExt)
		}
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, lic)
	})

}
