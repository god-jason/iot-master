// Package ruleengine 规则引擎核心：作用域匹配 → 触发过滤 → 条件求值（结构化条件或 expr 表达式，
// 支持边沿/防抖/限流）→ 报警状态机 → 动作执行，全程异步，不阻塞设备上报主链路。
package ruleengine

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/god-jason/iot-master/internal/room"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"xorm.io/builder"
)

// 内存缓存容量/保留时长上限，避免长期运行时相关 map 单调增长
const (
	exprCacheMax = 1024  // 表达式编译缓存最大条数
	prevTTL      = 3600  // 设备上一值快照保留时长（秒）
	prevMax      = 10000 // 设备上一值快照最多保留的设备数
)

// Engine 规则引擎实例（只负责规则，场景由 scene.SceneEngine 独立管理）
type Engine struct {
	caller DeviceAdapter   // 设备侧能力：物模型查询 / 设备控制 / 最新数据
	scopes *scopeMatcher   // 作用域匹配器（含分组设备缓存）
	acts   *actionExecutor // 动作执行器

	reloadMu sync.Mutex   // 串行化重载，避免并发重载互相覆盖出旧数据
	rulesMu  sync.RWMutex // 保护 rules 切片，重载与遍历并发安全
	rules    []types.Rule // 当前生效（未禁用）的规则列表

	exprMu    sync.Mutex             // 保护 exprCache，编译与读取序列化
	exprCache map[string]*vm.Program // 表达式原文 → 已编译程序，避免重复编译

	prevMu sync.Mutex               // 保护 prev
	prev   map[string]*prevSnapshot // 设备 ID → 上一帧属性值快照（边沿触发用）

	stateMu sync.Mutex             // 保护 states
	states  map[string]*alarmState // "规则ID:设备ID" → 告警状态机
}

// prevSnapshot 设备上一帧属性值快照，at 记录最近更新时间（秒），用于 TTL 清理。
type prevSnapshot struct {
	values map[string]interface{}
	at     int64
}

// alarmState 单条规则在单个设备上的告警状态机上下文。
// phase 在 ok / pending / firing 之间流转；since 记录进入 pending 的时刻（用于去抖计时），
// lastFire 记录最近一次实际触发时刻（用于限流窗口判断）。
type alarmState struct {
	phase    string // ok / pending / firing
	since    int64  // 进入 pending 的时间（秒）
	lastFire int64  // 最近一次告警时间（秒）
}

// defaultEngine 全局默认引擎实例，由 Start 赋值；各接入层通过 TriggerProperty/TriggerEvent/TriggerStatus 推送事件。
var defaultEngine *Engine

// Start 创建并启动规则引擎（规则定时刷新、定时规则调度）。
// 不再依赖 hub 事件总线 — 设备事件由各接入层（broker/mqtt/gateway）直接调 Trigger 推送。
// 场景/定时任务由 scene.Start(caller, mode) 独立启动。
func Start(caller DeviceAdapter, mode string) *Engine {
	e := &Engine{
		caller:    caller,
		scopes:    newScopeMatcher(caller),
		acts:      newActionExecutor(caller),
		exprCache: make(map[string]*vm.Program),
		prev:      make(map[string]*prevSnapshot),
		states:    make(map[string]*alarmState),
	}
	defaultEngine = e

	e.reloadRules()
	go e.reloadLoop()
	go e.scheduleLoop()

	logger.Info("规则引擎已启动", "mode", mode)
	return e
}

// TriggerProperty 手动推送属性事件（供 broker/mqtt/gateway 接入层调）
func TriggerProperty(deviceID string, values map[string]interface{}) {
	if defaultEngine == nil {
		return
	}
	defaultEngine.processEvent(deviceID, "property", values, nil, "")
}

// TriggerEvent 手动推送设备事件
func TriggerEvent(deviceID, eventName string, params map[string]interface{}) {
	if defaultEngine == nil {
		return
	}
	defaultEngine.processEvent(deviceID, "event", nil, params, eventName)
}

// TriggerStatus 手动推送在线/离线事件
func TriggerStatus(deviceID string, online bool) {
	if defaultEngine == nil {
		return
	}
	status := "offline"
	if online {
		status = "online"
	}
	defaultEngine.processEvent(deviceID, status, nil, map[string]interface{}{"status": boolToInt(online)}, "")
}

