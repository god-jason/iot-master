// 维保管理接口：维保计划 CRUD/启停/立即生成工单，维保工单 CRUD/指派/状态流转、统计汇总与 CSV 导出。
package api

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/maintenance"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// ============ 维保计划 ============

// MaintenancePlanEx 维保计划扩展实体（附带范围名称与负责人名称）
type MaintenancePlanEx struct {
	types.MaintenancePlan `xorm:"extends"`

	ScopeName string `json:"scope_name,omitempty"` // 目标范围名称
	OwnerName string `json:"owner_name,omitempty"` // 负责人名称
}

// PlanHandler 维保计划处理器
type PlanHandler struct {
	curd.Handler[types.MaintenancePlan]
}

// NewPlanHandler 创建维保计划处理器
func NewPlanHandler() *PlanHandler {
	return &PlanHandler{
		Handler: curd.Handler[types.MaintenancePlan]{
			Collection:   "maintenance_plan",
			IsInt64ID:    true,
			Keywords:     []string{"name"},
			UpdateFields: []string{"name", "scope_type", "scope_id", "cycle_type", "cycle_value", "remind_days", "items", "owner_id", "next_due", "disabled"},
		},
	}
}

// planFilterFields 计划列表允许的过滤字段
var planFilterFields = map[string]bool{
	"scope_type": true, "scope_id": true, "cycle_type": true, "owner_id": true, "disabled": true,
}

// loadPlan 按 ID 读取维保计划
func loadPlan(raw string) (*types.MaintenancePlan, bool) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return nil, false
	}
	var plan types.MaintenancePlan
	found, err := db.FindOne("maintenance_plan", builder.Eq{"id": id}, &plan)
	if err != nil || !found || plan.ID == 0 {
		return nil, false
	}
	return &plan, true
}

// canViewPlan 判断当前用户是否可查看计划：管理员 / 创建人 / 组织范围可管理
func canViewPlan(c *gin.Context, plan *types.MaintenancePlan) bool {
	user := GetCurrentUser(c)
	if user.Admin || plan.CreatedBy == user.UserID {
		return true
	}
	if plan.ScopeType != types.MaintenanceScopeOrg {
		return false
	}
	for _, id := range userOrgScopeIDs(c) {
		if strconv.FormatInt(id, 10) == plan.ScopeID {
			return true
		}
	}
	return false
}

// planScopeCond 计划列表权限条件
func planScopeCond(c *gin.Context) builder.Cond {
	user := GetCurrentUser(c)
	if user.Admin {
		return nil
	}
	orgArgs := make([]interface{}, 0)
	for _, id := range userOrgScopeIDs(c) {
		orgArgs = append(orgArgs, strconv.FormatInt(id, 10))
	}
	if len(orgArgs) == 0 {
		return builder.Eq{"created_by": user.UserID}
	}
	return builder.Or(
		builder.Eq{"created_by": user.UserID},
		builder.And(
			builder.Eq{"scope_type": types.MaintenanceScopeOrg},
			builder.In("scope_id", orgArgs...),
		),
	)
}

// validPlanScope 目标范围类型是否合法
func validPlanScope(t string) bool {
	switch t {
	case types.MaintenanceScopeModel, types.MaintenanceScopeGroup, types.MaintenanceScopeOrg, types.MaintenanceScopeDevice:
		return true
	}
	return false
}

// validPlanCycle 周期类型是否合法
func validPlanCycle(t string) bool {
	switch t {
	case types.MaintenanceCycleDay, types.MaintenanceCycleMonth, types.MaintenanceCycleYear:
		return true
	}
	return false
}

// userDisplayNames 批量查询用户展示名（优先昵称，其次用户名）
func userDisplayNames(ids []string) map[string]string {
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out
	}
	args := make([]interface{}, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		args = append(args, id)
	}
	if len(args) == 0 {
		return out
	}
	var users []types.User
	if err := db.Find("user", builder.In("id", args...), &users, 0, 0); err != nil {
		return out
	}
	for _, u := range users {
		if u.Nickname != "" {
			out[u.ID] = u.Nickname
		} else {
			out[u.ID] = u.Username
		}
	}
	return out
}

