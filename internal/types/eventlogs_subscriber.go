package types

import (
	"github.com/vpnhouse/common-lib-go/xtime"
)

type EventlogSubscriber struct {
	SubscriberID string      `db:"subscriber_id"`
	LogID        string      `db:"log_id"`
	Offset       int64       `db:"offset"`
	Updated      *xtime.Time `db:"updated"`
}
