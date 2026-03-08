package smart

import (
	"github.com/spf13/cast"
	"xorm.io/xorm/schemas"
)

func TypeToSqlType(t string) (st schemas.SQLType) {
	switch t {
	case "int", "int8", "int16", "int32":
		st = schemas.SQLType{schemas.Int, 0, 0}
	case "uint", "uint8", "uint16", "uint32":
		st = schemas.SQLType{schemas.UnsignedInt, 0, 0}
	case "int64":
		st = schemas.SQLType{schemas.BigInt, 0, 0}
	case "uint64":
		st = schemas.SQLType{schemas.UnsignedBigInt, 0, 0}
	case "float32", "float":
		st = schemas.SQLType{schemas.Float, 0, 0}
	case "float64", "double":
		st = schemas.SQLType{schemas.Double, 0, 0}
	case "decimal":
		st = schemas.SQLType{schemas.Decimal, 0, 0}
	case "complex64", "complex128":
		st = schemas.SQLType{schemas.Varchar, 64, 0}
	case "blob":
		st = schemas.SQLType{schemas.Blob, 0, 0}
	case "text":
		st = schemas.SQLType{schemas.Text, 0, 0}
	case "longtext":
		st = schemas.SQLType{schemas.LongText, 0, 0}
	case "bool", "boolean":
		st = schemas.SQLType{schemas.Bool, 0, 0}
	case "string":
		st = schemas.SQLType{schemas.Varchar, 255, 0}
	case "datetime":
		st = schemas.SQLType{schemas.DateTime, 0, 0}
	case "timestamp":
		st = schemas.SQLType{schemas.TimeStamp, 0, 0}
	default:
		st = schemas.SQLType{schemas.Text, 0, 0}
	}
	return
}

// Column 数据库列定义，放这里有点怪，但是又没办法
type Column struct {
	Name      string `json:"name,omitempty"`
	Comment   string `json:"comment,omitempty"`
	Type      string `json:"type,omitempty"`
	Default   string `json:"default,omitempty"`
	NotNull   bool   `json:"not_null,omitempty"`
	Length    int64  `json:"length,omitempty"`
	Length2   int64  `json:"length2,omitempty"`
	Primary   bool   `json:"primary,omitempty"`
	Increment bool   `json:"increment,omitempty"`
	Indexed   bool   `json:"indexed,omitempty"`
	Created   bool   `json:"created,omitempty"`
	Updated   bool   `json:"updated,omitempty"`
	Json      bool   `json:"json,omitempty"`
}

func (f *Column) ToColumn() *schemas.Column {
	col := schemas.NewColumn(f.Name, "", TypeToSqlType(f.Type), f.Length, f.Length2, !f.NotNull)
	col.IsPrimaryKey = f.Primary
	col.IsAutoIncrement = f.Increment
	col.Default = f.Default
	col.IsCreated = f.Created
	col.IsUpdated = f.Updated
	if f.Indexed {
		col.Indexes[f.Name] = schemas.IndexType
	}
	if f.Primary {
		col.Nullable = false //主键不能为空
	}
	col.Comment = f.Comment
	return col
}

func (f *Column) Cast(v any) (ret any, err error) {
	switch f.Type {
	case "int", "int8", "int16", "int32", "int64":
		ret, err = cast.ToInt64E(v)
	case "uint", "uint8", "uint16", "uint32", "uint64":
		ret, err = cast.ToUint64E(v)
	case "float32", "float64", "float", "double":
		ret, err = cast.ToFloat64E(v)
	default: //string
		ret = v
	}
	return
}
