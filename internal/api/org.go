// 组织管理接口
// 提供多级组织（代理商 / 经销商 / 门店）的树形查询与增删改，以及组织成员管理
// 数据范围：平台管理员可管理所有组织；组织负责人（owner）与组织成员可管理该组织及其子级组织
package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// OrgNode 组织树节点
type OrgNode struct {
	types.Org  `xorm:"extends"`
	Children   []*OrgNode `json:"children,omitempty"`
	OwnerName  string     `json:"owner_name,omitempty"` // 负责人昵称
	IsOwner    bool       `json:"is_owner,omitempty"`   // 当前用户是否为该组织负责人
	IsMember   bool       `json:"is_member,omitempty"`  // 当前用户是否为该组织成员
	CanManage  bool       `json:"can_manage"`           // 是否可编辑/删除（子级组织或管理员）
	CanMembers bool       `json:"can_members"`          // 是否可管理该组织成员
}

// OrgDetail 组织详情（附带负责人昵称）
type OrgDetail struct {
	types.Org `xorm:"extends"`
	OwnerName string `json:"owner_name,omitempty"`
}

// UserBrief 用户简要信息（用于负责人/成员选择器）
type UserBrief struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar,omitempty"`
	Cellphone string `json:"cellphone,omitempty"`
}

// OrgMemberItem 组织成员列表项
type OrgMemberItem struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar,omitempty"`
	Cellphone string    `json:"cellphone,omitempty"`
	Created   time.Time `json:"created,omitempty"`
}

// OrgHandler 组织管理处理器
type OrgHandler struct{}

// NewOrgHandler 创建组织处理器
func NewOrgHandler() *OrgHandler { return &OrgHandler{} }

// loadOrg 按 ID 读取组织
func loadOrg(id int64) (*types.Org, bool) {
	if id <= 0 {
		return nil, false
	}
	var org types.Org
	found, err := db.FindOne("org", builder.Eq{"id": id}, &org)
	if err != nil || !found || org.ID == 0 {
		return nil, false
	}
	return &org, true
}

// orgSubtreeIDs 返回组织子树内的全部组织 ID
func orgSubtreeIDs(root *types.Org) []int64 {
	if root == nil || root.Path == "" {
		return nil
	}
	var list []types.Org
	if err := db.Find("org", builder.Like{"path", root.Path + "%"}, &list, 0, 0); err != nil {
		return nil
	}
	ids := make([]int64, 0, len(list))
	for _, o := range list {
		ids = append(ids, o.ID)
	}
	return ids
}

// orgSubtreeIDsByID 按组织 ID 返回子树 ID；组织不存在时返回 nil
func orgSubtreeIDsByID(orgID int64) []int64 {
	org, ok := loadOrg(orgID)
	if !ok {
		return nil
	}
	return orgSubtreeIDs(org)
}

// int64ToInterfaces 将 []int64 转为 builder.In 需要的 []interface{}
func int64ToInterfaces(ids []int64) []interface{} {
	out := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		out = append(out, id)
	}
	return out
}

