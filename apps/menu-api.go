package apps

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/app"
)

func init() {
	//api.Register("GET", "/menu/:domain", menuGet)
	api.Register("GET", "menus", menuGet)
}

// @Summary 获取菜单
// @Schemes
// @Description 获取菜单
// @Tags plugin
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[[]Menu] 返回插件信息
// @Router /menus [get]
func menuGet(ctx *gin.Context) {
	var ms []*app.Menu

	_apps.Range(func(name string, a *App) bool {
		if len(a.Menus) > 0 {
			ms = append(ms, a.Menus...)
		}
		return true
	})

	//排序
	slices.SortFunc(ms, func(a, b *app.Menu) int {
		return a.Index - b.Index
	})

	api.OK(ctx, ms)
}
