package export

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/curd"
	"github.com/god-jason/iot-master/db"
)

func ApiExport(table, filename string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//id := ctx.MustGet("id")
		ids := ctx.QueryArray("id")

		var datum []map[string]any

		query := db.Engine.Table(table)
		if ids != nil && len(ids) > 0 {
			query = query.In("id", ids)
		}
		err := query.Find(&datum)
		if err != nil {
			curd.Error(ctx, err)
			return
		}

		//下载头
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Disposition", "attachment; filename="+filename+".zip") // 用来指定下载下来的文件名
		ctx.Header("Content-Transfer-Encoding", "binary")

		writer := zip.NewWriter(ctx.Writer)

		for _, data := range datum {
			id := data["id"]
			fn := fmt.Sprintf("%v.json", id)
			f, err := writer.Create(fn)
			if err != nil {
				return
			}

			buf, _ := json.Marshal(data)
			_, err = f.Write(buf)
			if err != nil {
				return
			}
		}

		err = writer.Close()
		if err != nil {
			return
		}
	}
}
