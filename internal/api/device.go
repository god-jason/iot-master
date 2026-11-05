// 设备管理接口
// 提供设备的增删改查、归属管理、设备控制（同进程直连内嵌 MQTT 网关）、设备实时状态从内存设备注册表读取等功能
// 重构说明：原独立的 binding 表已合并进 device，归属字段（family_id / room_id / group_id / user_id）直接挂载在设备记录上
package api

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/lib"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/tsdb"
	"github.com/god-jason/iot-master/pkg/tsdbapi"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// DeviceEx 设备扩展实体：联表查询设备 + 物模型 + 分组 + 房间 + 家庭 + 用户
type DeviceEx struct {
	types.Device `xorm:"extends"`

	Status     string                 `json:"status,omitempty"`      // 实时状态（"1" 在线 / "0" 离线，来自内存设备注册表）
	StatusData map[string]interface{} `json:"status_data,omitempty"` // 实时属性快照（来自内存设备注册表）
	Online     bool                   `json:"online"`                // 是否在线
	ModelName  string                 `json:"model_name,omitempty"`  // 物模型名称
	ModelIcon  string                 `json:"model_icon,omitempty"`  // 物模型图标
	ModelType  string                 `json:"model_type,omitempty"`  // 物模型设备类型
	GroupName  string                 `json:"group_name,omitempty"`  // 分组名称
	RoomName   string                 `json:"room_name,omitempty"`   // 房间名称
	FamilyName string                 `json:"family_name,omitempty"` // 家庭名称
	UserName   string                 `json:"user_name,omitempty"`   // 归属用户名称
	OrgIDs     []int64                `json:"org_ids,omitempty"`     // 已绑定的组织ID列表
	OrgNames   []string               `json:"org_names,omitempty"`   // 已绑定的组织名称列表
}

// attachDeviceOrgs 批量填充设备已绑定的组织信息
func attachDeviceOrgs(list []*DeviceEx) {
	if len(list) == 0 {
		return
	}
	deviceIDs := make([]interface{}, 0, len(list))
	for _, d := range list {
		deviceIDs = append(deviceIDs, d.ID)
	}
	var rels []types.DeviceOrg
	if err := db.Find("device_org", builder.In("device_id", deviceIDs...), &rels, 0, 0); err != nil || len(rels) == 0 {
		return
	}

	orgIDs := make([]interface{}, 0, len(rels))
	seenOrg := make(map[int64]bool, len(rels))
	for _, rel := range rels {
		if !seenOrg[rel.OrgID] {
			seenOrg[rel.OrgID] = true
			orgIDs = append(orgIDs, rel.OrgID)
		}
	}
	var orgs []types.Org
	nameMap := make(map[int64]string, len(orgIDs))
	if err := db.Find("org", builder.In("id", orgIDs...), &orgs, 0, 0); err == nil {
		for _, o := range orgs {
			nameMap[o.ID] = o.Name
		}
	}

	relMap := make(map[string][]int64, len(rels))
	for _, rel := range rels {
		relMap[rel.DeviceID] = append(relMap[rel.DeviceID], rel.OrgID)
	}
	for _, d := range list {
		for _, oid := range relMap[d.ID] {
			d.OrgIDs = append(d.OrgIDs, oid)
			d.OrgNames = append(d.OrgNames, nameMap[oid])
		}
	}
}

// deviceExSelect 联表查询的字段列表（标识符统一用反引号，xorm 会按方言替换）
const deviceExSelect = "`device`.*, " +
	"COALESCE(`model`.`name`, '') AS model_name, " +
	"COALESCE(`model`.`icon`, '') AS model_icon, " +
	"COALESCE(`model`.`type`, '') AS model_type, " +
	"COALESCE(`group`.`name`, '') AS group_name, " +
	"COALESCE(`room`.`name`, '') AS room_name, " +
	"COALESCE(`family`.`name`, '') AS family_name, " +
	"COALESCE(NULLIF(`user`.`nickname`, ''), `user`.`username`, '') AS user_name"

// deviceExSession 构建设备联表查询（每个请求使用独立 Session）
func deviceExSession() *xorm.Session {
	return db.Engine.Table("device").
		Select(deviceExSelect).
		Join("LEFT", "model", "`model`.`id` = `device`.`model_id`").
		Join("LEFT", "group", "`group`.`id` = `device`.`group_id`").
		Join("LEFT", "room", "`room`.`id` = `device`.`room_id`").
		Join("LEFT", "family", "`family`.`id` = `device`.`family_id`").
		Join("LEFT", "user", "`user`.`id` = `device`.`user_id`")
}

