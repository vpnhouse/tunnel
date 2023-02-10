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

type offsetSyncRedis struct {
	redisClient redis.UniversalClient
}

func NewOffsetSyncRedis(redisClient redis.UniversalClient) (*offsetSyncRedis, error) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("failed connect to redis: %w", err)
	}
	return &offsetSyncRedis{
		redisClient: redisClient,
	}, nil
}

func (s *offsetSyncRedis) Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error) {
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

func (s *offsetSyncRedis) Release(instanceID string, tunnelID string) error {
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

func (s *offsetSyncRedis) GetOffset(tunnelID string) (Offset, error) {
	key := buildOffsetKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	res, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Offset{TunnelID: tunnelID}, nil
		}
		return Offset{}, fmt.Errorf("failed to read and parse offset data: %w", err)
	}

	return offsetFromJson(res)
}

func (s *offsetSyncRedis) PutOffset(offset Offset) error {
	key := buildOffsetKey(offset.TunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	val := offset.ToJson()
	// Never be expired
	err := s.redisClient.Set(ctx, key, val, offsetKeepTimeout).Err()
	if err != nil {
		return fmt.Errorf("failed to store offset data: %w", err)
	}
	return nil
}
