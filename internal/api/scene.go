// 智能家居场景与定时任务管理接口。
// 场景为跨设备动作集合，支持手动一键执行；定时任务按 cron 触发场景。
// 二者均支持 online（master 服务器执行）/ offline（本地网关执行）两种运行模式，
// 引擎按模式各自加载，互不重复执行。
package api

import (
	"strconv"
	"strings"
	"time"

	sceneengine "github.com/god-jason/iot-master/internal/room"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// SceneHandler 场景管理处理器
type SceneHandler struct {
	curd.Handler[types.Scene]
}

// NewSceneHandler 创建场景处理器
func NewSceneHandler() *SceneHandler {
	return &SceneHandler{
		Handler: curd.Handler[types.Scene]{
			Collection: "scene",
			IsInt64ID:  true,
			Keywords:   []string{"name"},
		},
	}
}

// Get 获取场景详情（校验场景归属）
func (h *SceneHandler) Get(c *gin.Context) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的场景ID")
		return
	}
	scene, ok := loadAccessibleScene(c, id64)
	if !ok {
		return
	}
	curd.Success(c, scene)
}

// Create 创建场景
func (h *SceneHandler) Create(c *gin.Context) {
	var scene types.Scene
	if err := c.ShouldBindJSON(&scene); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if scene.Mode == "" {
		scene.Mode = types.SceneModeOnline
	}
	if !requireFamilyScope(c, scene.FamilyID) {
		return
	}

	scene.Created = time.Now()
	scene.Updated = scene.Created

	id, err := db.InsertOne("scene", &scene)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	scene.ID = id
	sceneengine.ReloadScenes()
	curd.Success(c, scene)
}

// Update 更新场景
func (h *SceneHandler) Update(c *gin.Context) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的场景ID")
		return
	}

	var scene types.Scene
	if err := c.ShouldBindJSON(&scene); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 先按原场景校验权限，再校验新的家庭归属
	if _, ok := loadAccessibleScene(c, id64); !ok {
		return
	}
	if scene.FamilyID != 0 || !IsAdminUser(c) {
		if !requireFamilyScope(c, scene.FamilyID) {
			return
		}
	}

	update := map[string]interface{}{
		"name":      scene.Name,
		"icon":      scene.Icon,
		"desc":      scene.Desc,
		"mode":      scene.Mode,
		"disabled":  scene.Disabled,
		"family_id": scene.FamilyID,
		"actions":   scene.Actions,
		"updated":   time.Now(),
	}
	if _, err := db.UpdateOne("scene", builder.Eq{"id": id64}, update); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	sceneengine.ReloadScenes()
	curd.Success(c, gin.H{"id": id64})
}

// Delete 删除场景
func (h *SceneHandler) Delete(c *gin.Context) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的场景ID")
		return
	}
	if _, ok := loadAccessibleScene(c, id64); !ok {
		return
	}
	h.Handler.Delete(c)
	sceneengine.ReloadScenes()
}

// Switch 启用/停用场景
func (h *SceneHandler) Switch(c *gin.Context) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的场景ID")
		return
	}
	if _, ok := loadAccessibleScene(c, id64); !ok {
		return
	}
	var req struct {
		Disabled bool `json:"disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if _, err := db.UpdateOne("scene", builder.Eq{"id": id64}, map[string]interface{}{"disabled": req.Disabled, "updated": time.Now()}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	sceneengine.ReloadScenes()
	curd.Success(c, gin.H{"id": id64, "disabled": req.Disabled})
}

// Execute 手动立即执行场景，返回逐动作结果
func (h *SceneHandler) Execute(c *gin.Context) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的场景ID")
		return
	}
	if _, ok := loadAccessibleScene(c, id64); !ok {
		return
	}
	results, err := sceneengine.ExecuteScene(id64)
	if err != nil {
		curd.Error(c, 400, err.Error())
		return
	}
	curd.Success(c, gin.H{"results": results})
}

// TaskHandler 定时任务管理处理器
type TaskHandler struct {
	curd.Handler[types.Task]
}

// NewTaskHandler 创建定时任务处理器
func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		Handler: curd.Handler[types.Task]{
			Collection: "task",
			IsInt64ID:  true,
			Keywords:   []string{"name"},
		},
	}
}

// Create 创建定时任务
func (h *TaskHandler) Create(c *gin.Context) {
	var task types.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if task.Mode == "" {
		task.Mode = types.SceneModeOnline
	}
	if task.Cron == "" || task.SceneID == 0 {
		curd.Error(c, 400, "cron表达式与关联场景不能为空")
		return
	}
	if !cronValid(task.Cron) {
		curd.Error(c, 400, "cron表达式格式错误（需5段：分 时 日 月 周）")
		return
	}
	// 关联场景必须存在且当前用户有权限
	if _, ok := loadAccessibleScene(c, task.SceneID); !ok {
		return
	}

	task.Created = time.Now()
	task.Updated = task.Created

	id, err := db.InsertOne("task", &task)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	task.ID = id
	sceneengine.ReloadScenes()
	curd.Success(c, task)
}

// Get 获取定时任务详情（校验任务归属）
func (h *TaskHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的ID")
		return
	}
	task, ok := loadAccessibleTask(c, id)
	if !ok {
		return
	}
	curd.Success(c, task)
}

// Update 更新定时任务（校验任务归属）
func (h *TaskHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的ID")
		return
	}
	if _, ok := loadAccessibleTask(c, id); !ok {
		return
	}

	var task types.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if task.Cron != "" && !cronValid(task.Cron) {
		curd.Error(c, 400, "cron表达式格式错误（需5段：分 时 日 月 周）")
		return
	}
	// 修改关联场景时，新场景也必须在当前用户权限范围内
	if task.SceneID != 0 {
		if _, ok := loadAccessibleScene(c, task.SceneID); !ok {
			return
		}
	}

	update := map[string]interface{}{
		"name":     task.Name,
		"desc":     task.Desc,
		"mode":     task.Mode,
		"disabled": task.Disabled,
		"cron":     task.Cron,
		"scene_id": task.SceneID,
		"updated":  time.Now(),
	}
	if _, err := db.UpdateOne("task", builder.Eq{"id": id}, update); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	sceneengine.ReloadScenes()
	curd.Success(c, gin.H{"id": id})
}

// Delete 删除定时任务（校验任务归属）
func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的ID")
		return
	}
	if _, ok := loadAccessibleTask(c, id); !ok {
		return
	}
	if _, err := db.DeleteOne("task", builder.Eq{"id": id}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	sceneengine.ReloadScenes()
	curd.Success(c, nil)
}

// Switch 启用/停用定时任务（校验任务归属）
func (h *TaskHandler) Switch(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的ID")
		return
	}
	if _, ok := loadAccessibleTask(c, id); !ok {
		return
	}
	var req struct {
		Disabled bool `json:"disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if _, err := db.UpdateOne("task", builder.Eq{"id": id}, map[string]interface{}{"disabled": req.Disabled, "updated": time.Now()}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	sceneengine.ReloadScenes()
	curd.Success(c, gin.H{"id": id, "disabled": req.Disabled})
}

// cronValid 轻量校验 cron 是否为 5 段
func cronValid(expr string) bool {
	return len(strings.Fields(strings.TrimSpace(expr))) == 5
}
