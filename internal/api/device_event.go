// 设备事件管理接口
// 提供设备事件（告警/通知/上报）的搜索查询和事件处理标记功能
package api

import (
	"fmt"
	"strings"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// DeviceEventHandler 设备事件处理器
// 负责设备事件的搜索、处理（标记为已处理）等操作
// 设备事件记录设备上报的告警、通知等信息
type DeviceEventHandler struct {
	curd.Handler[types.DeviceEvent] // 继承CRUD基础处理器
}

// NewDeviceEventHandler 创建设备事件处理器实例
func NewDeviceEventHandler() *DeviceEventHandler {
	return &DeviceEventHandler{
		Handler: curd.Handler[types.DeviceEvent]{
			Collection:   "device_event",
			IsInt64ID:    true,
			UpdateFields: []string{"device_id", "type", "code", "message", "data", "status"},
		},
	}
}

// Search 搜索设备事件列表
func (h *DeviceEventHandler) Search(c *gin.Context) {
	var params curd.ListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	params.Normalize()

	user := GetCurrentUser(c)

	// 从 params.Filter 中提取 device_id 与其他条件
	var deviceIDValue string
	hasDeviceID := false
	type extraCond struct {
		field string
		op    string
		value interface{}
	}
	var extraConds []extraCond
	for _, f := range params.Filter {
		if f.Field == "device_id" {
			hasDeviceID = true
			deviceIDValue = fmt.Sprint(f.Value)
			continue
		}
		extraConds = append(extraConds, extraCond{f.Field, f.Operator, f.Value})
	}

	// 指定 device_id 时必须校验该设备归属，避免读取他人设备事件
	if hasDeviceID {
		device, ok := loadDevice(deviceIDValue)
		if !ok {
			curd.Error(c, 404, "设备不存在")
			return
		}
		if !canAccessDevice(c, device) {
			curd.Error(c, 403, "无权限查看该设备事件")
			return
		}
	}

	// 根据是否指定 device_id 选择不同的查询策略
	var whereParts []string
	var args []interface{}

	if hasDeviceID {
		whereParts = append(whereParts, "device_event.device_id = ?")
		args = append(args, deviceIDValue)
	} else if user.Admin {
		// 系统管理员可查看全部事件
		whereParts = append(whereParts, "1=1")
	} else {
		// 未指定 device_id：限当前用户直接归属或所在家庭的设备
		whereParts = append(whereParts, `
			device_event.device_id IN (
				SELECT id FROM device WHERE user_id = ?
				OR family_id IN (SELECT family_id FROM member WHERE user_id = ?)
			)
		`)
		args = append(args, user.UserID, user.UserID)
	}

	// 额外过滤字段必须来自白名单，防止字段名拼接造成 SQL 注入
	allowedFields := map[string]string{
		"type": "type", "code": "code", "message": "message",
		"data": "data", "status": "status", "created": "created",
	}
	for _, cond := range extraConds {
		col, ok := allowedFields[cond.field]
		if !ok {
			curd.Error(c, 400, "不支持的过滤字段: "+cond.field)
			return
		}
		switch cond.op {
		case "ne":
			whereParts = append(whereParts, "device_event."+col+" <> ?")
			args = append(args, cond.value)
		case "like":
			whereParts = append(whereParts, "device_event."+col+" LIKE ?")
			args = append(args, "%"+fmt.Sprint(cond.value)+"%")
		case "gt":
			whereParts = append(whereParts, "device_event."+col+" > ?")
			args = append(args, cond.value)
		case "lt":
			whereParts = append(whereParts, "device_event."+col+" < ?")
			args = append(args, cond.value)
		case "gte":
			whereParts = append(whereParts, "device_event."+col+" >= ?")
			args = append(args, cond.value)
		case "lte":
			whereParts = append(whereParts, "device_event."+col+" <= ?")
			args = append(args, cond.value)
		default: // eq / ""
			whereParts = append(whereParts, "device_event."+col+" = ?")
			args = append(args, cond.value)
		}
	}

	fullWhere := strings.Join(whereParts, " AND ")

	// 计数
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM device_event
		WHERE %s
	`, fullWhere)
	total, err := db.ExecRawCount(countSQL, args...)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 查询（JOIN device 获取设备信息）
	querySQL := fmt.Sprintf(`
		SELECT 
			device_event.*,
			device.name as device_name,
			device.sn as device_sn
		FROM device_event
		LEFT JOIN device ON device_event.device_id = device.id
		WHERE %s
		ORDER BY device_event.id DESC
		LIMIT ? OFFSET ?
	`, fullWhere)
	queryArgs := append(args, params.PageSize, params.Offset())

	events, err := db.RawMaps(querySQL, queryArgs...)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	curd.SuccessTotal(c, events, total)
}

// checkEventAccess 校验当前用户是否有权访问指定事件（通过设备归属判断）
func (h *DeviceEventHandler) checkEventAccess(c *gin.Context, id string) bool {
	var event types.DeviceEvent
	found, err := db.FindOne("device_event", builder.Eq{"id": id}, &event)
	if err != nil || !found || event.ID == 0 {
		curd.Error(c, 404, "事件不存在")
		return false
	}
	device, ok := loadDevice(event.DeviceID)
	if !ok {
		curd.Error(c, 404, "关联设备不存在")
		return false
	}
	if !canAccessDevice(c, device) {
		curd.Error(c, 403, "无权限操作该事件")
		return false
	}
	return true
}

// Get 获取事件详情（先校验设备归属）
func (h *DeviceEventHandler) Get(c *gin.Context) {
	if !h.checkEventAccess(c, c.Param("id")) {
		return
	}
	h.Handler.Get(c)
}

// Update 更新事件信息（先校验设备归属）
func (h *DeviceEventHandler) Update(c *gin.Context) {
	if !h.checkEventAccess(c, c.Param("id")) {
		return
	}
	h.Handler.Update(c)
}

// Delete 删除事件（先校验设备归属）
func (h *DeviceEventHandler) Delete(c *gin.Context) {
	if !h.checkEventAccess(c, c.Param("id")) {
		return
	}
	h.Handler.Delete(c)
}

// Handle 处理设备事件（标记为已处理）
func (h *DeviceEventHandler) Handle(c *gin.Context) {
	id := c.Param("id")
	if !h.checkEventAccess(c, id) {
		return
	}

	affected, err := db.UpdateOne("device_event", builder.Eq{"id": id}, map[string]interface{}{"status": 1})
	if err != nil {
		logger.Error("处理设备事件失败", "error", err, "id", id)
		curd.Error(c, 500, "处理失败")
		return
	}
	if affected == 0 {
		curd.Error(c, 404, "事件不存在")
		return
	}

	curd.Success(c, nil)
}
