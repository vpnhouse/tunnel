// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"context"

	"github.com/comradevpn/tunnel/pkg/control"
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
