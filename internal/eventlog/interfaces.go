// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"context"
	"errors"

	"github.com/vpnhouse/tunnel/pkg/control"
)

var ErrNotFound = errors.New("not found")
var ErrSubscriptionStopped = errors.New("subscription stopped")
var ErrAlreadySubscribed = errors.New("already subscribed")

type EventPusher interface {
	Push(eventType EventType, data interface{}) error
}

type EventSubscriber interface {
	Subscribe(ctx context.Context, subscriberId string, opts ...SubscribeOption) (*Subscription, error)
	Unsubscribe(ctx context.Context, subscriberId string) error
}

type EventManager interface {
	EventPusher
	EventSubscriber
	control.ServiceController
}
