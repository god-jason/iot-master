// 子设备管理接口
// 提供网关下挂子设备的增删改查功能，直接操作 device 表并强制过滤 link_id
// 子设备通过 device.link_id 字段关联到网关链路（不再需要 binding 表）
package api

import (
	"fmt"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// SubDeviceHandler 子设备管理处理器
// 子设备是通过网关连接的设备，与网关存在从属关系
// 子设备CRUD操作直接针对device表，强制过滤link_id
type SubDeviceHandler struct {
	curd.Handler[types.Device] // 继承CRUD基础处理器
}

// NewSubDeviceHandler 创建子设备处理器实例
func NewSubDeviceHandler() *SubDeviceHandler {
	return &SubDeviceHandler{
		Handler: curd.Handler[types.Device]{
			Collection:   "device",                                            // 直接操作device表
			UpdateFields: []string{"model_id", "link", "address", "disabled"}, // 允许更新的字段
		},
	}
}

// checkGatewayAccess 检查用户对网关的访问权限
// 返回: gateway - 网关设备实体, ok - 是否有权限访问
func (h *SubDeviceHandler) checkGatewayAccess(c *gin.Context) (*types.Device, bool) {
	// URL 参数 :id 即为网关设备的 device_id
	gatewayDeviceID := c.Param("id")

	var gateway types.Device
	found, err := db.FindOne("device", builder.Eq{"id": gatewayDeviceID}, &gateway)
	if err != nil || !found || gateway.ID == "" {
		curd.Error(c, 404, "网关设备不存在")
		return nil, false
	}
	if gateway.LinkID == "" {
		curd.Error(c, 404, "网关设备未绑定链路")
		return nil, false
	}

	// 归属校验：仅设备归属用户、家庭成员或系统管理员可操作其子设备
	if !canAccessDevice(c, &gateway) {
		curd.Error(c, 403, "无权限操作该网关设备")
		return nil, false
	}

	return &gateway, true
}

// Search 查询子设备列表（直接查 device 表，强制 link_id 过滤）
func (h *SubDeviceHandler) Search(c *gin.Context) {
	gateway, ok := h.checkGatewayAccess(c)
	if !ok {
		return
	}
	linkID := gateway.LinkID

	var params curd.ListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	params.Normalize()

	// 构建过滤条件，强制过滤 link_id
	baseCond := params.BuildCond()
	var cond builder.Cond
	if baseCond != nil {
		cond = builder.And(baseCond, builder.Eq{"link_id": linkID})
	} else {
		cond = builder.Eq{"link_id": linkID}
	}
	if params.HasKeyword() {
		cond = builder.And(cond, builder.Like{"name", "%" + params.Keyword + "%"})
	}

	// 先取子设备列表
	var devices []types.Device
	total, err := db.PagedFind("device", cond, params.Page, params.PageSize, params.OrderBy, params.Order != "desc", &devices)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 批量获取 model 信息（一次 IN 查询，避免 N+1）
	modelMap := map[string]*types.Model{}
	modelIDs := make([]interface{}, 0, len(devices))
	seen := map[string]bool{}
	for _, d := range devices {
		if d.ModelID != "" && !seen[d.ModelID] {
			seen[d.ModelID] = true
			modelIDs = append(modelIDs, d.ModelID)
		}
	}
	if len(modelIDs) > 0 {
		var models []types.Model
		if err := db.Find("model", builder.In("id", modelIDs...), &models, 0, 0); err == nil {
			for i := range models {
				modelMap[models[i].ID] = &models[i]
			}
		}
	}

	// 构建返回结果，拼接 model 信息 + 内存实时状态
	type deviceWithStatus struct {
		types.Device
		ModelName  string                 `json:"model_name"`
		ModelIcon  string                 `json:"model_icon"`
		ModelType  string                 `json:"model_type"`
		Status     string                 `json:"status,omitempty"`
		StatusData map[string]interface{} `json:"status_data,omitempty"`
	}

	result := make([]deviceWithStatus, 0, len(devices))
	for _, d := range devices {
		ds := deviceWithStatus{Device: d}
		if m, ok := modelMap[d.ModelID]; ok {
			ds.ModelName = m.Name
			ds.ModelIcon = m.Icon
			ds.ModelType = m.Type
		}
		ds.Status, ds.StatusData = runtimeStatus(d.ID)
		result = append(result, ds)
	}

	curd.SuccessTotal(c, result, total)
}

// Get 获取子设备详情（直接查 device 表）
func (h *SubDeviceHandler) Get(c *gin.Context) {
	gateway, ok := h.checkGatewayAccess(c)
	if !ok {
		return
	}
	linkID := gateway.LinkID

	// 解析子设备ID
	subDeviceID := c.Param("sub_device_id")

	// 查询条件：子设备ID + 链路ID（确保只能查询该网关下的子设备）
	var device types.Device
	found, err := db.FindOne("device", builder.Eq{"id": subDeviceID, "link_id": linkID}, &device)
	if err != nil || !found || device.ID == "" {
		curd.Error(c, 404, "子设备不存在")
		return
	}

	// 拼接 model 信息 + 内存实时状态
	type deviceWithStatus struct {
		types.Device
		ModelName  string                 `json:"model_name"`
		ModelIcon  string                 `json:"model_icon"`
		ModelType  string                 `json:"model_type"`
		Status     string                 `json:"status,omitempty"`
		StatusData map[string]interface{} `json:"status_data,omitempty"`
	}
	result := deviceWithStatus{Device: device}
	if device.ModelID != "" {
		var m types.Model
		found, _ := db.FindOne("model", builder.Eq{"id": device.ModelID}, &m)
		if found {
			result.ModelName = m.Name
			result.ModelIcon = m.Icon
			result.ModelType = m.Type
		}
	}

	// 从内存设备注册表读取实时状态
	result.Status, result.StatusData = runtimeStatus(device.ID)

	curd.Success(c, result)
}

// Create 创建子设备（自动带上 link_id）
func (h *SubDeviceHandler) Create(c *gin.Context) {
	gateway, ok := h.checkGatewayAccess(c)
	if !ok {
		return
	}
	linkID := gateway.LinkID

	var req struct {
		types.Device
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 构建设备对象
	device := req.Device
	if device.ID == "" {
		device.ID = db.NewID()
	}
	device.LinkID = linkID // 设置链路ID
	device.Created = time.Now()
	if device.SN == "" {
		device.SN = fmt.Sprintf("%s-%s", linkID, device.Address)
	}

	// 归属必须与网关一致，禁止客户端把子设备挂到他人名下
	user := GetCurrentUser(c)
	if user.Admin {
		if device.UserID == "" {
			device.UserID = gateway.UserID
		}
	} else {
		device.UserID = gateway.UserID
	}
	if device.FamilyID == 0 {
		device.FamilyID = gateway.FamilyID
	}
	if !requireDeviceRefs(c, &device) {
		return
	}

	// 插入设备记录
	_, err := db.InsertOne("device", device)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	curd.Success(c, device)
}

// Update 更新子设备
// 只允许更新指定字段（model_id, link, address, disabled），禁止更新id, created, link_id
func (h *SubDeviceHandler) Update(c *gin.Context) {
	gateway, ok := h.checkGatewayAccess(c)
	if !ok {
		return
	}
	linkID := gateway.LinkID

	subDeviceID := c.Param("sub_device_id")

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 过滤可更新字段，构建更新数据
	updateSet := map[string]interface{}{}
	for k, v := range updateData {
		// 禁止更新的字段
		if k == "id" || k == "created" || k == "link_id" || k == "user_id" || k == "family_id" || k == "room_id" || k == "group_id" {
			continue
		}
		// 跳过空字符串值
		if str, ok := v.(string); ok && str == "" {
			continue
		}
		// 如果指定了UpdateFields，则只更新指定的字段
		if len(h.UpdateFields) > 0 {
			allowed := false
			for _, field := range h.UpdateFields {
				if k == field {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		updateSet[k] = v
	}

	if len(updateSet) == 0 {
		curd.Error(c, 400, "没有可更新的字段")
		return
	}

	// 更新条件：子设备ID + 链路ID（确保只能更新该网关下的子设备）
	cond := builder.Eq{"id": subDeviceID, "link_id": linkID}
	affected, err := db.UpdateOne("device", cond, updateSet)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if affected == 0 {
		curd.Error(c, 404, "子设备不存在")
		return
	}

	curd.Success(c, nil)
}

// Delete 删除子设备（同时清空该子设备可能存在的归属关系）
func (h *SubDeviceHandler) Delete(c *gin.Context) {
	gateway, ok := h.checkGatewayAccess(c)
	if !ok {
		return
	}
	linkID := gateway.LinkID

	subDeviceID := c.Param("sub_device_id")

	// 删除条件：子设备ID + 链路ID（确保只能删除该网关下的子设备）
	cond := builder.Eq{"id": subDeviceID, "link_id": linkID}

	// 删除设备记录
	affected, err := db.DeleteOne("device", cond)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if affected == 0 {
		curd.Error(c, 404, "子设备不存在")
		return
	}

	curd.Success(c, nil)
}
