package proto

import (
	"time"
)

func (x *Timestamp) IntoTime() time.Time {
	return time.Unix(x.Sec, x.Nsec)
}

// TimestampFromTime converts given std time.Time into Timestamp
func TimestampFromTime(t time.Time) *Timestamp {
	v := t.UnixNano()
	return &Timestamp{
		Sec:  v / 1e9,
		Nsec: v % 1e9,
	}
}
