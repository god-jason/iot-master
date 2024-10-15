package db

import (
	"github.com/zgwit/iot-master/pkg/exception"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

func ParseObjectId(id any) (primitive.ObjectID, error) {
	switch val := id.(type) {
	case primitive.ObjectID:
		return val, nil
	case string:
		return primitive.ObjectIDFromHex(val)
	default:
		return primitive.NilObjectID, exception.New("invalid object id")
	}
}

func StringifyObjectId(id any) (string, error) {
	switch val := id.(type) {
	case primitive.ObjectID:
		return val.Hex(), nil
	case string:
		return val, nil
	default:
		return "", exception.New("invalid object id")
	}
}

func ParseDocumentObjectId(doc any) {
	switch val := doc.(type) {
	case map[string]any:
		for k, v := range val {
			if strings.HasSuffix(k, "_id") {
				if vv, ok := v.(string); ok {
					oid, err := primitive.ObjectIDFromHex(vv)
					if err == nil {
						val[k] = oid
					}
				}
				//vv, err := ParseObjectId(v) //todo 检查数组
				continue
			}
			ParseDocumentObjectId(v)
		}
	case bson.E:
		if strings.HasSuffix(val.Key, "_id") {
			oid, err := ParseObjectId(val.Value)
			if err == nil {
				val.Value = oid
			}
		}
	case bson.D:
		for _, v := range val {
			ParseDocumentObjectId(v)
		}
	case []any: //todo is slice,  reflect.TypeOf(i).Kind()==reflect.Slice
		for _, v := range val {
			ParseDocumentObjectId(v)
		}
	case []map[string]any: //todo is slice,  reflect.TypeOf(i).Kind()==reflect.Slice
		for _, v := range val {
			ParseDocumentObjectId(v)
		}
	}
}

func StringifyDocumentObjectId(doc any) {
	switch val := doc.(type) {
	case map[string]any:
		for k, v := range val {
			if strings.HasSuffix(k, "_id") {
				if vv, ok := v.(primitive.ObjectID); ok {
					val[k] = vv.Hex()
				}
				continue
			}
			StringifyDocumentObjectId(v)
		}
	case []any:
		for _, v := range val {
			StringifyDocumentObjectId(v)
		}
	case []map[string]any:
		for _, v := range val {
			StringifyDocumentObjectId(v)
		}
	}
}
