package curd

import (
	"reflect"

	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/web"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

func GenerateUuidKey(data interface{}) error {
	value := reflect.ValueOf(data).Elem()
	field := value.FieldByName("id")
	//使用UUId作为Id
	//field.IsZero() 如果为空串时，生成UUID
	if field.Len() == 0 {
		field.SetString(uuid.NewString())
	}
	return nil
}

func GenerateXID[T any]() func(data *T) error {
	return func(data *T) error {
		value := reflect.ValueOf(data).Elem()
		field := value.FieldByName("Id")
		//使用UUId作为Id
		//field.IsZero() 如果为空串时，生成UUID
		if field.Len() == 0 {
			key := xid.New().String()
			field.SetString(key)
		}
		return nil
	}
}

func GenerateKSUID[T any]() func(data *T) error {
	return func(data *T) error {
		value := reflect.ValueOf(data).Elem()
		field := value.FieldByName("Id")
		//使用UUId作为Id
		//field.IsZero() 如果为空串时，生成UUID
		if field.Len() == 0 {
			key := ksuid.New().String()
			field.SetString(key)
		}
		return nil
	}
}

func GenerateID[T any]() func(data *T) error {
	return func(data *T) error {
		value := reflect.ValueOf(data).Elem()
		field := value.FieldByName("Id")
		//使用UUId作为Id
		//field.IsZero() 如果为空串时，生成UUID
		if field.Len() == 0 {
			var key string
			switch config.GetString(web.MODULE, "id") {
			case "uuid":
				key = uuid.NewString()
			case "shortuuid":
				key = shortuuid.New()
			case "ksuid":
				key = ksuid.New().String()
			case "xid":
				key = xid.New().String()
			default:
				key = xid.New().String()
			}
			field.SetString(key)
		}
		return nil
	}
}
