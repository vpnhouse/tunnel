package eventlog

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

const (
	offsetKeepTimeout = 7 * 24 * time.Hour
)

type Offset struct {
	TunnelID string `json:"tunnel_id"`
	LogID    string `json:"log_id"`
	Offset   int64  `json:"offset"`
}

func (s *Offset) ToJson() string {
	// Don't expect any error on this
	data, _ := json.Marshal(s)
	return string(data)
}

type OffsetSync interface {
	Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error)
	Release(instanceID string, tunnelID string) error

	GetOffset(tunnelID string) (Offset, error)
	PutOffset(offset Offset) error
}

func offsetFromJson(data string) (Offset, error) {
	return offsetFromJsonBytes([]byte(data))
}

func offsetFromJsonBytes(data []byte) (Offset, error) {
	var offset Offset
	err := json.Unmarshal([]byte(data), &offset)
	return offset, err
}

func buildSyncKey(value string) string {
	return fmt.Sprintf("eventlogs.sync.%s", base64.RawStdEncoding.EncodeToString([]byte(value)))
}

func buildOffsetKey(value string) string {
	return fmt.Sprintf("eventlogs.offset.%s", base64.RawStdEncoding.EncodeToString([]byte(value)))
}
