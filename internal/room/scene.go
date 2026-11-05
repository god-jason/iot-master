// Package room 房间与场景引擎：离线/在线模式均可用，由调用方注入设备能力（types.DeviceCaller）。
// Room（房间）类型定义在 types 包；SceneEngine 负责场景执行和定时调度。
package room

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"xorm.io/builder"
)

// SceneEngine 场景引擎：离线/在线模式均可用，由调用方注入设备能力（types.DeviceCaller）
type SceneEngine struct {
	caller types.DeviceCaller // 设备能力：在线状态 / 属性下发 / 服务调用
	mode   string             // 运行模式（offline / online），加载数据时按此过滤

	scenesMu sync.RWMutex  // 保护 scenes / tasks
	scenes   []types.Scene // 当前模式下的启用场景
	tasks    []types.Task  // 当前模式下的启用定时任务
}

var (
	defaultEngine    *SceneEngine          // 全局场景引擎实例，由 Start 赋值
	errSceneNotFound = errors.New("场景不存在") // 场景 ID 查不到
	errSceneDisabled = errors.New("场景未启用") // 场景已被禁用
)

// Start 启动场景引擎：加载场景+定时任务，启动定时调度循环
func Start(caller types.DeviceCaller, mode string) *SceneEngine {
	e := &SceneEngine{caller: caller, mode: mode}
	defaultEngine = e

	e.reloadScenes()
	go e.sceneScheduleLoop()

	logger.Info("场景引擎已启动", "mode", mode)
	return e
}

// ExecuteScene 手动执行场景
func ExecuteScene(sceneID int64) ([]types.SceneActionResult, error) {
	if defaultEngine == nil {
		return nil, errors.New("场景引擎未启动")
	}
	return defaultEngine.ExecuteScene(sceneID)
}

// EnabledScenes 返回当前引擎模式下全部启用场景
func EnabledScenes() []types.Scene {
	if defaultEngine == nil {
		return nil
	}
	return defaultEngine.EnabledScenes()
}

// ReloadScenes 立即从数据库刷新场景与定时任务
func ReloadScenes() {
	if defaultEngine != nil {
		defaultEngine.reloadScenes()
	}
}

// —— Engine 方法 ——

// reloadScenes 从 scene / task 表加载当前模式（离线/在线）下未禁用的场景与定时任务，
// 整体替换内存快照。任一查询失败则放弃本次刷新，保留旧数据。
func (e *SceneEngine) reloadScenes() {
	var scenes []types.Scene
	if err := db.Find("scene", builder.Eq{"disabled": false, "mode": e.mode}, &scenes, 0, 0); err != nil {
		logger.Warn("加载场景失败", "error", err, "mode", e.mode)
		return
	}
	var tasks []types.Task
	if err := db.Find("task", builder.Eq{"disabled": false, "mode": e.mode}, &tasks, 0, 0); err != nil {
		logger.Warn("加载定时任务失败", "error", err, "mode", e.mode)
		return
	}
	e.scenesMu.Lock()
	e.scenes = scenes
	e.tasks = tasks
	e.scenesMu.Unlock()
}

// ExecuteScene 执行指定场景：先查缓存，未命中回源库；场景不存在或已禁用分别返回对应错误。
func (e *SceneEngine) ExecuteScene(sceneID int64) ([]types.SceneActionResult, error) {
	scene, ok := e.lookupScene(sceneID)
	if !ok {
		return nil, errSceneNotFound
	}
	if scene.Disabled {
		return nil, errSceneDisabled
	}
	return e.runScene(scene), nil
}

// runScene 顺序执行场景的全部动作，单个动作失败不影响后续动作。
func (e *SceneEngine) runScene(scene types.Scene) []types.SceneActionResult {
	results := make([]types.SceneActionResult, 0, len(scene.Actions))
	for _, act := range scene.Actions {
		results = append(results, e.runSceneAction(act))
	}
	return results
}

// runSceneAction 执行单个场景动作：先校验目标设备、在线状态与动作参数，
// 再按类型下发属性或调用服务，结果统一包装成 SceneActionResult。
func (e *SceneEngine) runSceneAction(act types.SceneAction) (r types.SceneActionResult) {
	r.DeviceID = act.DeviceID
	r.Type = act.Type
	if act.DeviceID == "" {
		r.Success, r.Message = false, "未配置目标设备"
		return
	}
	if !e.caller.IsOnline(act.DeviceID) {
		r.Success, r.Message = false, "设备不在线"
		return
	}
	switch act.Type {
	case types.SceneActionProperty:
		if len(act.Properties) == 0 {
			r.Success, r.Message = false, "未配置下发属性"
			return
		}
		resp, err := e.caller.SetProperty(act.DeviceID, act.Properties)
		r.Success, r.Message = types.ResolveCallResult(resp, err)
	case types.SceneActionService:
		if act.Service == "" {
			r.Success, r.Message = false, "未配置服务名"
			return
		}
		resp, err := e.caller.Control(act.DeviceID, act.Service, act.Params)
		r.Success, r.Message = types.ResolveCallResult(resp, err)
	default:
		r.Success, r.Message = false, "未知动作类型: "+act.Type
	}
	return
}

