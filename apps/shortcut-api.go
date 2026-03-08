package apps

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/app"
)

func init() {
	api.Register("GET", "shortcuts", shortcutGet)
}

// @Summary 获取快捷方式
// @Schemes
// @Description 获取快捷方式
// @Tags plugin
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[[]Entry] 返回插件信息
// @Router /shortcuts [get]
func shortcutGet(ctx *gin.Context) {
	var ms []*app.Entry

	_apps.Range(func(name string, a *App) bool {
		if len(a.Shortcuts) > 0 {
			for _, m := range a.Shortcuts {
				mm := *m
				mm.Icon = "/app/assets/" + a.Id + "/" + m.Icon //转换路径
				ms = append(ms, &mm)
			}
		}
		return true
	})

	//排序
	//slices.SortFunc(ms, func(a, b *app.Entry) int {
	//	return a.Index - b.Index
	//})

	api.OK(ctx, ms)
}
