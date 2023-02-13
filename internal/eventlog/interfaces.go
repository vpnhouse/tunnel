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
