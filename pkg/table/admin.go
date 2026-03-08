package table

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var Path = "tables"

func ApiTableList(ctx *gin.Context) {

	var infos []*Table
	//&InfoEx{
	//	Name:          tab.Name,
	//	Label:         tab.Name,
	//	Columns:        len(tab.Columns),
	//	Schema:        tab.Schema != nil,
	//	Scripts:       len(tab.Scripts),
	//	Accumulations: len(tab.Accumulations),
	//	TimeSeries:    tab.TimeSeries != nil,
	//	Snapshot:      tab.Snapshot != nil,
	//}
	tables.Range(func(name string, tab *Table) bool {
		infos = append(infos, tab)
		return true
	})

	OK(ctx, infos)
}

func ApiTableReload(ctx *gin.Context) {
	tab := ctx.Param("table")
	err := Load(tab)
	if err != nil {
		Error(ctx, err)
		return
	}

	OK(ctx, nil)
}

type rename struct {
	Name string `json:"name"`
}

func ApiTableRename(ctx *gin.Context) {
	var r rename
	err := ctx.ShouldBindJSON(&r)
	if err != nil {
		Error(ctx, err)
		return
	}

	tab := ctx.Param("table")
	old := filepath.Join(viper.GetString("data"), Path, tab)
	name := filepath.Join(viper.GetString("data"), Path, r.Name)
	err = os.Rename(old, name)
	if err != nil {
		Error(ctx, err)
		return
	}

	//直接修改map，不雅
	t := tables.LoadAndDelete(tab)
	//t.Id = r.Name
	tables.Store(r.Name, t)

	OK(ctx, nil)
}

func ApiTableRemove(ctx *gin.Context) {
	tab := ctx.Param("table")
	dir := filepath.Join(viper.GetString("data"), Path, tab)
	err := os.RemoveAll(dir)
	if err != nil {
		Error(ctx, err)
		return
	}

	OK(ctx, nil)
}

func ApiTableCreate(ctx *gin.Context) {
	tab := ctx.Param("table")
	dir := filepath.Join(viper.GetString("data"), Path, tab)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		Error(ctx, err)
		return
	}

	//TODO 空白的
	_ = Load(tab)

	OK(ctx, nil)

}
