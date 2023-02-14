package eventlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	waitOutputWriteTimeout      = time.Second * 5
	waitInputReadAttemptTimeout = time.Second * 10
	maxInputReadAttempts        = 5
)

var errOutputEventStucked = errors.New("output event stucked")

func (s *Client) readAndPublishEvents() {

	fetchEventsClient, cancel, err := s.fetchEventsClient()

	if err != nil {
		s.publishOrDrop(&Event{PeerInfo: nil, Err: err})
		return
	}

	var numReadAttempts atomic.Int32
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			evt, err := fetchEventsClient.Recv()
			numReadAttempts.Store(0)
			if err != nil {
				if status, ok := status.FromError(err); !ok || status.Code() != codes.Canceled {
					s.publishOrDrop(&Event{PeerInfo: nil, Err: err})
				}
				return
			}

			peerInfo, offset, err := parseEvent(evt)
			zap.L().Debug("event", zap.Any("peer_info", peerInfo), zap.Any("offset", offset), zap.Error(err))

			s.publishOrError(&Event{PeerInfo: peerInfo, Err: err})
		}
	}()

	ticker := time.NewTicker(waitInputReadAttemptTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			attempt := numReadAttempts.Add(1)
			if attempt >= maxInputReadAttempts {
				cancel()
			}
			zap.L().Debug("waiting event", zap.Int32("attempt", attempt))
		case <-s.stop:
			cancel()
		case <-done:
			zap.L().Info("listen and publish events stopped")
			return
		}
	}
}

func (s *Client) publishOrDrop(event *Event) {
	timer := time.NewTimer(waitOutputWriteTimeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		if event.Err != nil {
			zap.L().Error("error publish event", zap.Error(event.Err))
		}
	case s.out <- event:
	}
}

func (s *Client) publishOrError(event *Event) error {
	timer := time.NewTimer(waitOutputWriteTimeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		return errOutputEventStucked
	case s.out <- event:
		return nil
	}
}

func parseEvent(evt *proto.FetchEventsResponse) (*proto.PeerInfo, Offset, error) {
	if evt == nil {
		return nil, Offset{}, nil
	}

	var peerInfo proto.PeerInfo
	err := json.Unmarshal(evt.Data, &peerInfo)
	if err != nil {
		return nil, Offset{}, fmt.Errorf("failed to parse peer info json data: %w", err)
	}

	offset := Offset{
		LogID:  evt.GetPosition().GetLogId(),
		Offset: evt.GetPosition().GetOffset(),
	}

	return &peerInfo, offset, nil
}