// uniqueInt64 去重
func uniqueInt64(ids []int64) []int64 {
	seen := make(map[int64]bool, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id > 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// containsInt64 判断切片是否包含指定值
func containsInt64(ids []int64, id int64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// userOrgRootIDs 返回用户关联的根组织：担任负责人的组织 + 加入的组织
func userOrgRootIDs(user *Claims) []int64 {
	if user == nil || user.UserID == "" {
		return nil
	}
	roots := make([]int64, 0, 4)

	var owned []types.Org
	if err := db.Find("org", builder.Eq{"owner_user_id": user.UserID}, &owned, 0, 0); err == nil {
		for _, o := range owned {
			roots = append(roots, o.ID)
		}
	}
	var rels []types.OrgMember
	if err := db.Find("org_member", builder.Eq{"user_id": user.UserID}, &rels, 0, 0); err == nil {
		for _, r := range rels {
			roots = append(roots, r.OrgID)
		}
	}
	return uniqueInt64(roots)
}

// userOrgScopeIDs 返回用户可管理范围（所有关联组织及其子树）的组织 ID；同一次请求内缓存
func userOrgScopeIDs(c *gin.Context) []int64 {
	if v, ok := c.Get("org_scope_ids"); ok {
		if ids, ok := v.([]int64); ok {
			return ids
		}
	}
	user := GetCurrentUser(c)
	var ids []int64
	if user.Admin {
		var all []types.Org
		if err := db.Find("org", nil, &all, 0, 0); err == nil {
			for _, o := range all {
				ids = append(ids, o.ID)
			}
		}
	} else {
		for _, rootID := range userOrgRootIDs(user) {
			ids = append(ids, orgSubtreeIDsByID(rootID)...)
		}
	}
	ids = uniqueInt64(ids)
	c.Set("org_scope_ids", ids)
	return ids
}

// orgDeviceIDsCond 构造“设备已绑定到指定组织集合”的查询条件
func orgDeviceIDsCond(ids []int64) builder.Cond {
	if len(ids) == 0 {
		return builder.Eq{"device.id": "__none__"}
	}
	sub := builder.Select("device_id").From("device_org").Where(builder.In("org_id", int64ToInterfaces(ids)...))
	return builder.In("device.id", sub)
}

// requireOrgAccess 校验当前用户是否有权访问目标组织（管理员 / 关联组织及其子树）
func requireOrgAccess(c *gin.Context, target *types.Org) bool {
	if IsAdminUser(c) {
		return true
	}
	if target == nil {
		curd.Error(c, 404, "组织不存在")
		return false
	}
	if containsInt64(userOrgScopeIDs(c), target.ID) {
		return true
	}
	curd.Error(c, 403, "无权限操作该组织")
	return false
}

// requireOrgAssignable 校验当前用户能否把资源（用户/设备）归属到目标组织
func requireOrgAssignable(c *gin.Context, orgID int64) bool {
	if orgID == 0 {
		return true
	}
	target, ok := loadOrg(orgID)
	if !ok {
		curd.Error(c, 404, "组织不存在")
		return false
	}
	if IsAdminUser(c) {
		return true
	}
	if containsInt64(userOrgScopeIDs(c), target.ID) {
		return true
	}
	curd.Error(c, 403, "只能归属到本组织及子级组织")
	return false
}

// requireOrgMemberManage 校验成员管理权限：管理员 / 该组织负责人 / 可管理该子级组织的人
func requireOrgMemberManage(c *gin.Context, org *types.Org) bool {
	if IsAdminUser(c) {
		return true
	}
	user := GetCurrentUser(c)
	if org.OwnerUserID != "" && org.OwnerUserID == user.UserID {
		return true
	}
	if !requireOrgAccess(c, org) {
		return false
	}
	// 关联组织本身只能由负责人管理成员，成员不能自行加人
	if containsInt64(userOrgRootIDs(user), org.ID) {
		curd.Error(c, 403, "只有组织负责人可以管理本组织成员")
		return false
	}
	return true
}

// userHasOrgAssociation 判断用户是否已与任一组织关联（负责人 / 成员），用于管理后台登录校验
func userHasOrgAssociation(userID string) bool {
	if userID == "" {
		return false
	}
	if n, _ := db.Count("org", builder.Eq{"owner_user_id": userID}); n > 0 {
		return true
	}
	if n, _ := db.Count("org_member", builder.Eq{"user_id": userID}); n > 0 {
		return true
	}
	return false
}

// ownerNames 批量查询用户昵称（负责人展示用）
func ownerNames(ids []string) map[string]string {
	out := make(map[string]string, len(ids))
	args := make([]interface{}, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id != "" && !seen[id] {
			seen[id] = true
			args = append(args, id)
		}
	}
	if len(args) == 0 {
		return out
	}
	var users []types.User
	if err := db.Find("user", builder.In("id", args...), &users, 0, 0); err != nil {
		return out
	}
	for _, u := range users {
		name := u.Nickname
		if name == "" {
			name = u.Username
		}
		out[u.ID] = name
	}
	return out
}

// Tree 组织树
// 路由: GET /org/tree
func (h *OrgHandler) Tree(c *gin.Context) {
	user := GetCurrentUser(c)
	var orgs []types.Org
	var err error
	if user.Admin {
		err = db.Find("org", nil, &orgs, 0, 0)
	} else {
		ids := userOrgScopeIDs(c)
		if len(ids) == 0 {
			curd.Success(c, []interface{}{})
			return
		}
		err = db.Find("org", builder.In("id", int64ToInterfaces(ids)...), &orgs, 0, 0)
	}
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	ownerIDs := make([]string, 0, len(orgs))
	for i := range orgs {
		ownerIDs = append(ownerIDs, orgs[i].OwnerUserID)
	}
	names := ownerNames(ownerIDs)

	// 用户关联的根组织 / 负责人组织 / 成员组织
	rootSet := make(map[int64]bool)
	ownerSet := make(map[int64]bool)
	memberSet := make(map[int64]bool)
	if !user.Admin {
		for _, id := range userOrgRootIDs(user) {
			rootSet[id] = true
		}
		var owned []types.Org
		if err := db.Find("org", builder.Eq{"owner_user_id": user.UserID}, &owned, 0, 0); err == nil {
			for _, o := range owned {
				ownerSet[o.ID] = true
			}
		}
		var rels []types.OrgMember
		if err := db.Find("org_member", builder.Eq{"user_id": user.UserID}, &rels, 0, 0); err == nil {
			for _, r := range rels {
				memberSet[r.OrgID] = true
			}
		}
	}

	// 构建树
	nodes := make(map[int64]*OrgNode, len(orgs))
	children := make(map[int64][]*OrgNode)
	for i := range orgs {
		node := &OrgNode{Org: orgs[i]}
		node.OwnerName = names[node.OwnerUserID]
		node.IsOwner = ownerSet[node.ID]
		node.IsMember = memberSet[node.ID]
		node.CanManage = user.Admin || !rootSet[node.ID]
		node.CanMembers = user.Admin || node.IsOwner || !rootSet[node.ID]
		nodes[node.ID] = node
	}
	roots := make([]*OrgNode, 0)
	for _, node := range nodes {
		if parent, ok := nodes[node.ParentID]; ok && node.ParentID != 0 {
			children[parent.ID] = append(children[parent.ID], node)
		} else {
			roots = append(roots, node)
		}
	}
	for id, list := range children {
		if node, ok := nodes[id]; ok {
			node.Children = list
		}
	}
	curd.Success(c, roots)
}

// Get 组织详情
// 路由: GET /org/:id
func (h *OrgHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		curd.Error(c, 400, "无效的组织ID")
		return
	}
	org, ok := loadOrg(id)
	if !ok {
		curd.Error(c, 404, "组织不存在")
		return
	}
	if !requireOrgAccess(c, org) {
		return
	}
	detail := OrgDetail{Org: *org}
	detail.OwnerName = ownerNames([]string{org.OwnerUserID})[org.OwnerUserID]
	curd.Success(c, detail)
}

// Create 创建组织
// 路由: POST /org
// 说明: 管理员可创建根组织（parent_id=0）；其他用户只能在自身可管理范围内创建子级组织
func (h *OrgHandler) Create(c *gin.Context) {
	var req struct {
		ParentID    int64  `json:"parent_id"`
		Name        string `json:"name"`
		Type        string `json:"type"`
		OwnerUserID string `json:"owner_user_id"`
		Region      string `json:"region"`
		Sort        int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		curd.Error(c, 400, "组织名称不能为空")
		return
	}

	var parent *types.Org
	if req.ParentID == 0 {
		if !IsAdminUser(c) {
			curd.Error(c, 403, "仅平台管理员可创建根组织")
			return
		}
	} else {
		var ok bool
		parent, ok = loadOrg(req.ParentID)
		if !ok {
			curd.Error(c, 404, "上级组织不存在")
			return
		}
		if !requireOrgAccess(c, parent) {
			return
		}
	}
	if req.OwnerUserID != "" {
		if found, _ := db.FindOne("user", builder.Eq{"id": req.OwnerUserID}, &types.User{}); !found {
			curd.Error(c, 400, "负责人用户不存在")
			return
		}
	}

	org := types.Org{
		ParentID:    req.ParentID,
		Name:        strings.TrimSpace(req.Name),
		Type:        req.Type,
		OwnerUserID: req.OwnerUserID,
		Region:      req.Region,
		Sort:        req.Sort,
		Created:     time.Now(),
	}
	if _, err := db.InsertOne("org", &org); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 自增 ID 回填后再生成物化路径
	if parent != nil {
		org.Level = parent.Level + 1
		org.Path = parent.Path + strconv.FormatInt(org.ID, 10) + "/"
	} else {
		org.Level = 1
		org.Path = "/" + strconv.FormatInt(org.ID, 10) + "/"
	}
	if _, err := db.UpdateOne("org", builder.Eq{"id": org.ID}, map[string]interface{}{
		"path": org.Path, "level": org.Level,
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	curd.Success(c, org)
}

// Update 更新组织
// 路由: PUT /org/:id
// 说明: 管理员可编辑任意组织；其他用户只能编辑其管理范围内的子级组织（不能编辑关联组织本身）
func (h *OrgHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		curd.Error(c, 400, "无效的组织ID")
		return
	}
	org, ok := loadOrg(id)
	if !ok {
		curd.Error(c, 404, "组织不存在")
		return
	}
	user := GetCurrentUser(c)
	if !user.Admin {
		if containsInt64(userOrgRootIDs(user), id) {
			curd.Error(c, 403, "只能管理子级组织")
			return
		}
		if !requireOrgAccess(c, org) {
			return
		}
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	allowed := map[string]bool{
		"name": true, "type": true, "owner_user_id": true,
		"region": true, "sort": true, "disabled": true,
	}
	updateSet := make(map[string]interface{})
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
	if ownerID, ok := updateSet["owner_user_id"].(string); ok && ownerID != "" {
		if found, _ := db.FindOne("user", builder.Eq{"id": ownerID}, &types.User{}); !found {
			curd.Error(c, 400, "负责人用户不存在")
			return
		}
	}
	if _, err := db.UpdateOne("org", builder.Eq{"id": id}, updateSet); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// Delete 删除组织
// 路由: DELETE /org/:id
// 说明: 存在子组织 / 成员 / 用户 / 设备绑定时禁止删除，避免产生孤儿数据
func (h *OrgHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		curd.Error(c, 400, "无效的组织ID")
		return
	}
	org, ok := loadOrg(id)
	if !ok {
		curd.Error(c, 404, "组织不存在")
		return
	}
	user := GetCurrentUser(c)
	if !user.Admin {
		if containsInt64(userOrgRootIDs(user), id) {
			curd.Error(c, 403, "只能管理子级组织")
			return
		}
		if !requireOrgAccess(c, org) {
			return
		}
	}

	if n, _ := db.Count("org", builder.Eq{"parent_id": id}); n > 0 {
		curd.Error(c, 400, "该组织下还有子组织，不能删除")
		return
	}
	if n, _ := db.Count("org_member", builder.Eq{"org_id": id}); n > 0 {
		curd.Error(c, 400, "该组织下还有成员，不能删除")
		return
	}
	if n, _ := db.Count("device_org", builder.Eq{"org_id": id}); n > 0 {
		curd.Error(c, 400, "该组织下还有绑定设备，不能删除")
		return
	}
	if _, err := db.DeleteOne("org", builder.Eq{"id": id}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// ListMembers 组织成员列表（含负责人信息）
// 路由: GET /org/:id/members
func (h *OrgHandler) ListMembers(c *gin.Context) {
	org, ok := h.loadOrgFromParam(c)
	if !ok {
		return
	}
	if !requireOrgAccess(c, org) {
		return
	}

	var rels []types.OrgMember
	if err := db.Find("org_member", builder.Eq{"org_id": org.ID}, &rels, 0, 0); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	userIDs := make([]interface{}, 0, len(rels)+1)
	if org.OwnerUserID != "" {
		userIDs = append(userIDs, org.OwnerUserID)
	}
	for _, r := range rels {
		userIDs = append(userIDs, r.UserID)
	}
	users := make(map[string]types.User)
	if len(userIDs) > 0 {
		var list []types.User
		if err := db.Find("user", builder.In("id", userIDs...), &list, 0, 0); err == nil {
			for _, u := range list {
				users[u.ID] = u
			}
		}
	}

	members := make([]OrgMemberItem, 0, len(rels))
	for _, r := range rels {
		u := users[r.UserID]
		members = append(members, OrgMemberItem{
			UserID:    r.UserID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Cellphone: u.Cellphone,
			Created:   r.Created,
		})
	}

	ownerName := ""
	if u, ok := users[org.OwnerUserID]; ok {
		ownerName = u.Nickname
		if ownerName == "" {
			ownerName = u.Username
		}
	}
	curd.Success(c, gin.H{
		"org_id":        org.ID,
		"owner_user_id": org.OwnerUserID,
		"owner_name":    ownerName,
		"members":       members,
	})
}

// AddMember 添加组织成员
// 路由: POST /org/:id/members
// 请求体: { "user_id": "xxx" }
func (h *OrgHandler) AddMember(c *gin.Context) {
	org, ok := h.loadOrgFromParam(c)
	if !ok {
		return
	}
	if !requireOrgMemberManage(c, org) {
		return
	}
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == "" {
		curd.Error(c, 400, "请选择要添加的成员")
		return
	}
	var user types.User
	found, err := db.FindOne("user", builder.Eq{"id": req.UserID}, &user)
	if err != nil || !found || user.ID == "" {
		curd.Error(c, 404, "用户不存在")
		return
	}

	// 幂等：已是成员则直接返回
	if n, _ := db.Count("org_member", builder.Eq{"org_id": org.ID, "user_id": req.UserID}); n > 0 {
		curd.Success(c, gin.H{"user_id": req.UserID, "added": true})
		return
	}
	rel := types.OrgMember{OrgID: org.ID, UserID: req.UserID, Created: time.Now()}
	if _, err := db.InsertOne("org_member", &rel); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, gin.H{"user_id": req.UserID, "added": true})
}

// RemoveMember 移除组织成员
// 路由: DELETE /org/:id/members/:user_id
func (h *OrgHandler) RemoveMember(c *gin.Context) {
	org, ok := h.loadOrgFromParam(c)
	if !ok {
		return
	}
	if !requireOrgMemberManage(c, org) {
		return
	}
	userID := c.Param("user_id")
	if userID == "" {
		curd.Error(c, 400, "无效的成员ID")
		return
	}
	affected, err := db.DeleteOne("org_member", builder.Eq{"org_id": org.ID, "user_id": userID})
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if affected == 0 {
		curd.Error(c, 404, "该用户不在组织成员中")
		return
	}
	curd.Success(c, nil)
}

// loadOrgFromParam 读取路径参数中的组织
func (h *OrgHandler) loadOrgFromParam(c *gin.Context) (*types.Org, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		curd.Error(c, 400, "无效的组织ID")
		return nil, false
	}
	org, ok := loadOrg(id)
	if !ok {
		curd.Error(c, 404, "组织不存在")
		return nil, false
	}
	return org, true
}

// SearchUsers 用户搜索（负责人/成员选择器）
// 路由: GET /user/search?keyword=
// 说明: 管理员可检索全部用户；其他用户只能检索自己创建的用户及可管理组织内的用户
func SearchUsers(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	var cond builder.Cond = builder.NewCond()
	if keyword != "" {
		kw := "%" + keyword + "%"
		cond = builder.Or(
			builder.Like{"username", kw},
			builder.Like{"nickname", kw},
			builder.Like{"cellphone", kw},
			builder.Like{"email", kw},
		)
	}

	user := GetCurrentUser(c)
	if !user.Admin {
		scope := userOrgScopeIDs(c)
		conds := []builder.Cond{
			builder.Eq{"parent_id": user.UserID}, // 自己创建的用户
		}
		if len(scope) > 0 {
			// 已加入可管理组织的成员用户
			var rels []types.OrgMember
			if err := db.Find("org_member", builder.In("org_id", int64ToInterfaces(scope)...), &rels, 0, 0); err == nil && len(rels) > 0 {
				memberIDs := make([]interface{}, 0, len(rels))
				for _, r := range rels {
					memberIDs = append(memberIDs, r.UserID)
				}
				conds = append(conds, builder.In("id", memberIDs...))
			}
		}
		cond = builder.And(cond, builder.Or(conds...))
	}

	var users []types.User
	if err := db.Find("user", cond, &users, 1, 20); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	list := make([]UserBrief, 0, len(users))
	for _, u := range users {
		list = append(list, UserBrief{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Cellphone: u.Cellphone,
		})
	}
	curd.Success(c, list)
}
