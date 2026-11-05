package maintenance

import (
	"time"

	"github.com/god-jason/iot-master/pkg/logger"
)

// scheduler 维保到期扫描调度器
type scheduler struct {
	stop chan struct{}
}

var current *scheduler

// Start boot 启动入口：按配置间隔扫描到期计划
func Start() error {
	if !Enabled() {
		logger.Info("维保模块已禁用")
		return nil
	}
	s := &scheduler{stop: make(chan struct{})}
	current = s
	go s.loop()
	logger.Info("维保计划调度器已启动", "scan_interval", ScanInterval().String())
	return nil
}

// Stop boot 关闭入口
func Stop() error {
	if current != nil {
		close(current.stop)
		current = nil
	}
	return nil
}

func (s *scheduler) loop() {
	// 启动时立即扫描一次
	s.scan()
	ticker := time.NewTicker(ScanInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.scan()
		case <-s.stop:
			return
		}
	}
}

func (s *scheduler) scan() {
	generated, err := GenerateDuePlans(time.Now())
	if err != nil {
		logger.Warn("维保计划扫描失败", "error", err)
		return
	}
	if generated > 0 {
		logger.Info("维保计划扫描完成", "generated", generated)
	}
}
