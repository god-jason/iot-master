// 设备-组织绑定接口
// 一台设备可以同时绑定多个组织；把设备分配到子级组织本质上是新增一条绑定记录
package api

import (
	"strconv"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// DeviceOrgHandler 设备组织绑定处理器
type DeviceOrgHandler struct{}

// NewDeviceOrgHandler 创建设备组织绑定处理器
func NewDeviceOrgHandler() *DeviceOrgHandler { return &DeviceOrgHandler{} }

// List 查询设备已绑定的组织
// 路由: GET /device/:id/orgs
func (h *DeviceOrgHandler) List(c *gin.Context) {
	device, ok := loadAccessibleDevice(c, c.Param("id"))
	if !ok {
		return
	}

	var rels []types.DeviceOrg
	if err := db.Find("device_org", builder.Eq{"device_id": device.ID}, &rels, 0, 0); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if len(rels) == 0 {
		curd.Success(c, []interface{}{})
		return
	}

	orgIDs := make([]interface{}, 0, len(rels))
	for _, rel := range rels {
		orgIDs = append(orgIDs, rel.OrgID)
	}
	var orgs []types.Org
	if err := db.Find("org", builder.In("id", orgIDs...), &orgs, 0, 0); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	orgMap := make(map[int64]types.Org, len(orgs))
	nameMap := make(map[int64]string, len(orgs))
	pathMap := make(map[int64]string, len(orgs))
	for _, o := range orgs {
		orgMap[o.ID] = o
		nameMap[o.ID] = o.Name
		pathMap[o.ID] = o.Path
	}

	list := make([]gin.H, 0, len(rels))
	for _, rel := range rels {
		list = append(list, gin.H{
			"org_id":   rel.OrgID,
			"org_name": nameMap[rel.OrgID],
			"org_path": pathMap[rel.OrgID],
			"created":  rel.Created,
			"level":    orgMap[rel.OrgID].Level,
			"disabled": orgMap[rel.OrgID].Disabled,
		})
	}
	curd.Success(c, list)
}

// Bind 绑定设备到组织（分配设备到子级组织）
// 路由: POST /device/:id/orgs
// 请求体: { "org_id": 123 }
func (h *DeviceOrgHandler) Bind(c *gin.Context) {
	device, ok := loadAccessibleDevice(c, c.Param("id"))
	if !ok {
		return
	}

	var req struct {
		OrgID int64 `json:"org_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OrgID <= 0 {
		curd.Error(c, 400, "请选择要绑定的组织")
		return
	}
	target, ok := loadOrg(req.OrgID)
	if !ok {
		curd.Error(c, 404, "组织不存在")
		return
	}
	if !requireOrgAccess(c, target) {
		return
	}

	// 幂等：已绑定则直接返回成功
	if n, _ := db.Count("device_org", builder.Eq{"device_id": device.ID, "org_id": req.OrgID}); n > 0 {
		curd.Success(c, gin.H{"org_id": req.OrgID, "bound": true})
		return
	}

	rel := types.DeviceOrg{DeviceID: device.ID, OrgID: req.OrgID}
	if _, err := db.InsertOne("device_org", &rel); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, gin.H{"org_id": req.OrgID, "bound": true})
}

// Unbind 解除设备与组织的绑定
// 路由: DELETE /device/:id/orgs/:org_id
func (h *DeviceOrgHandler) Unbind(c *gin.Context) {
	device, ok := loadAccessibleDevice(c, c.Param("id"))
	if !ok {
		return
	}
	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil || orgID <= 0 {
		curd.Error(c, 400, "无效的组织ID")
		return
	}
	target, ok := loadOrg(orgID)
	if !ok {
		curd.Error(c, 404, "组织不存在")
		return
	}
	if !requireOrgAccess(c, target) {
		return
	}

	affected, err := db.DeleteOne("device_org", builder.Eq{"device_id": device.ID, "org_id": orgID})
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if affected == 0 {
		curd.Error(c, 404, "该设备未绑定此组织")
		return
	}
	curd.Success(c, nil)
}
