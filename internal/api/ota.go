// OTA 固件与升级任务管理接口
// 固件文件只走 HTTP(S) 下载，MQTT 仅负责下发指令与接收进度上报。
package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/ota"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// OTAHandler OTA 固件与升级任务处理器
type OTAHandler struct{}

// NewOTAHandler 创建 OTA 处理器
func NewOTAHandler() *OTAHandler { return &OTAHandler{} }

// FirmwareItem 固件列表项（附带物模型名称与完整下载地址）
type FirmwareItem struct {
	types.Firmware
	ModelName   string `json:"model_name,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

// OTATaskItem 任务列表项（附带固件名称与版本）
type OTATaskItem struct {
	types.OTATask
	FirmwareName    string `json:"firmware_name,omitempty"`
	FirmwareVersion string `json:"firmware_version,omitempty"`
}

// TaskDeviceItem 任务设备明细项（附带设备名称/序列号/在线状态）
type TaskDeviceItem struct {
	types.OTATaskDevice
	DeviceName      string `json:"device_name,omitempty"`
	DeviceSN        string `json:"device_sn,omitempty"`
	ModelID         string `json:"model_id,omitempty"`
	Online          bool   `json:"online"`
	CurrentFirmware string `json:"current_firmware,omitempty"`
}

// types.Device 使用集中检查，避免重复加载
func loadFirmware(id string) (*types.Firmware, bool) {
	if id == "" {
		return nil, false
	}
	var fw types.Firmware
	found, err := db.FindOne("ota_firmware", builder.Eq{"id": id}, &fw)
	if err != nil || !found || fw.ID == "" {
		return nil, false
	}
	return &fw, true
}

// ============ 固件上传 ============

// allowedFirmwareExt 固件文件扩展名白名单
func allowedFirmwareExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".bin", ".hex", ".img", ".fw", ".zip", ".tar", ".gz", ".tgz", ".rar", ".7z":
		return true
	}
	return false
}

// UploadFirmware 上传固件文件并创建草稿记录
// 路由: POST /upload/firmware，form-data: file + name/version/model_id/channel/min_version/signature/notes
func (h *OTAHandler) UploadFirmware(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	maxSize := ota.MaxSizeBytes()
	// 预留 multipart 表单开销，避免合法大文件被提前截断
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize+16*1024*1024)

	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			curd.Error(c, 400, fmt.Sprintf("固件大小不能超过 %d MB", maxSize/1024/1024))
			return
		}
		curd.Error(c, 400, "请选择要上传的固件文件")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedFirmwareExt(ext) {
		curd.Error(c, 400, "不支持的固件文件类型")
		return
	}
	if file.Size <= 0 || file.Size > maxSize {
		curd.Error(c, 400, fmt.Sprintf("固件大小不能超过 %d MB", maxSize/1024/1024))
		return
	}

	version := strings.TrimSpace(c.PostForm("version"))
	if version == "" {
		curd.Error(c, 400, "请填写固件版本号")
		return
	}
	channel := strings.TrimSpace(c.PostForm("channel"))
	if channel != "beta" {
		channel = "stable"
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		name = strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	}

	dir := ota.StorageDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		curd.Error(c, 500, "无法创建固件存储目录")
		return
	}

	id := db.NewID()
	fileKey := id + ext
	fullPath := filepath.Join(dir, fileKey)

	src, err := file.Open()
	if err != nil {
		curd.Error(c, 500, "无法读取上传文件")
		return
	}
	defer src.Close()

	out, err := os.Create(fullPath)
	if err != nil {
		curd.Error(c, 500, "无法保存固件文件")
		return
	}

	hasher := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(out, hasher), src)
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil || written == 0 {
		_ = os.Remove(fullPath)
		curd.Error(c, 500, "固件文件写入失败")
		return
	}

	now := time.Now()
	fw := types.Firmware{
		ID:         id,
		ModelID:    strings.TrimSpace(c.PostForm("model_id")),
		Name:       name,
		Version:    version,
		FileKey:    fileKey,
		FileURL:    ota.PublicFirmwareURL(id),
		FileSize:   written,
		SHA256:     hex.EncodeToString(hasher.Sum(nil)),
		Signature:  strings.TrimSpace(c.PostForm("signature")),
		MinVersion: strings.TrimSpace(c.PostForm("min_version")),
		Channel:    channel,
		Notes:      strings.TrimSpace(c.PostForm("notes")),
		Status:     types.FirmwareStatusDraft,
		CreatedBy:  GetCurrentUserID(c),
		Created:    now,
		Updated:    now,
	}
	if _, err := db.InsertOne("ota_firmware", fw); err != nil {
		_ = os.Remove(fullPath)
		logger.Error("保存固件记录失败", "id", fw.ID, "error", err)
		curd.Error(c, 500, "保存固件记录失败")
		return
	}
	curd.Success(c, FirmwareItem{Firmware: fw, DownloadURL: ota.DownloadURL(&fw)})
}

// ============ 固件管理 ============

// buildFilterCond 将前端过滤条件转换为 xorm 条件
func buildFilterCond(filters []curd.FilterCondition) builder.Cond {
	conds := make([]builder.Cond, 0, len(filters))
	for _, f := range filters {
		if f.Field == "" {
			continue
		}
		switch f.Operator {
		case "", "eq":
			conds = append(conds, builder.Eq{f.Field: f.Value})
		case "ne":
			conds = append(conds, builder.Neq{f.Field: f.Value})
		case "gt":
			conds = append(conds, builder.Gt{f.Field: f.Value})
		case "gte":
			conds = append(conds, builder.Gte{f.Field: f.Value})
		case "lt":
			conds = append(conds, builder.Lt{f.Field: f.Value})
		case "lte":
			conds = append(conds, builder.Lte{f.Field: f.Value})
		case "like":
			if s, ok := f.Value.(string); ok {
				conds = append(conds, builder.Like{f.Field, "%" + s + "%"})
			}
		case "in":
			if arr, ok := f.Value.([]interface{}); ok && len(arr) > 0 {
				conds = append(conds, builder.In(f.Field, arr...))
			}
		}
	}
	if len(conds) == 0 {
		return nil
	}
	return builder.And(conds...)
}

// SearchFirmware 分页搜索固件
// 路由: POST /firmware/search
func (h *OTAHandler) SearchFirmware(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var params curd.ListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	params.Normalize()

	conds := []builder.Cond{}
	if fc := buildFilterCond(params.Filter); fc != nil {
		conds = append(conds, fc)
	}
	if params.HasKeyword() {
		kw := "%" + params.Keyword + "%"
		conds = append(conds, builder.Or(
			builder.Like{"name", kw},
			builder.Like{"fw_version", kw},
		))
	}
	var cond builder.Cond
	if len(conds) > 0 {
		cond = builder.And(conds...)
	}

	orderBy := params.OrderBy
	if orderBy == "id" {
		orderBy = "created"
	}
	var rows []types.Firmware
	total, err := db.PagedFind("ota_firmware", cond, params.Page, params.PageSize, orderBy, params.Order != "desc", &rows)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	modelNames := modelNamesOf(rows)
	items := make([]FirmwareItem, 0, len(rows))
	for i := range rows {
		items = append(items, FirmwareItem{
			Firmware:    rows[i],
			ModelName:   modelNames[rows[i].ModelID],
			DownloadURL: ota.DownloadURL(&rows[i]),
		})
	}
	curd.SuccessTotal(c, items, total)
}

// modelNamesOf 批量查询物模型名称
func modelNamesOf(rows []types.Firmware) map[string]string {
	ids := make([]interface{}, 0, len(rows))
	for _, r := range rows {
		if r.ModelID != "" {
			ids = append(ids, r.ModelID)
		}
	}
	out := map[string]string{}
	if len(ids) == 0 {
		return out
	}
	var models []types.Model
	if err := db.Find("model", builder.In("id", ids...), &models, 0, 0); err != nil {
		return out
	}
	for _, m := range models {
		out[m.ID] = m.Name
	}
	return out
}

// GetFirmware 固件详情
// 路由: GET /firmware/:id
func (h *OTAHandler) GetFirmware(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	fw, ok := loadFirmware(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	names := modelNamesOf([]types.Firmware{*fw})
	curd.Success(c, FirmwareItem{
		Firmware:    *fw,
		ModelName:   names[fw.ModelID],
		DownloadURL: ota.DownloadURL(fw),
	})
}

// UpdateFirmware 更新固件元信息（不含文件本体）
// 路由: PUT /firmware/:id
func (h *OTAHandler) UpdateFirmware(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	fw, ok := loadFirmware(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	var req struct {
		Name       *string `json:"name"`
		Version    *string `json:"version"`
		ModelID    *string `json:"model_id"`
		Channel    *string `json:"channel"`
		MinVersion *string `json:"min_version"`
		Signature  *string `json:"signature"`
		Notes      *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	fields := map[string]interface{}{"updated": time.Now()}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			curd.Error(c, 400, "固件名称不能为空")
			return
		}
		fields["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Version != nil {
		if strings.TrimSpace(*req.Version) == "" {
			curd.Error(c, 400, "版本号不能为空")
			return
		}
		fields["fw_version"] = strings.TrimSpace(*req.Version)
	}
	if req.ModelID != nil {
		fields["model_id"] = strings.TrimSpace(*req.ModelID)
	}
	if req.Channel != nil {
		channel := strings.TrimSpace(*req.Channel)
		if channel != "beta" {
			channel = "stable"
		}
		fields["channel"] = channel
	}
	if req.MinVersion != nil {
		fields["min_version"] = strings.TrimSpace(*req.MinVersion)
	}
	if req.Signature != nil {
		fields["signature"] = strings.TrimSpace(*req.Signature)
	}
	if req.Notes != nil {
		fields["notes"] = strings.TrimSpace(*req.Notes)
	}

	if _, err := db.UpdateOne("ota_firmware", builder.Eq{"id": fw.ID}, fields); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	updated, _ := loadFirmware(fw.ID)
	curd.Success(c, FirmwareItem{Firmware: *updated, DownloadURL: ota.DownloadURL(updated)})
}

// DeleteFirmware 删除固件（已被任务引用时禁止删除）
// 路由: DELETE /firmware/:id
func (h *OTAHandler) DeleteFirmware(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	fw, ok := loadFirmware(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	if n, _ := db.Count("ota_task", builder.Eq{"firmware_id": fw.ID}); n > 0 {
		curd.Error(c, 400, "该固件已被升级任务引用，无法删除")
		return
	}
	if _, err := db.DeleteOne("ota_firmware", builder.Eq{"id": fw.ID}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if fw.FileKey != "" {
		_ = os.Remove(filepath.Join(ota.StorageDir(), filepath.Base(fw.FileKey)))
	}
	curd.Success(c, nil)
}

// ReleaseFirmware 发布固件
// 路由: PUT /firmware/:id/release
func (h *OTAHandler) ReleaseFirmware(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	fw, ok := loadFirmware(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	if fw.FileKey == "" || fw.SHA256 == "" || fw.Version == "" {
		curd.Error(c, 400, "固件文件或版本信息不完整，无法发布")
		return
	}
	if _, err := db.UpdateOne("ota_firmware", builder.Eq{"id": fw.ID}, map[string]interface{}{
		"status":  types.FirmwareStatusReleased,
		"updated": time.Now(),
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// DisableFirmware 停用固件
// 路由: PUT /firmware/:id/disable
func (h *OTAHandler) DisableFirmware(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	fw, ok := loadFirmware(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	if _, err := db.UpdateOne("ota_firmware", builder.Eq{"id": fw.ID}, map[string]interface{}{
		"status":  types.FirmwareStatusDisabled,
		"updated": time.Now(),
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// ============ 固件黑白名单 ============

// FirmwareRuleItem 固件名单规则（附带目标名称与覆盖设备数）
type FirmwareRuleItem struct {
	types.FirmwareRule
	TargetName  string `json:"target_name,omitempty"`
	TargetSN    string `json:"target_sn,omitempty"`
	DeviceCount int    `json:"device_count"`
}

// buildFirmwareRuleItems 为名单规则补充目标名称与覆盖设备数
func buildFirmwareRuleItems(rules []types.FirmwareRule) []FirmwareRuleItem {
	deviceIDs := make([]interface{}, 0, len(rules))
	groupIDs := make([]interface{}, 0, len(rules))
	orgIDs := make([]interface{}, 0, len(rules))
	for _, r := range rules {
		switch r.TargetType {
		case types.RuleTargetDevice:
			deviceIDs = append(deviceIDs, r.TargetID)
		case types.RuleTargetGroup:
			if id, err := strconv.ParseInt(r.TargetID, 10, 64); err == nil {
				groupIDs = append(groupIDs, id)
			}
		case types.RuleTargetOrg:
			if id, err := strconv.ParseInt(r.TargetID, 10, 64); err == nil {
				orgIDs = append(orgIDs, id)
			}
		}
	}

	devices := map[string]types.Device{}
	if len(deviceIDs) > 0 {
		var list []types.Device
		if err := db.Find("device", builder.In("id", deviceIDs...), &list, 0, 0); err == nil {
			for _, d := range list {
				devices[d.ID] = d
			}
		}
	}
	groups := map[int64]string{}
	if len(groupIDs) > 0 {
		var list []types.Group
		if err := db.Find("group", builder.In("id", groupIDs...), &list, 0, 0); err == nil {
			for _, g := range list {
				groups[g.ID] = g.Name
			}
		}
	}
	orgs := map[int64]string{}
	if len(orgIDs) > 0 {
		var list []types.Org
		if err := db.Find("org", builder.In("id", orgIDs...), &list, 0, 0); err == nil {
			for _, o := range list {
				orgs[o.ID] = o.Name
			}
		}
	}

	items := make([]FirmwareRuleItem, 0, len(rules))
	for _, r := range rules {
		item := FirmwareRuleItem{FirmwareRule: r}
		switch r.TargetType {
		case types.RuleTargetDevice:
			if d, ok := devices[r.TargetID]; ok {
				item.TargetName = d.Name
				item.TargetSN = d.SN
			}
		case types.RuleTargetGroup:
			if id, err := strconv.ParseInt(r.TargetID, 10, 64); err == nil {
				item.TargetName = groups[id]
			}
		case types.RuleTargetOrg:
			if id, err := strconv.ParseInt(r.TargetID, 10, 64); err == nil {
				item.TargetName = orgs[id]
			}
		}
		if ids, err := ota.RuleTargetDeviceIDs(r.TargetType, r.TargetID); err == nil {
			item.DeviceCount = len(ids)
		}
		items = append(items, item)
	}
	return items
}

// ListFirmwareRules 固件的黑白名单规则
// 路由: GET /firmware/:id/rules
func (h *OTAHandler) ListFirmwareRules(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	fw, ok := loadFirmware(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	var rules []types.FirmwareRule
	if err := db.Find("ota_firmware_rule", builder.Eq{"firmware_id": fw.ID}, &rules, 0, 0); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	sort.SliceStable(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })

	items := buildFirmwareRuleItems(rules)
	white, black := 0, 0
	for _, r := range rules {
		if r.ListType == types.FirmwareListWhite {
			white++
		} else if r.ListType == types.FirmwareListBlack {
			black++
		}
	}
	curd.Success(c, gin.H{
		"rules":       items,
		"white_count": white,
		"black_count": black,
	})
}

// AddFirmwareRules 批量添加黑白名单规则
// 路由: POST /firmware/:id/rules，body: list_type + target_type + target_ids
func (h *OTAHandler) AddFirmwareRules(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	fw, ok := loadFirmware(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	var req struct {
		ListType   string   `json:"list_type"`
		TargetType string   `json:"target_type"`
		TargetIDs  []string `json:"target_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	listType := strings.TrimSpace(req.ListType)
	if listType != types.FirmwareListWhite && listType != types.FirmwareListBlack {
		curd.Error(c, 400, "名单类型只能是 white 或 black")
		return
	}
	targetType := strings.TrimSpace(req.TargetType)
	if targetType != types.RuleTargetDevice && targetType != types.RuleTargetGroup && targetType != types.RuleTargetOrg {
		curd.Error(c, 400, "不支持的名单目标类型")
		return
	}

	// 去重并校验目标存在
	ids := make([]string, 0, len(req.TargetIDs))
	seen := map[string]bool{}
	for _, raw := range req.TargetIDs {
		id := strings.TrimSpace(raw)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		curd.Error(c, 400, "请选择要加入名单的设备/分组/组织")
		return
	}
	if len(ids) > 500 {
		curd.Error(c, 400, "单次最多添加 500 项")
		return
	}

	args := make([]interface{}, 0, len(ids))
	switch targetType {
	case types.RuleTargetDevice:
		for _, id := range ids {
			args = append(args, id)
		}
		n, err := db.Count("device", builder.In("id", args...))
		if err != nil {
			curd.Error(c, 500, err.Error())
			return
		}
		if n != int64(len(ids)) {
			curd.Error(c, 400, "部分设备不存在")
			return
		}
	case types.RuleTargetGroup:
		for _, id := range ids {
			gid, err := strconv.ParseInt(id, 10, 64)
			if err != nil || gid <= 0 {
				curd.Error(c, 400, "无效的分组ID")
				return
			}
			args = append(args, gid)
		}
		n, err := db.Count("group", builder.In("id", args...))
		if err != nil {
			curd.Error(c, 500, err.Error())
			return
		}
		if n != int64(len(ids)) {
			curd.Error(c, 400, "部分分组不存在")
			return
		}
	case types.RuleTargetOrg:
		for _, id := range ids {
			oid, err := strconv.ParseInt(id, 10, 64)
			if err != nil || oid <= 0 {
				curd.Error(c, 400, "无效的组织ID")
				return
			}
			args = append(args, oid)
		}
		n, err := db.Count("org", builder.In("id", args...))
		if err != nil {
			curd.Error(c, 500, err.Error())
			return
		}
		if n != int64(len(ids)) {
			curd.Error(c, 400, "部分组织不存在")
			return
		}
	}

	// 过滤已存在的规则，避免重复添加
	var existing []types.FirmwareRule
	if err := db.Find("ota_firmware_rule", builder.And(
		builder.Eq{"firmware_id": fw.ID},
		builder.Eq{"list_type": listType},
		builder.Eq{"target_type": targetType},
	), &existing, 0, 0); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	for _, r := range existing {
		delete(seen, r.TargetID)
	}

	now := time.Now()
	rows := make([]types.FirmwareRule, 0, len(seen))
	for _, id := range ids {
		if !seen[id] {
			continue
		}
		rows = append(rows, types.FirmwareRule{
			FirmwareID: fw.ID,
			ListType:   listType,
			TargetType: targetType,
			TargetID:   id,
			CreatedBy:  GetCurrentUserID(c),
			Created:    now,
		})
	}
	if len(rows) > 0 {
		if _, err := db.Engine.Table("ota_firmware_rule").Insert(&rows); err != nil {
			curd.Error(c, 500, err.Error())
			return
		}
	}
	curd.Success(c, gin.H{"added": len(rows), "skipped": len(ids) - len(rows)})
}

