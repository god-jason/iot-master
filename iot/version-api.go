package iot

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/table"
)

func init() {
	api.RegisterUnAuthorized("GET", "product/:id/version", func(ctx *gin.Context) {

		tab, err := table.Get("version")
		if err != nil {
			api.Error(ctx, err)
			return
		}
		rows, err := tab.Find(&table.ParamSearch{
			Limit:  9999,
			Filter: map[string]interface{}{"product_id": ctx.Param("id")},
			Sort: map[string]int{
				"id": -1,
			},
		})
		if err != nil {
			api.Error(ctx, err)
			return
		}

		//兼容uni-app uni-select nz-select组件
		for _, row := range rows {
			row["label"] = row["name"]
			row["text"] = row["name"]
			row["value"] = row["url"]
		}

		api.OK(ctx, rows)
	})
}
