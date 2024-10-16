package xlimits

import (
	"context"
	"sync"

	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type Blocker struct {
	lock      sync.Mutex
	maxConn   int64
	consumers map[string]*consumer
}

type consumer struct {
	limit *semaphore.Weighted
	usage int
}

func NewBlocker(maxConn int) *Blocker {
	return &Blocker{
		maxConn:   int64(maxConn),
		consumers: make(map[string]*consumer),
	}
}

func (s *Blocker) take(id string) *consumer {
	s.lock.Lock()
	defer s.lock.Unlock()

	c, loaded := s.consumers[id]
	if !loaded {
		c = &consumer{
			limit: semaphore.NewWeighted(s.maxConn),
		}
		s.consumers[id] = c
	}

	c.usage += 1
	return c

}

func (s *Blocker) put(id string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	c, loaded := s.consumers[id]
	if !loaded {
		zap.L().Error("Can't put unknown consumer")
		return
	}

	c.usage -= 1
	if c.usage == 0 {
		delete(s.consumers, id)
	}
}

func (s *Blocker) Acquire(ctx context.Context, id string) (*consumer, error) {
	c := s.take(id)
	err := c.limit.Acquire(ctx, 1)
	if err != nil {
		s.put(id)
		return nil, xerror.EUnavailable("unavailable", err)
	}

	return c, nil
}

func (s *Blocker) Release(id string, c *consumer) {
	c.limit.Release(1)
	s.put(id)
}
