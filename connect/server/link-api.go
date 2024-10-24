package server

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/api"
	"github.com/god-jason/iot-master/curd"
)

func init() {
	api.Register("POST", "/link/count", curd.ApiCount[Link]())
	api.Register("POST", "/link/search", curd.ApiSearchWithHook[Link](
		[]*curd.With{{"server", "server_id", "id", "name", "server"}},
		func(links []*Link) error {
			for k, link := range links {
				c := GetLink(link.Id)
				if c != nil {
					links[k].Status = c.Status
				}
			}
			return nil
		}))

	api.Register("GET", "/link/list", curd.ApiList[Link]())

	api.Register("POST", "/link/create", curd.ApiCreateHook[Link](curd.GenerateID[Link](), nil))

	api.Register("GET", "/link/:id", curd.ParseParamStringId, curd.ApiGetHook[Link](func(link *Link) error {
		c := GetLink(link.Id)
		if c != nil {
			link.Status = c.Status
		}
		return nil
	}))

	api.Register("POST", "/link/:id", curd.ParseParamStringId, curd.ApiUpdateHook[Link](nil, nil))

	api.Register("GET", "/link/:id/delete", curd.ParseParamStringId, curd.ApiDeleteHook[Link](nil, func(m *Link) error {
		l := GetLink(m.Id)
		if l != nil {
			links.Delete(l.Id)
			return l.Close()
		}
		return nil
	}))

	api.Register("GET", "/link/:id/disable", curd.ParseParamStringId, curd.ApiDisableHook[Link](true, nil, func(value interface{}) error {
		id := value.(string)
		c := GetLink(id)
		return c.Close()
	}))

	api.Register("GET", "/link/:id/enable", curd.ParseParamStringId, curd.ApiDisableHook[Link](false, nil, nil))

	api.Register("GET", "/link/:id/close", curd.ParseParamStringId, func(ctx *gin.Context) {
		id := ctx.GetString("id")
		c := GetLink(id)
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

// @Summary 查询连接数量
// @Schemes
// @Description 查询连接数量
// @Tags link
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int64] 返回连接数量
// @Router /link/count [post]
func noopLinkCount() {}

// @Summary 查询连接
// @Schemes
// @Description 查询连接
// @Tags link
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Link] 返回连接信息
// @Router /link/search [post]
func noopLinkSearch() {}

// @Summary 查询连接
// @Schemes
// @Description 查询连接
// @Tags link
// @Param search query curd.ParamList true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Link] 返回连接信息
// @Router /link/list [get]
func noopLinkList() {}

// @Summary 创建连接
// @Schemes
// @Description 创建连接
// @Tags link
// @Param link body Link true "连接信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Link] 返回连接信息
// @Router /link/create [post]
func noopLinkCreate() {}

// @Summary 修改连接
// @Schemes
// @Description 修改连接
// @Tags link
// @Param id path int true "连接ID"
// @Param link body Link true "连接信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Link] 返回连接信息
// @Router /link/{id} [post]
func noopLinkUpdate() {}

// @Summary 获取连接
// @Schemes
// @Description 获取连接
// @Tags link
// @Param id path int true "连接ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Link] 返回连接信息
// @Router /link/{id} [get]
func noopLinkGet() {}

// @Summary 删除连接
// @Schemes
// @Description 删除连接
// @Tags link
// @Param id path int true "连接ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Link] 返回连接信息
// @Router /link/{id}/delete [get]
func noopLinkDelete() {}

// @Summary 启用连接
// @Schemes
// @Description 启用连接
// @Tags link
// @Param id path int true "连接ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Link] 返回连接信息
// @Router /link/{id}/enable [get]
func noopLinkEnable() {}

// @Summary 禁用连接
// @Schemes
// @Description 禁用连接
// @Tags link
// @Param id path int true "连接ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Link] 返回连接信息
// @Router /link/{id}/disable [get]
func noopLinkDisable() {}

// @Summary 停止连接
// @Schemes
// @Description 停止连接
// @Tags link
// @Param id path int true "连接ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Link] 返回连接信息
// @Router /link/{id}/close [get]
func noopLinkStop() {}
