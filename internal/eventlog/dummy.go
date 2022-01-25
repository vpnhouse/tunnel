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
