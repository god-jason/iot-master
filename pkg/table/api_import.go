package table

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// FormFiles 从请求中获取文件
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

// ApiImport 导入数据
func ApiImport(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	var docs []Document

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

	tid := ctx.GetString("tenant")
	if tid != "" {
		column := table.Column("tenant_id")
		if column != nil {
			for _, doc := range docs {
				if _, ok := doc["tenant_id"]; !ok {
					doc["tenant_id"] = tid
				}
			}
		}
	}

	var errs []error
	var ids []any
	for _, doc := range docs {
		id, err := table.Insert(doc)
		if err != nil {
			errs = append(errs, err)
		}
		ids = append(ids, id)
	}

	if len(errs) > 0 {
		Error(ctx, errors.Join(errs...))
		return
	}
	OK(ctx, ids)
}

// ApiExport 导出数据
func ApiExport(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	var body ParamSearch
	err = ctx.ShouldBindJSON(&body)
	if err != nil {
		Error(ctx, err)
		return
	}

	if viper.GetBool("tenant") {
		tid := ctx.GetString("tenant")
		if tid != "" {
			column := table.Column("tenant_id")
			if column != nil {
				if body.Filter == nil {
					body.Filter = make(map[string]any)
				}
				if _, ok := body.Filter["tenant_id"]; !ok {
					body.Filter["tenant_id"] = tid
				}
			}
		}
	}

	results, err := table.Find(&body)
	if err != nil {
		Error(ctx, err)
		return
	}

	buf, err := json.Marshal(results)
	if err != nil {
		Error(ctx, err)
		return
	}

	filename := table.Name + "-export-" + time.Now().Format("20060102150405") + ".json"
	ctx.Status(http.StatusOK)
	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Length", strconv.Itoa(len(buf)))
	_, _ = ctx.Writer.Write(buf)
}
