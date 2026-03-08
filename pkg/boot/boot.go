package boot

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/busy-cloud/boat/lib"
	"go.uber.org/multierr"
)

type Task struct {
	Startup  func() error
	Shutdown func() error
	Depends  []string

	booting atomic.Bool
	booted  atomic.Bool
}

var tasks lib.Map[Task]
var queue []string

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
		err = Open(name, nil)
		if err != nil {
			return false
		}
		return true
	})

	end := time.Now().UnixMilli()
	fmt.Printf("[boot] startup finished in %dms\n", end-start)
	return
}

func Shutdown() (err error) {
	start := time.Now().UnixMilli()

	//逆序关闭
	for i := len(queue) - 1; i >= 0; i-- {
		task := queue[i]
		println("close task", task)
		err = multierr.Append(err, Close(task))
	}

	//tasks.Range(func(name string, task *Task) bool {
	//	err = Close(name)
	//	if err != nil {
	//		//return false
	//		//log.Error(err)
	//		println(err.Error())
	//	}
	//	return true
	//})

	end := time.Now().UnixMilli()
	fmt.Printf("[boot] shutdown %dms\n", end-start)
	return
}

func Open(name string, parent []string) error {
	//重复检查
	if len(parent) > 0 {
		//fmt.Printf("[boot] open %s, depended by %s \n", name, strings.Join(parent, "->"))

		last := parent[len(parent)-1]
		if last == name {
			return fmt.Errorf("任务循环依赖 %s", strings.Join(parent, "->"))
		}
	}

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
			t := tasks.Load(n) //没有找到的依赖项
			if t != nil {
				//parent = append(parent, name)
				err := Open(n, append(parent, name)) //没有递归检查，可能会死循环
				if err != nil {
					return err
				}
			}
		}
	}

	//log.Info("[boot] open", name)
	start := time.Now().UnixMilli()
	//println("[startup]", name)

	//正式启动
	err := task.Startup()
	queue = append(queue, name)

	//计算时间
	end := time.Now().UnixMilli()
	fmt.Printf("[boot] open %s \t %dms\n", name, end-start)

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
		//println("[boot] close", name)
		fmt.Printf("[boot] close %s\n", name)
		return task.Shutdown()
	}
	return nil
}
