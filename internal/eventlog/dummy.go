// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"context"

	"github.com/vpnhouse/tunnel/pkg/xerror"
)

func NewDummy() *dummyEventManager {
	return &dummyEventManager{}
}

type dummyEventManager struct {
}

func (d *dummyEventManager) Push(_ EventType, _ interface{}) error {
	return nil
}
func (d *dummyEventManager) Subscribe(ctx context.Context, subscriberId string, opts ...SubscribeOption) (*Subscription, error) {
	return nil, xerror.EInternalError("Attempt to receive events from dummy event manager", nil)
}

func (d *dummyEventManager) Running() bool {
	return false
}

func (d *dummyEventManager) Shutdown() error {
	return nil
}