// DeleteFirmwareRule 删除单条名单规则
// 路由: DELETE /firmware/:id/rules/:rule_id
func (h *OTAHandler) DeleteFirmwareRule(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	fw, ok := loadFirmware(c.Param("id"))
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	ruleID, err := strconv.ParseInt(c.Param("rule_id"), 10, 64)
	if err != nil || ruleID <= 0 {
		curd.Error(c, 400, "无效的规则ID")
		return
	}
	if _, err := db.DeleteOne("ota_firmware_rule", builder.And(
		builder.Eq{"id": ruleID},
		builder.Eq{"firmware_id": fw.ID},
	)); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// DownloadFirmware 固件文件下载（公开接口，无 token，支持 Range）
// 路由: GET /ota/firmware/:id/file
func DownloadFirmware(c *gin.Context) {
	fw, ok := loadFirmware(c.Param("id"))
	if !ok || fw.FileKey == "" {
		c.String(http.StatusNotFound, "firmware not found")
		return
	}
	fullPath := filepath.Join(ota.StorageDir(), filepath.Base(fw.FileKey))
	f, err := os.Open(fullPath)
	if err != nil {
		c.String(http.StatusNotFound, "firmware file not found")
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		c.String(http.StatusInternalServerError, "firmware file unavailable")
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(fw.FileKey)))
	c.Header("X-Firmware-SHA256", fw.SHA256)
	http.ServeContent(c.Writer, c.Request, fw.FileKey, stat.ModTime(), f)
}