// boolToInt 布尔转 0/1，用于统一上下线事件的状态取值。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// processEvent 规则引擎事件入口：把接入层的属性/事件/上下线消息归一成规则触发类型，
// 依次做触发过滤 → 作用域匹配 → 条件求值 → 状态机推进，最后更新属性上一值快照。
func (e *Engine) processEvent(deviceID, evType string, values, params map[string]interface{}, eventName string) {
	if deviceID == "" {
		return
	}
	// 1. 事件归一：evType → types.RuleTrigger* 触发类型，并提取参与条件求值的数据
	var triggerType string
	var data map[string]interface{}
	switch evType {
	case "property":
		triggerType = types.RuleTriggerProperty
		data = values
	case "event":
		triggerType = types.RuleTriggerEvent
		data = params
	case "online", "offline":
		triggerType = types.RuleTriggerStatus
		data = map[string]interface{}{"status": boolToInt(evType == "online")}
	default:
		return
	}

	// 2. 属性事件才有"上一帧"语义，边沿条件（cond.Edge）需要与上一值比较
	var prevData map[string]interface{}
	if evType == "property" {
		prevData = e.getPrev(deviceID)
	}

	e.rulesMu.RLock()
	rules := e.rules
	e.rulesMu.RUnlock()

	for _, rule := range rules {
		// 3. 三道门槛：触发类型/触发字段 → 作用域 → 条件表达式
		if rule.Trigger.Type != triggerType || !e.triggerMatch(rule, evType, values, eventName) {
			continue
		}
		if !e.scopes.match(rule.Scope, deviceID) {
			continue
		}
		matched := e.evaluate(rule, deviceID, data, prevData)
		e.applyState(rule, deviceID, data, matched)
	}

	// 4. 本帧处理结束后才更新上一值，保证本帧所有规则看到的是同一份历史快照
	if evType == "property" && len(values) > 0 {
		e.updatePrev(deviceID, values)
	}
}

// reloadLoop 每 10 秒从数据库重新加载一次规则，实现规则热更新（新增/修改/禁用无需重启）。
func (e *Engine) reloadLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		e.reloadRules()
	}
}

// reloadRules 从 rule 表加载全部未禁用规则并整体替换内存列表。
// 查询失败时保留旧规则，避免数据库抖动导致规则短暂失效；
// 重载成功后按当前规则集合裁剪告警状态，并按 TTL/容量清理上一值快照。
func (e *Engine) reloadRules() {
	e.reloadMu.Lock()
	defer e.reloadMu.Unlock()

	var rules []types.Rule
	if err := db.Find("rule", builder.Eq{"disabled": false}, &rules, 0, 0); err != nil {
		logger.Warn("加载规则失败", "error", err)
		return
	}
	// 按优先级降序排序，同优先级按创建时间升序；评估与动作执行均按此顺序进行
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority
		}
		return rules[i].Created.Before(rules[j].Created)
	})
	e.rulesMu.Lock()
	e.rules = rules
	e.rulesMu.Unlock()

	e.pruneStates(rules)
	e.prunePrev()
}

// ReloadRules 立即从数据库重载规则（供业务层规则 CRUD 后调用）。
// 异步执行避免阻塞调用方；内部通过 reloadMu 串行化，保证并发安全。
func ReloadRules() {
	e := defaultEngine
	if e == nil {
		return
	}
	go e.reloadRules()
}

// pruneStates 按当前规则集合裁剪告警状态：规则删除后遗留的 "规则ID:设备ID" 状态一并清理。
func (e *Engine) pruneStates(rules []types.Rule) {
	alive := make(map[string]struct{}, len(rules))
	for _, r := range rules {
		alive[r.ID] = struct{}{}
	}
	e.stateMu.Lock()
	defer e.stateMu.Unlock()
	for key := range e.states {
		// key 以首个 ":" 分隔规则 ID 与设备 ID，规则 ID 不在当前集合即视为失效
		idx := strings.IndexByte(key, ':')
		ruleID := key
		if idx >= 0 {
			ruleID = key[:idx]
		}
		if _, ok := alive[ruleID]; !ok {
			delete(e.states, key)
		}
	}
}

// prunePrev 清理设备上一值快照：先按 TTL 删除过期条目，仍超容量上限时淘汰最久未更新的设备，
// 避免长期运行时 prev 内存单调增长。
func (e *Engine) prunePrev() {
	now := time.Now().Unix()
	e.prevMu.Lock()
	defer e.prevMu.Unlock()
	for id, p := range e.prev {
		if now-p.at > prevTTL {
			delete(e.prev, id)
		}
	}
	for len(e.prev) > prevMax {
		var oldestID string
		oldestAt := int64(1 << 62)
		for id, p := range e.prev {
			if p.at < oldestAt {
				oldestID, oldestAt = id, p.at
			}
		}
		if oldestID == "" {
			break
		}
		delete(e.prev, oldestID)
	}
}

