package db

import (
	"database/sql/driver"
	"time"
)

const localDateTimeFormat string = "2006-01-02 15:04:05"

type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(localDateTimeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, localDateTimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	now, err := time.ParseInLocation(`"`+localDateTimeFormat+`"`, string(b), time.Local)
	*t = Time(now)
	return err
}

func (t Time) String() string {
	return time.Time(t).Format(localDateTimeFormat)
}

func (t Time) Now() Time {
	return Time(time.Now())
}

func (t Time) ParseTime(tt time.Time) Time {
	return Time(tt)
}

func (t Time) format() string {
	return time.Time(t).Format(localDateTimeFormat)
}

func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.format()), nil
}

func (t *Time) FromDB(b []byte) error {
	if nil == b || len(b) == 0 {
		t = nil
		return nil
	}
	var now time.Time
	var err error
	now, err = time.ParseInLocation(localDateTimeFormat, string(b), time.Local)
	if nil == err {
		*t = Time(now)
		return nil
	}
	now, err = time.ParseInLocation("2006-01-02T15:04:05Z", string(b), time.Local)
	if nil == err {
		*t = Time(now)
		return nil
	}
	panic("自己定义个layout日期格式处理一下数据库里面的日期型数据解析!")
	return err
}

//func (t *Time) Scan(v interface{}) error {
// // Should be more strictly to check this type.
// vt, err := time.Parse("2006-01-02 15:04:05", string(v.([]byte)))
// if err != nil {
// return err
// }
// *t = Time(vt)
// return nil
//}

func (t *Time) ToDB() ([]byte, error) {
	if nil == t {
		return nil, nil
	}
	return []byte(time.Time(*t).Format(localDateTimeFormat)), nil
}

func (t *Time) Value() (driver.Value, error) {
	if nil == t {
		return nil, nil
	}
	return time.Time(*t).Format(localDateTimeFormat), nil
}