// ============ 升级任务 ============

// taskRequest 创建/预览任务请求
type taskRequest struct {
	Name         string        `json:"name"`
	FirmwareID   string        `json:"firmware_id"`
	ScopeType    string        `json:"scope_type"`
	ScopeID      string        `json:"scope_id"`
	Strategy     *ota.Strategy `json:"strategy"`
	ScheduledAt  int64         `json:"scheduled_at"` // 毫秒时间戳，0 表示立即
	AutoRollback bool          `json:"auto_rollback"`
}

// scopeLabel 目标范围中文名
func scopeLabel(scopeType string) string {
	switch scopeType {
	case types.OTAScopeModel:
		return "型号"
	case types.OTAScopeGroup:
		return "分组"
	case types.OTAScopeOrg:
		return "组织"
	case types.OTAScopeDevice:
		return "设备"
	}
	return scopeType
}

// normalizeTaskRequest 校验并补全任务参数
func normalizeTaskRequest(req *taskRequest) (ota.Strategy, time.Time, error) {
	req.FirmwareID = strings.TrimSpace(req.FirmwareID)
	if req.FirmwareID == "" {
		return ota.Strategy{}, time.Time{}, fmt.Errorf("请选择升级固件")
	}
	switch req.ScopeType {
	case types.OTAScopeModel, types.OTAScopeGroup, types.OTAScopeOrg, types.OTAScopeDevice:
	default:
		return ota.Strategy{}, time.Time{}, fmt.Errorf("不支持的升级目标类型")
	}
	if strings.TrimSpace(req.ScopeID) == "" {
		return ota.Strategy{}, time.Time{}, fmt.Errorf("请选择升级目标")
	}

	st := ota.DefaultStrategy()
	if req.Strategy != nil {
		st = *req.Strategy
	}
	ota.NormalizeStrategy(&st)
	if err := ota.ValidateStrategy(st); err != nil {
		return ota.Strategy{}, time.Time{}, err
	}

	var scheduledAt time.Time
	if req.ScheduledAt > 0 {
		scheduledAt = time.UnixMilli(req.ScheduledAt)
	}
	return st, scheduledAt, nil
}