// attachPlanNames 批量补充计划的范围名称与负责人名称
func attachPlanNames(items []*MaintenancePlanEx) {
	if len(items) == 0 {
		return
	}

	modelIDs := make([]interface{}, 0)
	groupIDs := make([]interface{}, 0)
	orgIDs := make([]interface{}, 0)
	deviceIDs := make([]interface{}, 0)
	ownerIDs := make([]string, 0)
	for _, it := range items {
		switch it.ScopeType {
		case types.MaintenanceScopeModel:
			if it.ScopeID != "" {
				modelIDs = append(modelIDs, it.ScopeID)
			}
		case types.MaintenanceScopeGroup:
			if n, err := strconv.ParseInt(it.ScopeID, 10, 64); err == nil {
				groupIDs = append(groupIDs, n)
			}
		case types.MaintenanceScopeOrg:
			if n, err := strconv.ParseInt(it.ScopeID, 10, 64); err == nil {
				orgIDs = append(orgIDs, n)
			}
		case types.MaintenanceScopeDevice:
			if it.ScopeID != "" {
				deviceIDs = append(deviceIDs, it.ScopeID)
			}
		}
		ownerIDs = append(ownerIDs, it.OwnerID)
	}

	modelNames := map[string]string{}
	if len(modelIDs) > 0 {
		var models []types.Model
		if err := db.Find("model", builder.In("id", modelIDs...), &models, 0, 0); err == nil {
			for _, m := range models {
				modelNames[m.ID] = m.Name
			}
		}
	}
	groupNames := map[int64]string{}
	if len(groupIDs) > 0 {
		var groups []types.Group
		if err := db.Find("group", builder.In("id", groupIDs...), &groups, 0, 0); err == nil {
			for _, g := range groups {
				groupNames[g.ID] = g.Name
			}
		}
	}
	orgNames := map[int64]string{}
	if len(orgIDs) > 0 {
		var orgs []types.Org
		if err := db.Find("org", builder.In("id", orgIDs...), &orgs, 0, 0); err == nil {
			for _, o := range orgs {
				orgNames[o.ID] = o.Name
			}
		}
	}
	deviceNames := map[string]string{}
	if len(deviceIDs) > 0 {
		var devices []types.Device
		if err := db.Find("device", builder.In("id", deviceIDs...), &devices, 0, 0); err == nil {
			for _, d := range devices {
				deviceNames[d.ID] = d.Name
			}
		}
	}
	owners := userDisplayNames(ownerIDs)

	for _, it := range items {
		if it.OwnerID != "" {
			it.OwnerName = owners[it.OwnerID]
		}
		switch it.ScopeType {
		case types.MaintenanceScopeModel:
			it.ScopeName = modelNames[it.ScopeID]
		case types.MaintenanceScopeGroup:
			if n, err := strconv.ParseInt(it.ScopeID, 10, 64); err == nil {
				it.ScopeName = groupNames[n]
			}
		case types.MaintenanceScopeOrg:
			if n, err := strconv.ParseInt(it.ScopeID, 10, 64); err == nil {
				it.ScopeName = orgNames[n]
			}
		case types.MaintenanceScopeDevice:
			it.ScopeName = deviceNames[it.ScopeID]
		}
	}
}

