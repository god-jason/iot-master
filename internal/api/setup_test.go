// Package api 安装向导接口测试
// 覆盖 /api/setup/* 的路由注册、参数校验、连接失败处理以及配置落盘行为。
// 所有用例均在临时目录中运行，不会写入仓库根目录的 master.yaml。
package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupResponse 安装接口统一响应结构
type setupResponse struct {
	OK        bool     `json:"ok"`
	Message   string   `json:"message"`
	Warnings  []string `json:"warnings"`
	Installed bool     `json:"installed"`
}

// newSetupTestServer 创建仅注册安装向导路由的测试服务器
func newSetupTestServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	RegisterInstall(e.Group("/api"))
	return e
}

// doSetupRequest 发送 JSON 请求并解析响应
func doSetupRequest(t *testing.T, e *gin.Engine, method, path string, body interface{}) (int, setupResponse) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("编码请求体失败: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	var resp setupResponse
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
		}
	}
	return w.Code, resp
}

// TestSetupStatus 验证安装状态随 master.yaml 是否存在而变化
func TestSetupStatus(t *testing.T) {
	t.Chdir(t.TempDir())
	e := newSetupTestServer()

	code, resp := doSetupRequest(t, e, http.MethodGet, "/api/setup/status", nil)
	if code != http.StatusOK {
		t.Fatalf("期望 HTTP 200，实际 %d", code)
	}
	if resp.Installed {
		t.Fatalf("空目录应返回未安装")
	}

	// 写入 master.yaml 后应识别为已安装
	if err := os.WriteFile("master.yaml", []byte("db:\n  driver: mysql\n"), 0600); err != nil {
		t.Fatalf("写入 master.yaml 失败: %v", err)
	}
	_, resp = doSetupRequest(t, e, http.MethodGet, "/api/setup/status", nil)
	if !resp.Installed {
		t.Fatalf("存在 master.yaml 时应返回已安装")
	}
}

// TestSetupTestDB 验证数据库连接测试：不支持的驱动与不可达地址
func TestSetupTestDB(t *testing.T) {
	e := newSetupTestServer()

	// 不支持的驱动：参数错误，HTTP 400
	code, resp := doSetupRequest(t, e, http.MethodPost, "/api/setup/test-db", setupDBConfig{
		Driver: "sqlite", Host: "127.0.0.1", Port: 1, Database: "dmp",
	})
	if code != http.StatusBadRequest || resp.OK {
		t.Fatalf("不支持的驱动应返回 400 且 ok=false，实际 code=%d ok=%v msg=%s", code, resp.OK, resp.Message)
	}

	// 不可达的 MySQL：连接失败，HTTP 200 + ok=false
	_, resp = doSetupRequest(t, e, http.MethodPost, "/api/setup/test-db", setupDBConfig{
		Driver: "mysql", Host: "127.0.0.1", Port: 1, Database: "dmp", Username: "root",
	})
	if resp.OK {
		t.Fatalf("不可达数据库应返回 ok=false")
	}
	if !strings.Contains(resp.Message, "无法") && !strings.Contains(resp.Message, "连接") {
		t.Logf("数据库失败提示: %s", resp.Message)
	}
}

// TestSetupTestRedis 验证 Redis 连通性测试的失败分支
func TestSetupTestRedis(t *testing.T) {
	e := newSetupTestServer()
	_, resp := doSetupRequest(t, e, http.MethodPost, "/api/setup/test-redis", setupRedisConfig{
		Host: "127.0.0.1", Port: 1,
	})
	if resp.OK {
		t.Fatalf("不可达 Redis 应返回 ok=false")
	}
}

// TestSetupTestInflux 验证 InfluxDB 连接测试：空 URL 与不可达地址
func TestSetupTestInflux(t *testing.T) {
	e := newSetupTestServer()

	_, resp := doSetupRequest(t, e, http.MethodPost, "/api/setup/test-influx", setupInfluxConfig{URL: ""})
	if resp.OK || resp.Message != "请输入 InfluxDB URL" {
		t.Fatalf("空 URL 应提示填写地址，实际 ok=%v msg=%s", resp.OK, resp.Message)
	}

	_, resp = doSetupRequest(t, e, http.MethodPost, "/api/setup/test-influx", setupInfluxConfig{URL: "http://127.0.0.1:1"})
	if resp.OK {
		t.Fatalf("不可达 InfluxDB 应返回 ok=false")
	}
}