// runtimeStatus 从内存设备注册表读取实时在线状态与属性快照。
// 内存中没有该设备（未接入 / 已离线 / 未注册）时一律返回离线状态。
func runtimeStatus(deviceID string) (string, map[string]interface{}) {
	dev, ok := iot.GetDevice(deviceID)
	if !ok {
		return "0", nil
	}
	status := "0"
	if dev.IsOnline() {
		status = "1"
	}
	props := dev.Properties()
	if len(props) == 0 {
		props = nil
	}
	return status, props
}

// applyDeviceStatus 把内存设备注册表中的实时状态合并进 DeviceEx
func applyDeviceStatus(d *DeviceEx) {
	d.Status, d.StatusData = runtimeStatus(d.ID)
	d.Online = d.Status == "1"
}

// DeviceHandler 设备管理处理器
// 负责设备的CRUD操作、归属管理、设备控制等功能
// 设备归属（family_id / room_id / group_id / user_id）直接存储在 device 表上
type DeviceHandler struct {
	curd.Handler[types.Device] // CRUD基础处理器，继承通用的增删改查能力
}

// NewDeviceHandler 创建设备处理器实例
func NewDeviceHandler() *DeviceHandler {
	return &DeviceHandler{
		Handler: curd.Handler[types.Device]{
			Collection: "device",
			UpdateFields: []string{
				"name", "sn", "model_id", "address",
				"lat", "lng", "geo_hash",
				"group_id", "room_id", "family_id", "user_id",
				"disabled", "link",
			},
		},
	}
}

// Search 查询设备列表（联表查询物模型 / 分组 / 房间 / 家庭 / 用户名称）
func (h *DeviceHandler) Search(c *gin.Context) {
	var params curd.ListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	params.Normalize()

	user := GetCurrentUser(c)

	// 构建基础条件：组织范围优先，其次家庭 / 房间（联表后字段加 device. 前缀避免歧义）
	var cond builder.Cond
	familyIDStr := c.Query("family_id")
	roomIDStr := c.Query("room_id")
	orgIDStr := c.Query("org_id")

	// 非管理员的组织范围：归属组织 + 负责人组织 + 成员组织及其子树
	var scopeIDs []int64
	if !user.Admin {
		scopeIDs = userOrgScopeIDs(c)
	}

	switch {
	case !user.Admin && len(scopeIDs) > 0:
		// 组织负责人 / 成员只能看到绑定到可管理组织范围内的设备
		cond = orgDeviceIDsCond(scopeIDs)
	case orgIDStr != "":
		orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
		if err != nil || orgID <= 0 {
			curd.Error(c, 400, "无效的组织ID")
			return
		}
		if !user.Admin {
			target, ok := loadOrg(orgID)
			if !ok {
				curd.Error(c, 404, "组织不存在")
				return
			}
			if !requireOrgAccess(c, target) {
				return
			}
		}
		cond = orgDeviceIDsCond(orgSubtreeIDsByID(orgID))
	case familyIDStr != "":
		familyID, err := strconv.ParseInt(familyIDStr, 10, 64)
		if err != nil || familyID == 0 {
			curd.Error(c, 400, "无效的家庭ID")
			return
		}
		if !canAccessFamily(user.UserID, familyID, user.Admin) {
			curd.Error(c, 403, "无权限查看该家庭设备")
			return
		}
		cond = builder.Eq{"device.family_id": familyID}
	case roomIDStr != "":
		roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
		if err != nil || roomID == 0 {
			curd.Error(c, 400, "无效的房间ID")
			return
		}
		if !canAccessRoom(user.UserID, roomID, user.Admin) {
			curd.Error(c, 403, "无权限查看该房间设备")
			return
		}
		cond = builder.Eq{"device.room_id": roomID}
	default:
		// 默认：当前用户有访问权限的设备（家庭成员 + 直接归属）
		if user.Admin {
			cond = builder.NewCond()
		} else {
			cond = builder.Eq{"device.user_id": user.UserID}
			var members []types.Member
			if err := db.Find("member", builder.Eq{"user_id": user.UserID}, &members, 0, 0); err == nil && len(members) > 0 {
				familyIDs := make([]interface{}, 0, len(members))
				for _, m := range members {
					familyIDs = append(familyIDs, m.FamilyID)
				}
				cond = builder.Or(builder.Eq{"device.user_id": user.UserID}, builder.In("device.family_id", familyIDs...))
			}
		}
	}

	// 关键词搜索
	if params.HasKeyword() {
		cond = builder.And(cond, builder.Or(
			builder.Like{"device.name", "%" + params.Keyword + "%"},
			builder.Like{"device.sn", "%" + params.Keyword + "%"},
		))
	}

	// 计数（与列表使用同一套联表条件）
	total, err := deviceExSession().Where(cond).Select("count(*)").Count()
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 排序字段加表前缀，避免 created/id 等多表同名列冲突
	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "device.created"
	} else if !strings.Contains(orderBy, ".") {
		orderBy = "device." + orderBy
	}
	dir := " DESC"
	if params.Order == "asc" {
		dir = " ASC"
	}

	list := make([]DeviceEx, 0, params.PageSize)
	if err := deviceExSession().Where(cond).
		OrderBy(orderBy+dir).
		Limit(params.PageSize, (params.Page-1)*params.PageSize).
		Find(&list); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 批量合并内存中的实时状态与组织绑定信息
	for i := range list {
		applyDeviceStatus(&list[i])
	}
	ptrs := make([]*DeviceEx, len(list))
	for i := range list {
		ptrs[i] = &list[i]
	}
	attachDeviceOrgs(ptrs)

	curd.SuccessTotal(c, list, total)
}

