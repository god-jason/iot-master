package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/database"
	"github.com/zgwit/iot-master/log"
	"github.com/zgwit/iot-master/master"
	"github.com/zgwit/iot-master/model"
	"github.com/zgwit/storm/v3/q"
	"golang.org/x/net/websocket"
)

func tunnelRoutes(app *gin.RouterGroup) {
	app.POST("list", tunnelList)
	app.POST("create", tunnelCreate)

	app.GET("event/clear", tunnelEventClearAll)

	app.Use(parseParamId)
	app.GET(":id", tunnelDetail)
	app.POST(":id", tunnelUpdate)
	app.GET(":id/delete", tunnelDelete)
	app.GET(":id/start", tunnelStart)
	app.GET(":id/stop", tunnelStop)
	app.GET(":id/enable", tunnelEnable)
	app.GET(":id/disable", tunnelDisable)
	app.GET(":id/watch", tunnelWatch)
	app.POST(":id/event/list", tunnelEvent)
	app.GET(":id/event/clear", tunnelEventClear)
}

func tunnelList(ctx *gin.Context) {
	records, cnt, err := normalSearch(ctx, database.Master, &model.Tunnel{})
	if err != nil {
		replyError(ctx, err)
		return
	}

	//补充信息
	tunnels := records.(*[]*model.Tunnel)
	ts := make([]*model.TunnelEx, 0) //len(tunnels)

	for _, d := range *tunnels {
		l := &model.TunnelEx{Tunnel: *d}
		ts = append(ts, l)
		d := master.GetTunnel(l.Id)
		if d != nil {
			l.Running = d.Instance.Running()
		}
	}

	replyList(ctx, ts, cnt)
}

func tunnelCreate(ctx *gin.Context) {
	var tunnel model.Tunnel
	err := ctx.ShouldBindJSON(&tunnel)
	if err != nil {
		replyError(ctx, err)
		return
	}

	err = database.Master.Save(&tunnel)
	if err != nil {
		replyError(ctx, err)
		return
	}

	replyOk(ctx, tunnel)

	//启动
	//tunnelStart(ctx)
	if !tunnel.Disabled {
		go func() {
			err := master.LoadTunnel(tunnel.Id)
			if err != nil {
				log.Error(err)
				return
			}
		}()
	}
}

func tunnelDetail(ctx *gin.Context) {
	var tunnel model.Tunnel
	err := database.Master.One("Id", ctx.GetInt("id"), &tunnel)
	if err != nil {
		replyError(ctx, err)
		return
	}
	tnl := &model.TunnelEx{Tunnel: tunnel}
	d := master.GetTunnel(tnl.Id)
	if d != nil {
		tnl.Running = d.Instance.Running()
	}
	replyOk(ctx, tnl)
}

func tunnelUpdate(ctx *gin.Context) {
	var tunnel model.Tunnel
	err := ctx.ShouldBindJSON(&tunnel)
	if err != nil {
		replyError(ctx, err)
		return
	}
	tunnel.Id = ctx.GetInt("id")

	err = database.Master.Update(&tunnel)
	if err != nil {
		replyError(ctx, err)
		return
	}

	replyOk(ctx, tunnel)

	//重新启动
	go func() {
		_ = master.RemoveTunnel(tunnel.Id)
		err := master.LoadTunnel(tunnel.Id)
		if err != nil {
			log.Error(err)
			return
		}
	}()
}

func tunnelDelete(ctx *gin.Context) {
	tunnel := model.Tunnel{Id: ctx.GetInt("id")}
	err := database.Master.DeleteStruct(&tunnel)
	if err != nil {
		replyError(ctx, err)
		return
	}

	replyOk(ctx, tunnel)

	//关闭
	go func() {
		err := master.RemoveTunnel(tunnel.Id)
		if err != nil {
			log.Error(err)
		}
	}()
}

func tunnelStart(ctx *gin.Context) {
	tunnel := master.GetTunnel(ctx.GetInt("id"))
	if tunnel == nil {
		replyFail(ctx, "not found")
		return
	}
	err := tunnel.Instance.Open()
	if err != nil {
		replyError(ctx, err)
		return
	}

	replyOk(ctx, nil)
}

func tunnelStop(ctx *gin.Context) {
	tunnel := master.GetTunnel(ctx.GetInt("id"))
	if tunnel == nil {
		replyFail(ctx, "not found")
		return
	}
	err := tunnel.Instance.Close()
	if err != nil {
		replyError(ctx, err)
		return
	}

	replyOk(ctx, nil)
}

func tunnelEnable(ctx *gin.Context) {
	err := database.Master.UpdateField(&model.Tunnel{Id: ctx.GetInt("id")}, "Disabled", false)
	if err != nil {
		replyError(ctx, err)
		return
	}
	replyOk(ctx, nil)

	//启动
	go func() {
		err := master.LoadTunnel(ctx.GetInt("id"))
		if err != nil {
			log.Error(err)
			return
		}
	}()
}

func tunnelDisable(ctx *gin.Context) {
	err := database.Master.UpdateField(&model.Tunnel{Id: ctx.GetInt("id")}, "Disabled", true)
	if err != nil {
		replyError(ctx, err)
		return
	}
	replyOk(ctx, nil)

	//关闭
	go func() {
		tunnel := master.GetTunnel(ctx.GetInt("id"))
		if tunnel == nil {
			return
		}
		err := tunnel.Instance.Close()
		if err != nil {
			log.Error(err)
			return
		}
	}()
}

func tunnelWatch(ctx *gin.Context) {
	tunnel := master.GetTunnel(ctx.GetInt("id"))
	if tunnel == nil {
		replyFail(ctx, "找不到通道")
		return
	}
	websocket.Handler(func(ws *websocket.Conn) {
		watchAllEvents(ws, tunnel.Instance)
	}).ServeHTTP(ctx.Writer, ctx.Request)
}

func tunnelEvent(ctx *gin.Context) {
	events, cnt, err := normalSearchById(ctx, database.History, "TunnelId", ctx.GetInt("id"), &model.TunnelEvent{})
	if err != nil {
		replyError(ctx, err)
		return
	}
	replyList(ctx, events, cnt)
}

func tunnelEventClear(ctx *gin.Context) {
	err := database.History.Select(q.Eq("TunnelId", ctx.GetInt("id"))).Delete(&model.TunnelEvent{})
	if err != nil {
		replyError(ctx, err)
		return
	}

	replyOk(ctx, nil)
}

func tunnelEventClearAll(ctx *gin.Context) {
	err := database.History.Drop(&model.TunnelEvent{})
	if err != nil {
		replyError(ctx, err)
		return
	}

	replyOk(ctx, nil)
}
