package table

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

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

	//多租户过滤
	if viper.GetBool("tenant") {
		tid := ctx.GetString("tenant")
		if tid != "" {
			column := table.Column("tenant_id")
			if column != nil {
				if body.Filter == nil {
					body.Filter = make(map[string]any)
				}
				//只有未传值tenant_id时，才会赋值用户所在的tenant_id
				if _, ok := body.Filter["tenant_id"]; !ok {
					body.Filter["tenant_id"] = tid
				}
			}
		}
	}

	//查询
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
	// 设置响应头
	ctx.Status(http.StatusOK)
	ctx.Header("Content-Type", "application/json")
	//ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Length", strconv.Itoa(len(buf)))
	_, _ = ctx.Writer.Write(buf)
	//ctx.Abort()
}
