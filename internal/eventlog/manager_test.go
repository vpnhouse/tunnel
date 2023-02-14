// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestInstance(cfg StorageConfig) *eventManager {
	// make log manager, each instance have their own FS.
	m, err := New(cfg, afero.NewMemMapFs())
	if err != nil {
		panic(err)
	}
	return m
}

func TestSinglePush(t *testing.T) {
	log := newTestInstance(StorageConfig{Dir: "/", Size: 1000})

	data := "foo bar 123"
	ts := time.Now().Unix()
	err := log.Push(PeerAdd, data)
	require.NoError(t, err)

	err = log.Shutdown()
	require.NoError(t, err)

	// directly inspect the filesystem content
	logid := log.storage.currentLog.String()
	lastLogName := "/" + logid
	last, err := log.storage._fs.Open(lastLogName)
	require.NoError(t, err)

	event, _, err := readEvent(last, 0, lastLogName)
	require.NoError(t, err)
	assert.Equal(t, PeerAdd, event.Type)
	assert.Equal(t, ts, event.Timestamp)

	var s string
	_ = json.Unmarshal(event.Data, &s)
	assert.Equal(t, data, s)
}

func TestReadBack(t *testing.T) {
	l := newTestInstance(StorageConfig{Dir: "/", Size: 100})

	err := l.Push(PeerAdd, "hello world")
	require.NoError(t, err)

	sub, err := l.Subscribe(context.Background(), "", EventlogPosition{
		Offset: 0,
		LogID:  "",
	})
	require.NoError(t, err)

	event := <-sub.Events()
	l.Shutdown()

	assert.Equal(t, PeerAdd, event.Type)
	var s string
	json.Unmarshal(event.Data, &s)
	assert.Equal(t, "hello world", s)
}

func TestSubscribeToInvalidOffset(t *testing.T) {
	log := newTestInstance(StorageConfig{Dir: "/", Size: 100, MaxFiles: 5})

	for i := 0; i < 10; i++ {
		err := log.Push(PeerAdd, "hello world")
		require.NoError(t, err)
	}

	log.storage.lock.Lock()
	logID := log.storage.currentLog.uuid
	log.storage.lock.Unlock()

	// assume we know the logID somehow
	sub, err := log.Subscribe(context.Background(), "", EventlogPosition{LogID: logID, Offset: 5})
	require.NoError(t, err)

	// must immediately get an error
	err = <-sub.Errors()
	require.Error(t, err)
}

func TestSubscribeToUnknownLog(t *testing.T) {
	log := newTestInstance(StorageConfig{Dir: "/", Size: 10_000})

	_, err := log.Subscribe(context.Background(), "", EventlogPosition{LogID: uuid.New().String(), Offset: 0})
	require.Error(t, err)
}

func TestSubscribeToOffset(t *testing.T) {
	log := newTestInstance(StorageConfig{Dir: "/", Size: 10_000})

	err := log.Push(PeerAdd, "hello world")
	require.NoError(t, err)
	err = log.Push(PeerRemove, "hello world")
	require.NoError(t, err)
	err = log.Push(PeerTraffic, "hello world")
	require.NoError(t, err)
	err = log.Push(PeerFirstConnect, "hello world")
	require.NoError(t, err)
	err = log.Push(PeerUpdate, "hello world")
	require.NoError(t, err)

	time.Sleep(2 * time.Millisecond)

	bs, err := marshalEvent(PeerAdd, time.Now().UTC().Unix(), "hello world")
	require.NoError(t, err)

	offset := int64(len(bs))

	logID := log.storage.currentLog.uuid
	sub, err := log.Subscribe(context.Background(), "", EventlogPosition{LogID: logID, Offset: offset})
	require.NoError(t, err)
	ev1 := <-sub.Events()
	sub.Close()
	assert.Equal(t, PeerRemove, ev1.Type)

	sub, err = log.Subscribe(context.Background(), "", EventlogPosition{LogID: logID, Offset: 4 * offset})
	require.NoError(t, err)
	ev2 := <-sub.Events()
	sub.Close()
	assert.Equal(t, PeerUpdate, ev2.Type)
}

func TestMultipleReads(t *testing.T) {
	log := newTestInstance(StorageConfig{Dir: "/", Size: 10_000})
	logID := log.storage.currentLog.uuid

	ts := time.Now().Unix()
	_ = log.Push(PeerAdd, "data here")

	const reads = 10
	events := make([]Event, reads)

	wg := &sync.WaitGroup{}
	wg.Add(reads)

	for i := 0; i < reads; i++ {
		sub, err := log.Subscribe(context.Background(), "", EventlogPosition{LogID: logID, Offset: 0})
		if err != nil {
			panic(err)
		}
		go func(i int) {
			defer wg.Done()
			ev := <-sub.Events()
			events[i] = ev
			<-sub.Close()
		}(i)
	}

	wg.Wait()
	_ = log.Shutdown()

	for i := 0; i < reads; i++ {
		assert.Equal(t, PeerAdd, events[i].Type)
		assert.Equal(t, ts, events[i].Timestamp)
	}
}

func TestWritesOrder(t *testing.T) {
	// note: will trigger 5 rotations with a log size = 25k bytes
	log := newTestInstance(StorageConfig{Dir: "/", Size: 25_000, MaxFiles: 4})

	ctx := context.Background()

	sub, err := log.Subscribe(ctx, "", EventlogPosition{LogID: "", Offset: 0})
	require.NoError(t, err)

	const writes = 5000
	reads := 0
	errCnt := 0

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		prevTimestamp := int64(0)
		for event := range sub.Events() {
			if event.Timestamp < prevTimestamp {
				panic(fmt.Sprintf("order mismatch: event=%d reads=%d\n", event.Timestamp, reads))
			}
			prevTimestamp = event.Timestamp
			reads++
		}
	}()

	var errArr []error
	go func() {
		defer wg.Done()
		for err := range sub.Errors() {
			t.Log("Error:", err)
			errCnt++
			errArr = append(errArr, err)
		}
	}()

	for i := 0; i < writes; i++ {
		err = log.Push(PeerAdd, "data")
		require.NoError(t, err)
	}

	time.Sleep(time.Second)

	log.Shutdown()

	wg.Wait()
	assert.Equal(t, writes, reads)
	assert.Equal(t, 1, len(errArr))
	assert.True(t, errors.Is(errArr[0], context.Canceled))
}

func TestReadsCount(t *testing.T) {
	log := newTestInstance(StorageConfig{Dir: "/", Size: 50_000})

	const writes = 100
	const concurrency = 50
	const batch = writes / concurrency

	wg := &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			for i := 0; i < batch; i++ {
				err := log.Push(PeerAdd, "data")
				require.NoError(t, err)
			}
		}(i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rets := make([]int, concurrency)
	for i := 0; i < concurrency; i++ {
		sub, err := log.Subscribe(ctx, "", EventlogPosition{LogID: "", Offset: 0})
		require.NoError(t, err)

		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			for range sub.Events() {
				rets[x]++
			}
		}(i)
	}

	wg.Wait()
	for i, v := range rets {
		assert.Equal(t, writes, v, "mismatch at %d-th, %d vs %d", i, v, writes)
	}
}
