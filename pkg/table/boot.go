package table

import (
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/log"
)

func init() {
	boot.Register("table", &boot.Task{
		Startup: Startup,
		Depends: []string{"config", "database", "apps"},
	})
}

func Startup() error {
	var err error

	//加载表
	path := config.GetString(MODULE, "path")
	if len(path) != 0 {
		err = Scan(path)
		if err != nil {
			return err
		}
	}

	//初始化，编译钩子
	tables.Range(func(key, value any) bool {
		err := value.(*Table).Init()
		if err != nil {
			log.Error(err)
		}
		return true
	})

	//同步表结构
	if config.GetBool(MODULE, "sync") {
		var tbs []*Table
		tables.Range(func(key, value any) bool {
			tbs = append(tbs, value.(*Table))
			return true
		})
		if len(tbs) > 0 {
			err = Sync(tbs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
