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

func (instance *userStorage) take(id string) *userInfo {
	instance.lock.Lock()
	defer instance.lock.Unlock()

	user, loaded := instance.users[id]
	if !loaded {
		user = &userInfo{
			limit: semaphore.NewWeighted(instance.maxConn),
		}
		instance.users[id] = user
	}

	user.usage += 1
	return user

}

func (instance *userStorage) put(id string) {
	instance.lock.Lock()
	defer instance.lock.Unlock()

	user, loaded := instance.users[id]
	if !loaded {
		zap.L().Error("Can't put unknown user")
		return
	}

	user.usage -= 1
	if user.usage == 0 {
		delete(instance.users, id)
	}
}

// TODO: Recover user limits

func (instance *userStorage) acquire(ctx context.Context, id string) (*userInfo, error) {
	user := instance.take(id)
	err := user.limit.Acquire(ctx, 1)
	if err != nil {
		instance.put(id)
		return nil, xerror.EUnavailable("unavailable", err)
	}

	return user, nil
}

func (instance *userStorage) release(id string, user *userInfo) {
	user.limit.Release(1)
	instance.put(id)
}