// PreviewTask 预览目标设备数量与批次切分
// 路由: POST /ota/task/preview
func (h *OTAHandler) PreviewTask(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var req taskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	st, _, err := normalizeTaskRequest(&req)
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}
	fw, ok := loadFirmware(req.FirmwareID)
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	if fw.Status != types.FirmwareStatusReleased {
		curd.Error(c, 400, "固件未发布，不能用于升级")
		return
	}
	if ota.ServerBase() == "" {
		curd.Error(c, 400, "请先在配置中设置 ota.server_url，设备需要可访问的固件下载地址")
		return
	}

	devices, stats, err := ota.PrepareTargets(fw, req.ScopeType, req.ScopeID, "")
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}

	online := 0
	sample := make([]gin.H, 0, 20)
	for i := range devices {
		isOnline := false
		if dev, ok := iot.GetDevice(devices[i].ID); ok {
			isOnline = dev.IsOnline()
		}
		if isOnline {
			online++
		}
		if i < 20 {
			sample = append(sample, gin.H{
				"id":               devices[i].ID,
				"name":             devices[i].Name,
				"sn":               devices[i].SN,
				"firmware_version": devices[i].FirmwareVersion,
				"online":           isOnline,
			})
		}
	}

	plans := ota.BuildBatchPlans(len(devices), st)
	curd.Success(c, gin.H{
		"total":            len(devices),
		"online":           online,
		"offline":          len(devices) - online,
		"excluded":         stats.Excluded,
		"white_filtered":   stats.WhiteFiltered,
		"black_filtered":   stats.BlackFiltered,
		"whitelist_active": stats.WhitelistActive,
		"rule_filtered":    stats.RuleFiltered(),
		"scope_label":      scopeLabel(req.ScopeType),
		"firmware_version": fw.Version,
		"strategy":         st,
		"batches":          plans,
		"devices":          sample,
	})
}