// Search 分页搜索维保计划
func (h *PlanHandler) Search(c *gin.Context) {
	var req struct {
		curd.ListParams
		ScopeType string `json:"scope_type"`
		Disabled  *bool  `json:"disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	req.ListParams.Normalize()

	conds := []builder.Cond{}
	if pc := planScopeCond(c); pc != nil {
		conds = append(conds, pc)
	}
	if req.ScopeType != "" {
		conds = append(conds, builder.Eq{"scope_type": req.ScopeType})
	}
	if req.Disabled != nil {
		conds = append(conds, builder.Eq{"disabled": *req.Disabled})
	}
	if req.HasKeyword() {
		conds = append(conds, builder.Like{"name", "%" + req.Keyword + "%"})
	}
	conds = append(conds, whitelistFilterConds(req.Filter, planFilterFields, "")...)

	var cond builder.Cond
	if len(conds) > 1 {
		cond = builder.And(conds...)
	} else if len(conds) == 1 {
		cond = conds[0]
	}

	orderBy := req.OrderBy
	if orderBy == "" || orderBy == "id" {
		orderBy = "created"
	}

	plans := make([]types.MaintenancePlan, 0, req.PageSize)
	total, err := db.PagedFind("maintenance_plan", cond, req.Page, req.PageSize, orderBy, req.Order != "desc", &plans)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	items := make([]MaintenancePlanEx, 0, len(plans))
	ptrs := make([]*MaintenancePlanEx, 0, len(plans))
	for i := range plans {
		items = append(items, MaintenancePlanEx{MaintenancePlan: plans[i]})
		ptrs = append(ptrs, &items[i])
	}
	attachPlanNames(ptrs)

	curd.SuccessTotal(c, items, total)
}

// Get 获取维保计划详情
func (h *PlanHandler) Get(c *gin.Context) {
	plan, ok := loadPlan(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保计划不存在")
		return
	}
	if !canViewPlan(c, plan) {
		curd.Error(c, 403, "无权限查看该维保计划")
		return
	}
	item := MaintenancePlanEx{MaintenancePlan: *plan}
	attachPlanNames([]*MaintenancePlanEx{&item})
	curd.Success(c, item)
}

// Create 创建维保计划
func (h *PlanHandler) Create(c *gin.Context) {
	var plan types.MaintenancePlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if !validPlanScope(plan.ScopeType) {
		curd.Error(c, 400, "请选择有效的目标范围类型")
		return
	}
	if strings.TrimSpace(plan.ScopeID) == "" {
		curd.Error(c, 400, "请选择目标范围")
		return
	}
	if plan.CycleType == "" {
		plan.CycleType = types.MaintenanceCycleMonth
	}
	if !validPlanCycle(plan.CycleType) {
		curd.Error(c, 400, "请选择有效的周期类型")
		return
	}
	if plan.CycleValue <= 0 {
		curd.Error(c, 400, "周期值必须大于 0")
		return
	}

	now := time.Now()
	if plan.NextDue.IsZero() {
		plan.NextDue = maintenance.NextDue(now, plan.CycleType, plan.CycleValue)
	}
	if plan.Items == "" {
		plan.Items = "[]"
	}
	if plan.OwnerID == "" {
		plan.OwnerID = GetCurrentUserID(c)
	}
	plan.ID = 0
	plan.CreatedBy = GetCurrentUserID(c)
	plan.Created = now
	plan.Updated = now

	if _, err := db.InsertOne("maintenance_plan", &plan); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, plan)
}

// Update 更新维保计划（仅管理员或创建人）
func (h *PlanHandler) Update(c *gin.Context) {
	plan, ok := loadPlan(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保计划不存在")
		return
	}
	user := GetCurrentUser(c)
	if !user.Admin && plan.CreatedBy != user.UserID {
		curd.Error(c, 403, "无权限修改该维保计划")
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	allowed := map[string]bool{
		"name": true, "scope_type": true, "scope_id": true, "cycle_type": true,
		"cycle_value": true, "remind_days": true, "items": true, "owner_id": true,
		"next_due": true, "disabled": true,
	}
	updateSet := make(map[string]interface{}, len(payload))
	for k, v := range payload {
		if !allowed[k] || v == nil {
			continue
		}
		switch k {
		case "name":
			s, ok := v.(string)
			if !ok || strings.TrimSpace(s) == "" {
				curd.Error(c, 400, "计划名称不能为空")
				return
			}
			updateSet[k] = strings.TrimSpace(s)
		case "scope_type":
			s, ok := v.(string)
			if !ok || !validPlanScope(s) {
				curd.Error(c, 400, "请选择有效的目标范围类型")
				return
			}
			updateSet[k] = s
		case "scope_id":
			s, ok := v.(string)
			if !ok || strings.TrimSpace(s) == "" {
				curd.Error(c, 400, "请选择目标范围")
				return
			}
			updateSet[k] = strings.TrimSpace(s)
		case "cycle_type":
			s, ok := v.(string)
			if !ok || !validPlanCycle(s) {
				curd.Error(c, 400, "请选择有效的周期类型")
				return
			}
			updateSet[k] = s
		case "cycle_value":
			n := toInt64(v)
			if n <= 0 {
				curd.Error(c, 400, "周期值必须大于 0")
				return
			}
			updateSet[k] = n
		case "remind_days":
			n := toInt64(v)
			if n < 0 {
				curd.Error(c, 400, "提前提醒天数不能为负数")
				return
			}
			updateSet[k] = n
		case "items":
			if s, ok := v.(string); ok {
				updateSet[k] = s
			} else {
				data, err := json.Marshal(v)
				if err != nil {
					curd.Error(c, 400, "无效的检查项模板")
					return
				}
				updateSet[k] = string(data)
			}
		case "owner_id":
			s, ok := v.(string)
			if !ok {
				curd.Error(c, 400, "无效的负责人")
				return
			}
			updateSet[k] = s
		case "next_due":
			s, ok := v.(string)
			if !ok || s == "" {
				continue
			}
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				curd.Error(c, 400, "无效的下次到期时间")
				return
			}
			updateSet[k] = t
		case "disabled":
			b, ok := v.(bool)
			if !ok {
				curd.Error(c, 400, "无效的启用状态")
				return
			}
			updateSet[k] = b
		}
	}
	if len(updateSet) == 0 {
		curd.Error(c, 400, "没有需要更新的字段")
		return
	}
	updateSet["updated"] = time.Now()

	if _, err := db.UpdateOne("maintenance_plan", builder.Eq{"id": plan.ID}, updateSet); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// Delete 删除维保计划（仅管理员或创建人）
func (h *PlanHandler) Delete(c *gin.Context) {
	plan, ok := loadPlan(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保计划不存在")
		return
	}
	user := GetCurrentUser(c)
	if !user.Admin && plan.CreatedBy != user.UserID {
		curd.Error(c, 403, "无权限删除该维保计划")
		return
	}
	if _, err := db.DeleteOne("maintenance_plan", builder.Eq{"id": plan.ID}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// Switch 启用/停用维保计划
func (h *PlanHandler) Switch(c *gin.Context) {
	plan, ok := loadPlan(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保计划不存在")
		return
	}
	user := GetCurrentUser(c)
	if !user.Admin && plan.CreatedBy != user.UserID {
		curd.Error(c, 403, "无权限操作该维保计划")
		return
	}
	var req struct {
		Disabled *bool `json:"disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Disabled == nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if _, err := db.UpdateOne("maintenance_plan", builder.Eq{"id": plan.ID}, map[string]interface{}{
		"disabled": *req.Disabled,
		"updated":  time.Now(),
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// Generate 立即为维保计划生成工单
func (h *PlanHandler) Generate(c *gin.Context) {
	plan, ok := loadPlan(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保计划不存在")
		return
	}
	user := GetCurrentUser(c)
	if !user.Admin && plan.CreatedBy != user.UserID {
		curd.Error(c, 403, "无权限操作该维保计划")
		return
	}
	generated, total, err := maintenance.GenerateForPlan(plan, time.Now())
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, gin.H{"generated": generated, "total": total})
}

// ============ 维保工单 ============

// MaintenanceOrderEx 维保工单扩展实体（附带设备、负责人、计划信息与超期标记）
type MaintenanceOrderEx struct {
	types.MaintenanceOrder `xorm:"extends"`

	DeviceName   string `json:"device_name,omitempty"`   // 设备名称
	DeviceSN     string `json:"device_sn,omitempty"`     // 设备序列号
	AssigneeName string `json:"assignee_name,omitempty"` // 负责人名称
	PlanName     string `json:"plan_name,omitempty"`     // 所属计划名称
	Overdue      bool   `json:"overdue" xorm:"-"`        // 是否已超期
}

// orderExSelect 工单联表查询字段（标识符统一用反引号）
const orderExSelect = "`maintenance_order`.*, " +
	"COALESCE(`device`.`name`, '') AS device_name, " +
	"COALESCE(`device`.`sn`, '') AS device_sn, " +
	"COALESCE(NULLIF(`user`.`nickname`, ''), `user`.`username`, '') AS assignee_name, " +
	"COALESCE(`maintenance_plan`.`name`, '') AS plan_name"

// orderExSession 构建工单联表查询（每个请求使用独立 Session）
func orderExSession() *xorm.Session {
	return db.Engine.Table("maintenance_order").
		Select(orderExSelect).
		Join("LEFT", "device", "`device`.`id` = `maintenance_order`.`device_id`").
		Join("LEFT", "user", "`user`.`id` = `maintenance_order`.`assignee_id`").
		Join("LEFT", "maintenance_plan", "`maintenance_plan`.`id` = `maintenance_order`.`plan_id`")
}

// OrderHandler 维保工单处理器
type OrderHandler struct {
	curd.Handler[types.MaintenanceOrder]
}

// NewOrderHandler 创建维保工单处理器
func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		Handler: curd.Handler[types.MaintenanceOrder]{
			Collection: "maintenance_order",
			IsInt64ID:  true,
			Keywords:   []string{"order_no", "title"},
		},
	}
}

