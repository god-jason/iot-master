package table

import "github.com/busy-cloud/boat/api"

// 表相关接口只能放这里了，否则会import circle: api->table->api
func init() {

	//表管理 限管理员权限
	//api.RegisterAdmin("GET", "table/list", ApiTableList)
	//api.RegisterAdmin("POST", "table/:table", ApiTableCreate)
	//api.RegisterAdmin("POST", "table/:table/rename", ApiTableRename)
	//api.RegisterAdmin("GET", "table/:table/remove", ApiTableRemove)
	//api.RegisterAdmin("GET", "table/:table/reload", ApiTableReload)
	//api.RegisterAdmin("GET", "table/:table/conf/*conf", ApiConf)
	//api.RegisterAdmin("POST", "table/:table/conf/*conf", ApiConfUpdate)

	//表接口 普通权限
	api.Register("POST", "table/:table/count", ApiCount)
	api.Register("POST", "table/:table/create", ApiCreate)
	api.Register("POST", "table/:table/update/*id", ApiUpdate)
	api.Register("GET", "table/:table/delete/*id", ApiDelete)
	api.Register("GET", "table/:table/detail/*id", ApiDetail)
	api.Register("POST", "table/:table/search", ApiSearch)
	api.Register("POST", "table/:table/import", ApiImport)
	api.Register("POST", "table/:table/export", ApiExport)
}
