// 安装向导接口：master.yaml 不存在时，主程序以安装模式启动，
// 仅提供本文件中的 /api/setup/* 接口与前端安装页面（admin/dist）。
package api

import (
	crand "crypto/rand"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
	"xorm.io/xorm"
)

// setupDBConfig 关系数据库配置（安装表单）
type setupDBConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"sslmode"`
	Params   string `json:"params"`
}

// setupRedisConfig Redis 配置（安装表单）
type setupRedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// setupInfluxConfig InfluxDB 配置（安装表单）
type setupInfluxConfig struct {
	URL    string `json:"url"`
	Token  string `json:"token"`
	Org    string `json:"org"`
	Bucket string `json:"bucket"`
}

// setupWebConfig HTTP 服务配置（安装表单）
type setupWebConfig struct {
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	HTTPS       bool     `json:"https"`
	LetsEncrypt bool     `json:"lets_encrypt"`
	Debug       bool     `json:"debug"`
	CertFile    string   `json:"cert_file"`
	KeyFile     string   `json:"key_file"`
	Domains     []string `json:"domains"`
}

// setupBrokerConfig MQTT Broker 配置（安装表单）：mode 为 builtin 或 external
type setupBrokerConfig struct {
	Mode      string `json:"mode"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Validator bool   `json:"validator"`
}

// setupAdminConfig 首个管理员账号（安装表单）
type setupAdminConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// setupOemConfig 品牌定制信息（安装表单）
type setupOemConfig struct {
	Logo       string `json:"logo"`
	SystemName string `json:"system_name"`
	Company    string `json:"company"`
	ICP        string `json:"icp"`
}

// SetupRequest 安装表单完整请求体
type SetupRequest struct {
	DB        setupDBConfig     `json:"db"`
	Redis     setupRedisConfig  `json:"redis"`
	Influx    setupInfluxConfig `json:"influx"`
	Web       setupWebConfig    `json:"web"`
	Broker    setupBrokerConfig `json:"broker"`
	Admin     setupAdminConfig  `json:"admin"`
	UploadDir string            `json:"upload_dir"`
	Oem       setupOemConfig    `json:"oem"`
}

// RegisterInstall 注册安装模式路由（无需认证，仅安装模式下调用）
func RegisterInstall(r *gin.RouterGroup) {
	r.GET("/setup/status", setupStatus)           // 查询是否已完成安装
	r.POST("/setup/test-db", setupTestDB)         // 测试关系数据库连接
	r.POST("/setup/test-redis", setupTestRedis)   // 测试 Redis 连通性
	r.POST("/setup/test-influx", setupTestInflux) // 测试 InfluxDB 连接
	r.POST("/setup/save", setupSave)              // 写入配置、建表并创建管理员
}

// setupTestDB 测试关系数据库连接：构造 DSN 后执行 SELECT 1
func setupTestDB(c *gin.Context) {
	var req setupDBConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"ok": false, "message": "参数错误: " + err.Error()})
		return
	}

	dsn, err := db.DSN(req.Driver, req.Host, req.Port, req.Database, req.Username, req.Password, req.SSLMode, req.Params)
	if err != nil {
		c.JSON(400, gin.H{"ok": false, "message": err.Error()})
		return
	}

	engine, err := xorm.NewEngine(req.Driver, dsn)
	if err != nil {
		c.JSON(200, gin.H{"ok": false, "message": "驱动加载失败: " + err.Error()})
		return
	}
	defer engine.Close()

	// 连接测试不保留连接，设置较短的存活时间
	engine.SetConnMaxLifetime(3 * time.Second)
	if _, err = engine.Exec("SELECT 1"); err != nil {
		c.JSON(200, gin.H{"ok": false, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "message": "连接成功"})
}

// setupTestRedis 测试 Redis 连通性：仅做 TCP 探测，不发送 PING 命令
func setupTestRedis(c *gin.Context) {
	var req setupRedisConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"ok": false, "message": "参数错误: " + err.Error()})
		return
	}
	addr := net.JoinHostPort(req.Host, fmt.Sprintf("%d", req.Port))
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		c.JSON(200, gin.H{"ok": false, "message": "无法连接: " + err.Error()})
		return
	}
	_ = conn.Close()
	c.JSON(200, gin.H{"ok": true, "message": "连接成功"})
}

