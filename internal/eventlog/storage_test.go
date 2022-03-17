// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotateBySize(t *testing.T) {
	storage, err := newFsStorage(StorageConfig{Dir: "/", Size: 100, MaxFiles: 3}, afero.NewMemMapFs())
	require.NoError(t, err)

	first := storage.currentLog

	large := make([]byte, 1000)
	// rotation must happen after this single write
	err = storage.Write(large)
	require.NoError(t, err)

	assert.Len(t, storage.rotated, 3, "must have 3 slots allocated (from config)")
	assert.Equal(t, int64(0), storage.currentWritten, "new log must be empty")
	assert.Equal(t, storage.rotated[2].uuid, first.uuid)
	assert.Equal(t, storage.rotated[2].timestamp, first.timestamp)
	assert.Equal(t, 1, first.seq)
	assert.Equal(t, 2, storage.currentLog.seq)
}

func TestRestoreFromDir(t *testing.T) {
	f := afero.NewMemMapFs()
	// create 3 files according to template
	var lastUUID string
	for i := 1; i < 4; i++ {
		id := uuid.New().String()
		_, err := f.Create(fmt.Sprintf("/%d_%d_%s", i, time.Now().Unix(), id))
		require.NoError(t, err)
		lastUUID = id
	}

	stor, err := newFsStorage(StorageConfig{Dir: "/", MaxFiles: 5}, f)
	require.NoError(t, err)
	assert.Len(t, stor.rotated, 5)
	assert.Equal(t, stor.currentLog.uuid, lastUUID)
	assert.Equal(t, stor.currentLog.seq, 3)
}

func TestRestoreCorruptedDir(t *testing.T) {
	f := afero.NewMemMapFs()
	cfg := StorageConfig{Dir: "/", MaxFiles: 5}

	f.Create("/1_2_notuuid")
	_, err := newFsStorage(cfg, f)
	require.Error(t, err)

	f.Create("/a_2_" + uuid.New().String())
	_, err = newFsStorage(cfg, f)
	require.Error(t, err)

	f.Create("/1_" + uuid.New().String())
	_, err = newFsStorage(cfg, f)
	require.Error(t, err)

	f.Create("/" + uuid.New().String())
	_, err = newFsStorage(cfg, f)
	require.Error(t, err)
}

func TestValidateLogSeq(t *testing.T) {
	tests := []struct {
		ns       []int
		mustFail bool
	}{
		{[]int{1, 2, 3}, false},
		{[]int{5, 6, 7}, false},
		{[]int{5, 7, 6}, true},
		{[]int{5, 6, 8}, true},
		{[]int{1, 1, 1}, true},
		{[]int{3, 2, 1}, true},
	}

	for _, tt := range tests {
		fseq := make([]namedFile, len(tt.ns))
		for i := range tt.ns {
			fseq[i].seq = tt.ns[i]
		}

		err := validateFilesSequence(fseq)
		if tt.mustFail {
			require.Error(t, err, "sequence %v must fail", tt.ns)
		} else {
			require.NoError(t, err, "sequence %v must survive", tt.ns)
		}
	}
}

func TestSortNamedFiles(t *testing.T) {
	f := afero.NewMemMapFs()
	now := time.Now().Unix()

	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/"+dirFileNameTemplate, i+1, now, uuid.New().String())
		f.Create(path)
	}

	_, err := newFsStorage(StorageConfig{Dir: "/", MaxFiles: 1001}, f)
	require.NoError(t, err)
}

func TestMustRotate_size(t *testing.T) {
	tests := []struct {
		maxSize  int64
		currentW int64
		must     bool
	}{
		{0, 1000, false},
		{0, math.MaxInt64, false},
		{1000, 500, false},
		{1000, 1000, true},
	}

	for _, tt := range tests {
		f := &fsStorage{
			config:         StorageConfig{Size: tt.maxSize},
			currentWritten: tt.currentW,
		}

		assert.Equal(t, tt.must, f.mustRotate())
	}
}

func TestMustRotate_time(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		btime time.Time
		ttl   time.Duration
		must  bool
	}{
		{now, 0, false},
		{now, 1 * time.Hour, false},
		{now.Add(-30 * time.Minute), 1 * time.Hour, false},
		{now.Add(-61 * time.Minute), 1 * time.Hour, true},
	}

	for _, tt := range tests {
		f := &fsStorage{
			config:       StorageConfig{Period: tt.ttl},
			currentBtime: tt.btime,
		}

		assert.Equal(t, tt.must, f.mustRotate())
	}
}
