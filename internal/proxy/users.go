package proxy

import (
	"context"
	"sync"

	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type userStorage struct {
	lock    sync.Mutex
	maxConn int64
	users   map[string]*userInfo
}

type userInfo struct {
	limit *semaphore.Weighted
	usage int
}

func newUserStorage(maxConn int) *userStorage {
	return &userStorage{
		maxConn: int64(maxConn),
		users:   make(map[string]*userInfo),
	}
}

func (s *userStorage) take(id string) *userInfo {
	s.lock.Lock()
	defer s.lock.Unlock()

	user, loaded := s.users[id]
	if !loaded {
		user = &userInfo{
			limit: semaphore.NewWeighted(s.maxConn),
		}
		s.users[id] = user
	}

	user.usage += 1
	return user

}

func (s *userStorage) put(id string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	user, loaded := s.users[id]
	if !loaded {
		zap.L().Error("Can't put unknown user")
		return
	}

	user.usage -= 1
	if user.usage == 0 {
		delete(s.users, id)
	}
}

// TODO: Recover user limits

func (s *userStorage) acquire(ctx context.Context, id string) (*userInfo, error) {
	user := s.take(id)
	err := user.limit.Acquire(ctx, 1)
	if err != nil {
		s.put(id)
		return nil, xerror.EUnavailable("unavailable", err)
	}

	return user, nil
}

func (s *userStorage) release(id string, user *userInfo) {
	user.limit.Release(1)
	s.put(id)
}
