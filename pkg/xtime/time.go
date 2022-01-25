package xtime

import (
	"database/sql/driver"
	"time"
)

type Time struct {
	Time time.Time
}

func (lt *Time) Scan(src interface{}) error {
	tm := time.Unix(src.(int64), 0)
	lt.Time = tm
	return nil
}

func (lt *Time) Value() (driver.Value, error) {
	if lt == nil {
		return nil, nil
	}
	return driver.Value(lt.Time.Unix()), nil
}
func Now() Time {
	return Time{time.Now()}
}

func FromTimePtr(t *time.Time) *Time {
	if t == nil {
		return nil
	}
	return &Time{*t}
}

func (lt *Time) TimePtr() *time.Time {
	if lt == nil {
		return nil
	}

	return &lt.Time
}