// Create 创建设备（直接在 device 记录上带归属字段）
// 自动生成 ID + Secret（如果前端没传）
func (h *DeviceHandler) Create(c *gin.Context) {
	var device types.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	if device.ID == "" {
		device.ID = db.NewID()
	}
	if device.Secret == "" {
		device.Secret = db.NewSecret()
	}
	// 自动生成 GeoHash（有经纬度就一定算，不管前端有没有传）
	if device.Lat != 0 && device.Lng != 0 {
		device.GeoHash = lib.GeoHashEncodeBase32(device.Lat, device.Lng, 12)
	}
	device.Created = time.Now()

	user := GetCurrentUser(c)
	// 普通用户只能把设备归属到自己名下，系统管理员可指定归属
	if user.Admin {
		if device.UserID == "" {
			device.UserID = user.UserID
		}
	} else {
		device.UserID = user.UserID
	}
	if !requireDeviceRefs(c, &device) {
		return
	}

	var err error
	if device.SN == "" {
		// SN 允许为空：省略该列让数据库写入 NULL。
		// sn 上有唯一索引，多个空字符串会被判定为重复，NULL 则不受唯一约束限制
		_, err = db.Engine.Table("device").Omit("sn").Insert(&device)
	} else {
		_, err = db.InsertOne("device", device)
	}
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	curd.Success(c, device)
}

// Get 获取设备详情（联表查询物模型 / 分组 / 房间 / 家庭 / 用户名称）
func (h *DeviceHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var device DeviceEx
	found, err := deviceExSession().Where(builder.Eq{"device.id": id}).Get(&device)
	if err != nil || !found || device.ID == "" {
		curd.Error(c, 404, "设备不存在")
		return
	}

	// 权限校验：直接归属用户、家庭成员或系统管理员
	if !canAccessDevice(c, &device.Device) {
		curd.Error(c, 403, "无权限访问该设备")
		return
	}

	// 实时状态（内存设备注册表）与组织绑定信息
	applyDeviceStatus(&device)
	attachDeviceOrgs([]*DeviceEx{&device})

	curd.Success(c, device)
}

