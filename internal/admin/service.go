package admin

import (
	"context"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/vpnhouse/common-lib-go/xcache"
	"github.com/vpnhouse/common-lib-go/xutils"

	"github.com/vpnhouse/tunnel/internal/storage"
	"github.com/vpnhouse/tunnel/internal/types"
)

type Handler interface {
	KillActiveUserSessions(userId string)
}

type Service struct {
	storage             *storage.Storage
	actionsCache        *lru.Cache[string, *types.ActionRule]
	usersToKillSessions *xcache.Cache
	lock                sync.Mutex
	handlers            []Handler
}

func New(storage *storage.Storage) (*Service, error) {
	actionsCache, err := lru.New[string, *types.ActionRule](1024)
	if err != nil {
		return nil, fmt.Errorf("failed to create lru cache for actions: %w", err)
	}

	s := &Service{
		storage:      storage,
		actionsCache: actionsCache,
	}

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
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create restricted users cache for actions: %w", err)
	}

	go s.run()

	return s, nil
}

func (s *Service) run() {
	ctx := context.Background()
	s.storage.CleanupExpiredActionRules(ctx)

	cleanupTicker := time.NewTicker(time.Hour)
	defer cleanupTicker.Stop()

	restrictUsersTicker := time.NewTicker(time.Minute)
	defer restrictUsersTicker.Stop()

	for {
		select {
		case <-cleanupTicker.C:
			s.storage.CleanupExpiredActionRules(ctx)
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
