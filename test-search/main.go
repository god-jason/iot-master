package main

import (
	"log"

	"github.com/goccy/go-json"
	"xorm.io/builder"
	"xorm.io/xorm"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	engine, err := xorm.NewEngine("mysql", "root:Flzx3000c@tcp(git.zgwit.com:3306)/boat?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}

	engine.ShowSQL(true)

	bdr := builder.Dialect(engine.DriverName()).Select("device.id", "device.name", "device.product_id", "product.name as product").From("device")
	bdr.LeftJoin("product", "product.id = device.product_id")

	rows, err := engine.QueryInterface(bdr)
	if err != nil {
		log.Fatal(err)
	}

	str, err := json.Marshal(rows)
	log.Println(string(str), err)
}
