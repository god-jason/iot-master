package apps

import (
	_ "embed"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/app"
)

//go:embed icon.png
var defaultIcon []byte

const gmtFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func init() {

	//启动时间作为默认图标的修改时间
	bootTime := time.Now()

	api.Register("GET", "app/list", func(ctx *gin.Context) {
		var as []*app.Base
		_apps.Range(func(name string, item *App) bool {
			as = append(as, &item.Base)
			return true
		})
		api.OK(ctx, as)
	})

	api.Register("GET", "app/:app", func(ctx *gin.Context) {
		app := _apps.Load(ctx.Param("app"))
		if app != nil {
			api.Fail(ctx, "找不到插件")
			return
		}
		api.OK(ctx, app)
	})

	api.Register("GET", "app/:app/assets/*asset", func(ctx *gin.Context) {

	})

	//api.Register("GET", "app/:app/assets/*asset", func(ctx *gin.Context) {
	//	app := _apps.Load(ctx.Param("app"))
	//	ctx.FileFromFS(ctx.Param("asset"), http.FS(app.Assets))
	//})

	api.Register("GET", "app/:app/icon", func(ctx *gin.Context) {
		app := _apps.Load(ctx.Param("app"))
		if app != nil {
			api.Fail(ctx, "找不到插件")
			return
		}

		icon := app.Icon
		if icon == "" {
			icon = IconName
		}

		file, err := os.Open(IconName)
		if err != nil {
			//return nil, err
			//return icon, nil //使用默认图片
			ctx.Header("Last-Modified", bootTime.UTC().Format(gmtFormat))
			ctx.Header("Content-Type", "image/png")
			_, _ = ctx.Writer.Write(defaultIcon)
			return
		}
		defer file.Close()

		st, _ := file.Stat()
		buf, err := io.ReadAll(file)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.Header("Last-Modified", st.ModTime().UTC().Format(gmtFormat))
		ctx.Header("Content-Type", "image/png")
		_, _ = ctx.Writer.Write(buf)
	})

	api.RegisterAdmin("POST", "app/import", func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			api.Error(ctx, err)
			return
		}
		f, err := file.Open()
		if err != nil {
			api.Error(ctx, err)
			return
		}
		defer f.Close()

		f2, err := os.CreateTemp("app", "install-*")
		if err != nil {
			api.Error(ctx, err)
			return
		}
		defer f2.Close()

		_, err = io.Copy(f2, f)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		err = app.Unpack(app.PublicKey(), f2.Name(), RootPath)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, nil)
	})

	api.RegisterAdmin("GET", "app/:app/install", func(ctx *gin.Context) {

	})

	api.RegisterAdmin("GET", "app/:app/delete", func(ctx *gin.Context) {

	})

	api.RegisterAdmin("GET", "app/privileges", func(ctx *gin.Context) {
		var ps []*app.Privilege
		_apps.Range(func(name string, item *App) bool {
			ps = append(ps, item.Privileges...)
			return true
		})
		api.OK(ctx, ps)
	})

}
