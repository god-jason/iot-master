package curd

import (
	"encoding/json"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/gin-gonic/gin"
)

func map2struct(m map[string]any, s any) error {
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, s)
}

func struct2map(s any) (m map[string]any, err error) {
	var buf []byte
	buf, err = json.Marshal(s)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &m)
	return
}

func ApiUpdate[T any](fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//写入ID
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		var data T

		var fs = fields
		if len(fs) > 0 {
			err := ctx.ShouldBindJSON(&data)
			if err != nil {
				api.Error(ctx, err)
				return
			}
		} else {
			//ctx.Body不能读两次，只能反复转换，代码有点丑陋
			var model map[string]any
			err := ctx.ShouldBindJSON(&model)
			if err != nil {
				api.Error(ctx, err)
				return
			}

			err = map2struct(model, &data)
			if err != nil {
				api.Error(ctx, err)
				return
			}

			//取所有键名
			for k, _ := range model {
				fs = append(fs, k)
			}
		}

		//value.Elem().FieldByName("id").Set(reflect.ValueOf(id))
		_, err = db.Engine().ID(id).Cols(fs...).Update(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, &data)
	}
}

func ApiUpdateHook[T any](before, after func(m *T) error, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		var data T

		var fs = fields
		if len(fs) > 0 {
			err := ctx.ShouldBindJSON(&data)
			if err != nil {
				api.Error(ctx, err)
				return
			}

		} else {
			//ctx.Body不能读两次，只能反复转换，代码有点丑陋
			var model map[string]any
			err := ctx.ShouldBindJSON(&model)
			if err != nil {
				api.Error(ctx, err)
				return
			}

			err = map2struct(model, &data)
			if err != nil {
				api.Error(ctx, err)
				return
			}

			//取所有键名
			for k, _ := range model {
				fs = append(fs, k)
			}
		}

		if before != nil {
			if err := before(&data); err != nil {
				api.Error(ctx, err)
				return
			}
		}

		//value.Elem().FieldByName("id").Set(reflect.ValueOf(id))
		_, err = db.Engine().ID(id).Cols(fs...).Update(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		if after != nil {
			//改为异常执行，减少前端错误
			go func() {
				if err := after(&data); err != nil {
					log.Error(err)
				}
			}()
		}

		api.OK(ctx, &data)
	}
}
