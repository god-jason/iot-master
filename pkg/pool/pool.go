package pool

import (
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/exception"
	"github.com/busy-cloud/boat/log"
	ants "github.com/panjf2000/ants/v2"
)

var Pool *ants.Pool

func Startup() (err error) {
	Pool, err = ants.NewPool(config.GetInt(MODULE, "size"), ants.WithPanicHandler(func(err any) {
		log.Error(err)
	}))
	return exception.Wrap(err)
}

func Shutdown() error {
	if Pool != nil {
		Pool.Release()
		Pool = nil
	}
	return nil
}

func Insert(task func()) error {
	if Pool == nil {
		go task()
	}
	err := Pool.Submit(task)
	return exception.Wrap(err)
}
