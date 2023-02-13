// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/vpnhouse/tunnel/proto"
)

const (
	MagicHI uint8 = 0xAA
	MagicLO uint8 = 0x55

	magicSize = 2
	sizeSize  = 2
	typeSize  = 4
	tsSize    = 8

	headerSize = magicSize + sizeSize + typeSize + tsSize
	maxBodyLen = 1<<16 - headerSize
)

type EventType int32

const (
	Unspecified      EventType = EventType(proto.EventType_Unspecified)
	PeerAdd          EventType = EventType(proto.EventType_PeerAdd)
	PeerRemove       EventType = EventType(proto.EventType_PeerRemove)
	PeerUpdate       EventType = EventType(proto.EventType_PeerUpdate)
	PeerTraffic      EventType = EventType(proto.EventType_PeerTraffic)
	PeerFirstConnect EventType = EventType(proto.EventType_PeerFirstConnect)
)

type Event struct {
	Type      EventType `json:"type"`
	Timestamp int64     `json:"ts"`
	LogID     string    `json:"log_id"`
	Offset    int64     `json:"offset"`
	Data      []byte    `json:"data"`
}

func (e Event) IntoProto() *proto.FetchEventsResponse {
	return &proto.FetchEventsResponse{
		EventType: proto.EventType(e.Type),
		Timestamp: proto.TimestampFromTime(time.Unix(e.Timestamp, 0)),
		Position:  &proto.EventLogPosition{LogId: e.LogID, Offset: e.Offset},
		Data:      e.Data,
	}
}

type eventHeader struct {
	bodySize  uint16
	eventType int32
	timestamp int64
}

// marshalEvent marshals events with a header
func marshalEvent(eventType uint32, timestamp int64, event interface{}) ([]byte, error) {
	// to prevent too much parsing and leave event body eve-readable in the file, store it as:
	// magic(2) + size(2) + event_type(4) + timestamp(8) + body(any)
	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %v", err)
	}

	// add \n here to get the coreutils-friendly line-by-line format.
	// we will strip it on unmarshaling.
	body = append(body, '\n')
	bodyLen := uint16(len(body))

	if bodyLen > maxBodyLen {
		return nil, fmt.Errorf("event is too large")
	}

	size := into2bytes(bodyLen)
	typeBytes := into4bytes(eventType)
	tsBytes := into8bytes(timestamp)

	header := make([]byte, 0, headerSize)
	header = append(header, MagicHI, MagicLO)
	header = append(header, size...)
	header = append(header, typeBytes...)
	header = append(header, tsBytes...)

	return append(header, body...), nil
}

// readAndParseEvent reads event from given r and fills the Event structure.
// atOffset and atLogID does not affect reading in any ways and using only
// to provide these fields to the resulting Event.
// Also returns nextOffset for a caller.
func readEvent(r io.Reader, atOffset int64, atLogID string) (Event, int64, error) {
	header, err := readEventHeader(r)
	if err != nil {
		return Event{}, 0, err
	}
	body, err := readEventBody(r, header)
	if err != nil {
		return Event{}, 0, err
	}

	event := Event{
		Type:      EventType(header.eventType),
		Timestamp: header.timestamp,
		Data:      body,
		LogID:     atLogID,
		Offset:    atOffset,
	}
	nextOffset := atOffset + headerSize + int64(header.bodySize)
	return event, nextOffset, nil
}

// readEventHeader reads exactly headerSize bytes and parse them into eventHeader.
func readEventHeader(r io.Reader) (eventHeader, error) {
	bs := make([]byte, headerSize)
	n, err := r.Read(bs)
	if err != nil {
		return eventHeader{}, err
	}

	if n != headerSize {
		return eventHeader{}, fmt.Errorf("header: too short read (got %d, expect %d)", n, headerSize)
	}

	if bs[0] != MagicHI || bs[1] != MagicLO {
		return eventHeader{}, fmt.Errorf("header: invalid magic number")
	}

	header := eventHeader{
		bodySize:  from2bytes(bs[2:4]),
		eventType: int32(from4bytes(bs[4:8])),
		timestamp: from8bytes(bs[8:16]),
	}
	return header, nil
}

// readEventBody reads event body described by the given header.
// The reader must supply exactly header.bodySize many bytes,
// and also be advanced to the start of the body (e.g used right after
// the readEventHeader method without any reads from r).
func readEventBody(r io.Reader, header eventHeader) ([]byte, error) {
	body := make([]byte, header.bodySize)
	n, err := r.Read(body)
	if err != nil {
		return nil, err
	}
	if n != int(header.bodySize) {
		return nil, fmt.Errorf("body: too short read (got %d, expect %d)", n, header.bodySize)
	}
	// strip \n we added in marshalEvent
	return body[:len(body)-1], nil
}

func into8bytes(i int64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(i))
	return bs
}

func into4bytes(i uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, i)
	return bs
}

func into2bytes(i uint16) []byte {
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, i)
	return bs
}

func from8bytes(bs []byte) int64 {
	return int64(binary.LittleEndian.Uint64(bs))
}

func from4bytes(bs []byte) uint32 {
	return binary.LittleEndian.Uint32(bs)
}

func from2bytes(bs []byte) uint16 {
	return binary.LittleEndian.Uint16(bs)
}
