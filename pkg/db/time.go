package db

import (
	"database/sql/driver"
	"time"
)

const localDateTimeFormat string = "2006-01-02 15:04:05"

// Time 本地时间类型，格式 YYYY-MM-DD HH:mm:ss
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
	return err
}

func (t *Time) ToDB() ([]byte, error) {
	if t == nil {
		return nil, nil
	}
	return []byte(time.Time(*t).Format(localDateTimeFormat)), nil
}

func (t *Time) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return time.Time(*t).Format(localDateTimeFormat), nil
}
