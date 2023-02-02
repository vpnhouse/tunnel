package eventlog

import (
	"sync"

	"github.com/vpnhouse/tunnel/proto"
)

const (
	federationAuthHeader = "X-VPNHOUSE-FEDERATION-KEY"
	tunnelAuthHeader     = "X-VPNHOUSE-TUNNEL-KEY"
)

type Client struct {
	opts   options
	client proto.EventLogServiceClient
	out    chan *Event
	once   sync.Once
	stop   chan struct{}
	done   chan struct{}
}

func NewClient(opt ...Option) (*Client, error) {
	var opts options
	for _, o := range opt {
		err := o(&opts)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		opts: opts,
		out:  make(chan *Event),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}, nil
}

func (s *Client) Read() chan *Event {
	s.once.Do(func() {
		go func() {
			defer func() {
				close(s.out)
				close(s.done)
			}()
			err := s.connect()
			if err != nil {
				s.publishOrDrop(&Event{Err: err})
				return
			}
			s.listenAndPublish()
		}()
	})
	return s.out
}

func (s *Client) Close() {
	close(s.stop)
	<-s.done
}
