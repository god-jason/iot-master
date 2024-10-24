package boot

import (
	"fmt"
	"github.com/god-jason/iot-master/lib"
	"sync/atomic"
	"time"
)

type Task struct {
	Startup  func() error
	Shutdown func() error
	Depends  []string

	booting atomic.Bool
	booted  atomic.Bool
}

var tasks lib.Map[Task]

func Load(name string) *Task {
	return tasks.Load(name)
}

func Register(name string, task *Task) {
	//println("[boot] register", name)
	tasks.Store(name, task)
}

func Unregister(name string) {
	tasks.Delete(name)
}

func Startup() (err error) {
	start := time.Now().UnixMilli()
	tasks.Range(func(name string, task *Task) bool {
		//过滤掉依赖启动
		if task.booting.Load() || task.booted.Load() {
			return true
		}
		//启动
		err = Open(name)
		if err != nil {
			return false
		}
		return true
	})
	end := time.Now().UnixMilli()
	fmt.Printf("[boot] finished in %dms\n", end-start)
	return
}

func Shutdown() (err error) {
	tasks.Range(func(name string, task *Task) bool {
		err = Close(name)
		if err != nil {
			return false
		}
		return true
	})
	return
}

func Open(name string) error {
	task := tasks.Load(name)
	if task == nil {
		return fmt.Errorf("找不到任务 %s", name)
	}

	//过滤掉依赖启动
	if task.booting.Load() || task.booted.Load() {
		return nil
	}

	task.booting.Store(true)
	defer task.booting.Store(false)

	//启动依赖
	if len(task.Depends) > 0 {
		for _, n := range task.Depends {
			err := Open(n) //TODO 没有递归检查，可能会死循环
			if err != nil {
				return err
			}
		}
	}

	//log.Info("[boot] open", name)
	start := time.Now().UnixMilli()
	//println("[startup]", name)

	//正式启动
	err := task.Startup()

	//计算时间
	end := time.Now().UnixMilli()
	fmt.Printf("[startup] %s \t %dms\n", name, end-start)

	task.booted.Store(true) //不管成功失败，都代表已经启动了
	if err != nil {
		return err
	}

	return nil
}

func Close(name string) error {
	task := tasks.Load(name)
	if task == nil {
		return fmt.Errorf("找不到任务 %s", name)
	}
	task.booted.Store(false)
	if task.Shutdown != nil {
		//log.Info("[boot] close", name)
		println("[shutdown]", name)
		return task.Shutdown()
	}
	return nil
}
