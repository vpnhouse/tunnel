package proxy

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type userStorage struct {
	ctx     context.Context
	lock    sync.Mutex
	maxConn int64
	users   map[string]*userEntry
}

type userEntry struct {
	storage *userStorage
	limit   *semaphore.Weighted
	conns   int
}

func newUserStorage(ctx context.Context, maxConn int) *userStorage {
	return &userStorage{
		ctx:     ctx,
		maxConn: int64(maxConn),
		users:   make(map[string]*userEntry),
	}
}

func (instance *userStorage) acquire(id string) error {
	instance.lock.Lock()
	user, ok := instance.users[id]

	if !ok {
		user = &userEntry{
			storage: instance,
			limit:   semaphore.NewWeighted(instance.maxConn),
		}
		instance.users[id] = user
		zap.L().Debug("Created user entry", zap.String("id", id))
	} else {
		zap.L().Debug("Using existing user entry", zap.String("id", id))
	}

	user.conns += 1
	instance.lock.Unlock()
	err := user.limit.Acquire(instance.ctx, 1)
	instance.lock.Lock()
	defer instance.lock.Unlock()

	if err != nil {
		user.conns -= 1
		if user.conns < 0 {
			zap.L().Error("Negative connections", zap.String("id", id))
		}
		if user.conns <= 0 {
			delete(instance.users, id)
		}
		return err
	}

	zap.L().Debug("Acquired limit", zap.String("id", id))
	return nil
}

func (instance *userStorage) release(id string) {
	instance.lock.Lock()
	defer instance.lock.Unlock()

	user, ok := instance.users[id]
	if !ok {
		zap.L().Error("Release for unknown user", zap.String("id", id))
		return
	}

	user.limit.Release(1)
	zap.L().Debug("Released limit", zap.String("id", id))

	user.conns -= 1
	if user.conns < 0 {
		zap.L().Error("Negative connections", zap.String("id", id))
	}
	if user.conns <= 0 {
		delete(instance.users, id)
	}
}
