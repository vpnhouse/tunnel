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

type offsetSyncEtcd struct {
	client *clientv3.Client
	kv     clientv3.KV
}

func NewOffsetSyncEtcd(client *clientv3.Client) (*offsetSyncEtcd, error) {
	kv := clientv3.NewKV(client)
	return &offsetSyncEtcd{
		client: client,
		kv:     kv,
	}, nil
}

func (s *offsetSyncEtcd) Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error) {
	key := buildSyncKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	defer cancel()

	lease, _ := s.client.Grant(ctx, int64(ttl.Seconds()))
	isEqual := clientv3.Compare(clientv3.Value(key), "=", instanceID)
	isNotExist := clientv3.Compare(clientv3.ModRevision(key), "=", 0)
	opPut := clientv3.OpPut(key, instanceID, clientv3.WithLease(lease.ID))

	txn := s.client.Txn(ctx)
	txn = txn.If(isEqual).Then(opPut)
	txn = txn.If(isNotExist).Then(opPut)

	resp, err := txn.Commit()
	if err != nil {
		return false, fmt.Errorf("aquire job lock key %s failed: %w", key, err)
	}
	return resp.Succeeded, nil
}

func (s *offsetSyncEtcd) Release(instanceID string, tunnelID string) error {
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

func (s *offsetSyncEtcd) GetOffset(tunnelID string) (Offset, error) {
	key := buildOffsetKey(tunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	defer cancel()

	resp, err := s.kv.Get(ctx, key)
	if err != nil {
		return Offset{}, fmt.Errorf("failed to read and parse offset data: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return Offset{}, nil
	}

	return offsetFromJsonBytes(resp.Kvs[0].Value)
}

func (s *offsetSyncEtcd) PutOffset(offset Offset) error {
	key := buildOffsetKey(offset.TunnelID)
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	defer cancel()

	val := offset.ToJson()
	lease, _ := s.client.Grant(ctx, int64(offsetKeepTimeout.Seconds()))
	_, err := s.kv.Put(ctx, key, val, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to read and parse offset data: %w", err)
	}
	return nil
}
