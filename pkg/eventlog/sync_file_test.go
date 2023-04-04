package eventlog

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupLogger() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
}

func TestEventlogSyncFileAcquireLockTtl(t *testing.T) {
	setupLogger()

	dir := os.TempDir()
	eventSync, err := NewEventlogSyncFile(dir)
	require.NoError(t, err, "failed to create offset sync file")

	instanceID1 := "instance_1"
	tunnelID1 := "tunnel_1"

	instanceID2 := "instance_2"

	var wg sync.WaitGroup
	wg.Add(2)
	acquired1 := 0
	acquired2 := 0
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			acquired, err := eventSync.Acquire(instanceID1, tunnelID1, time.Second)
			require.NoError(t, err, "failed to acquire offset sync lock du to error: %s", instanceID1)
			if acquired {
				acquired1++
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			acquired, err := eventSync.Acquire(instanceID2, tunnelID1, time.Second)
			require.NoError(t, err, "failed to acquire offset sync lock due to error: %s", instanceID2)
			if acquired {
				acquired2++
			}
		}
	}()

	wg.Wait()

	require.Equal(t, 100, acquired1+acquired2, "incorrect number of acquired sync locks: %v(instance1) != %v(instance2)", acquired1, acquired2)
	t.Logf("number of acquired sync locks: %v(instance1) <-> %v(instance2)", acquired1, acquired2)

	t.Log("waiting for a while")
	time.Sleep(time.Second)
	t.Log("stop waiting")

	acquired, err := eventSync.Acquire(instanceID2, tunnelID1, time.Second)
	require.NoError(t, err, "failed to acquire offset sync lock due to error: %s", instanceID2)
	require.True(t, acquired, "failed to acquire offset sync lock: %s", instanceID2)

	err = eventSync.Release(instanceID1, tunnelID1)
	require.NoError(t, err, "failed to release offset sync lock due to error: %s", instanceID1)

	err = eventSync.Release(instanceID2, tunnelID1)
	require.NoError(t, err, "failed to release offset sync lock due to error: %s", instanceID1)
}