// triggerMatch 触发过滤：按规则触发类型校验本次事件是否命中其关心的触发点。
// property 校验上报中是否带该属性字段、event 校验事件名、status 校验 online/offline；
// 未配置具体字段/事件名/状态时，视为该类触发全部命中。
func (e *Engine) triggerMatch(rule types.Rule, evType string, values map[string]interface{}, eventName string) bool {
	switch rule.Trigger.Type {
	case types.RuleTriggerProperty:
		if rule.Trigger.Property != "" {
			_, ok := values[rule.Trigger.Property]
			return ok
		}
		return true
	case types.RuleTriggerEvent:
		if rule.Trigger.EventName != "" {
			return eventName == rule.Trigger.EventName
		}
		return true
	case types.RuleTriggerStatus:
		if rule.Trigger.Status != "" {
			return evType == rule.Trigger.Status
		}
		return true
	default:
		return false
	}
}

// evaluate 条件求值：优先走 expr 表达式（编译缓存 + 运行时环境），
// 否则回退到结构化条件组 evalConditions。编译或执行失败一律按"未命中"处理，不向上抛错。
func (e *Engine) evaluate(rule types.Rule, deviceID string, value, prev map[string]interface{}) bool {
	if rule.Expr != "" {
		prog, err := e.compile(rule.Expr)
		if err != nil {
			logger.Warn("规则表达式编译失败", "ruleID", rule.ID, "expr", rule.Expr, "error", err)
			return false
		}
		// 表达式运行环境：value=当前值快照，prev=上一帧快照，device_id/ts 供表达式引用
		env := map[string]interface{}{
			"value":     value,
			"prev":      prev,
			"device_id": deviceID,
			"ts":        time.Now().Unix(),
		}
		out, err := expr.Run(prog, env)
		if err != nil {
			logger.Debug("规则表达式执行失败", "ruleID", rule.ID, "error", err)
			return false
		}
		// 表达式返回值必须是 bool，其它类型（如 nil）一律视为未命中
		b, _ := out.(bool)
		return b
	}
	return evalConditions(rule.Conditions, rule.Logic, value, prev)
}

// compile 编译表达式并缓存编译结果。cache 以表达式原文为 key，
// 避免同一条规则在每条事件上都重复编译；expr.Env 传空 map，允许任意变量名。
func (e *Engine) compile(code string) (*vm.Program, error) {
	e.exprMu.Lock()
	defer e.exprMu.Unlock()
	if prog, ok := e.exprCache[code]; ok {
		return prog, nil
	}
	prog, err := expr.Compile(code, expr.Env(map[string]interface{}{}), expr.AllowUndefinedVariables())
	if err != nil {
		return nil, err
	}
	// 缓存达到上限时整体重建，保证编译缓存容量有界，避免规则增删后无限增长
	if len(e.exprCache) >= exprCacheMax {
		e.exprCache = make(map[string]*vm.Program)
	}
	e.exprCache[code] = prog
	return prog, nil
}

// applyState 告警状态机：按本次条件命中结果推进 ok / pending / firing 三种状态，
// 实现去抖（Debounce：需持续满足 N 秒才告警）、限流（Throttle：告警后 N 秒内不重复触发动作）
// 以及恢复通知（firing → ok 时回调 onRecovered）。状态按 "规则ID:设备ID" 维度独立维护。
func (e *Engine) applyState(rule types.Rule, deviceID string, data map[string]interface{}, matched bool) {
	key := rule.ID + ":" + deviceID
	now := time.Now().Unix()

	e.stateMu.Lock()
	st, ok := e.states[key]
	if !ok {
		// 首次见到该规则+设备组合，初始化为正常态
		st = &alarmState{phase: types.RuleStateOK}
		e.states[key] = st
	}

	if matched {
		switch st.phase {
		case types.RuleStateOK:
			// 首次命中：配置了去抖则先进入 pending 观察期，否则直接告警
			if rule.Debounce > 0 {
				st.phase = types.RuleStatePending
				st.since = now
				e.stateMu.Unlock()
				return
			}
			st.phase = types.RuleStateFiring
			st.lastFire = now
		case types.RuleStatePending:
			// 去抖观察期：持续满足达到 Debounce 秒才真正告警，未达标则继续等待
			if now-st.since < rule.Debounce {
				e.stateMu.Unlock()
				return
			}
			st.phase = types.RuleStateFiring
			st.lastFire = now
		case types.RuleStateFiring:
			// 已处于告警中：受 Throttle 限流，窗口内不重复执行动作；Throttle<=0 表示不限流
			if rule.Throttle <= 0 || now-st.lastFire < rule.Throttle {
				e.stateMu.Unlock()
				return
			}
			st.lastFire = now
		}
	} else {
		// 条件不再满足：回到 ok；若此前处于告警中，则补发一次恢复通知
		wasFiring := st.phase == types.RuleStateFiring
		st.phase = types.RuleStateOK
		e.stateMu.Unlock()

		if wasFiring {
			e.onRecovered(rule, deviceID, data)
		}
		return
	}
	e.stateMu.Unlock()

	e.onFire(rule, deviceID, data)
}

