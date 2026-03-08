package apis

import (
	"time"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "backup", backup)
	api.Register("POST", "recovery", recovery)
}

// 备份数据库
func backup(ctx *gin.Context) {
	// 设置响应头
	filename := "backup_" + time.Now().Format("20060102150405") + ".sql"
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Type", "application/sql")

	err := db.Engine().DumpAll(ctx.Writer)
	if err != nil {
		api.Error(ctx, err)
		return
	}
}

// 恢复数据库
func recovery(ctx *gin.Context) {
	// 获取上传文件
	header, err := ctx.FormFile("header")
	if err != nil {
		api.Error(ctx, err)
		return
	}

	// 打开文件
	file, err := header.Open()
	if err != nil {
		api.Error(ctx, err)
		return
	}
	defer file.Close()

	// 恢复数据
	rs, err := db.Engine().Import(file)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//统计结果
	var count int64
	for _, r := range rs {
		n, _ := r.RowsAffected()
		count += n
	}

	api.OK(ctx, count)
}