// orderFilterFields 工单列表允许的过滤字段
var orderFilterFields = map[string]bool{
	"status": true, "type": true, "device_id": true, "assignee_id": true, "priority": true,
}

// activeOrderStatuses 未完成工单状态（用于超期判断）
var activeOrderStatuses = []interface{}{
	types.MaintenanceOrderStatusPending,
	types.MaintenanceOrderStatusProcessing,
	types.MaintenanceOrderStatusReviewing,
}

// loadOrder 按 ID 读取维保工单
func loadOrder(raw string) (*types.MaintenanceOrder, bool) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return nil, false
	}
	var order types.MaintenanceOrder
	found, err := db.FindOne("maintenance_order", builder.Eq{"id": id}, &order)
	if err != nil || !found || order.ID == 0 {
		return nil, false
	}
	return &order, true
}

// canAccessOrder 判断当前用户是否可操作工单：管理员 / 负责人 / 设备可访问
func canAccessOrder(c *gin.Context, order *types.MaintenanceOrder) bool {
	user := GetCurrentUser(c)
	if user.Admin {
		return true
	}
	if order.AssigneeID != "" && order.AssigneeID == user.UserID {
		return true
	}
	if order.DeviceID != "" {
		if device, ok := loadDevice(order.DeviceID); ok && canAccessDevice(c, device) {
			return true
		}
	}
	return false
}

// visibleDeviceIDs 当前用户可见的设备ID（直接归属/家庭共享/组织绑定）
func visibleDeviceIDs(c *gin.Context) []interface{} {
	user := GetCurrentUser(c)
	seen := make(map[string]bool)
	out := make([]interface{}, 0)

	add := func(m map[string]interface{}, key string) {
		v, ok := m[key].(string)
		if !ok || v == "" || seen[v] {
			return
		}
		seen[v] = true
		out = append(out, v)
	}

	rows, err := db.RawMaps(
		"SELECT id FROM device WHERE user_id = ? OR family_id IN (SELECT family_id FROM member WHERE user_id = ?)",
		user.UserID, user.UserID,
	)
	if err == nil {
		for _, row := range rows {
			add(row, "id")
		}
	}

	if orgIDs := userOrgScopeIDs(c); len(orgIDs) > 0 {
		args := int64ToInterfaces(orgIDs)
		placeholders := make([]string, 0, len(args))
		for range args {
			placeholders = append(placeholders, "?")
		}
		sql := "SELECT device_id FROM device_org WHERE org_id IN (" + strings.Join(placeholders, ",") + ")"
		if rows, err := db.RawMaps(sql, args...); err == nil {
			for _, row := range rows {
				add(row, "device_id")
			}
		}
	}

	if len(out) == 0 {
		return []interface{}{"none"}
	}
	return out
}

// orderScopeCond 工单权限条件：管理员无限制，其他用户限本人负责或可见设备的工单
func orderScopeCond(c *gin.Context) builder.Cond {
	user := GetCurrentUser(c)
	if user.Admin {
		return nil
	}
	deviceIDs := visibleDeviceIDs(c)
	return builder.Or(
		builder.Eq{"maintenance_order.assignee_id": user.UserID},
		builder.In("maintenance_order.device_id", deviceIDs...),
	)
}

