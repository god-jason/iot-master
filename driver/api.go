package driver

import (
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/api"
	"github.com/zgwit/iot-master/curd"
)

func init() {

	api.Register("GET", "/driver/list", func(ctx *gin.Context) {
		var ps []*Driver
		for _, p := range drivers {
			ps = append(ps, p)
		}
		curd.OK(ctx, ps)
	})

	api.Register("GET", "/driver/:name/mapper", func(ctx *gin.Context) {
		name := ctx.Param("name")
		if p, ok := drivers[name]; ok {
			curd.OK(ctx, p.MapperForm)
		} else {
			curd.Fail(ctx, "协议找不到")
		}
	})

	api.Register("GET", "/driver/:name/poller", func(ctx *gin.Context) {
		name := ctx.Param("name")
		if p, ok := drivers[name]; ok {
			curd.OK(ctx, p.PollersForm)
		} else {
			curd.Fail(ctx, "协议找不到")
		}
	})

	api.Register("GET", "/driver/:name/option", func(ctx *gin.Context) {
		name := ctx.Param("name")
		if p, ok := drivers[name]; ok {
			curd.OK(ctx, p.OptionForm)
		} else {
			curd.Fail(ctx, "协议找不到")
		}
	})

	api.Register("GET", "/driver/:name/station", func(ctx *gin.Context) {
		name := ctx.Param("name")
		if p, ok := drivers[name]; ok {
			curd.OK(ctx, p.StationForm)
		} else {
			curd.Fail(ctx, "协议找不到")
		}
	})
}

// @Summary 协议列表
// @Schemes
// @Description 协议列表
// @Tags driver
// @Produce json
// @Success 200 {object} curd.ReplyData[Driver] 返回协议列表
// @Router /driver/list [get]
func noopProtocolList() {}

// @Summary 协议参数
// @Schemes
// @Description 协议参数
// @Tags driver
// @Produce json
// @Success 200 {object} curd.ReplyData[[]types.FormItem] 返回协议参数
// @Router /driver/option [get]
func noopProtocolOptions() {}

// @Summary 协议轮询器
// @Schemes
// @Description 协议轮询器
// @Tags driver
// @Produce json
// @Success 200 {object} curd.ReplyData[[]types.FormItem] 返回协议轮询器
// @Router /driver/poller [get]
func noopProtocolPollers() {}

// @Summary 协议映射
// @Schemes
// @Description 协议映射
// @Tags driver
// @Produce json
// @Success 200 {object} curd.ReplyData[[]types.FormItem] 返回协议映射
// @Router /driver/mapper [get]
func noopProtocolMappers() {}

// @Summary 协议设备站号
// @Schemes
// @Description 协议设备站号
// @Tags driver
// @Produce json
// @Success 200 {object} curd.ReplyData[[]types.FormItem] 返回协议映射
// @Router /driver/station [get]
func noopProtocolStations() {}