// createTaskFromDevices 创建任务、批次与设备明细
func createTaskFromDevices(fw *types.Firmware, devices []types.Device, name, scopeType, scopeID string, st ota.Strategy, scheduledAt time.Time, autoRollback bool, createdBy string) (*types.OTATask, error) {
	if len(devices) == 0 {
		return nil, fmt.Errorf("没有可升级的设备")
	}
	ota.NormalizeStrategy(&st)
	if err := ota.ValidateStrategy(st); err != nil {
		return nil, err
	}
	plans := ota.BuildBatchPlans(len(devices), st)
	rawStrategy, _ := json.Marshal(st)

	now := time.Now()
	task := types.OTATask{
		ID:           db.NewID(),
		Name:         name,
		FirmwareID:   fw.ID,
		ScopeType:    scopeType,
		ScopeID:      scopeID,
		Strategy:     string(rawStrategy),
		ScheduledAt:  scheduledAt,
		Status:       types.OTATaskStatusRunning,
		Total:        len(devices),
		AutoRollback: autoRollback,
		CreatedBy:    createdBy,
		StartedAt:    now,
		Created:      now,
	}
	if !scheduledAt.IsZero() && scheduledAt.After(now) {
		task.Status = types.OTATaskStatusScheduled
		task.StartedAt = time.Time{}
	}
	if _, err := db.InsertOne("ota_task", &task); err != nil {
		return nil, err
	}

	batches := make([]types.OTATaskBatch, 0, len(plans))
	for _, p := range plans {
		batches = append(batches, types.OTATaskBatch{
			TaskID:      task.ID,
			Seq:         p.Seq,
			Percent:     p.Percent,
			WaitMinutes: p.WaitMinutes,
			Status:      types.OTABatchStatusPending,
		})
	}
	if len(batches) > 0 {
		if _, err := db.Engine.Table("ota_task_batch").Insert(&batches); err != nil {
			return nil, err
		}
	}

	// 按批次顺序切分设备
	rows := make([]types.OTATaskDevice, 0, len(devices))
	index := 0
	for _, p := range plans {
		for n := 0; n < p.Count && index < len(devices); n++ {
			d := devices[index]
			index++
			rows = append(rows, types.OTATaskDevice{
				TaskID:      task.ID,
				BatchSeq:    p.Seq,
				DeviceID:    d.ID,
				FromVersion: d.FirmwareVersion,
				ToVersion:   fw.Version,
				Status:      types.OTADeviceStatusPending,
				Updated:     now,
			})
		}
	}
	// 兜底：策略切分异常时把剩余设备放入最后一批
	if index < len(devices) && len(plans) > 0 {
		lastSeq := plans[len(plans)-1].Seq
		for ; index < len(devices); index++ {
			d := devices[index]
			rows = append(rows, types.OTATaskDevice{
				TaskID:      task.ID,
				BatchSeq:    lastSeq,
				DeviceID:    d.ID,
				FromVersion: d.FirmwareVersion,
				ToVersion:   fw.Version,
				Status:      types.OTADeviceStatusPending,
				Updated:     now,
			})
		}
	}
	for start := 0; start < len(rows); start += 200 {
		end := start + 200
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[start:end]
		if _, err := db.Engine.Table("ota_task_device").Insert(&chunk); err != nil {
			return nil, err
		}
	}
	return &task, nil
}