// onFire 告警触发：执行规则配置的全部动作，并写入一条 matched=true 的执行日志。
func (e *Engine) onFire(rule types.Rule, deviceID string, data map[string]interface{}) {
	logger.Info("规则命中", "rule", rule.Name, "deviceID", deviceID)
	results := e.acts.execute(rule, deviceID, data)
	e.writeLog(rule, deviceID, data, true, types.RuleStateFiring, results)
}

// onRecovered 告警恢复：只发送一条恢复通知（复用 notify 动作），并写入 resolved 状态的执行日志。
func (e *Engine) onRecovered(rule types.Rule, deviceID string, data map[string]interface{}) {
	logger.Info("告警恢复", "rule", rule.Name, "deviceID", deviceID)
	// 恢复通知只保留 notify 动作，避免恢复时重复执行控制类动作
	recoverRule := rule
	recoverRule.Actions = []types.RuleAction{{
		Type:   types.RuleActionNotify,
		Params: map[string]interface{}{"message": rule.Name + " 告警已恢复"},
	}}
	e.acts.execute(recoverRule, deviceID, data)
	e.writeLog(rule, deviceID, data, false, types.RuleStateResolved, nil)
}

// writeLog 写入 rule_log 执行日志；写库失败只记告警，不影响规则主流程。
func (e *Engine) writeLog(rule types.Rule, deviceID string, data map[string]interface{}, matched bool, state string, results []types.RuleActionResult) {
	logEntry := types.RuleLog{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		DeviceID: deviceID,
		Trigger:  rule.Trigger.Type,
		Values:   data,
		Matched:  matched,
		State:    state,
		Results:  results,
		Created:  time.Now(),
	}
	if _, err := db.InsertOne("rule_log", logEntry); err != nil {
		logger.Warn("写入规则执行日志失败", "error", err, "ruleID", rule.ID)
	}
}

// getPrev 返回设备上一帧属性快照的副本（无历史时返回空 map），供边沿条件比较。
func (e *Engine) getPrev(deviceID string) map[string]interface{} {
	e.prevMu.Lock()
	defer e.prevMu.Unlock()
	if p, ok := e.prev[deviceID]; ok {
		cp := make(map[string]interface{}, len(p.values))
		for k, v := range p.values {
			cp[k] = v
		}
		return cp
	}
	return map[string]interface{}{}
}

// updatePrev 用本帧属性值增量更新设备上一值快照（本帧未上报的字段保留旧值），并刷新更新时间。
func (e *Engine) updatePrev(deviceID string, values map[string]interface{}) {
	e.prevMu.Lock()
	defer e.prevMu.Unlock()
	p, ok := e.prev[deviceID]
	if !ok {
		p = &prevSnapshot{values: make(map[string]interface{})}
		e.prev[deviceID] = p
	}
	for k, v := range values {
		p.values[k] = v
	}
	p.at = time.Now().Unix()
}

// scheduleLoop 每分钟整点检查定时规则并对范围内设备做一次评估
func (e *Engine) scheduleLoop() {
	types.WaitUntilNextMinute()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for t := range ticker.C {
		e.runScheduled(t)
	}
}

// runScheduled 执行当前分钟命中的定时规则：cron 匹配后枚举作用域内的全部设备，
// 用各设备最新时序值做一次条件求值并推进状态机（与事件触发共用同一套状态机和动作）。
func (e *Engine) runScheduled(t time.Time) {
	e.rulesMu.RLock()
	rules := e.rules
	e.rulesMu.RUnlock()

	for _, rule := range rules {
		if rule.Trigger.Type != types.RuleTriggerSchedule || !room.CronMatches(rule.Trigger.Cron, t) {
			continue
		}
		for _, deviceID := range e.scopes.enumerate(rule.Scope) {
			values := e.latestValues(deviceID)
			// 定时规则没有"上一帧"语义，prev 传 nil
			matched := e.evaluate(rule, deviceID, values, nil)
			e.applyState(rule, deviceID, values, matched)
		}
	}
}

// latestValues 取设备各属性最新值（来自时序库 LatestData），供定时规则求值。
func (e *Engine) latestValues(deviceID string) map[string]interface{} {
	points, err := e.caller.LatestData(deviceID)
	if err != nil || len(points) == 0 {
		return map[string]interface{}{}
	}
	values := make(map[string]interface{}, len(points))
	for field, p := range points {
		values[field] = p.Value
	}
	return values
}
