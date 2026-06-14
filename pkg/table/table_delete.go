package table

import (
	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// Delete 根据条件删除数据
func (t *Table) Delete(filter map[string]any) (rows int64, err error) {
	bdr := builder.Dialect(db.Engine().DriverName()).Delete().From(db.Engine().Quote(t.Name))

	cs, err := t.condWhere(filter, false)
	if err != nil {
		return 0, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	res, err := db.Engine().Exec(bdr)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
