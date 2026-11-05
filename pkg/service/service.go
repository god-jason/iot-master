// Package service 提供 Windows / Linux 跨平台服务注册能力，
// 封装 github.com/kardianos/service，支持 --install / --uninstall 等命令行操作。
package service

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
)

var (
	svc    service.Service
	logger service.Logger
)

// Config 服务注册配置项
type Config struct {
	Name        string   // 服务注册名（系统内部用）
	DisplayName string   // 服务显示名（服务管理器显示）
	Description string   // 服务描述
	Arguments   []string // 启动参数（Windows 服务启动时附加到可执行文件后的参数）
}

// Program 包装启动/关闭函数，实现 service.Program 接口
type Program struct {
	Startup  func() error
	Shutdown func() error
}

// Start 实现 service.Program.Start
func (p *Program) Start(s service.Service) error {
	return p.Startup()
}

// Stop 实现 service.Program.Stop
func (p *Program) Stop(s service.Service) error {
	return p.Shutdown()
}

// Register 注册一个服务实例，必须在 Run / Install / Uninstall 之前调用
func Register(cfg *Config, startup, shutdown func() error) error {
	sc := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   cfg.Arguments,
	}

	p := &Program{
		Startup:  startup,
		Shutdown: shutdown,
	}

	var err error
	svc, err = service.New(p, sc)
	if err != nil {
		return fmt.Errorf("service.New: %w", err)
	}

	logger, err = svc.Logger(nil)
	if err != nil {
		// logger 获取失败不致命，降级为 nil（下面的 Info/Error 会自动 fallback 到 stdout）
		logger = nil
	}

	return nil
}

// Run 以服务模式启动（阻塞直到服务停止）
func Run() error {
	if svc == nil {
		return fmt.Errorf("service not registered, call Register first")
	}
	return svc.Run()
}

// Install 注册为系统服务
func Install() error {
	if svc == nil {
		return fmt.Errorf("service not registered")
	}
	return svc.Install()
}

// Uninstall 从系统中注销服务
func Uninstall() error {
	if svc == nil {
		return fmt.Errorf("service not registered")
	}
	return svc.Uninstall()
}

// Start 启动已注册的服务（通常由 sc.exe 调用）
func Start() error {
	if svc == nil {
		return fmt.Errorf("service not registered")
	}
	return svc.Start()
}

// Stop 停止正在运行的服务
func Stop() error {
	if svc == nil {
		return fmt.Errorf("service not registered")
	}
	return svc.Stop()
}

// Restart 重启服务
func Restart() error {
	if svc == nil {
		return fmt.Errorf("service not registered")
	}
	return svc.Restart()
}

// Info 写服务日志（优先用系统 logger，fallback 到 stdout）
func Info(v ...interface{}) {
	if logger != nil {
		_ = logger.Info(v...)
	} else {
		fmt.Fprintln(os.Stdout, v...)
	}
}

// Warn 写警告日志
func Warn(v ...interface{}) {
	if logger != nil {
		_ = logger.Warning(v...)
	} else {
		fmt.Fprintln(os.Stderr, v...)
	}
}

// Error 写错误日志
func Error(v ...interface{}) {
	if logger != nil {
		_ = logger.Error(v...)
	} else {
		fmt.Fprintln(os.Stderr, v...)
	}
}
