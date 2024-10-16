package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/api"
	"github.com/zgwit/iot-master/curd"
	"github.com/zgwit/iot-master/db"
)

func init() {

	api.Register("POST", "/server/count", curd.ApiCount[Server]())

	api.Register("POST", "/server/search", curd.ApiSearchHook[Server](func(servers []*Server) error {
		for k, server := range servers {
			c := GetServer(server.Id)
			if c != nil {
				servers[k].Status = c.Status
			}
		}
		return nil
	}))

	api.Register("GET", "/server/list", curd.ApiList[Server]())

	api.Register("POST", "/server/create", curd.ApiCreateHook[Server](curd.GenerateID[Server](), func(value *Server) error {
		return LoadServer(value)
	}))

	api.Register("GET", "/server/:id", curd.ParseParamStringId, curd.ApiGetHook[Server](func(server *Server) error {
		c := GetServer(server.Id)
		if c != nil {
			server.Status = c.Status
		}
		return nil
	}))

	api.Register("POST", "/server/:id", curd.ParseParamStringId, curd.ApiUpdateHook[Server](nil, func(value *Server) error {
		c := GetServer(value.Id)
		err := c.Close()
		if err != nil {
			log.Error(err)
		}
		return LoadServer(value)
	}))

	api.Register("GET", "/server/:id/delete", curd.ParseParamStringId, curd.ApiDeleteHook[Server](nil, func(value *Server) error {
		c := GetServer(value.Id)
		if c != nil {
			servers.Delete(value.Id)
			return c.Close()
		}
		return nil
	}))

	api.Register("GET", "/server/:id/disable", curd.ParseParamStringId, curd.ApiDisableHook[Server](true, nil, func(value interface{}) error {
		id := value.(string)
		c := GetServer(id)
		return c.Close()
	}))

	api.Register("GET", "/server/:id/enable", curd.ParseParamStringId, curd.ApiDisableHook[Server](false, nil, func(value interface{}) error {
		id := value.(string)
		var m Server
		has, err := db.Engine.ID(id).Get(&m)
		if err != nil {
			return err
		}
		if !has {
			return fmt.Errorf("找不到 %s", id)
		}
		return LoadServer(&m)
	}))

	api.Register("GET", "/server/:id/open", curd.ParseParamStringId, func(ctx *gin.Context) {
		id := ctx.GetString("id")
		c := GetServer(id)
		if c == nil {
			curd.Fail(ctx, "找不到连接")
			return
		}
		err := c.Open()
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		curd.OK(ctx, nil)
	})

	api.Register("GET", "/server/:id/close", curd.ParseParamStringId, func(ctx *gin.Context) {
		id := ctx.GetString("id")
		c := GetServer(id)
		if c == nil {
			curd.Fail(ctx, "找不到连接")
			return
		}
		err := c.Close()
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		curd.OK(ctx, nil)
	})

}

// @Summary 查询服务器数量
// @Schemes
// @Description 查询服务器数量
// @Tags server
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int64] 返回服务器数量
// @Router /server/count [post]
func noopServerCount() {}

// @Summary 查询服务器
// @Schemes
// @Description 查询服务器
// @Tags server
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Server] 返回服务器信息
// @Router /server/search [post]
func noopServerSearch() {}

// @Summary 查询服务器
// @Schemes
// @Description 查询服务器
// @Tags server
// @Param search query curd.ParamList true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Server] 返回服务器信息
// @Router /server/list [get]
func noopServerList() {}

// @Summary 创建服务器
// @Schemes
// @Description 创建服务器
// @Tags server
// @Param search body Server true "服务器信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Server] 返回服务器信息
// @Router /server/create [post]
func noopServerCreate() {}

// @Summary 修改服务器
// @Schemes
// @Description 修改服务器
// @Tags server
// @Param id path int true "服务器ID"
// @Param server body Server true "服务器信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Server] 返回服务器信息
// @Router /server/{id} [post]
func noopServerUpdate() {}

// @Summary 获取服务器
// @Schemes
// @Description 获取服务器
// @Tags server
// @Param id path int true "服务器ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Server] 返回服务器信息
// @Router /server/{id} [get]
func noopServerGet() {}

// @Summary 删除服务器
// @Schemes
// @Description 删除服务器
// @Tags server
// @Param id path int true "服务器ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Server] 返回服务器信息
// @Router /server/{id}/delete [get]
func noopServerDelete() {}

// @Summary 启用服务器
// @Schemes
// @Description 启用服务器
// @Tags server
// @Param id path int true "服务器ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Server] 返回服务器信息
// @Router /server/{id}/enable [get]
func noopServerEnable() {}

// @Summary 禁用服务器
// @Schemes
// @Description 禁用服务器
// @Tags server
// @Param id path int true "服务器ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Server] 返回服务器信息
// @Router /server/{id}/disable [get]
func noopServerDisable() {}

// @Summary 启动服务端
// @Schemes
// @Description 启动服务端
// @Tags server
// @Param id path int true "服务端ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Server] 返回服务端信息
// @Router /server/{id}/open [get]
func noopServerStart() {}

// @Summary 停止服务端
// @Schemes
// @Description 停止服务端
// @Tags server
// @Param id path int true "服务端ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Server] 返回服务端信息
// @Router /server/{id}/close [get]
func noopServerStop() {}
