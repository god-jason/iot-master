// 规则引擎管理接口：规则的增删改查、启停、模拟试跑，以及规则执行日志查询。
package api

import (
	"time"

	"github.com/god-jason/iot-master/internal/rule-engine"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// RuleHandler 规则管理处理器
type RuleHandler struct {
	curd.Handler[types.Rule]
}

// NewRuleHandler 创建规则处理器
func NewRuleHandler() *RuleHandler {
	return &RuleHandler{
		Handler: curd.Handler[types.Rule]{
			Collection: "rule",
			Keywords:   []string{"name"},
		},
	}
}

// Create 创建规则（仅系统管理员）
func (h *RuleHandler) Create(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var rule types.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	if rule.ID == "" {
		rule.ID = db.NewID()
	}
	rule.Created = time.Now()
	rule.Updated = rule.Created

	if _, err := db.InsertOne("rule", rule); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	ruleengine.ReloadRules()
	curd.Success(c, rule)
}

// Update 更新规则（仅系统管理员）
func (h *RuleHandler) Update(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id := c.Param("id")

	var rule types.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	update := map[string]interface{}{
		"name":       rule.Name,
		"desc":       rule.Desc,
		"disabled":   rule.Disabled,
		"priority":   rule.Priority,
		"scope":      rule.Scope,
		"trigger":    rule.Trigger,
		"expr":       rule.Expr,
		"conditions": rule.Conditions,
		"logic":      rule.Logic,
		"actions":    rule.Actions,
		"debounce":   rule.Debounce,
		"throttle":   rule.Throttle,
		"updated":    time.Now(),
	}
	if _, err := db.UpdateOne("rule", builder.Eq{"id": id}, update); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	ruleengine.ReloadRules()
	curd.Success(c, gin.H{"id": id})
}

// Delete 删除规则（仅系统管理员），删除后热重载规则引擎
func (h *RuleHandler) Delete(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id := c.Param("id")
	affected, err := db.DeleteOne("rule", builder.Eq{"id": id})
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if affected == 0 {
		curd.Error(c, 404, "规则不存在")
		return
	}
	ruleengine.ReloadRules()
	curd.Success(c, nil)
}

// Switch 启用/停用规则（仅系统管理员）
func (h *RuleHandler) Switch(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id := c.Param("id")
	var req struct {
		Disabled bool `json:"disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if _, err := db.UpdateOne("rule", builder.Eq{"id": id}, map[string]interface{}{"disabled": req.Disabled, "updated": time.Now()}); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	ruleengine.ReloadRules()
	curd.Success(c, gin.H{"id": id, "disabled": req.Disabled})
}

// Test 用模拟数据试跑规则条件，返回是否命中（不执行任何动作，仅系统管理员）
func (h *RuleHandler) Test(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id := c.Param("id")
	var req struct {
		Values map[string]interface{} `json:"values"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	var rule types.Rule
	found, err := db.FindOne("rule", builder.Eq{"id": id}, &rule)
	if err != nil || !found || rule.ID == "" {
		curd.Error(c, 404, "规则不存在")
		return
	}

	matched, err := ruleengine.EvaluateRule(rule, req.Values)
	if err != nil {
		curd.Error(c, 400, "条件求值失败: "+err.Error())
		return
	}
	curd.Success(c, gin.H{"matched": matched})
}

// NewRuleLogHandler 创建规则执行日志处理器（只读查询）
func NewRuleLogHandler() *curd.Handler[types.RuleLog] {
	return &curd.Handler[types.RuleLog]{
		Collection: "rule_log",
		IsInt64ID:  true,
		Keywords:   []string{"rule_name", "device_id"},
	}
}
