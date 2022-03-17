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
	"time"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type eventManager struct {
	lock    sync.Mutex
	cancel  context.CancelFunc
	running bool
	storage *fsStorage

	// closes when shutdown is complete
	unblockDone chan struct{}

	// buffered chan for incoming events
	incoming chan []byte
	// subscribers track callers (see the Subscribe() method)
	subscribers map[*Subscription]struct{}
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
		subscribers: map[*Subscription]struct{}{},
		storage:     storage,
		cancel:      cancel,
		running:     true,
	}

	go m.run(ctx)
	return m, nil
}

// Push adds the event to the log.
func (em *eventManager) Push(eventType uint32, timestamp int64, data interface{}) error {
	em.lock.Lock()
	running := em.running
	em.lock.Unlock()

	if !running {
		return fmt.Errorf("service is not running")
	}

	// TODO(nikonov): actually, we CAN (un)marshal nil,
	//  so should it be considered as API misuse?
	if data == nil {
		return fmt.Errorf("cannot push nil event")
	}
	if timestamp == 0 {
		// TODO(nikonov): or UnixNano?
		timestamp = time.Now().Unix()
	}

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
	labels map[string]string
	cancel context.CancelFunc
	events chan Event
	errors chan error

	// notify closes when the subscriber done
	stopChan chan struct{}
}

func (s *Subscription) Events() <-chan Event {
	return s.events
}

func (s *Subscription) Errors() <-chan error {
	return s.errors
}

// Close the subscriber.
// Might be called multiple times.
func (s *Subscription) Close() <-chan struct{} {
	s.cancel()
	return s.stopChan
}

// SubscriptionOpts describe where we want to start reading the log.
// Zero offset means the beginning of the file.
// Empty logID means the beginning of the whole journal.
type SubscriptionOpts struct {
	LogID  string
	Offset int64
	Labels map[string]string
}

func (s SubscriptionOpts) validate() error {
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
func (em *eventManager) Subscribe(ctx context.Context, opts SubscriptionOpts) (*Subscription, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	em.lock.Lock()
	defer em.lock.Unlock()

	if !em.running {
		return nil, fmt.Errorf("manager is not running")
	}

	if len(opts.LogID) == 0 {
		opts.LogID = em.storage.FirstLog()
	} else if !em.storage.HasLog(opts.LogID) {
		return nil, fmt.Errorf("unknown log file `%s`", opts.LogID)
	}

	ctx, cancel := context.WithCancel(ctx)
	sub := &Subscription{
		cancel:   cancel,
		labels:   opts.Labels,
		events:   make(chan Event),
		errors:   make(chan error),
		stopChan: make(chan struct{}),
	}

	// add ourselves to the subscribers map,
	// will be removed in the goroutine right below
	em.subscribers[sub] = struct{}{}

	go func() {
		err := em.tail(ctx, opts.LogID, opts.Offset, sub.events)
		select {
		// dont block if nobody consumes the chan
		case sub.errors <- err:
		default:
		}

		// since tail exited we can be sure that nobody
		// operates the sub.events chan anymore.
		close(sub.events)
		close(sub.errors)

		// finally tell em we're done here
		close(sub.stopChan)

		em.lock.Lock()
		defer em.lock.Unlock()
		delete(em.subscribers, sub)
	}()

	return sub, nil
}

func (em *eventManager) Running() bool {
	em.lock.Lock()
	defer em.lock.Unlock()

	return em.running
}

func (em *eventManager) Shutdown() error {
	em.lock.Lock()

	if !em.running {
		return fmt.Errorf("manager is not running")
	}

	em.running = false
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
			em.running = false

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

	for sub := range em.subscribers {
		zap.L().Debug("closing the subscriber", zap.Any("labels", sub.labels))
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
func (em *eventManager) tail(ctx context.Context, logID string, offset int64, data chan<- Event) error {
	zap.L().Debug("start tailing", zap.String("log_id", logID), zap.Int64("offset", offset))

	var reader io.ReadCloser
	defer func() {
		if reader != nil {
			_ = reader.Close()
		}
	}()

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

			select {
			case <-ctx.Done():
				return ctx.Err()
			case data <- event:
			}
		}
	}
}
