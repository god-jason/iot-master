package client

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/api"
	"github.com/zgwit/iot-master/curd"
	"github.com/zgwit/iot-master/db"
)

func init() {

	api.Register("POST", "/client/count", curd.ApiCount[Client]())

	api.Register("POST", "/client/search", curd.ApiSearchHook[Client](func(clients []*Client) error {
		for k, client := range clients {
			c := GetClient(client.Id)
			if c != nil {
				clients[k].Status = c.Status
			}
		}
		return nil
	}))

	api.Register("GET", "/client/list", curd.ApiList[Client]())

	api.Register("POST", "/client/create", curd.ApiCreateHook[Client](curd.GenerateID[Client](), func(value *Client) error {
		return LoadClient(value)
	}))

	api.Register("GET", "/client/:id", curd.ParseParamStringId, curd.ApiGetHook[Client](func(client *Client) error {
		c := GetClient(client.Id)
		if c != nil {
			client.Status = c.Status
		}
		return nil
	}))

	api.Register("POST", "/client/:id", curd.ParseParamStringId, curd.ApiUpdateHook[Client](nil, func(value *Client) error {
		c := GetClient(value.Id)
		if c != nil {
			err := c.Close()
			if err != nil {
				return err
			}
		}
		return LoadClient(value)
	}))

	api.Register("GET", "/client/:id/delete", curd.ParseParamStringId, curd.ApiDeleteHook[Client](nil, func(value *Client) error {
		c := GetClient(value.Id)
		if c != nil {
			clients.Delete(value.Id)
			return c.Close()
		}
		return nil
	}))

	api.Register("GET", "/client/:id/disable", curd.ParseParamStringId, curd.ApiDisableHook[Client](true, nil, func(value interface{}) error {
		id := value.(string)
		c := GetClient(id)
		return c.Close()
	}))

	api.Register("GET", "/client/:id/enable", curd.ParseParamStringId, curd.ApiDisableHook[Client](false, nil, func(value interface{}) error {
		id := value.(string)
		var m Client
		has, err := db.Engine.ID(id).Get(&m)
		if err != nil {
			return err
		}
		if !has {
			return fmt.Errorf("找不到 %s", id)
		}
		return LoadClient(&m)
	}))

	api.Register("GET", "/client/:id/open", curd.ParseParamStringId, func(ctx *gin.Context) {
		id := ctx.GetString("id")
		c := GetClient(id)
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

	api.Register("GET", "/client/:id/close", curd.ParseParamStringId, func(ctx *gin.Context) {
		id := ctx.GetString("id")
		c := GetClient(id)
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

// @Summary 查询客户端数量
// @Schemes
// @Description 查询客户端数量
// @Tags client
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int64] 返回客户端数量
// @Router /client/count [post]
func noopClientCount() {}

// @Summary 查询客户端
// @Schemes
// @Description 查询客户端
// @Tags client
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Client] 返回客户端信息
// @Router /client/search [post]
func noopClientSearch() {}

// @Summary 查询客户端
// @Schemes
// @Description 查询客户端
// @Tags client
// @Param search query curd.ParamList true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Client] 返回客户端信息
// @Router /client/list [get]
func noopClientList() {}

// @Summary 创建客户端
// @Schemes
// @Description 创建客户端
// @Tags client
// @Param search body Client true "客户端信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Client] 返回客户端信息
// @Router /client/create [post]
func noopClientCreate() {}

// @Summary 修改客户端
// @Schemes
// @Description 修改客户端
// @Tags client
// @Param id path int true "客户端ID"
// @Param client body Client true "客户端信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Client] 返回客户端信息
// @Router /client/{id} [post]
func noopClientUpdate() {}

// @Summary 获取客户端
// @Schemes
// @Description 获取客户端
// @Tags client
// @Param id path int true "客户端ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Client] 返回客户端信息
// @Router /client/{id} [get]
func noopClientGet() {}

// @Summary 删除客户端
// @Schemes
// @Description 删除客户端
// @Tags client
// @Param id path int true "客户端ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Client] 返回客户端信息
// @Router /client/{id}/delete [get]
func noopClientDelete() {}

// @Summary 启用客户端
// @Schemes
// @Description 启用客户端
// @Tags client
// @Param id path int true "客户端ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Client] 返回客户端信息
// @Router /client/{id}/enable [get]
func noopClientEnable() {}

// @Summary 禁用客户端
// @Schemes
// @Description 禁用客户端
// @Tags client
// @Param id path int true "客户端ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Client] 返回客户端信息
// @Router /client/{id}/disable [get]
func noopClientDisable() {}

// @Summary 启动客户端
// @Schemes
// @Description 启动客户端
// @Tags client
// @Param id path int true "客户端ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Client] 返回客户端信息
// @Router /client/{id}/open [get]
func noopClientStart() {}

// @Summary 停止客户端
// @Schemes
// @Description 停止客户端
// @Tags client
// @Param id path int true "客户端ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Client] 返回客户端信息
// @Router /client/{id}/close [get]
func noopClientStop() {}
