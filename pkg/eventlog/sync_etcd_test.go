package eventlog

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestOffsetEtcdLock(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
	client, err := clientv3.New(clientv3.Config{
		DialTimeout: time.Second,
		Endpoints:   []string{"127.0.0.1:2379"},
	})

	require.NoError(t, err, "failed to create etcd client")

	offset, err := NewEventlogSyncEtcd(client)
	require.NoError(t, err, "failed to create etcd offset")

	acquired, err := offset.Acquire("instance_1", "tunnel_1", 2*time.Second)
	require.NoError(t, err, "acquire lock failed by error")
	require.True(t, acquired, "acquire lock is failed")

	acquired, err = offset.Acquire("instance_1", "tunnel_1", 2*time.Second)
	require.NoError(t, err, "2nd acquire lock failed by error")
	require.True(t, acquired, "2nd acquire lock is failed")

	acquired, err = offset.Acquire("instance_2", "tunnel_1", 2*time.Second)
	require.NoError(t, err, "acquire lock for instance_2 failed by error")
	require.False(t, acquired, "acquire lock for instance_2 must fail")

	time.Sleep(3 * time.Second)
	acquired, err = offset.Acquire("instance_2", "tunnel_1", 2*time.Second)
	require.NoError(t, err, "acquire lock for instance_2 failed by error")
	require.True(t, acquired, "acquire lock for instance_2 is failed")

	err = offset.Release("instance_2", "tunnel_1")
	require.NoError(t, err, "release lock for instance_2 failed by error")

	acquired, err = offset.Acquire("instance_1", "tunnel_1", 2*time.Second)
	require.NoError(t, err, "acquire lock failed by error")
	require.True(t, acquired, "acquire lock is failed")

}