// TestSetupSaveFailureKeepsConfigAbsent 验证数据库不可用时不会写出 master.yaml
func TestSetupSaveFailureKeepsConfigAbsent(t *testing.T) {
	t.Chdir(t.TempDir())
	e := newSetupTestServer()

	code, resp := doSetupRequest(t, e, http.MethodPost, "/api/setup/save", SetupRequest{
		DB: setupDBConfig{
			Driver: "mysql", Host: "127.0.0.1", Port: 1,
			Database: "dmp", Username: "root", Password: "x",
		},
		Admin: setupAdminConfig{Username: "admin", Password: "test123456"},
	})
	if code != http.StatusOK || resp.OK {
		t.Fatalf("数据库不可用时应返回 ok=false，实际 code=%d ok=%v msg=%s", code, resp.OK, resp.Message)
	}
	if _, err := os.Stat("master.yaml"); !os.IsNotExist(err) {
		t.Fatalf("数据库连接失败时不应写出 master.yaml")
	}
}

// TestSetupSaveInvalidDriver 验证不支持的驱动不会落盘配置
func TestSetupSaveInvalidDriver(t *testing.T) {
	t.Chdir(t.TempDir())
	e := newSetupTestServer()

	code, _ := doSetupRequest(t, e, http.MethodPost, "/api/setup/save", SetupRequest{
		DB: setupDBConfig{Driver: "sqlite", Host: "127.0.0.1", Port: 1, Database: "dmp"},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("不支持的驱动应返回 400，实际 %d", code)
	}
	if _, err := os.Stat("master.yaml"); !os.IsNotExist(err) {
		t.Fatalf("参数错误时不应写出 master.yaml")
	}
}

// TestBuildYAML 验证安装表单到配置结构的映射与默认值兜底
func TestBuildYAML(t *testing.T) {
	cfg := buildYAML(SetupRequest{
		DB:     setupDBConfig{Driver: "mysql", Host: "127.0.0.1", Port: 3306, Database: "dmp", Username: "root"},
		Redis:  setupRedisConfig{Host: "127.0.0.1", Port: 6379},
		Web:    setupWebConfig{Host: "0.0.0.0", Port: 8080},
		Broker: setupBrokerConfig{Mode: "builtin", Host: "0.0.0.0", Port: 1883},
	})

	if cfg["db"].(map[string]interface{})["driver"] != "mysql" {
		t.Fatalf("db.driver 映射错误")
	}
	broker := cfg["broker"].(map[string]interface{})
	if broker["host"] != "0.0.0.0" || broker["port"] != 1883 {
		t.Fatalf("内置 broker 地址映射错误: %v", broker)
	}
	if cfg["upload"].(map[string]interface{})["dir"] != "./uploads" {
		t.Fatalf("upload.dir 应使用默认值 ./uploads")
	}
	if cfg["oem"].(map[string]interface{})["system_name"] != "小白白" {
		t.Fatalf("oem.system_name 应使用默认值")
	}
	if len(cfg["jwt"].(map[string]interface{})["api"].(string)) != 32 {
		t.Fatalf("jwt.api 应为 32 位随机串")
	}

	// 外部 broker 不写监听地址
	external := buildYAML(SetupRequest{Broker: setupBrokerConfig{Mode: "external", Host: "mqtt.example.com", Port: 1883}})
	extBroker := external["broker"].(map[string]interface{})
	if extBroker["host"] != "" || extBroker["port"] != 0 {
		t.Fatalf("外部 broker 不应写入监听地址: %v", extBroker)
	}
}

// TestRandomString 验证随机串长度与字符集
func TestRandomString(t *testing.T) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, n := range []int{0, 1, 16, 32, 64} {
		s := randomString(n)
		if len(s) != n {
			t.Fatalf("randomString(%d) 长度应为 %d，实际 %d", n, n, len(s))
		}
		for _, r := range s {
			if !strings.ContainsRune(letters, r) {
				t.Fatalf("randomString(%d) 含非法字符 %q", n, r)
			}
		}
	}
	if randomString(32) == randomString(32) {
		t.Fatalf("两次随机串不应相同")
	}
}

// TestWriteFileAtomic 验证原子写入成功后文件内容完整
func TestWriteFileAtomic(t *testing.T) {
	t.Chdir(t.TempDir())
	want := []byte("db:\n  driver: mysql\n")
	if err := writeFileAtomic("master.yaml", want, 0600); err != nil {
		t.Fatalf("原子写入失败: %v", err)
	}
	got, err := os.ReadFile("master.yaml")
	if err != nil {
		t.Fatalf("读取写入文件失败: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("文件内容不一致")
	}
	// 临时文件应已清理
	matches, _ := filepath.Glob(".master.yaml.tmp-*")
	if len(matches) != 0 {
		t.Fatalf("临时文件未清理: %v", matches)
	}
}
