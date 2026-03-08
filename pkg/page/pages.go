package page

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/store"
)

var PagesFS store.FS

func init() {
	api.RegisterUnAuthorized("GET", "page/*page", func(ctx *gin.Context) {
		ctx.FileFromFS(ctx.Param("page")+".json", http.FS(PagesFS))
	})
}