// lookupScene 查找场景：优先读内存快照，未命中再回源 DB（限定当前模式），
// 避免场景热更新前缓存缺失导致执行失败。
func (e *SceneEngine) lookupScene(sceneID int64) (types.Scene, bool) {
	e.scenesMu.RLock()
	for _, s := range e.scenes {
		if s.ID == sceneID {
			e.scenesMu.RUnlock()
			return s, true
		}
	}
	e.scenesMu.RUnlock()
	var scene types.Scene
	found, err := db.FindOne("scene", builder.Eq{"id": sceneID, "mode": e.mode}, &scene)
	if err != nil || !found || scene.ID == 0 {
		return types.Scene{}, false
	}
	return scene, true
}

// EnabledScenes 返回当前模式下启用场景的拷贝，调用方可安全遍历，不影响内部切片。
func (e *SceneEngine) EnabledScenes() []types.Scene {
	e.scenesMu.RLock()
	defer e.scenesMu.RUnlock()
	out := make([]types.Scene, len(e.scenes))
	copy(out, e.scenes)
	return out
}

// sceneScheduleLoop 场景定时调度循环：对齐整分钟后每分钟扫描一次到期任务。
func (e *SceneEngine) sceneScheduleLoop() {
	types.WaitUntilNextMinute()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		e.runDueTasks()
	}
}

// runDueTasks 扫描全部定时任务：cron 命中当前分钟则执行其关联场景，
// 执行完毕后回写 task.last_run（无论动作成败都回写，避免同分钟内重复触发）。
func (e *SceneEngine) runDueTasks() {
	// 复制任务切片后立即释放读锁，避免执行场景期间长时间持锁
	e.scenesMu.RLock()
	tasks := make([]types.Task, len(e.tasks))
	copy(tasks, e.tasks)
	e.scenesMu.RUnlock()

	now := time.Now()
	for _, task := range tasks {
		if !CronMatches(task.Cron, now) {
			continue
		}
		scene, ok := e.lookupScene(task.SceneID)
		if !ok {
			logger.Warn("定时任务关联的场景不存在", "task", task.Name, "sceneID", task.SceneID)
			continue
		}
		logger.Info("定时任务触发场景", "task", task.Name, "scene", scene.Name)
		results := e.runScene(scene)

		// 汇总失败动作，仅记录告警，不中断后续任务
		failCount := 0
		for _, r := range results {
			if !r.Success {
				failCount++
			}
		}
		if failCount > 0 {
			logger.Warn("场景部分动作执行失败", "scene", scene.Name, "fail", failCount, "total", len(results))
		}
		if _, err := db.UpdateOne("task", builder.Eq{"id": task.ID}, map[string]interface{}{"last_run": now}); err != nil {
			logger.Warn("回写任务执行时间失败", "error", err, "task", task.Name)
		}
	}
}

// —— cron 匹配器（导出供 rule-engine 包定时规则使用） ——

// CronMatches 判断给定时间是否命中 5 段 cron 表达式（分 时 日 月 周）。
// 日（day-of-month）与周（day-of-week）同时受限时按标准 cron 语义取"或"，
// 否则两者都必须满足；表达式不是 5 段时直接视为不命中。
func CronMatches(expr string, t time.Time) bool {
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 {
		return false
	}
	minuteOK := matchCronField(fields[0], t.Minute(), 0, 59)
	hourOK := matchCronField(fields[1], t.Hour(), 0, 23)
	monthOK := matchCronField(fields[3], int(t.Month()), 1, 12)
	// 仅当对应字段不是 "*" 时视为受限，用于判定日/周的"或"语义
	domRestricted := fields[2] != "*"
	dowRestricted := fields[4] != "*"
	domOK := matchCronField(fields[2], t.Day(), 1, 31)
	dowOK := matchCronField(fields[4], int(t.Weekday()), 0, 6)
	dayOK := domOK && dowOK
	if domRestricted && dowRestricted {
		dayOK = domOK || dowOK
	}
	return minuteOK && hourOK && dayOK && monthOK
}

// matchCronField 匹配逗号分隔的多个 cron 片段（如 "1,15,30"），任一命中即可。
func matchCronField(spec string, value, min, max int) bool {
	for _, part := range strings.Split(spec, ",") {
		if matchCronPart(part, value, min, max) {
			return true
		}
	}
	return false
}

// matchCronPart 匹配单个 cron 片段，支持 "*"、"a"、"a-b"、"*/n"、"a-b/n"：
// 先确定取值区间 [lo,hi]，再判断 value 是否落在区间内且与 lo 的差值能被步长整除。
func matchCronPart(part string, value, min, max int) bool {
	step := 1
	rangeSpec := part
	// 解析 "/step" 后缀：步长必须是正整数，非法直接判不命中
	if idx := strings.Index(part, "/"); idx >= 0 {
		n, err := strconv.Atoi(part[idx+1:])
		if err != nil || n <= 0 {
			return false
		}
		step = n
		rangeSpec = part[:idx]
	}
	lo, hi := min, max
	if rangeSpec != "" && rangeSpec != "*" {
		// 解析区间：a-b 取 [a,b]，单值 a 取 [a,a]，均需落在字段合法范围内
		if idx := strings.Index(rangeSpec, "-"); idx >= 0 {
			l, err1 := strconv.Atoi(rangeSpec[:idx])
			h, err2 := strconv.Atoi(rangeSpec[idx+1:])
			if err1 != nil || err2 != nil || l < min || h > max || l > h {
				return false
			}
			lo, hi = l, h
		} else {
			v, err := strconv.Atoi(rangeSpec)
			if err != nil || v < min || v > max {
				return false
			}
			lo, hi = v, v
		}
	}
	// 先做范围判断，再用取模判断是否落在步长网格上
	if value < lo || value > hi {
		return false
	}
	return (value-lo)%step == 0
}
