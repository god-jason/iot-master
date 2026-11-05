// Package boot 模块化启动框架
// 每个 pkg 的 boot.go 通过 init() 注册自己，声明依赖。boot.Startup() 按依赖拓扑排序启动。
// 关闭时逆序调用 Shutdown()。
// 用法（在每个 pkg/boot.go）:
//
//	func init() {
//	    boot.Register("db", &boot.Task{
//	        Startup:  func() error { return InitEngine(cfg) },
//	        Shutdown: func() error { Close(); return nil },
//	    })
//	}
package boot

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/multierr"
)

// Task 启动任务
type Task struct {
	Startup  func() error
	Shutdown func() error
	Depends  []string

	booting atomic.Bool
	booted  atomic.Bool
}

var (
	mu     sync.RWMutex
	tasks  = make(map[string]*Task)
	queue  []string
)

// Register 注册启动任务
func Register(name string, task *Task) {
	mu.Lock()
	defer mu.Unlock()
	tasks[name] = task
}

// Startup 启动所有已注册任务，按依赖拓扑排序
func Startup() error {
	start := time.Now()
	mu.Lock()
	defer mu.Unlock()

	for name, task := range tasks {
		if task.booting.Load() || task.booted.Load() {
			continue
		}
		if err := openLocked(name); err != nil {
			return err
		}
	}

	fmt.Printf("[boot] all started in %dms\n", time.Since(start).Milliseconds())
	return nil
}

// Shutdown 逆序关闭所有已启动任务
func Shutdown() error {
	start := time.Now()
	var err error
	for i := len(queue) - 1; i >= 0; i-- {
		name := queue[i]
		if t, ok := tasks[name]; ok && t.booted.Load() {
			fmt.Printf("[boot] close %s\n", name)
			if t.Shutdown != nil {
				err = multierr.Append(err, t.Shutdown())
			}
			t.booted.Store(false)
		}
	}
	queue = nil
	fmt.Printf("[boot] shutdown in %dms\n", time.Since(start).Milliseconds())
	return err
}

func openLocked(name string) error {
	task := tasks[name]
	if task == nil {
		return fmt.Errorf("[boot] task not found: %s", name)
	}
	if task.booting.Load() || task.booted.Load() {
		return nil
	}
	task.booting.Store(true)
	defer task.booting.Store(false)

	// 启动依赖
	for _, dep := range task.Depends {
		if _, ok := tasks[dep]; !ok {
			fmt.Printf("[boot] warn: dependency %q not registered (skip)\n", dep)
			continue
		}
		if err := openLocked(dep); err != nil {
			return err
		}
	}

	start := time.Now()
	if err := task.Startup(); err != nil {
		return fmt.Errorf("[boot] %s: %w", name, err)
	}
	queue = append(queue, name)
	task.booted.Store(true)
	fmt.Printf("[boot] open %-12s  %dms\n", name, time.Since(start).Milliseconds())
	return nil
}
