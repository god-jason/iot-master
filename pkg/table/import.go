package table

import (
	"encoding/json"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

func FormFiles(ctx *gin.Context) (files []*multipart.FileHeader, err error) {
	form, err := ctx.MultipartForm()
	if err != nil {
		return nil, err
	}
	for _, f := range form.File {
		files = append(files, f...)
	}
	return
}

func ApiImport(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	var docs []Document

	//支持文件上传
	if ctx.ContentType() == "multipart/form-data" {
		files, err := FormFiles(ctx)
		if err != nil {
			Error(ctx, err)
			return
		}

		if len(files) != 1 {
			Fail(ctx, "仅支持一个文件")
			return
		}

		file, err := files[0].Open()
		defer file.Close()

		buf, err := io.ReadAll(file)
		if err != nil {
			Error(ctx, err)
			return
		}

		err = json.Unmarshal(buf, &docs)
		if err != nil {
			Error(ctx, err)
			return
		}
	} else {
		err := ctx.ShouldBind(&docs)
		if err != nil {
			Error(ctx, err)
			return
		}
	}

	//多租户默认
	tid := ctx.GetString("tenant")
	if tid != "" {
		column := table.Column("tenant_id")
		if column != nil {
			for _, doc := range docs {
				//只有未传值tenant_id时，才会赋值用户所在的tenant_id
				if _, ok := doc["tenant_id"]; !ok {
					doc["tenant_id"] = tid
				}
			}
		}
	}

	//依次写入
	var ids []any
	for _, doc := range docs {
		id, err := table.Insert(doc)
		if err != nil {
			Error(ctx, err)
			return
		}
		ids = append(ids, id)
	}

	OK(ctx, ids)
}