// isOverdue 是否为超期工单
func isOverdue(order *types.MaintenanceOrder, now time.Time) bool {
	if order.DueAt.IsZero() || !order.DueAt.Before(now) {
		return false
	}
	switch order.Status {
	case types.MaintenanceOrderStatusPending, types.MaintenanceOrderStatusProcessing, types.MaintenanceOrderStatusReviewing:
		return true
	}
	return false
}

// validOrderType 工单类型是否合法
func validOrderType(t string) bool {
	switch t {
	case types.MaintenanceOrderTypePlan, types.MaintenanceOrderTypeFault,
		types.MaintenanceOrderTypePatrol, types.MaintenanceOrderTypeInstall:
		return true
	}
	return false
}

// validOrderResult 工单结果是否合法
func validOrderResult(r string) bool {
	switch r {
	case types.MaintenanceResultNormal, types.MaintenanceResultFixed,
		types.MaintenanceResultReplaced, types.MaintenanceResultUnresolved:
		return true
	}
	return false
}

// whitelistFilterConds 按白名单字段构造过滤条件
func whitelistFilterConds(filters []curd.FilterCondition, allowed map[string]bool, prefix string) []builder.Cond {
	out := make([]builder.Cond, 0, len(filters))
	for _, f := range filters {
		if !allowed[f.Field] {
			continue
		}
		field := prefix + f.Field
		switch f.Operator {
		case "", "eq":
			out = append(out, builder.Eq{field: f.Value})
		case "ne":
			out = append(out, builder.Neq{field: f.Value})
		case "gt":
			out = append(out, builder.Gt{field: f.Value})
		case "gte":
			out = append(out, builder.Gte{field: f.Value})
		case "lt":
			out = append(out, builder.Lt{field: f.Value})
		case "lte":
			out = append(out, builder.Lte{field: f.Value})
		case "like":
			if s, ok := f.Value.(string); ok {
				out = append(out, builder.Like{field, "%" + s + "%"})
			}
		case "in":
			if arr, ok := f.Value.([]interface{}); ok && len(arr) > 0 {
				out = append(out, builder.In(field, arr...))
			}
		}
	}
	return out
}

// orderSearchReq 工单查询参数（兼容 curd.ListParams）
type orderSearchReq struct {
	curd.ListParams
	Status     string `json:"status"`
	Type       string `json:"type"`
	DeviceID   string `json:"device_id"`
	AssigneeID string `json:"assignee_id"`
	Overdue    bool   `json:"overdue"`
	DueFrom    string `json:"due_from"`
	DueTo      string `json:"due_to"`
}

// buildOrderCond 构建工单查询条件（搜索与导出共用）
func buildOrderCond(c *gin.Context, req *orderSearchReq) (builder.Cond, error) {
	conds := []builder.Cond{}
	if oc := orderScopeCond(c); oc != nil {
		conds = append(conds, oc)
	}
	if req.Status != "" {
		conds = append(conds, builder.Eq{"maintenance_order.status": req.Status})
	}
	if req.Type != "" {
		conds = append(conds, builder.Eq{"maintenance_order.type": req.Type})
	}
	if req.DeviceID != "" {
		conds = append(conds, builder.Eq{"maintenance_order.device_id": req.DeviceID})
	}
	if req.AssigneeID != "" {
		conds = append(conds, builder.Eq{"maintenance_order.assignee_id": req.AssigneeID})
	}
	now := time.Now()
	if req.Overdue {
		conds = append(conds, builder.And(
			builder.Lt{"maintenance_order.due_at": now},
			builder.In("maintenance_order.status", activeOrderStatuses...),
		))
	}
	if req.DueFrom != "" {
		t, err := time.Parse(time.RFC3339, req.DueFrom)
		if err != nil {
			return nil, fmt.Errorf("无效的起始时间")
		}
		conds = append(conds, builder.Gte{"maintenance_order.due_at": t})
	}
	if req.DueTo != "" {
		t, err := time.Parse(time.RFC3339, req.DueTo)
		if err != nil {
			return nil, fmt.Errorf("无效的结束时间")
		}
		conds = append(conds, builder.Lte{"maintenance_order.due_at": t})
	}
	if req.HasKeyword() {
		kw := "%" + req.Keyword + "%"
		conds = append(conds, builder.Or(
			builder.Like{"maintenance_order.order_no", kw},
			builder.Like{"maintenance_order.title", kw},
		))
	}
	conds = append(conds, whitelistFilterConds(req.Filter, orderFilterFields, "maintenance_order.")...)

	if len(conds) == 0 {
		// 无过滤条件时使用空条件，避免 db.Count/Find 收到 nil 报 Unsupported condition type
		return builder.NewCond(), nil
	}
	if len(conds) == 1 {
		return conds[0], nil
	}
	return builder.And(conds...), nil
}

