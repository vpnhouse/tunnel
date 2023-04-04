// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalTooShort(t *testing.T) {
	bs, err := marshalEvent(42, 1101, "my data...")
	require.NoError(t, err)

	bs = bs[:headerSize-1]
	_, next, err := readEvent(bytes.NewBuffer(bs), 0, "")
	assert.Contains(t, err.Error(), "header: too short read")
	assert.Equal(t, int64(0), next)
}

func TestUnmarshalShortWithHeader(t *testing.T) {
	bs, err := marshalEvent(42, 1101, "my data...")
	require.NoError(t, err)

	bs = bs[:len(bs)-3]
	_, next, err := readEvent(bytes.NewBuffer(bs), 0, "")
	assert.Contains(t, err.Error(), "body: too short read")
	assert.Equal(t, int64(0), next)
}

func TestNoValidMagic(t *testing.T) {
	bs, err := marshalEvent(42, 1101, "my data...")
	require.NoError(t, err)

	bs[0] = 0xff
	_, next, err := readEvent(bytes.NewBuffer(bs), 0, "")
	assert.EqualError(t, err, "header: invalid magic number")
	assert.Equal(t, int64(0), next)
}

func TestBitOps(t *testing.T) {
	t16 := []uint16{
		0,
		1,
		0xdead,
		0xca00,
		0x00b1,
		math.MaxUint16,
	}

	for _, tt := range t16 {
		bs := into2bytes(tt)
		assert.Len(t, bs, 2)
		v := from2bytes(bs[:])
		assert.Equal(t, tt, v)
	}

	t32 := []uint32{
		0,
		1,
		0xdeadface,
		0xbeefd00b,
		math.MaxUint32,
	}

	for _, tt := range t32 {
		bs := into4bytes(tt)
		assert.Len(t, bs, 4)
		v := from4bytes(bs)
		assert.Equal(t, tt, v)
	}

	t64 := []int64{
		0,
		1,
		math.MaxInt32 + 1,
		math.MaxInt64,
	}
	for _, tt := range t64 {
		bs := into8bytes(tt)
		assert.Len(t, bs, 8)
		v := from8bytes(bs[:])
		assert.Equal(t, tt, v)
	}
}

func TestReadEvent(t *testing.T) {
	ts := time.Now().UTC().Unix()
	bs, err := marshalEvent(PeerAdd, ts, "my string")
	require.NoError(t, err)

	var offset int64 = 100
	event, next, err := readEvent(bytes.NewBuffer(bs), offset, "log_id")
	require.NoError(t, err)

	assert.Equal(t, offset+int64(len(bs)), next)
	assert.Equal(t, PeerAdd, event.Type)
	assert.Equal(t, ts, event.Timestamp)
	assert.Equal(t, "log_id", event.LogID)
	assert.Equal(t, offset, event.Offset)
	var body string
	_ = json.Unmarshal(event.Data, &body)
	assert.Equal(t, "my string", body)
}

func TestReadEvent_EOF(t *testing.T) {
	ts := time.Now().UTC().Unix()
	bs, err := marshalEvent(PeerAdd, ts, "my string")
	require.NoError(t, err)

	r := bytes.NewBuffer(bs)
	_, next, err := readEvent(r, 0, "log_id")
	require.NoError(t, err)
	assert.Equal(t, int64(len(bs)), next)

	_, _, err = readEvent(r, next, "log_id")
	assert.Equal(t, io.EOF, err)
}

func TestReadEventHeaderBody(t *testing.T) {
	body := "my string"
	expectedBodySize := len(body) + 2 + 1
	ts := time.Now().UTC().Unix()
	bs, err := marshalEvent(PeerAdd, ts, body)
	require.NoError(t, err)

	r := bytes.NewBuffer(bs)
	header, err := readEventHeader(r)
	require.NoError(t, err)
	assert.Equal(t, int32(PeerAdd), header.eventType)
	assert.Equal(t, ts, header.timestamp)
	assert.Equal(t, uint16(expectedBodySize), header.bodySize)

	b, err := readEventBody(r, header)
	require.NoError(t, err)
	assert.Equal(t, expectedBodySize-1, len(b))
	var s string
	_ = json.Unmarshal(b, &s)
	assert.Equal(t, body, s)
}

func BenchmarkMarshalEvent(b *testing.B) {
	b.ReportAllocs()

	ts := time.Now().Unix()
	data := map[string]interface{}{
		"id":   1234,
		"user": "root",
		"foo":  "bar",
	}
	for i := 0; i < b.N; i++ {
		_, _ = marshalEvent(PeerAdd, ts, data)
	}
}

func BenchmarkUnmarshalEvent(b *testing.B) {
	b.ReportAllocs()

	bs, _ := marshalEvent(42, time.Now().Unix(), map[string]interface{}{
		"id":   1234,
		"user": "root",
		"foo":  "bar",
	})
	for i := 0; i < b.N; i++ {
		r := bytes.NewBuffer(bs)
		_, _, _ = readEvent(r, 0, "log")
	}
}
