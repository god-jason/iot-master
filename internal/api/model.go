package api

import (
	"regexp"
	"strings"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// 物模型ID格式：字母、数字、下划线、中划线，长度不超过 32（对应 id char(32)）
var modelKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)

// 允许更新的物模型字段（与 registerModelRoutes 的 UpdateFields 保持一致）
var modelUpdateFields = map[string]bool{
	"name": true, "icon": true, "type": true, "protocol": true,
	"properties": true, "pollers": true, "expressions": true, "filters": true,
	"events": true, "services": true, "settings": true, "controls": true,
}

// updateModel 更新物模型。允许修改主键（物模型ID），并同步更新设备引用的 model_id。
func updateModel(c *gin.Context) {
	oldID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 先确认记录存在，避免"值未变化导致影响行数为 0"被误判为不存在
	var model types.Model
	found, err := db.FindOne("model", builder.Eq{"id": oldID}, &model)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if !found {
		curd.Error(c, 404, "数据不存在")
		return
	}

	newID := ""
	if v, ok := body["id"].(string); ok {
		newID = strings.TrimSpace(v)
	}

	if newID != "" && newID != oldID {
		if !modelKeyPattern.MatchString(newID) {
			curd.Error(c, 400, "物模型ID只能包含字母、数字、下划线、中划线，且长度不超过32")
			return
		}
		var exist types.Model
		existFound, err := db.FindOne("model", builder.Eq{"id": newID}, &exist)
		if err != nil {
			curd.Error(c, 500, err.Error())
			return
		}
		if existFound {
			curd.Error(c, 400, "物模型ID已存在")
			return
		}
	}

	updateSet := map[string]interface{}{}
	for k, v := range body {
		if k == "id" || k == "created" || k == "created_at" {
			continue
		}
		if str, ok := v.(string); ok && str == "" {
			continue
		}
		if !modelUpdateFields[k] {
			continue
		}
		updateSet[k] = v
	}

	if len(updateSet) > 0 {
		if _, err := db.UpdateOne("model", builder.Eq{"id": oldID}, updateSet); err != nil {
			curd.Error(c, 500, err.Error())
			return
		}
	}

	if newID == "" || newID == oldID {
		if len(updateSet) == 0 {
			curd.Error(c, 400, "没有可更新的字段")
			return
		}
		curd.Success(c, nil)
		return
	}

	// 修改主键：db.UpdateOne 会跳过 id，这里直接更新主键列
	if _, err := db.Engine.Table("model").Where(builder.Eq{"id": oldID}).Cols("id").
		Update(map[string]interface{}{"id": newID}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 同步更新设备表对该物模型的引用
	if _, err := db.UpdateOne("device", builder.Eq{"model_id": oldID}, map[string]interface{}{"model_id": newID}); err != nil {
		curd.Error(c, 500, "更新设备关联的物模型ID失败: "+err.Error())
		return
	}

	curd.Success(c, nil)
}