// setupTestInflux 测试 InfluxDB 连接：请求 /health 端点
func setupTestInflux(c *gin.Context) {
	var req setupInfluxConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"ok": false, "message": "参数错误: " + err.Error()})
		return
	}
	if req.URL == "" {
		c.JSON(200, gin.H{"ok": false, "message": "请输入 InfluxDB URL"})
		return
	}
	resp, err := http.Get(req.URL + "/health")
	if err != nil {
		c.JSON(200, gin.H{"ok": false, "message": "无法连接: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		c.JSON(200, gin.H{"ok": true, "message": "连接成功"})
		return
	}
	c.JSON(200, gin.H{"ok": false, "message": fmt.Sprintf("HTTP %d", resp.StatusCode)})
}

// setupSave 保存安装配置：初始化 DB → 建表 → 创建管理员 → 原子写入 master.yaml
// 先验证数据库可用再落盘配置，避免连接失败时留下不可用的配置文件
func setupSave(c *gin.Context) {
	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"ok": false, "message": "参数错误: " + err.Error()})
		return
	}

	// 1. 初始化 DB + 建表 + 创建 admin
	dsn, err := db.DSN(req.DB.Driver, req.DB.Host, req.DB.Port, req.DB.Database, req.DB.Username, req.DB.Password, req.DB.SSLMode, req.DB.Params)
	if err != nil {
		c.JSON(400, gin.H{"ok": false, "message": err.Error()})
		return
	}
	if err := db.InitEngine(req.DB.Driver, dsn); err != nil {
		c.JSON(200, gin.H{"ok": false, "message": "数据库连接失败: " + err.Error()})
		return
	}
	defer db.Close()

	// 显式探测连接：xorm 为懒连接，仅 NewEngine 成功不代表数据库真实可用
	if err := db.Engine.Ping(); err != nil {
		c.JSON(200, gin.H{"ok": false, "message": "数据库连接失败: " + err.Error()})
		return
	}

	var warnings []string
	// 表同步失败不中断安装（如目标库已存在同名表），仅作为警告返回
	if err := db.SyncTables(
		new(types.User), new(types.Family), new(types.Member),
		new(types.Room), new(types.Group), new(types.Device),
		new(types.Model), new(types.Rule), new(types.RuleLog),
		new(types.Scene), new(types.Task), new(types.DeviceEvent),
		new(types.DeviceLog), new(types.Scada),
		new(types.GatewayServer), new(types.Link), new(types.Component),
	); err != nil {
		warnings = append(warnings, "表同步警告: "+err.Error())
	}

	// admin 创建失败同样只记录警告，不阻断安装流程
	if err := createAdminUser(req.Admin.Username, req.Admin.Password, req.Admin.Nickname); err != nil {
		warnings = append(warnings, "admin 用户创建警告: "+err.Error())
	}

	// 2. 数据库就绪后原子写入 master.yaml（临时文件 + rename），避免写入中断产生半截配置
	yamlCfg := buildYAML(req)
	data, err := yaml.Marshal(yamlCfg)
	if err != nil {
		c.JSON(500, gin.H{"ok": false, "message": "配置序列化失败: " + err.Error()})
		return
	}
	if err := writeFileAtomic("master.yaml", data, 0600); err != nil {
		c.JSON(500, gin.H{"ok": false, "message": "写入 master.yaml 失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"ok":       true,
		"message":  "安装完成！配置已保存，请重启服务以正常模式运行。",
		"warnings": warnings,
	})
}

// writeFileAtomic 原子写入文件：先写同目录临时文件，再重命名覆盖目标文件。
// 避免进程中断时留下不完整的 master.yaml。
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // 正常情况下已重命名，异常时清理残留

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// setupStatus 查询 master.yaml 是否已存在，用于前端判断是否进入安装向导
func setupStatus(c *gin.Context) {
	_, err := os.Stat("master.yaml")
	c.JSON(200, gin.H{"installed": !os.IsNotExist(err)})
}