// CreateTask 创建升级任务
// 路由: POST /ota/task
func (h *OTAHandler) CreateTask(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var req taskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	st, scheduledAt, err := normalizeTaskRequest(&req)
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}
	fw, ok := loadFirmware(req.FirmwareID)
	if !ok {
		curd.Error(c, 404, "固件不存在")
		return
	}
	if fw.Status != types.FirmwareStatusReleased {
		curd.Error(c, 400, "固件未发布，不能用于升级")
		return
	}
	if ota.ServerBase() == "" {
		curd.Error(c, 400, "请先在配置中设置 ota.server_url，设备需要可访问的固件下载地址")
		return
	}

	devices, stats, err := ota.PrepareTargets(fw, req.ScopeType, req.ScopeID, "")
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}
	if len(devices) == 0 {
		switch {
		case stats.RuleFiltered() > 0 && stats.Excluded > 0:
			curd.Error(c, 400, "目标设备均被固件黑白名单过滤或已有进行中的升级任务")
		case stats.RuleFiltered() > 0:
			curd.Error(c, 400, "目标设备均被固件黑白名单排除")
		case stats.Excluded > 0:
			curd.Error(c, 400, "目标设备均已有进行中的升级任务")
		default:
			curd.Error(c, 400, "目标范围内没有可升级的设备")
		}
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("%s %s 升级", fw.Name, fw.Version)
	}

	task, err := createTaskFromDevices(fw, devices, name, req.ScopeType, req.ScopeID, st, scheduledAt, req.AutoRollback, GetCurrentUserID(c))
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, gin.H{
		"task":             task,
		"excluded":         stats.Excluded,
		"white_filtered":   stats.WhiteFiltered,
		"black_filtered":   stats.BlackFiltered,
		"rule_filtered":    stats.RuleFiltered(),
		"whitelist_active": stats.WhitelistActive,
	})
}

// SearchTasks 分页搜索升级任务
// 路由: POST /ota/task/search
func (h *OTAHandler) SearchTasks(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var params curd.ListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	params.Normalize()

	conds := []builder.Cond{}
	if fc := buildFilterCond(params.Filter); fc != nil {
		conds = append(conds, fc)
	}
	if params.HasKeyword() {
		conds = append(conds, builder.Like{"name", "%" + params.Keyword + "%"})
	}
	var cond builder.Cond
	if len(conds) > 0 {
		cond = builder.And(conds...)
	}
	orderBy := params.OrderBy
	if orderBy == "id" {
		orderBy = "created"
	}
	var rows []types.OTATask
	total, err := db.PagedFind("ota_task", cond, params.Page, params.PageSize, orderBy, params.Order != "desc", &rows)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	fwIDs := make([]interface{}, 0, len(rows))
	for _, r := range rows {
		if r.FirmwareID != "" {
			fwIDs = append(fwIDs, r.FirmwareID)
		}
	}
	firmwares := map[string]types.Firmware{}
	if len(fwIDs) > 0 {
		var list []types.Firmware
		if err := db.Find("ota_firmware", builder.In("id", fwIDs...), &list, 0, 0); err == nil {
			for _, fw := range list {
				firmwares[fw.ID] = fw
			}
		}
	}

	items := make([]OTATaskItem, 0, len(rows))
	for _, r := range rows {
		fw := firmwares[r.FirmwareID]
		items = append(items, OTATaskItem{
			OTATask:         r,
			FirmwareName:    fw.Name,
			FirmwareVersion: fw.Version,
		})
	}
	curd.SuccessTotal(c, items, total)
}

