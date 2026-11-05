package pool

import (
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/logger"

	ants "github.com/panjf2000/ants/v2"
)

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}

// Pool 全局协程池实例
var Pool *ants.Pool

// Startup 启动协程池。size 从 viper 读，默认 10000。
func Startup() error {
	var err error
	Pool, err = ants.NewPool(FromViper(), ants.WithPanicHandler(func(err any) {
		logger.Error("协程池任务 panic", "error", err)
	}))
	return err
}

// Shutdown 关闭协程池
func Shutdown() error {
	if Pool != nil {
		Pool.Release()
		Pool = nil
	}
	return nil
}

// Submit 投递任务；Pool 未启动时回退到 go 关键字
func Submit(task func()) {
	if Pool == nil {
		go task()
		return
	}
	if err := Pool.Submit(task); err != nil {
		logger.Error("协程池投递失败，回退到 goroutine", "error", err)
		go task()
	}
}

// Stats 返回协程池当前统计
func Stats() *Info {
	if Pool == nil {
		return nil
	}
	return &Info{
		Cap:     Pool.Cap(),
		Free:    Pool.Free(),
		Running: Pool.Running(),
		Waiting: Pool.Waiting(),
	}
}
