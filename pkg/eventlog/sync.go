package eventlog

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	offsetKeepTimeout = 7 * 24 * time.Hour
)

var ErrPositionNotFound = errors.New("position not found")

type Position struct {
	LogID  string `json:"log_id"`
	Offset int64  `json:"offset"`
}

func (s *Position) ToJson() string {
	// Don't expect any error on this
	data, _ := json.Marshal(s)
	return string(data)
}

type EventlogSync interface {
	Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error)
	Release(instanceID string, tunnelID string) error

	GetPosition(tunnelID string) (Position, error)
	PutPosition(tunnelID string, position Position) error
	DeletePosition(tunnelID string) error
}

func positionFromJson(data string) (Position, error) {
	return positionFromJsonBytes([]byte(data))
}

func positionFromJsonBytes(data []byte) (Position, error) {
	var pos Position
	err := json.Unmarshal([]byte(data), &pos)
	return pos, err
}

func buildSyncKey(value string) string {
	return fmt.Sprintf("eventlog.sync.%s", base64.RawStdEncoding.EncodeToString([]byte(value)))
}

func buildPositionKey(value string) string {
	return fmt.Sprintf("eventlog.position.%s", base64.RawStdEncoding.EncodeToString([]byte(value)))
}
