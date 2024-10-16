package export

import (
	"archive/zip"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/curd"
	"github.com/zgwit/iot-master/db"
	"io"
)

func ApiImport(table string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
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

		reader, err := zip.NewReader(file, formFile.Size)
		if err != nil {
			curd.Error(ctx, err)
			return
		}

		//数据解析
		//var datum []map[string]any
		var in int64
		for _, file := range reader.File {
			if file.FileInfo().IsDir() {
				continue
			}

			reader, err := file.Open()
			buf, err := io.ReadAll(reader)
			if err != nil {
				curd.Error(ctx, err)
				return
			}

			var data map[string]any
			err = json.Unmarshal(buf, &data)
			if err != nil {
				curd.Error(ctx, err)
				return
			}

			has, err := db.Engine.Table(table).ID(data["id"]).Exist()
			if err != nil {
				curd.Error(ctx, err)
				return
			}
			if has {
				_, _ = db.Engine.Table(table).Where("id=?", data["id"]).Delete()
			}

			//插入数据
			n, err := db.Engine.Table(table).Insert(data)
			if err != nil {
				curd.Error(ctx, err)
				return
			}
			in += n
		}

		curd.OK(ctx, in)
	}
}
