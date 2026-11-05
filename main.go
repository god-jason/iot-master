// Package main Master 服务可执行程序入口包。
// 启动流程：
//  1. 检查 master.yaml 是否存在
//     - 不存在 → 安装模式：提供 /api/setup/* 路由与 Vue 安装向导页面（admin/dist）
//     - 存在   → 正常模式：按 boot 框架拓扑排序启动所有组件
//
// 启动模块依赖顺序：config → logger/db/memstore/broker/... → web → api
// 各 pkg 模块自带 boot.Register(Depends: ["config"])，main.go 只注册进程特有的任务
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/god-jason/iot-master/internal/api"
	_ "github.com/god-jason/iot-master/internal/broker"
	_ "github.com/god-jason/iot-master/internal/gateway"
	"github.com/god-jason/iot-master/internal/iot"
	_ "github.com/god-jason/iot-master/internal/maintenance"
	_ "github.com/god-jason/iot-master/internal/mqtt"
	_ "github.com/god-jason/iot-master/internal/ota"
	sceneengine "github.com/god-jason/iot-master/internal/room"
	ruleengine "github.com/god-jason/iot-master/internal/rule-engine"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/service"
	"github.com/god-jason/iot-master/pkg/web"

	"github.com/spf13/pflag"
)

// 全局引用，供 boot 任务访问
var (
	engine *web.Engine
)

func main() {
	// 0. 全局 panic recover
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			fmt.Fprintf(os.Stderr, "[PANIC] %v\n%s\n", r, stack)
			logger.Error("程序异常退出 (panic)", "error", r, "stack", stack)
			os.Exit(2)
		}
	}()

	// 0.1 命令行参数（Windows 服务注册）
	var (
		installFlag   bool
		uninstallFlag bool
	)
	pflag.BoolVar(&installFlag, "install", false, "注册为 Windows 系统服务")
	pflag.BoolVar(&uninstallFlag, "uninstall", false, "从 Windows 系统服务中注销")
	pflag.Parse()

	// 1. master.yaml 不存在 → 安装模式
	if _, err := os.Stat("master.yaml"); os.IsNotExist(err) {
		if installFlag || uninstallFlag {
			fmt.Fprintln(os.Stderr, "错误：--install/--uninstall 需要 master.yaml 存在")
			os.Exit(1)
		}
		runInstallMode()
		return
	}

	// 3. Windows 服务注册/注销分支
	if installFlag {
		runServiceInstall()
		return
	}
	if uninstallFlag {
		runServiceUninstall()
		return
	}

	// 4. 正常模式：按拓扑排序启动所有模块（pkg 模块自带 boot，按 Depends 排序）
	// 注册数据模型，db 模块在 sync=true 时会据此自动同步表结构（补齐缺失的表/字段）
	db.Register(
		new(types.User), new(types.Family), new(types.Member),
		new(types.Room), new(types.Group), new(types.Device),
		new(types.Model), new(types.Rule), new(types.RuleLog),
		new(types.Scene), new(types.Task), new(types.DeviceEvent),
		new(types.DeviceLog), new(types.Scada),
		new(types.GatewayServer), new(types.Link), new(types.Component),
		new(types.Org), new(types.DeviceOrg), new(types.OrgMember),
		new(types.Firmware), new(types.FirmwareRule),
		new(types.OTATask), new(types.OTATaskBatch), new(types.OTATaskDevice),
		new(types.MaintenancePlan), new(types.MaintenanceOrder),
	)
	if err := boot.Startup(); err != nil {
		logger.Error("启动失败", "error", err)
		_ = boot.Shutdown()
		os.Exit(1)
	}

	// 启动规则引擎与场景引擎：master 以 online 模式执行场景与定时任务
	adapter := iot.NewEngineAdapter()
	mode := config.GetString("app", "mode")
	if mode == "" {
		mode = types.SceneModeOnline
	}
	ruleengine.Start(adapter, mode)
	sceneengine.Start(adapter, mode)

	// 5. 创建 Web 引擎并注册 API 路由
	engine = web.New()
	api.InitRouter(engine.Group("/api"))

	// 6. 监听 OS 信号优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("收到退出信号，开始清理", "signal", sig.String())
		_ = boot.Shutdown()
		time.AfterFunc(3*time.Second, func() { os.Exit(0) })
	}()

	// 6. 启动 Web 服务（阻塞）— web 模块已由 pkg/web boot 初始化 Engine
	addr := fmt.Sprintf("%s:%d", config.GetString("web", "host"), config.GetInt("web", "port"))
	logger.Info("Master 服务启动中", "address", addr)
	if err := engine.Run(addr); err != nil {
		logger.Error("Web 服务错误", "error", err)
	}
}

// runServiceInstall 注册为 Windows 系统服务
func runServiceInstall() {
	// pkg/config boot.Startup 还没跑，手动 Load
	config.Name("master")
	config.AddPath(".")
	if err := config.Load(false); err != nil {
		fmt.Fprintf(os.Stderr, "加载 master.yaml 失败: %v\n", err)
		os.Exit(1)
	}

	logger.Init(config.GetString("log", "level"))

	if err := service.Register(&service.Config{
		Name:        "iot-dmp",
		DisplayName: config.GetString("oem", "system_name"),
		Description: config.GetString("oem", "company"),
		Arguments:   []string{},
	}, boot.Startup, boot.Shutdown); err != nil {
		fmt.Fprintf(os.Stderr, "服务注册失败: %v\n", err)
		os.Exit(1)
	}

	if err := service.Install(); err != nil {
		fmt.Fprintf(os.Stderr, "服务安装失败（可能需要管理员权限）: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 服务注册成功。可在「服务」管理器中启动，或执行 master.exe --service")
}

// runServiceUninstall 从 Windows 系统服务中注销
func runServiceUninstall() {
	logger.Init("info")

	if err := service.Register(&service.Config{
		Name:        "iot-dmp",
		DisplayName: "IoT DMP",
	}, nil, nil); err != nil {
		fmt.Fprintf(os.Stderr, "服务注册失败: %v\n", err)
		os.Exit(1)
	}

	if err := service.Uninstall(); err != nil {
		fmt.Fprintf(os.Stderr, "服务卸载失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 服务已注销。")
}

// runInstallMode 安装模式：提供 /api/setup/* 接口，并托管 admin/dist 中的 Vue 安装向导
func runInstallMode() {
	logger.Init("info")
	logger.Info("未检测到 master.yaml，进入安装模式")

	engine = web.New()
	// 安装模式下没有 master.yaml，固定从 admin/dist 提供前端页面；
	// 目录不存在（开发环境）时跳过，改由 Vite 开发服务器访问安装向导
	if err := engine.ServeFrontend("admin/dist"); err != nil {
		logger.Warn("未找到 admin/dist，安装向导页面需通过前端开发服务器访问", "error", err)
	}
	api.RegisterInstall(engine.Group("/api"))

	logger.Info("安装向导监听 http://localhost:8080")

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		logger.Info("收到退出信号", "signal", syscall.SIGTERM.String())
		os.Exit(0)
	}()

	_ = engine.Run(":8080")
}