// GetTask 任务详情（含批次）
// 路由: GET /ota/task/:id
func (h *OTAHandler) GetTask(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var task types.OTATask
	found, err := db.FindOne("ota_task", builder.Eq{"id": c.Param("id")}, &task)
	if err != nil || !found || task.ID == "" {
		curd.Error(c, 404, "任务不存在")
		return
	}
	var fw types.Firmware
	_, _ = db.FindOne("ota_firmware", builder.Eq{"id": task.FirmwareID}, &fw)

	var batches []types.OTATaskBatch
	_ = db.Find("ota_task_batch", builder.Eq{"task_id": task.ID}, &batches, 0, 0)
	sort.SliceStable(batches, func(i, j int) bool { return batches[i].Seq < batches[j].Seq })

	strategy := ota.ParseStrategy(task.Strategy)
	curd.Success(c, gin.H{
		"task":     task,
		"firmware": fw,
		"batches":  batches,
		"strategy": strategy,
	})
}

// ListTaskDevices 分页查询任务设备明细
// 路由: GET /ota/task/:id/devices
func (h *OTAHandler) ListTaskDevices(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	taskID := c.Param("id")
	var task types.OTATask
	found, err := db.FindOne("ota_task", builder.Eq{"id": taskID}, &task)
	if err != nil || !found || task.ID == "" {
		curd.Error(c, 404, "任务不存在")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}
	status := strings.TrimSpace(c.Query("status"))
	keyword := strings.TrimSpace(c.Query("keyword"))

	conds := []builder.Cond{builder.Eq{"task_id": taskID}}
	if status != "" {
		conds = append(conds, builder.Eq{"status": status})
	}
	if keyword != "" {
		var devices []types.Device
		kw := "%" + keyword + "%"
		_ = db.Find("device", builder.Or(builder.Like{"name", kw}, builder.Like{"sn", kw}), &devices, 0, 0)
		if len(devices) == 0 {
			curd.SuccessTotal(c, []TaskDeviceItem{}, 0)
			return
		}
		ids := make([]interface{}, 0, len(devices))
		for _, d := range devices {
			ids = append(ids, d.ID)
		}
		conds = append(conds, builder.In("device_id", ids...))
	}

	var rows []types.OTATaskDevice
	total, err := db.PagedFind("ota_task_device", builder.And(conds...), page, pageSize, "id", true, &rows)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	devices := loadDevicesByIDs(rows)
	items := make([]TaskDeviceItem, 0, len(rows))
	for _, r := range rows {
		d := devices[r.DeviceID]
		isOnline := false
		if dev, ok := iot.GetDevice(r.DeviceID); ok {
			isOnline = dev.IsOnline()
		}
		items = append(items, TaskDeviceItem{
			OTATaskDevice:   r,
			DeviceName:      d.Name,
			DeviceSN:        d.SN,
			ModelID:         d.ModelID,
			Online:          isOnline,
			CurrentFirmware: d.FirmwareVersion,
		})
	}
	curd.SuccessTotal(c, items, total)
}

// loadDevicesByIDs 批量加载设备，返回 ID → Device
func loadDevicesByIDs(rows []types.OTATaskDevice) map[string]types.Device {
	ids := make([]interface{}, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.DeviceID)
	}
	out := map[string]types.Device{}
	if len(ids) == 0 {
		return out
	}
	var devices []types.Device
	if err := db.Find("device", builder.In("id", ids...), &devices, 0, 0); err != nil {
		return out
	}
	for _, d := range devices {
		out[d.ID] = d
	}
	return out
}

// loadTask 读取任务，失败时写入响应
func loadTask(c *gin.Context) (*types.OTATask, bool) {
	var task types.OTATask
	found, err := db.FindOne("ota_task", builder.Eq{"id": c.Param("id")}, &task)
	if err != nil || !found || task.ID == "" {
		curd.Error(c, 404, "任务不存在")
		return nil, false
	}
	return &task, true
}

