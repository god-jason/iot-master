package attach

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/zgwit/iot-master/curd"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type attachInfo struct {
	Name   string    `json:"name,omitempty"`
	Mime   string    `json:"mime,omitempty"`
	Time   time.Time `json:"time"`
	Size   int64     `json:"size,omitempty"`
	Folder bool      `json:"folder,omitempty"`
}

type RenameBody struct {
	Name string `json:"name,omitempty"`
}

type MoveBody struct {
	Path string `json:"path,omitempty"`
}

func getParams(ctx *gin.Context, params []string) []string {
	var ps []string
	for _, p := range params {
		ps = append(ps, ctx.Param(p))
	}
	return ps
}

func ApiList(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//列出目录
		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		files, err := os.ReadDir(filename)
		if err != nil {
			curd.Error(ctx, err)
			return
		}

		var items []*attachInfo
		for _, stat := range files {
			info, _ := stat.Info()
			item := &attachInfo{
				Name:   info.Name(),
				Time:   info.ModTime(),
				Size:   info.Size(),
				Folder: info.IsDir(),
			}
			//识别类型
			if !stat.IsDir() {
				item.Mime = mime.TypeByExtension(filepath.Ext(stat.Name()))
			}
			items = append(items, item)
		}
		curd.OK(ctx, items)
	}
}

func ApiInfo(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//列出目录
		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		info, err := os.Stat(filename)
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		item := &attachInfo{
			Name:   info.Name(),
			Time:   info.ModTime(),
			Size:   info.Size(),
			Folder: info.IsDir(),
		}

		curd.OK(ctx, item)
	}
}

func ApiUpload(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		dir := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		_ = os.MkdirAll(dir, os.ModePerm) //创建目录

		form, err := ctx.MultipartForm()
		if err != nil {
			curd.Error(ctx, err)
			return
		}

		//接收所有附件
		for _, files := range form.File {
			for _, header := range files {
				filename := filepath.Join(dir, header.Filename)
				err := ctx.SaveUploadedFile(header, filename)
				if err != nil {
					curd.Error(ctx, err)
					return
				}
			}
		}

		curd.OK(ctx, nil)
	}
}

func ApiWrite(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		_ = os.MkdirAll(filepath.Dir(filename), os.ModePerm) //创建目录
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		defer f.Close()

		_, err = io.Copy(f, ctx.Request.Body)
		if err != nil {
			curd.Error(ctx, err)
			return
		}

		curd.OK(ctx, nil)
	}
}

func ApiRead(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		ctx.File(filename)
	}
}

func ApiDownload(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		ctx.Header("Content-Disposition", "attachment; filename="+ctx.Param("name"))
		http.ServeFile(ctx.Writer, ctx.Request, filename)
	}
}

func ApiRename(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var rename RenameBody
		err := ctx.ShouldBindJSON(&rename)
		if err != nil {
			curd.Error(ctx, err)
			return
		}

		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		newPath := filepath.Join(filepath.Dir(filename), rename.Name)

		err = os.Rename(filename, newPath)
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		curd.OK(ctx, nil)
	}
}

func ApiRemove(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		err := os.Remove(filename)
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		curd.OK(ctx, nil)
	}
}

func ApiMove(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var move MoveBody
		err := ctx.ShouldBindJSON(&move)
		if err != nil {
			curd.Error(ctx, err)
			return
		}

		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		newPath := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), move.Path, filepath.Base(filename))

		err = os.Rename(filename, newPath)
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		curd.OK(ctx, nil)
	}
}

func ApiMakeDir(root string, params ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filename := filepath.Join(viper.GetString("data"), root, filepath.Join(getParams(ctx, params)...), ctx.Param("name"))
		err := os.MkdirAll(filename, os.ModePerm)
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		curd.OK(ctx, nil)
	}
}

func Routers(group *gin.RouterGroup, root string, params ...string) {

	group.GET("/list/*name", ApiList(root, params...))

	group.GET("/info/*name", ApiInfo(root, params...))

	group.GET("/read/*name", ApiRead(root, params...))

	group.POST("/write/*name", ApiWrite(root, params...))

	group.POST("/upload/*name", ApiUpload(root, params...))

	group.GET("/download/*name", ApiDownload(root, params...))

	group.POST("/rename/*name", ApiRename(root, params...))

	group.GET("/remove/*name", ApiRemove(root, params...))

	group.POST("/move/*name", ApiMove(root, params...))

	group.GET("/mkdir/*name", ApiMakeDir(root, params...))
}
