package eventlog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/vpnhouse/tunnel/pkg/human"
	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type positionAck struct {
	Position      Position
	ResetPosition bool
}

func (s *Client) readAndPublishEvents() {
	ctx, cancel := context.WithCancel(context.Background())

	fetchEventsClient, err := s.fetchEventsClient(ctx)
	if err != nil {
		cancel()
		s.publishOrDrop(&Event{Err: err})
		return
	}

	// Sending offsets is not intercepted by context cancel as we have to report the latest
	// offset on eventlog read cancelling
	eventFetchedClient, err := s.eventFetchedClient(context.Background())
	if err != nil {
		cancel()
		s.publishOrDrop(&Event{Err: err})
		return
	}

	done := make(chan struct{})
	positionAckChan := make(chan positionAck)

	var lastReadSec atomic.Uint64
	lastReadSec.Store(uint64(time.Now().Unix()))

	// Loop to read events from tunnel node
	go func() {
		defer func() {
			close(positionAckChan)
			close(done)
		}()
		for {
			evt, err := fetchEventsClient.Recv()

			if s.opts.StopIdleTimeout > 0 {
				lastReadSec.Store(uint64(time.Now().UTC().Unix()))
			}

			if err != nil {
				if status, ok := status.FromError(err); ok {
					switch status.Code() {
					case codes.Canceled:
						return
					case codes.NotFound:
						// log is odd and not found
						// so clear the currently saved position to be able start from active log
						zap.L().Info("log offset not found, reset odd position and exit", zap.Error(err))
						select {
						case <-time.After(reportOffsetTimeout * 2):
							s.publishOrDrop(&Event{Err: errors.New("cannot handle reset event position")})
							return
						case positionAckChan <- positionAck{ResetPosition: true}:
						}
						return
					}
				}
				s.publishOrDrop(&Event{Err: err})
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
			case <-time.After(reportOffsetTimeout * 2):
				s.publishOrDrop(&Event{Err: errors.New("cannot handle read event offset position")})
				return
			case positionAckChan <- positionAck{Position: offset}:
			}
		}
	}()

	// Loop to notify offset positions
	go func() {
		ticker := time.NewTicker(reportOffsetTimeout)
		var sentPos positionAck
		var currPos positionAck

		defer func() {
			ticker.Stop()
			err := eventFetchedClient.CloseSend()
			zap.L().Debug("send read event position stopped", zap.Error(err))
		}()

		for {
			select {
			case <-ticker.C:
				if sentPos == currPos {
					continue
				}

				pos := &proto.EventFetchedRequest{}
				if currPos.ResetPosition {
					pos.ResetEventlogPosition = currPos.ResetPosition
				} else {
					pos.Position = &proto.EventLogPosition{
						LogId:  currPos.Position.LogID,
						Offset: currPos.Position.Offset,
					}
				}

				err := eventFetchedClient.Send(pos)
				if err != nil {
					zap.L().Error("failed to send read event offset position", zap.Error(err))
					return
				}

				if currPos.ResetPosition {
					err = s.eventlogSync.DeletePosition(s.opts.TunnelID)
				} else {
					err = s.eventlogSync.PutPosition(s.opts.TunnelID, currPos.Position)
				}
				if err != nil {
					zap.L().Error("failed to keep store read event offset position", zap.Error(err))
					return
				}
				sentPos = currPos

			case newPos, ok := <-positionAckChan:
				if ok {
					currPos = newPos
					continue
				}

				// Report the latest offset back to the node prior exiting
				if sentPos == currPos {
					return
				}

				pos := &proto.EventFetchedRequest{}
				if currPos.ResetPosition {
					pos.ResetEventlogPosition = currPos.ResetPosition
				} else {
					pos.Position = &proto.EventLogPosition{
						LogId:  currPos.Position.LogID,
						Offset: currPos.Position.Offset,
					}
				}

				err := eventFetchedClient.Send(pos)
				if err != nil {
					zap.L().Error("failed to send read event offset position", zap.Error(err))
					return
				}

				if currPos.ResetPosition {
					err = s.eventlogSync.DeletePosition(s.opts.TunnelID)
				} else {
					err = s.eventlogSync.PutPosition(s.opts.TunnelID, currPos.Position)
				}
				if err != nil {
					zap.L().Error("failed to keep store read event offset position", zap.Error(err))
					return
				}

				return
			}
		}
	}()

	ticker := time.NewTicker(s.getProlongateLockTimeout())
	defer func() {
		ticker.Stop()
		err := s.eventlogSync.Release(s.instanceID, s.opts.TunnelID)
		if err != nil {
			zap.L().Error("failed to release sync lock to process events",
				zap.String("instance_id", s.instanceID),
				zap.String("tunnel", s.tunnelHost),
				zap.Error(err),
			)
		} else {
			zap.L().Info("release sync lock to process events",
				zap.String("instance_id", s.instanceID),
				zap.String("tunnel", s.tunnelHost),
			)
		}
	}()

	lockTimeout := s.getLockTtl()

	// min control loop
	for {
		select {
		case <-ticker.C:
			if s.opts.StopIdleTimeout > 0 && time.Unix(int64(lastReadSec.Load()), 0).Add(s.opts.StopIdleTimeout).Before(time.Now().UTC()) {
				cancel()
				zap.L().Info("stop reading events as timeout to wait events is exceeded", zap.Stringer("timeout", human.Interval(s.opts.StopIdleTimeout)))
			} else {
				// Prolongate lock
				acquired, err := s.eventlogSync.Acquire(s.instanceID, s.opts.TunnelID, lockTimeout)
				if !acquired {
					s.publishOrDrop(&Event{Err: fmt.Errorf("stop reading events as failed to extend lock to process events: %w", ErrLockNotAcquired)})
					cancel()
					zap.L().Info("stop reading events as failed to extend lock to process events",
						zap.String("instance_id", s.instanceID),
						zap.String("tunnel_id", s.opts.TunnelID),
						zap.Error(err),
					)
				} else {
					zap.L().Debug("extend sync lock",
						zap.String("instance_id", s.instanceID),
						zap.String("tunnel_id", s.opts.TunnelID),
						zap.Stringer("ttl", human.Interval(lockTimeout)),
					)
				}
			}
		case <-s.stop:
			cancel()
		case <-done:
			cancel()
			zap.L().Info("listen and publish events stopped")
			return
		}
	}
}