// Update 更新设备信息（先校验归属，再按白名单字段更新）
func (h *DeviceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if _, ok := loadAccessibleDevice(c, id); !ok {
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	allowed := map[string]bool{
		"name": true, "sn": true, "model_id": true, "address": true,
		"lat": true, "lng": true, "geo_hash": true,
		"group_id": true, "room_id": true, "family_id": true, "user_id": true,
		"disabled": true, "link": true,
	}

	updateSet := make(map[string]interface{}, len(payload))
	for k, v := range payload {
		if !allowed[k] {
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		updateSet[k] = v
	}
	if len(updateSet) == 0 {
		curd.Error(c, 400, "没有需要更新的字段")
		return
	}

	user := GetCurrentUser(c)
	if !user.Admin {
		// 普通用户不能修改设备归属用户
		if v, ok := updateSet["user_id"]; ok && v != nil {
			if s, ok := v.(string); !ok || s != user.UserID {
				curd.Error(c, 403, "无权限修改设备归属用户")
				return
			}
		}
		// 普通用户只能指定自己有权访问的家庭 / 房间
		if v, ok := updateSet["family_id"]; ok && v != nil {
			fid := toInt64(v)
			if fid != 0 && !canAccessFamily(user.UserID, fid, false) {
				curd.Error(c, 403, "无权限指定该家庭")
				return
			}
		}
		if v, ok := updateSet["room_id"]; ok && v != nil {
			rid := toInt64(v)
			if rid != 0 && !canAccessRoom(user.UserID, rid, false) {
				curd.Error(c, 403, "无权限指定该房间")
				return
			}
		}
	}

	if _, err := db.UpdateOne("device", builder.Eq{"id": id}, updateSet); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	curd.Success(c, nil)
}

// Delete 删除设备（先校验归属）
func (h *DeviceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if _, ok := loadAccessibleDevice(c, id); !ok {
		return
	}
	h.Handler.Delete(c)
}

// Unbind 解绑设备：清空设备的归属字段
func (h *DeviceHandler) Unbind(c *gin.Context) {
	id := c.Param("id")

	var device types.Device
	found, err := db.FindOne("device", builder.Eq{"id": id}, &device)
	if err != nil || !found || device.ID == "" {
		curd.Error(c, 404, "设备不存在")
		return
	}

	// 权限：设备归属用户、设备所属家庭成员或系统管理员
	if !canAccessDevice(c, &device) {
		curd.Error(c, 403, "无权限解绑该设备")
		return
	}

	// 清空归属字段
	_, err = db.UpdateOne("device", builder.Eq{"id": id}, map[string]interface{}{
		"family_id": nil,
		"room_id":   nil,
		"group_id":  nil,
	})
	if err != nil {
		logger.Error("解绑设备失败", "error", err)
		curd.Error(c, 500, "解绑失败")
		return
	}

	curd.Success(c, nil)
}

// Property 获取设备实时属性（内存设备注册表，未接入设备返回空快照）
func (h *DeviceHandler) Property(c *gin.Context) {
	id := c.Param("id")
	if _, ok := loadAccessibleDevice(c, id); !ok {
		return
	}

	data := map[string]gin.H{}
	if dev, ok := iot.GetDevice(id); ok {
		now := time.Now().Unix()
		for k, v := range dev.Properties() {
			data[k] = gin.H{"value": v, "time": now}
		}
	}
	curd.Success(c, data)
}

// History 查询设备属性历史（时序库按物模型维度存储，tag id = 设备ID）
func (h *DeviceHandler) History(c *gin.Context) {
	id := c.Param("id")
	device, ok := loadAccessibleDevice(c, id)
	if !ok {
		return
	}

	start, err := time.Parse(time.RFC3339, c.Query("start"))
	if err != nil {
		start = time.Now().Add(-time.Hour)
	}
	end, err := time.Parse(time.RFC3339, c.Query("end"))
	if err != nil {
		end = time.Now()
	}
	if end.Before(start) {
		start, end = end, start
	}

	if tsdb.DefaultStore == nil {
		curd.Success(c, gin.H{})
		return
	}

	series, err := tsdb.DefaultStore.Query(context.Background(), tsdbapi.QuerySpec{
		Measurement: device.ModelID,
		Tags:        map[string]string{"id": id},
		Start:       start,
		End:         end,
		Window:      c.Query("window"),
	})
	if err != nil {
		logger.Error("查询设备历史数据失败", "device", id, "error", err)
		curd.Error(c, 500, "查询历史数据失败")
		return
	}

	out := make(map[string][]gin.H, len(series))
	for field, s := range series {
		points := make([]gin.H, 0, len(s.Times))
		for i := range s.Times {
			if i >= len(s.Values) {
				break
			}
			points = append(points, gin.H{"time": s.Times[i], "value": s.Values[i]})
		}
		out[field] = points
	}
	curd.Success(c, out)
}

// Control 调用设备服务（需要设备在线并回 ACK）
func (h *DeviceHandler) Control(c *gin.Context) {
	id := c.Param("id")
	if _, ok := loadAccessibleDevice(c, id); !ok {
		return
	}

	var req struct {
		Service string                 `json:"service"`
		Params  map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Service == "" {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	dev, ok := iot.GetDevice(id)
	if !ok || !dev.IsOnline() {
		curd.Error(c, 503, "设备未在线")
		return
	}
	res, err := dev.Control(req.Service, req.Params)
	if err != nil {
		curd.Error(c, 500, "设备操作失败: "+err.Error())
		return
	}
	curd.Success(c, res)
}

// Setting 下发设备配置
func (h *DeviceHandler) Setting(c *gin.Context) {
	id := c.Param("id")
	if _, ok := loadAccessibleDevice(c, id); !ok {
		return
	}

	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	dev, ok := iot.GetDevice(id)
	if !ok || !dev.IsOnline() {
		curd.Error(c, 503, "设备未在线")
		return
	}
	if err := dev.SetConfig(config); err != nil {
		curd.Error(c, 500, "配置下发失败: "+err.Error())
		return
	}
	curd.Success(c, nil)
}

// ============ 辅助函数 ============

// joinStrings 用指定分隔符拼接字符串切片（空切片返回空串）
func joinStrings(arr []string, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	result := arr[0]
	for i := 1; i < len(arr); i++ {
		result += sep + arr[i]
	}
	return result
}

// getStr 从 map 中安全取出字符串值（键不存在或类型不符时返回空串）
func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