// PauseTask 暂停任务
// 路由: PUT /ota/task/:id/pause
func (h *OTAHandler) PauseTask(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	task, ok := loadTask(c)
	if !ok {
		return
	}
	if task.Status != types.OTATaskStatusRunning && task.Status != types.OTATaskStatusScheduled {
		curd.Error(c, 400, "当前任务状态不支持暂停")
		return
	}
	if _, err := db.UpdateOne("ota_task", builder.Eq{"id": task.ID}, map[string]interface{}{
		"status":       types.OTATaskStatusPaused,
		"pause_reason": "manual",
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// ResumeTask 恢复任务
// 路由: PUT /ota/task/:id/resume
func (h *OTAHandler) ResumeTask(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	task, ok := loadTask(c)
	if !ok {
		return
	}
	if task.Status != types.OTATaskStatusPaused {
		curd.Error(c, 400, "当前任务状态不支持恢复")
		return
	}
	if _, err := db.UpdateOne("ota_task", builder.Eq{"id": task.ID}, map[string]interface{}{
		"status":       types.OTATaskStatusRunning,
		"pause_reason": "",
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// CancelTask 取消任务：未终端设备标记为已取消
// 路由: PUT /ota/task/:id/cancel
func (h *OTAHandler) CancelTask(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	task, ok := loadTask(c)
	if !ok {
		return
	}
	switch task.Status {
	case types.OTATaskStatusCompleted, types.OTATaskStatusPartial,
		types.OTATaskStatusFailed, types.OTATaskStatusCancelled:
		curd.Error(c, 400, "任务已结束，无法取消")
		return
	}

	now := time.Now()
	active := []interface{}{
		types.OTADeviceStatusPending,
		types.OTADeviceStatusDispatched,
		types.OTADeviceStatusDownloading,
		types.OTADeviceStatusVerifying,
		types.OTADeviceStatusInstalling,
		types.OTADeviceStatusRebooting,
	}
	if _, err := db.UpdateOne("ota_task_device",
		builder.And(builder.Eq{"task_id": task.ID}, builder.In("status", active...)),
		map[string]interface{}{
			"status":      types.OTADeviceStatusCancelled,
			"finished_at": now,
			"updated":     now,
		}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	_, _ = db.UpdateOne("ota_task_batch",
		builder.And(builder.Eq{"task_id": task.ID}, builder.In("status", []interface{}{types.OTABatchStatusPending, types.OTABatchStatusRunning})),
		map[string]interface{}{
			"status":      types.OTABatchStatusCompleted,
			"finished_at": now,
		})
	if _, err := db.UpdateOne("ota_task", builder.Eq{"id": task.ID}, map[string]interface{}{
		"status":      types.OTATaskStatusCancelled,
		"finished_at": now,
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// RetryFailed 重试失败/超时设备：创建新的重试批次，避免破坏原有灰度顺序
// 路由: PUT /ota/task/:id/retry-failed
func (h *OTAHandler) RetryFailed(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	task, ok := loadTask(c)
	if !ok {
		return
	}
	switch task.Status {
	case types.OTATaskStatusRunning, types.OTATaskStatusScheduled:
		curd.Error(c, 400, "任务执行中，请先暂停再重试失败设备")
		return
	}

	var failed []types.OTATaskDevice
	if err := db.Find("ota_task_device",
		builder.And(
			builder.Eq{"task_id": task.ID},
			builder.In("status", []interface{}{types.OTADeviceStatusFailed, types.OTADeviceStatusTimeout}),
		), &failed, 0, 0); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if len(failed) == 0 {
		curd.Error(c, 400, "没有需要重试的失败设备")
		return
	}

	// 取当前最大批次序号
	var batches []types.OTATaskBatch
	_ = db.Find("ota_task_batch", builder.Eq{"task_id": task.ID}, &batches, 0, 0)
	maxSeq := 0
	for _, b := range batches {
		if b.Seq > maxSeq {
			maxSeq = b.Seq
		}
	}
	nextSeq := maxSeq + 1
	now := time.Now()
	batch := types.OTATaskBatch{
		TaskID:    task.ID,
		Seq:       nextSeq,
		Percent:   0,
		Status:    types.OTABatchStatusRunning,
		StartedAt: now,
	}
	if _, err := db.InsertOne("ota_task_batch", &batch); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	ids := make([]interface{}, 0, len(failed))
	for _, r := range failed {
		ids = append(ids, r.ID)
	}
	if _, err := db.UpdateOne("ota_task_device", builder.In("id", ids...), map[string]interface{}{
		"batch_seq":     nextSeq,
		"status":        types.OTADeviceStatusPending,
		"retry_count":   0,
		"error":         "",
		"dispatched_at": nil,
		"finished_at":   nil,
		"updated":       now,
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if _, err := db.UpdateOne("ota_task", builder.Eq{"id": task.ID}, map[string]interface{}{
		"status":      types.OTATaskStatusRunning,
		"finished_at": nil,
	}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, gin.H{"retry_count": len(failed), "batch_seq": nextSeq})
}

// UpgradeDevice 单设备升级（复用任务模型，scope=device）
// 路由: POST /device/:id/ota
func (h *OTAHandler) UpgradeDevice(c *gin.Context) {
	device, ok := loadAccessibleDevice(c, c.Param("id"))
	if !ok {
		return
	}
	var req struct {
		FirmwareID string `json:"firmware_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	fw, found := loadFirmware(strings.TrimSpace(req.FirmwareID))
	if !found {
		curd.Error(c, 404, "固件不存在")
		return
	}
	if fw.Status != types.FirmwareStatusReleased {
		curd.Error(c, 400, "固件未发布，不能用于升级")
		return
	}
	if fw.ModelID != "" && fw.ModelID != device.ModelID {
		curd.Error(c, 400, "固件与设备型号不匹配")
		return
	}
	rules, err := ota.LoadRuleSet(fw.ID)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if !rules.Allowed(device.ID) {
		curd.Error(c, 400, "该设备被固件黑白名单限制，不允许升级")
		return
	}
	if device.Disabled {
		curd.Error(c, 400, "设备已禁用")
		return
	}
	if ota.ServerBase() == "" {
		curd.Error(c, 400, "请先在配置中设置 ota.server_url")
		return
	}
	busy, err := ota.ActiveDeviceIDs([]string{device.ID}, "")
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if busy[device.ID] {
		curd.Error(c, 400, "该设备已有进行中的升级任务")
		return
	}

	st := ota.DefaultStrategy()
	name := fmt.Sprintf("%s %s 升级", device.Name, fw.Version)
	task, err := createTaskFromDevices(fw, []types.Device{*device}, name, types.OTAScopeDevice, device.ID, st, time.Time{}, false, GetCurrentUserID(c))
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, task)
}
