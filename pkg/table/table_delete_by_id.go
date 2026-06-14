package table

import (
	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// DeleteById 根据ID删除数据
func (t *Table) DeleteById(id any) (rows int64, err error) {
	if t.BeforeDelete != nil {
		err = t.BeforeDelete(id)
		if err != nil {
			return
		}
	}

	cs, err := t.condId(id)
	if err != nil {
		return 0, err
	}
	bdr := builder.Dialect(db.Engine().DriverName()).Delete().From(db.Engine().Quote(t.Name))
	for _, c := range cs {
		bdr.Where(c)
	}

	res, err := db.Engine().Exec(bdr)
	if err != nil {
		return 0, err
	}

	rows, _ = res.RowsAffected()

	if t.AfterDelete != nil {
		// 获取删除前的文档
		doc, _ := t.Get(id, nil)
		err = t.AfterDelete(id, doc)
	}

	return rows, err
}

// DeleteByIdEx 根据ID删除数据（带额外过滤条件）
func (t *Table) DeleteByIdEx(id any, filter map[string]any) (rows int64, err error) {
	if t.BeforeDelete != nil {
		err = t.BeforeDelete(id)
		if err != nil {
			return
		}
	}

	cs, err := t.condId(id)
	if err != nil {
		return 0, err
	}
	bdr := builder.Dialect(db.Engine().DriverName()).Delete().From(db.Engine().Quote(t.Name))
	for _, c := range cs {
		bdr.Where(c)
	}

	if len(filter) > 0 {
		cs, err = t.condWhere(filter, false)
		if err != nil {
			return 0, err
		}
		for _, c := range cs {
			bdr.Where(c)
		}
	}

	res, err := db.Engine().Exec(bdr)
	if err != nil {
		return 0, err
	}

	rows, _ = res.RowsAffected()

	if t.AfterDelete != nil {
		// 获取删除前的文档
		doc, _ := t.Get(id, nil)
		err = t.AfterDelete(id, doc)
	}

	return rows, err
}