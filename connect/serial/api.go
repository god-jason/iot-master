package serial

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/bucket/log"
	"github.com/god-jason/iot-master/api"
	"github.com/god-jason/iot-master/curd"
	"github.com/god-jason/iot-master/db"
	"go.bug.st/serial"
)

func init() {

	api.Register("POST", "/serial/count", curd.ApiCount[Serial]())

	api.Register("POST", "/serial/search", curd.ApiSearchHook[Serial](func(serials []*Serial) error {
		for k, ser := range serials {
			c := GetSerial(ser.Id)
			if c != nil {
				serials[k].Status = c.Status
			}
		}
		return nil
	}))

	api.Register("GET", "/serial/list", curd.ApiList[Serial]())

	api.Register("POST", "/serial/create", curd.ApiCreateHook[Serial](curd.GenerateID[Serial](), func(value *Serial) error {
		return LoadSerial(value)
	}))

	api.Register("GET", "/serial/:id", curd.ParseParamStringId, curd.ApiGetHook[Serial](func(ser *Serial) error {
		c := GetSerial(ser.Id)
		if c != nil {
			ser.Status = c.Status
		}
		return nil
	}))

	api.Register("POST", "/serial/:id", curd.ParseParamStringId, curd.ApiUpdateHook[Serial](nil, func(value *Serial) error {
		c := GetSerial(value.Id)
		err := c.Close()
		if err != nil {
			log.Error(err)
		}
		return LoadSerial(value)
	}))

	api.Register("GET", "/serial/:id/delete", curd.ParseParamStringId, curd.ApiDeleteHook[Serial](nil, func(value *Serial) error {
		c := GetSerial(value.Id)
		if c != nil {
			serials.Delete(value.Id)
			return c.Close()
		}
		return nil
	}))

	api.Register("GET", "/serial/:id/disable", curd.ParseParamStringId, curd.ApiDisableHook[Serial](true, nil, func(value interface{}) error {
		id := value.(string)
		c := GetSerial(id)
		return c.Close()
	}))

	api.Register("GET", "/serial/:id/enable", curd.ParseParamStringId, curd.ApiDisableHook[Serial](false, nil, func(value interface{}) error {
		id := value.(string)
		var m Serial
		has, err := db.Engine.ID(id).Get(&m)
		if err != nil {
			return err
		}
		if !has {
			return fmt.Errorf("找不到 %s", id)
		}
		return LoadSerial(&m)
	}))

	api.Register("GET", "/serial/:id/open", curd.ParseParamStringId, func(ctx *gin.Context) {
		id := ctx.GetString("id")
		c := GetSerial(id)
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

	api.Register("GET", "/serial/:id/close", curd.ParseParamStringId, func(ctx *gin.Context) {
		id := ctx.GetString("id")
		c := GetSerial(id)
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

	api.Register("GET", "/serial/ports", func(ctx *gin.Context) {
		list, err := serial.GetPortsList()
		if err != nil {
			curd.Error(ctx, err)
			return
		}
		curd.OK(ctx, list)
	})

}
