package types

import (
	"time"
)

const localDateTimeFormat string = "2006-01-02 15:04:05"

type Time time.Time

func (t *Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(localDateTimeFormat)+2)
	b = append(b, '"')
	b = time.Time(*t).AppendFormat(b, localDateTimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	now, err := time.ParseInLocation(`"`+localDateTimeFormat+`"`, string(b), time.Local)
	*t = Time(now)
	return err
}

func (t *Time) String() string {
	return time.Time(*t).Format(localDateTimeFormat)
}

func (t *Time) Now() Time {
	return Time(time.Now())
}

func (t *Time) ParseTime(tm time.Time) Time {
	return Time(tm)
}

func (t *Time) format() string {
	return time.Time(*t).Format(localDateTimeFormat)
}

func (t *Time) MarshalText() ([]byte, error) {
	return []byte(t.format()), nil
}
