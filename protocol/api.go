package protocol

import (
	"encoding/json"
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
	"path/filepath"
)

func init() {

	api.Register("GET", "iot/protocol/list", func(ctx *gin.Context) {
		var ps []*Base
		for _, item := range protocolsStore.Items {
			entries, err := item.ReadDir("")
			if err != nil {
				api.Error(ctx, err)
				return
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				ext := filepath.Ext(entry.Name())
				if ext == ".json" {
					buf, err := item.ReadFile(entry.Name())
					if err != nil {
						api.Error(ctx, err)
						return
					}
					var menu Base
					err = json.Unmarshal(buf, &menu)
					if err != nil {
						api.Error(ctx, err)
						return
					}

					ps = append(ps, &menu)
				}
			}
		}
		api.OK(ctx, ps)
	})

	api.Register("GET", "iot/protocol/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		ctx.FileFromFS(name+".json", &protocolsStore)
	})
}
