package types

import (
	"github.com/vpnhouse/tunnel/pkg/xtime"
)

type EventlogSubscriber struct {
	SubscriberID string      `db:"subscriber_id"`
	LogID        string      `db:"log_id"`
	Offset       int64       `db:"offset"`
	Updated      *xtime.Time `db:"updated"`
}