// Search 分页搜索维保工单
func (h *OrderHandler) Search(c *gin.Context) {
	var req orderSearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	req.ListParams.Normalize()

	cond, err := buildOrderCond(c, &req)
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}

	total, err := db.Count("maintenance_order", cond)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	orderBy := req.OrderBy
	dir := " ASC"
	if orderBy == "" || orderBy == "id" {
		orderBy = "maintenance_order.due_at"
	} else {
		if !strings.Contains(orderBy, ".") {
			orderBy = "maintenance_order." + orderBy
		}
		dir = " DESC"
		if req.Order == "asc" {
			dir = " ASC"
		}
	}

	list := make([]MaintenanceOrderEx, 0, req.PageSize)
	if err := orderExSession().Where(cond).
		OrderBy(orderBy+dir).
		Limit(req.PageSize, req.Offset()).
		Find(&list); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	now := time.Now()
	for i := range list {
		list[i].Overdue = isOverdue(&list[i].MaintenanceOrder, now)
	}
	curd.SuccessTotal(c, list, total)
}

// Get 获取维保工单详情
func (h *OrderHandler) Get(c *gin.Context) {
	order, ok := loadOrder(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保工单不存在")
		return
	}
	if !canAccessOrder(c, order) {
		curd.Error(c, 403, "无权限查看该维保工单")
		return
	}

	var item MaintenanceOrderEx
	found, err := orderExSession().Where(builder.Eq{"maintenance_order.id": order.ID}).Get(&item)
	if err != nil || !found || item.ID == 0 {
		curd.Error(c, 404, "维保工单不存在")
		return
	}
	item.Overdue = isOverdue(&item.MaintenanceOrder, time.Now())
	curd.Success(c, item)
}

// Create 手工创建维保工单
func (h *OrderHandler) Create(c *gin.Context) {
	var order types.MaintenanceOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if order.DeviceID == "" {
		curd.Error(c, 400, "请选择设备")
		return
	}
	device, ok := loadAccessibleDevice(c, order.DeviceID)
	if !ok {
		return
	}

	now := time.Now()
	if order.Type == "" {
		order.Type = types.MaintenanceOrderTypePatrol
	}
	if !validOrderType(order.Type) {
		curd.Error(c, 400, "无效的工单类型")
		return
	}
	if order.Priority <= 0 {
		order.Priority = 2
	}
	if order.Title == "" {
		order.Title = device.Name + "维保"
	}
	userID := GetCurrentUserID(c)
	if order.AssigneeID == "" {
		order.AssigneeID = userID
	}
	if order.Items == "" {
		order.Items = "[]"
	}
	order.ID = 0
	order.OrderNo = maintenance.OrderNo()
	order.Status = types.MaintenanceOrderStatusPending
	order.CreatedBy = userID
	order.Created = now
	order.Updated = now

	if _, err := db.InsertOne("maintenance_order", &order); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, order)
}

// normalizeOrderFields 按白名单提取工单可更新字段，并统一 JSON/时间/数值类型
func normalizeOrderFields(payload map[string]interface{}, allowed map[string]bool) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(payload))
	for k, v := range payload {
		if !allowed[k] || v == nil {
			continue
		}
		switch k {
		case "title", "type", "assignee_id", "remark", "items", "attachments", "parts", "result":
			if s, ok := v.(string); ok {
				if s == "" && (k == "title" || k == "type" || k == "assignee_id") {
					continue
				}
				out[k] = s
				continue
			}
			data, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("无效的字段: %s", k)
			}
			out[k] = string(data)
		case "priority":
			out[k] = toInt64(v)
		case "cost":
			out[k] = toFloat(v)
		case "due_at":
			s, ok := v.(string)
			if !ok || s == "" {
				continue
			}
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				t, err = time.Parse("2006-01-02 15:04:05", s)
				if err != nil {
					return nil, fmt.Errorf("无效的应完成时间")
				}
			}
			out[k] = t
		}
	}
	return out, nil
}

// toFloat 将 JSON 反序列化后的数值安全转换为 float64
func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	}
	return 0
}

