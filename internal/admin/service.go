package admin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vpnhouse/common-lib-go/xcache"
	"github.com/vpnhouse/common-lib-go/xutils"

	"github.com/vpnhouse/tunnel/internal/storage"
)

type Handler interface {
	KillActiveUserSessions(userId string)
}

type Service struct {
	storage             *storage.Storage
	actionsCache        *xcache.Cache
	usersToKillSessions *xcache.Cache
	lock                sync.Mutex
	handlers            []Handler
}

func New(storage *storage.Storage) (*Service, error) {
	s := &Service{
		storage: storage,
	}

	var err error
	s.usersToKillSessions, err = xcache.New(
		32<<20, // 32 Mb
		func(items *xcache.Items) {
			// This is triggered periodically by Reset() call (see run() method)
			// Also can be called once the cache got full and start internal cleaning
			s.lock.Lock()
			handlers := s.handlers
			s.lock.Unlock()
			for i := range items.Keys {
				for _, h := range handlers {
					h.KillActiveUserSessions(xutils.BytesToString(items.Keys[i]))
				}
			}
		},
	)

	s.actionsCache, err = xcache.New(
		32<<20, // 32 Mb
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create restricted users cache for actions: %w", err)
	}

	go s.run()

	return s, nil
}

func (s *Service) run() {
	ctx := context.Background()
	s.storage.CleanupExpiredActionRules(ctx)

	cleanupExpiredTicker := time.NewTicker(time.Hour)
	defer cleanupExpiredTicker.Stop()

	cleanupCacheTicker := time.NewTicker(10 * time.Second)
	defer cleanupCacheTicker.Stop()

	restrictUsersTicker := time.NewTicker(time.Minute)
	defer restrictUsersTicker.Stop()

	for {
		select {
		case <-cleanupExpiredTicker.C:
			numCleaned := s.storage.CleanupExpiredActionRules(ctx)
			if numCleaned > 0 {
				s.actionsCache.Reset()
			}
		case <-cleanupCacheTicker.C:
			s.actionsCache.Reset()
		case <-restrictUsersTicker.C:
			s.usersToKillSessions.Reset()
		}
	}
}

func (s *Service) AddHandler(handler Handler) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.handlers = append(s.handlers, handler)
}
