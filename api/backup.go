package api

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/curd"
	"github.com/god-jason/iot-master/db"
	"github.com/god-jason/iot-master/pkg/build"
	"time"
)

// @Summary 导出所有数据
// @Schemes
// @Description 导出所有数据
// @Tags backup
// @Accept json
// @Produce octet-stream
// @Success 200 {object} string 返回SQL文件
// @Router /backup/export [get]
func backupExport(ctx *gin.Context) {
	tm := time.Now().Format("2006-01-02-15-04-05")
	fn := "iot-master-" + build.Version + "-" + tm + ".sql"
	//下载头
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+fn) // 用来指定下载下来的文件名
	ctx.Header("Content-Transfer-Encoding", "binary")

	err := db.Engine.DumpAll(ctx.Writer)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
}

// @Summary 导入所有数据
// @Schemes
// @Description 导入所有数据，文件内容为SQL文本，
// @Tags backup
// @Param file formData file true "SQL"
// @Accept mpfd
// @Produce json
// @Success 200 {object} curd.ReplyData[int64] 返回数据数量
// @Router /backup/import [post]
func backupImport(ctx *gin.Context) {
	formFile, err := ctx.FormFile("file")

	if err != nil {
		curd.Error(ctx, err)
		return
	}

	file, err := formFile.Open()
	if err != nil {
		curd.Error(ctx, err)
		return
	}
	defer file.Close()

	results, err := db.Engine.Import(file)
	if err != nil {
		curd.Error(ctx, err)
		return
	}

	var count int64
	for _, r := range results {
		cnt, _ := r.RowsAffected()
		count += cnt
	}

	curd.OK(ctx, count)
}

func backupRouter(app *gin.RouterGroup) {
	app.GET("/export", backupExport)
	app.POST("/import", backupImport)
}