// Update 更新维保工单信息
func (h *OrderHandler) Update(c *gin.Context) {
	order, ok := loadOrder(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保工单不存在")
		return
	}
	if !canAccessOrder(c, order) {
		curd.Error(c, 403, "无权限修改该维保工单")
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	allowed := map[string]bool{
		"title": true, "type": true, "priority": true, "assignee_id": true,
		"due_at": true, "remark": true, "items": true, "attachments": true,
		"parts": true, "cost": true,
	}
	updateSet, err := normalizeOrderFields(payload, allowed)
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}
	if t, ok := updateSet["type"].(string); ok && !validOrderType(t) {
		curd.Error(c, 400, "无效的工单类型")
		return
	}
	if len(updateSet) == 0 {
		curd.Error(c, 400, "没有需要更新的字段")
		return
	}
	updateSet["updated"] = time.Now()

	if _, err := db.UpdateOne("maintenance_order", builder.Eq{"id": order.ID}, updateSet); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// Delete 删除维保工单（管理员 / 创建人 / 负责人）
func (h *OrderHandler) Delete(c *gin.Context) {
	order, ok := loadOrder(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保工单不存在")
		return
	}
	user := GetCurrentUser(c)
	if !user.Admin && order.CreatedBy != user.UserID && order.AssigneeID != user.UserID {
		curd.Error(c, 403, "无权限删除该维保工单")
		return
	}
	if _, err := db.DeleteOne("maintenance_order", builder.Eq{"id": order.ID}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// Assign 指派维保工单负责人
func (h *OrderHandler) Assign(c *gin.Context) {
	order, ok := loadOrder(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保工单不存在")
		return
	}
	if !canAccessOrder(c, order) {
		curd.Error(c, 403, "无权限指派该维保工单")
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AssigneeID == "" {
		curd.Error(c, 400, "请选择负责人")
		return
	}
	var assignee types.User
	found, err := db.FindOne("user", builder.Eq{"id": req.AssigneeID}, &assignee)
	if err != nil || !found || assignee.ID == "" {
		curd.Error(c, 400, "负责人不存在")
		return
	}

	if _, err := db.UpdateOne("maintenance_order", builder.Eq{"id": order.ID}, map[string]interface{}{
		"assignee_id": req.AssigneeID,
		"updated":     time.Now(),
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// StartOrder 开始处理工单（pending → processing）
func (h *OrderHandler) StartOrder(c *gin.Context) {
	order, ok := loadOrder(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保工单不存在")
		return
	}
	if !canAccessOrder(c, order) {
		curd.Error(c, 403, "无权限操作该维保工单")
		return
	}
	if order.Status != types.MaintenanceOrderStatusPending {
		curd.Error(c, 400, "只有待处理工单可以开始")
		return
	}

	now := time.Now()
	if _, err := db.UpdateOne("maintenance_order", builder.Eq{"id": order.ID}, map[string]interface{}{
		"status":     types.MaintenanceOrderStatusProcessing,
		"started_at": now,
		"updated":    now,
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// Complete 完成维保工单
func (h *OrderHandler) Complete(c *gin.Context) {
	order, ok := loadOrder(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保工单不存在")
		return
	}
	if !canAccessOrder(c, order) {
		curd.Error(c, 403, "无权限操作该维保工单")
		return
	}
	if order.Status == types.MaintenanceOrderStatusCompleted || order.Status == types.MaintenanceOrderStatusCancelled {
		curd.Error(c, 400, "工单已结束，不能重复完成")
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil && !errors.Is(err, io.EOF) {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	allowed := map[string]bool{
		"result": true, "items": true, "attachments": true, "parts": true, "cost": true, "remark": true,
	}
	updateSet, err := normalizeOrderFields(payload, allowed)
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}

	result, _ := updateSet["result"].(string)
	if result == "" {
		result = types.MaintenanceResultNormal
	}
	if !validOrderResult(result) {
		curd.Error(c, 400, "无效的处理结果")
		return
	}

	now := time.Now()
	updateSet["result"] = result
	updateSet["status"] = types.MaintenanceOrderStatusCompleted
	updateSet["finished_at"] = now
	updateSet["updated"] = now

	if _, err := db.UpdateOne("maintenance_order", builder.Eq{"id": order.ID}, updateSet); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// Cancel 取消维保工单
func (h *OrderHandler) Cancel(c *gin.Context) {
	order, ok := loadOrder(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "维保工单不存在")
		return
	}
	if !canAccessOrder(c, order) {
		curd.Error(c, 403, "无权限操作该维保工单")
		return
	}
	if order.Status == types.MaintenanceOrderStatusCompleted {
		curd.Error(c, 400, "已完成的工单不能取消")
		return
	}

	updateSet := map[string]interface{}{}
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil && !errors.Is(err, io.EOF) {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if remark, ok := payload["remark"].(string); ok && remark != "" {
		updateSet["remark"] = remark
	}
	now := time.Now()
	updateSet["status"] = types.MaintenanceOrderStatusCancelled
	updateSet["updated"] = now

	if _, err := db.UpdateOne("maintenance_order", builder.Eq{"id": order.ID}, updateSet); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// fmtTime 格式化导出时间（零值输出空串）
func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// Export 导出维保工单 CSV（查询参数与搜索一致，不分页）
func (h *OrderHandler) Export(c *gin.Context) {
	req := orderSearchReq{
		Status:     c.Query("status"),
		Type:       c.Query("type"),
		DeviceID:   c.Query("device_id"),
		AssigneeID: c.Query("assignee_id"),
		Overdue:    c.Query("overdue") == "true" || c.Query("overdue") == "1",
		DueFrom:    c.Query("due_from"),
		DueTo:      c.Query("due_to"),
	}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	req.ListParams.Normalize()

	cond, err := buildOrderCond(c, &req)
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}

	list := make([]MaintenanceOrderEx, 0)
	if err := orderExSession().Where(cond).OrderBy("maintenance_order.due_at ASC").Find(&list); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="maintenance-orders-%s.csv"`, time.Now().Format("20060102")))
	// UTF-8 BOM，避免 Excel 打开中文乱码
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"工单号", "设备名称", "设备SN", "类型", "状态", "负责人", "应完成时间", "完成时间", "结果", "成本", "备注"})
	for i := range list {
		o := &list[i]
		_ = w.Write([]string{
			o.OrderNo, o.DeviceName, o.DeviceSN, o.Type, o.Status, o.AssigneeName,
			fmtTime(o.DueAt), fmtTime(o.FinishedAt), o.Result,
			strconv.FormatFloat(o.Cost, 'f', 2, 64), o.Remark,
		})
	}
	w.Flush()
}

// Summary 维保统计汇总
func (h *OrderHandler) Summary(c *gin.Context) {
	plans, err := db.Count("maintenance_plan", planScopeCond(c))
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	var orders []types.MaintenanceOrder
	if err := db.Find("maintenance_order", orderScopeCond(c), &orders, 0, 0); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var pending, processing, reviewing, completed, cancelled, overdue int
	var monthCompleted int
	var monthCost float64
	var hoursSum float64
	var hoursCount int
	byTypeCount := map[string]int{}
	monthAssignees := map[string]*struct {
		count     int
		completed int
	}{}

	for i := range orders {
		o := &orders[i]
		switch o.Status {
		case types.MaintenanceOrderStatusPending:
			pending++
		case types.MaintenanceOrderStatusProcessing:
			processing++
		case types.MaintenanceOrderStatusReviewing:
			reviewing++
		case types.MaintenanceOrderStatusCompleted:
			completed++
		case types.MaintenanceOrderStatusCancelled:
			cancelled++
		}
		if isOverdue(o, now) {
			overdue++
		}
		byTypeCount[o.Type]++
		if o.Status == types.MaintenanceOrderStatusCompleted {
			if !o.StartedAt.IsZero() && !o.FinishedAt.IsZero() {
				hoursSum += o.FinishedAt.Sub(o.StartedAt).Hours()
				hoursCount++
			}
			if !o.FinishedAt.IsZero() && !o.FinishedAt.Before(monthStart) {
				monthCompleted++
				monthCost += o.Cost
			}
		}
		// 本月工单（按创建时间）按负责人统计
		if !o.Created.IsZero() && !o.Created.Before(monthStart) {
			stat := monthAssignees[o.AssigneeID]
			if stat == nil {
				stat = &struct {
					count     int
					completed int
				}{}
				monthAssignees[o.AssigneeID] = stat
			}
			stat.count++
			if o.Status == types.MaintenanceOrderStatusCompleted {
				stat.completed++
			}
		}
	}

	avgHours := 0.0
	if hoursCount > 0 {
		avgHours = math.Round(hoursSum/float64(hoursCount)*100) / 100
	}

	// 按类型统计，按数量降序
	byType := make([]gin.H, 0, len(byTypeCount))
	type typeCount struct {
		Type  string `json:"type"`
		Count int    `json:"count"`
	}
	typeCounts := make([]typeCount, 0, len(byTypeCount))
	for t, n := range byTypeCount {
		typeCounts = append(typeCounts, typeCount{Type: t, Count: n})
	}
	sort.Slice(typeCounts, func(i, j int) bool {
		if typeCounts[i].Count == typeCounts[j].Count {
			return typeCounts[i].Type < typeCounts[j].Type
		}
		return typeCounts[i].Count > typeCounts[j].Count
	})
	for _, tc := range typeCounts {
		byType = append(byType, gin.H{"type": tc.Type, "count": tc.Count})
	}

	// 本月工单按负责人统计，取前 5
	type assigneeCount struct {
		AssigneeID string `json:"assignee_id"`
		Count      int    `json:"count"`
		Completed  int    `json:"completed"`
	}
	assigneeIDs := make([]string, 0, len(monthAssignees))
	assigneeCounts := make([]assigneeCount, 0, len(monthAssignees))
	for id, stat := range monthAssignees {
		assigneeIDs = append(assigneeIDs, id)
		assigneeCounts = append(assigneeCounts, assigneeCount{AssigneeID: id, Count: stat.count, Completed: stat.completed})
	}
	sort.Slice(assigneeCounts, func(i, j int) bool {
		if assigneeCounts[i].Count == assigneeCounts[j].Count {
			return assigneeCounts[i].AssigneeID < assigneeCounts[j].AssigneeID
		}
		return assigneeCounts[i].Count > assigneeCounts[j].Count
	})
	if len(assigneeCounts) > 5 {
		assigneeCounts = assigneeCounts[:5]
	}
	names := userDisplayNames(assigneeIDs)
	byAssignee := make([]gin.H, 0, len(assigneeCounts))
	for _, ac := range assigneeCounts {
		byAssignee = append(byAssignee, gin.H{
			"assignee_id":   ac.AssigneeID,
			"assignee_name": names[ac.AssigneeID],
			"count":         ac.Count,
			"completed":     ac.Completed,
		})
	}

	curd.Success(c, gin.H{
		"plans":           plans,
		"pending":         pending,
		"processing":      processing,
		"reviewing":       reviewing,
		"completed":       completed,
		"cancelled":       cancelled,
		"overdue":         overdue,
		"month_completed": monthCompleted,
		"month_cost":      monthCost,
		"avg_hours":       avgHours,
		"by_type":         byType,
		"by_assignee":     byAssignee,
	})
}
