package table

import (
	"context"

	"github.com/busy-cloud/boat/db"
	"xorm.io/xorm/schemas"
)

func Sync(tables []*Table) error {
	//此函数仿照xorm.Engine.Sync

	//每次把所有表都查出来了
	originTables, err := db.Engine().Dialect().GetTables(db.Engine().DB(), context.Background())
	if err != nil {
		return err
	}

	for _, table := range tables {

		var originTable *schemas.Table
		for _, ot := range originTables {

			if ot.Name == table.Name {
				originTable = ot
				break
			}
		}

		//新表
		if originTable == nil {
			err = table.Create()
			if err != nil {
				return err
			}
			continue
		}

		_, columns, err := db.Engine().Dialect().GetColumns(db.Engine().DB(), context.Background(), originTable.Name)
		if err != nil {
			return err
		}

		//更新列
		schema := table.Schema()
		for _, col := range schema.Columns() {
			found := false
			for _, oc := range columns {
				if col.Name == oc.Name {
					found = true
					break
				}
			}

			//添加新列
			if !found {
				originTable.AddColumn(col)
				sql := db.Engine().Dialect().AddColumnSQL(originTable.Name, col)
				_, err = db.Engine().Exec(sql)
				if err != nil {
					return err
				}
			}

			//TODO 处理修改字段类型
		}

		indexes, err := db.Engine().Dialect().GetIndexes(db.Engine().DB(), context.Background(), originTable.Name)
		if err != nil {
			return err
		}

		//更新索引
		for _, index := range schema.Indexes {
			found := false
			for _, oi := range indexes {
				if index.Equal(oi) {
					found = true
					break
				}
			}

			//添加索引
			if !found {
				sql := db.Engine().Dialect().CreateIndexSQL(originTable.Name, index)
				_, err = db.Engine().Exec(sql)
				if err != nil {
					return err
				}
			}

			//TODO 删除多余索引
		}

	}

	return err
}
