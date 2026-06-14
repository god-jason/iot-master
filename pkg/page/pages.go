package page

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/store"
)

var PagesFS store.FS

func init() {
	api.RegisterUnAuthorized("GET", "page/*page", func(ctx *gin.Context) {
		page := ctx.Param("page")

		// 优先查找 .js 文件
		file, err := PagesFS.Open(page + ".js")
		if err == nil {
			defer file.Close()
			data, err := io.ReadAll(file)
			if err == nil {
				ctx.Header("Content-Type", "application/javascript")
				ctx.Data(http.StatusOK, "application/javascript", data)
				return
			}
		}

		// 其次查找 .json 文件
		file, err = PagesFS.Open(page + ".json")
		if err == nil {
			defer file.Close()
			data, err := io.ReadAll(file)
			if err == nil {
				ctx.Header("Content-Type", "application/json")
				ctx.Data(http.StatusOK, "application/json", data)
				return
			}
		}

		ctx.Status(http.StatusNotFound)
	})
}
