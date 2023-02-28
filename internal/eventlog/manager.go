// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

var ErrServiceStopped = errors.New("service stopped")
var ErrNilEvent = errors.New("event is nil")

type eventManager struct {
	stopped atomic.Bool

	lock    sync.Mutex
	cancel  context.CancelFunc
	storage *fsStorage

	// closes when shutdown is complete
	unblockDone chan struct{}

	// buffered chan for incoming events
	incoming chan []byte
	// subscribers track callers (see the Subscribe() method)
	subscribers map[string]*Subscription
}

// New initializes and starts the event log manager
func New(cfg StorageConfig, fss ...afero.Fs) (*eventManager, error) {
	storage, err := newFsStorage(cfg, fss...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	m := &eventManager{
		incoming:    make(chan []byte, 100),
		unblockDone: make(chan struct{}),
		subscribers: map[string]*Subscription{},
		storage:     storage,
		cancel:      cancel,
	}

	go m.run(ctx)
	return m, nil
}

// Push adds the event to the log.
func (em *eventManager) Push(eventType EventType, data interface{}) error {

	if em.stopped.Load() {
		return ErrServiceStopped
	}

	// TODO(nikonov): actually, we CAN (un)marshal nil,
	//  so should it be considered as API misuse?
	if data == nil {
		return ErrNilEvent
	}

	// timestamp in seconds
	timestamp := time.Now().UTC().Unix()

	// marshal event here, this way we block the caller, not the
	// processEvent method, which, in turn, will block the whole queue.
	bs, err := marshalEvent(eventType, timestamp, data)
	if err != nil {
		return err
	}

	em.incoming <- bs
	return nil
}

type Subscription struct {
	subscriberID string
	cancel       context.CancelFunc
	events       chan Event

	// notify closes when the subscriber done
	stopChan chan struct{}
}

func (s *Subscription) Events() <-chan Event {
	return s.events
}

// Close the subscriber.
// Might be called multiple times.
func (s *Subscription) Close() <-chan struct{} {
	s.cancel()
	return s.stopChan
}

// EventlogPosition describe where we want to start reading the log.
// Zero offset means the beginning of the file.
// Empty logID means the beginning of the whole journal.
type EventlogPosition struct {
	LogID  string
	Offset int64
}

func (s *EventlogPosition) validate() error {
	if s.Offset < 0 {
		return fmt.Errorf("negative offset is not supported")
	}
	if len(s.LogID) == 0 && s.Offset > 0 {
		return fmt.Errorf("fileID if required for a given offset")
	}

	if len(s.LogID) > 0 {
		if _, err := uuid.Parse(s.LogID); err != nil {
			return fmt.Errorf("failed to parse given uuid: %v", err)
		}
	}
	return nil
}

// Subscribe allocates and starts new subscription to a log.
// The caller must consume channels given by .Events() and .Errors() methods.
// Context cancellation leads to subscription destruction, as well as calls of
// .Close() method.
func (em *eventManager) Subscribe(ctx context.Context, subscriberID string, opts ...SubscribeOption) (*Subscription, error) {
	var options subscribeOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	if em.stopped.Load() {
		return nil, ErrServiceStopped
	}

	em.lock.Lock()
	defer em.lock.Unlock()

	evenlogPosition := options.Position

	if options.ActiveLog {
		evenlogPosition = EventlogPosition{
			LogID:  em.storage.CurrentLog(),
			Offset: 0,
		}
	} else {
		if len(evenlogPosition.LogID) == 0 {
			evenlogPosition = EventlogPosition{
				LogID:  em.storage.FirstLog(),
				Offset: 0,
			}
		} else if !em.storage.HasLog(evenlogPosition.LogID) {
			return nil, fmt.Errorf("no such log %s: %w", evenlogPosition.LogID, ErrNotFound)
		}
	}

	if _, ok := em.subscribers[subscriberID]; ok {
		return nil, fmt.Errorf("failed to subscribe %s: %w", subscriberID, ErrAlreadySubscribed)
	}

	ctx, cancel := context.WithCancel(ctx)
	sub := &Subscription{
		cancel:       cancel,
		subscriberID: subscriberID,
		events:       make(chan Event),
		stopChan:     make(chan struct{}),
	}

	// add ourselves to the subscribers map,
	// will be removed in the goroutine right below
	em.subscribers[subscriberID] = sub

	go func() {
		err := em.tail(ctx, evenlogPosition, options.SkipEventAtPosition, sub)
		if err != nil {
			zap.L().Info("subscription stopped", zap.String("subscriber_id", sub.subscriberID), zap.Error(err))
		}
		em.deleteSubscription(sub)
	}()

	return sub, nil
}

// Unsubscribe by id
func (em *eventManager) Unsubscribe(ctx context.Context, subscriberID string) error {
	if em.stopped.Load() {
		return ErrServiceStopped
	}

	em.lock.Lock()
	sub, ok := em.subscribers[subscriberID]
	em.lock.Unlock()

	if !ok {
		return nil
	}

	<-sub.Close()
	em.deleteSubscription(sub)
	return nil
}

func (em *eventManager) deleteSubscription(sub *Subscription) {

	// since tail exited we can be sure that nobody
	// operates the sub.events chan anymore.
	close(sub.events)

	// finally tell em we're done here
	close(sub.stopChan)

	em.lock.Lock()
	defer em.lock.Unlock()
	delete(em.subscribers, sub.subscriberID)
}

func (em *eventManager) Running() bool {
	return !em.stopped.Load()
}

func (em *eventManager) Shutdown() error {

	if em.stopped.Load() {
		return fmt.Errorf("manager is not running")
	}

	em.lock.Lock()
	em.cancel()
	em.lock.Unlock()

	<-em.unblockDone
	return nil
}

func (em *eventManager) run(ctx context.Context) {
	for {
		select {
		case event := <-em.incoming:
			em.storeEvent(event)

		case <-ctx.Done():
			zap.L().Info("got termination signal")
			// stop receiving new messages
			em.stopped.Store(true)

			// sink the chan first, we have to store
			// remaining events before we go down.
		ffor:
			for {
				select {
				case event := <-em.incoming:
					em.storeEvent(event)
				default:
					//
					break ffor
				}
			}

			em.teardown()
			return
		}
	}
}

// storeEvent writes the event in the underlying file
func (em *eventManager) storeEvent(eventData []byte) {
	if err := em.storage.Write(eventData); err != nil {
		// TODO(nikonov): that's critical. What we gonna do?
	}
}

func (em *eventManager) teardown() {
	em.lock.Lock()
	defer em.lock.Unlock()

	for subscriberId, sub := range em.subscribers {
		zap.L().Debug("closing the subscriber", zap.String("subscriber_id", subscriberId))
		// wait for subscriber termination
		<-sub.Close()
	}

	// .Sync and close the log file(s)
	em.storage.Close()
	// notify that we're done.
	close(em.unblockDone)
}

// tail sequentially reads a given log at given offset, and all the following files, if any.
// tail does not validate given arguments expecting them to be verified by a caller.
func (em *eventManager) tail(ctx context.Context, eventlogPosition EventlogPosition, skipEventAtPosition bool, sub *Subscription) error {
	zap.L().Debug("start tailing",
		zap.String("log_id", eventlogPosition.LogID),
		zap.Int64("offset", eventlogPosition.Offset),
		zap.String("subscriber_id", sub.subscriberID),
	)

	var reader io.ReadCloser
	defer func() {
		if reader != nil {
			_ = reader.Close()
		}
	}()

	logID := eventlogPosition.LogID
	offset := eventlogPosition.Offset

	for {
		var err error
		reader, err = em.storage.OpenLog(logID, offset)
		if err != nil {
			return err
		}

		for {
			// check that we're still waiting for a data
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			event, nextOffset, err := readEvent(reader, offset, logID)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					// probably we shouldn't ever see this,
					// but anyway, better to be annoyed by this warning.
					return xerror.EInternalError("got non-EOF error while reading log", err,
						zap.String("log_id", logID),
						zap.Int64("offset", offset))
				}

				nextLog, err := em.storage.NextLog(logID)
				if err != nil {
					return xerror.EInternalError("failed to switch to a next log", err)
				}

				if nextLog != logID {
					// go reading from the next file
					zap.L().Debug("switching to a next log",
						zap.String("current", logID),
						zap.String("next", nextLog))

					logID = nextLog
					offset = 0
					_ = reader.Close()
					break
				}

				// in case of unintensive writing it may loop here forever
				// eating lots of CPU. Better wait for a bit before asking
				// for more data.
				time.Sleep(10 * time.Millisecond)
				continue
			}

			offset = nextOffset

			if skipEventAtPosition {
				// Skip event (once)
				skipEventAtPosition = false
				continue
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case sub.events <- event:
			}
		}
	}
}
