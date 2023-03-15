package eventlog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisTimeout = 5 * time.Second
)

type eventlogSyncRedis struct {
	redisClient redis.UniversalClient
}

func NewEventlogSyncRedis(redisClient redis.UniversalClient) (*eventlogSyncRedis, error) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("failed connect to redis: %w", err)
	}
	return &eventlogSyncRedis{
		redisClient: redisClient,
	}, nil
}

func (s *eventlogSyncRedis) Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error) {
	key := buildSyncKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	acquired, err := s.redisClient.SetNX(ctx, key, instanceID, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("aquire job lock key %s failed: %w", key, err)
	}

	if acquired {
		return true, nil
	}

	value, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Potentially key can disappear so that assume we still not acquire the lock
			// for the sake of simplicity
			return false, nil
		}
		return false, fmt.Errorf("aquire job lock key value %s failed: %w", key, err)
	}

	if value == instanceID {
		err = s.redisClient.Set(ctx, key, instanceID, ttl).Err()
		if err != nil {
			return false, fmt.Errorf("extend job lock key ttl %s failed: %w", key, err)
		}
		acquired = true
	}

	return acquired, nil
}

func (s *eventlogSyncRedis) Release(instanceID string, tunnelID string) error {
	key := buildSyncKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	value, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("release job lock key %s failed: %w", key, err)
	}

	if value != instanceID {
		return nil
	}

	_, err = s.redisClient.Del(ctx, key).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("release job lock key %s failed: %w", key, err)
	}

	return nil
}

func (s *eventlogSyncRedis) GetPosition(tunnelID string) (Position, error) {
	key := buildPositionKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	res, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Position{}, ErrPositionNotFound
		}
		return Position{}, fmt.Errorf("failed to read and parse position data: %w", err)
	}

	return positionFromJson(res)
}

func (s *eventlogSyncRedis) PutPosition(tunnelID string, position Position) error {
	key := buildPositionKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	val := position.ToJson()
	// Never be expired
	err := s.redisClient.Set(ctx, key, val, offsetKeepTimeout).Err()
	if err != nil {
		return fmt.Errorf("failed to store position data: %w", err)
	}
	return nil
}

func (s *eventlogSyncRedis) DeletePosition(tunnelID string) error {
	key := buildPositionKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	err := s.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete offset data: %w", err)
	}
	return nil
}
