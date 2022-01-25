package eventlog

import (
	"context"

	"github.com/Codename-Uranium/tunnel/pkg/control"
)

type EventPusher interface {
	Push(eventType uint32, timestamp int64, data interface{}) error
}

type EventSubscriber interface {
	Subscribe(ctx context.Context, opts SubscriptionOpts) (*Subscription, error)
}

type EventManager interface {
	EventPusher
	EventSubscriber
	control.ServiceController
}
