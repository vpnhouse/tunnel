package eventlog

import (
	"fmt"
	"sync"

	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
)

const (
	federationAuthHeader = "X-VPNHOUSE-FEDERATION-KEY"
	tunnelAuthHeader     = "X-VPNHOUSE-TUNNEL-KEY"
)

type Client struct {
	opts       options
	client     proto.EventLogServiceClient
	out        chan *Event
	once       sync.Once
	stop       chan struct{}
	done       chan struct{}
	offsetSync OffsetSync
	tunnelID   string
	instanceID string
}

func NewClient(instanceID string, offsetSync OffsetSync, opt ...Option) (*Client, error) {
	var opts options
	for _, o := range opt {
		err := o(&opts)
		if err != nil {
			return nil, err
		}
	}

	tunnelID := opts.TunnelID()
	if tunnelID == "" {
		return nil, fmt.Errorf("tunnel host is not defined")
	}

	if instanceID == "" {
		return nil, fmt.Errorf("instance id is not defined")
	}

	return &Client{
		opts:       opts,
		out:        make(chan *Event),
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
		tunnelID:   tunnelID,
		instanceID: instanceID,
		offsetSync: offsetSync,
	}, nil
}

func (s *Client) Events() chan *Event {
	s.once.Do(func() {
		go func() {
			defer func() {
				close(s.out)
				close(s.done)
			}()
			lockTimeout := s.getLockTtl()
			acquired, err := s.offsetSync.Acquire(s.instanceID, s.tunnelID, lockTimeout)
			if !acquired {
				s.publishOrDrop(&Event{Err: fmt.Errorf("stop reading events as failed to acquire lock to process events: %w", errLockNotAcquired)})
				zap.L().Info("stop reading events as failed to acquire lock to process events",
					zap.String("instance_id", s.instanceID),
					zap.String("tunnel_id", s.tunnelID),
					zap.Error(err),
				)
				return
			}
			err = s.connect()
			if err != nil {
				s.publishOrDrop(&Event{Err: err})
				return
			}
			s.readAndPublishEvents()
		}()
	})
	return s.out
}

func (s *Client) Close() {
	close(s.stop)
	<-s.done
}
