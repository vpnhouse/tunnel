package eventlog

import (
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())

	fetchEventsClient, err := s.fetchEventsClient(ctx)
	if err != nil {
		cancel()
		s.publishOrDrop(&Event{Err: err})
		return
	}

	eventFetchedClient, err := s.eventFetchedClient(ctx)
	if err != nil {
		cancel()
		s.publishOrDrop(&Event{Err: err})
		return
	}

	done := make(chan struct{})
	offsetChan := make(chan Offset)

	var numReadAttempts atomic.Int32

	go func() {
		defer func() {
			close(offsetChan)
			close(done)
		}()
		for {
			evt, err := fetchEventsClient.Recv()
			numReadAttempts.Store(0)
			if err != nil {
				if status, ok := status.FromError(err); !ok || status.Code() != codes.Canceled {
					s.publishOrDrop(&Event{Err: err})
				}
				return
			}

			peerInfo, offset, err := parseEvent(evt)
			zap.L().Debug("event", zap.Any("peer_info", peerInfo), zap.Any("offset", offset), zap.Error(err))

			err = s.publishOrError(&Event{PeerInfo: peerInfo, Err: err})
			if err != nil {
				zap.L().Error("failed to publish event", zap.Error(err))
				return
			}

			select {
			case <-time.After(time.Second):
				s.publishOrDrop(&Event{Err: errors.New("timeout to send read event position offset to tunnel node")})
				return
			case offsetChan <- offset:
			}
		}
	}()

	// Loop to notify offset positions
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		var sentOffset Offset
		var currOffset Offset

		for {
			select {
			case <-ticker.C:
				if sentOffset != currOffset {
					err := eventFetchedClient.Send(&proto.EventFetchedRequest{Position: &proto.EventLogPosition{
						LogId:  currOffset.LogID,
						Offset: currOffset.Offset,
					}})
					if err != nil {
						zap.L().Error("failed to send read event offset position", zap.Error(err))
						return
					}

					sentOffset = currOffset

					err = s.offsetSync.PutOffset(s.tunnelID, sentOffset)
					if err != nil {
						zap.L().Error("failed to keep store read event offset position", zap.Error(err))
						return
					}
				}
			case newOffset, ok := <-offsetChan:
				if !ok {
					zap.L().Error("send read event position intercepted", zap.Error(err))
					err := eventFetchedClient.CloseSend()
					if err != nil {
						zap.L().Error("failed to stop send read event offset position", zap.Error(err))
					}
					return
				}
				currOffset = newOffset
			}
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
				zap.L().Debug(
					"exceed number of attempts to wait events, trigger to stop reading events",
					zap.Int32("attempt", attempt),
					zap.Int32("max_attempts", maxInputReadAttempts),
				)
			} else {
				zap.L().Debug("waiting event", zap.Int32("attempt", attempt))
			}
		case <-s.stop:
			cancel()
			zap.L().Debug("trigger to stop reading events")
		case <-done:
			cancel()
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