func (s *Client) publishOrDrop(event *Event) {
	select {
	case s.out <- event:
	default:
		if event.Err != nil {
			zap.L().Error("failed to publish event error", zap.Error(event.Err))
		}
	}
}

func (s *Client) publishOrError(event *Event) error {
	timer := time.NewTimer(waitOutputWriteTimeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		return ErrOutputEventStucked
	case s.out <- event:
		return nil
	}
}

func (s *Client) getLockTtl() time.Duration {
	return s.getProlongateLockTimeout() + lockTtl
}

func (s *Client) getProlongateLockTimeout() time.Duration {
	if s.opts.StopIdleTimeout > 0 {
		return s.opts.StopIdleTimeout
	}
	return defaultLockProlongateTimeout
}

func parseEvent(evt *proto.FetchEventsResponse) (*proto.PeerInfo, Position, error) {
	if evt == nil {
		return nil, Position{}, nil
	}

	var peerInfo proto.PeerInfo
	err := json.Unmarshal(evt.Data, &peerInfo)
	if err != nil {
		return nil, Position{}, fmt.Errorf("failed to parse peer info json data: %w", err)
	}

	offset := Position{
		LogID:  evt.GetPosition().GetLogId(),
		Offset: evt.GetPosition().GetOffset(),
	}

	return &peerInfo, offset, nil
}
