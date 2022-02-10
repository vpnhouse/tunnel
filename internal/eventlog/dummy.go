// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"context"

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
)

func NewDummy() *dummyEventManager {
	return &dummyEventManager{
		running: true,
	}
}

type dummyEventManager struct {
	running bool
}

func (d *dummyEventManager) Push(_ uint32, _ int64, _ interface{}) error {
	return nil
}
func (d *dummyEventManager) Subscribe(_ context.Context, _ SubscriptionOpts) (*Subscription, error) {
	return nil, xerror.EInternalError("Attempt to receive events from dummy event manager", nil)
}

func (d *dummyEventManager) Running() bool {
	return d.running
}

func (d *dummyEventManager) Shutdown() error {
	d.running = false
	return nil
}
