package table

import (
	"errors"

	"github.com/god-jason/iot-master/pkg/db"
	"github.com/spf13/cast"
	"xorm.io/builder"
)

// Count 统计符合条件的数据数量
func (t *Table) Count(filter map[string]any) (cnt int64, err error) {
	bdr := builder.Dialect(db.Engine().DriverName()).Select("count(*)").From(db.Engine().Quote(t.Name))

	cs, err := t.condWhere(filter, false)
	if err != nil {
		return 0, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	res, err := db.Engine().QueryInterface(bdr)
	if err != nil {
		return 0, err
	}

	if len(res) == 0 {
		return 0, errors.New("没有可计数的值")
	}

	for _, v := range res[0] {
		return cast.ToInt64(v), nil
	}

	return 0, errors.New("没有可计数的值")
}