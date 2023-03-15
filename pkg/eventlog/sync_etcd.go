package eventlog

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	etcdTimeout = 5 * time.Second
)

type eventlogSyncEtcd struct {
	client *clientv3.Client
	kv     clientv3.KV
}

func NewEventlogSyncEtcd(client *clientv3.Client) (*eventlogSyncEtcd, error) {
	kv := clientv3.NewKV(client)
	return &eventlogSyncEtcd{
		client: client,
		kv:     kv,
	}, nil
}

func (s *eventlogSyncEtcd) Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error) {
	key := buildSyncKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	defer cancel()

	lease, err := s.client.Grant(ctx, int64(ttl.Seconds()))
	if err != nil {
		return false, fmt.Errorf("aquire job lock key %s failed: %w", key, err)
	}

	opPut := clientv3.OpPut(key, instanceID, clientv3.WithLease(lease.ID))

	isEqual := clientv3.Compare(clientv3.Value(key), "=", instanceID)
	resp, err := s.client.Txn(ctx).If(isEqual).Then(opPut).Commit()
	if err != nil {
		return false, fmt.Errorf("aquire job lock key %s failed: %w", key, err)
	}

	if !resp.Succeeded {
		isNotExist := clientv3.Compare(clientv3.ModRevision(key), "=", 0)
		resp, err = s.client.Txn(ctx).If(isNotExist).Then(opPut).Commit()
		if err != nil {
			return false, fmt.Errorf("aquire job lock key %s failed: %w", key, err)
		}
	}

	return resp.Succeeded, nil
}

func (s *eventlogSyncEtcd) Release(instanceID string, tunnelID string) error {
	key := buildSyncKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	defer cancel()

	opDel := clientv3.OpDelete(key)
	isEqual := clientv3.Compare(clientv3.Value(key), "=", instanceID)
	txn := s.client.Txn(ctx)
	txn = txn.If(isEqual).Then(opDel)

	_, err := txn.Commit()

	if err != nil {
		return fmt.Errorf("release job lock key %s failed: %w", key, err)
	}

	return nil
}

func (s *eventlogSyncEtcd) GetPosition(tunnelID string) (Position, error) {
	key := buildPositionKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	defer cancel()

	resp, err := s.kv.Get(ctx, key)
	if err != nil {
		return Position{}, fmt.Errorf("failed to read and parse offset data: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return Position{}, ErrPositionNotFound
	}

	return positionFromJsonBytes(resp.Kvs[0].Value)
}

func (s *eventlogSyncEtcd) PutPosition(tunnelId string, position Position) error {
	key := buildPositionKey(tunnelId)
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	defer cancel()

	val := position.ToJson()
	lease, err := s.client.Grant(ctx, int64(offsetKeepTimeout.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to grant ttl lease offset data: %w", err)
	}
	_, err = s.kv.Put(ctx, key, val, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to read and parse offset data: %w", err)
	}
	return nil
}

func (s *eventlogSyncEtcd) DeletePosition(tunnelId string) error {
	key := buildPositionKey(tunnelId)
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	defer cancel()

	_, err := s.kv.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete offset data: %w", err)
	}

	return nil
}