// buildYAML 把安装表单转换为写入 master.yaml 的配置结构。
// 拼装各段配置，并对 upload 目录、OEM 信息等字段做默认值兜底，JWT 密钥随机生成。
func buildYAML(req SetupRequest) map[string]interface{} {
	brokerCfg := map[string]interface{}{"validator": req.Broker.Validator}
	// 仅内置 broker 需要记录监听地址；外部 broker 不写 host/port
	if req.Broker.Mode == "builtin" {
		brokerCfg["host"] = req.Broker.Host
		brokerCfg["port"] = req.Broker.Port
	} else {
		brokerCfg["host"] = ""
		brokerCfg["port"] = 0
	}

	dbCfg := map[string]interface{}{
		"driver": req.DB.Driver, "host": req.DB.Host, "port": req.DB.Port,
		"database": req.DB.Database, "username": req.DB.Username, "password": req.DB.Password,
	}
	// sslmode / params 仅在用户填写时写入，避免影响不支持这些参数的驱动
	if req.DB.SSLMode != "" {
		dbCfg["sslmode"] = req.DB.SSLMode
	}
	if req.DB.Params != "" {
		dbCfg["params"] = req.DB.Params
	}

	return map[string]interface{}{
		"db": dbCfg,
		"redis": map[string]interface{}{
			"host": req.Redis.Host, "port": req.Redis.Port,
			"password": req.Redis.Password, "db": req.Redis.DB,
		},
		"influxdb": map[string]interface{}{
			"url": req.Influx.URL, "token": req.Influx.Token,
			"org": req.Influx.Org, "bucket": req.Influx.Bucket,
		},
		"broker": brokerCfg,
		"web": map[string]interface{}{
			"host": req.Web.Host, "port": req.Web.Port,
			"https": req.Web.HTTPS, "cert_file": req.Web.CertFile,
			"key_file": req.Web.KeyFile, "lets_encrypt": req.Web.LetsEncrypt,
			"domains": req.Web.Domains, "debug": req.Web.Debug,
		},
		"jwt":    map[string]interface{}{"api": randomString(32)},
		"upload": map[string]interface{}{"dir": pickStr(req.UploadDir, "./uploads")},
		"log":    map[string]interface{}{"level": "debug"},
		"oem": map[string]interface{}{
			"logo":        pickStr(req.Oem.Logo, "/logo.jpg"),
			"system_name": pickStr(req.Oem.SystemName, "小白白"),
			"company":     pickStr(req.Oem.Company, "本易物联"),
			"icp":         req.Oem.ICP,
		},
	}
}

// createAdminUser 创建首个平台管理员：密码 bcrypt 加密后直接插入 user 表。
// 用户名为空时兜底为 admin，密码为空则直接报错。
func createAdminUser(username, password, nickname string) error {
	// 密码必须有值，否则会创建出无法登录的管理员
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	// 用户名兜底
	if username == "" {
		username = "admin"
	}
	_, err = db.Engine.Exec(
		"INSERT INTO user (id, username, password, nickname, admin, created) VALUES (?, ?, ?, ?, ?, ?)",
		db.NewID(), username, string(hashed), nickname, true, time.Now(),
	)
	return err
}

// pickStr 字段兜底：s 为空时返回默认值 def。
func pickStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// randomString 生成长度为 n 的随机字符串，用于安装时生成 JWT 密钥与初始密码。
// 使用 crypto/rand 生成安全随机数（字符集映射仅带来可忽略的取模偏差）。
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if n <= 0 {
		return ""
	}
	raw := make([]byte, n)
	if _, err := crand.Read(raw); err != nil {
		// crypto/rand 不可用时退化为 uuid（其内部同样使用 crypto/rand），两种 uuid 共 64 字符足够覆盖 n<=64
		raw = []byte(strings.ReplaceAll(uuid.NewString()+uuid.NewString(), "-", ""))
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = letters[int(raw[i])%len(letters)]
	}
	return string(out)
}
