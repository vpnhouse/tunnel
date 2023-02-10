package eventlog

import (
	"encoding/json"
	"time"
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

func OffsetFromJson(data string) (Offset, error) {
	var offset Offset
	err := json.Unmarshal([]byte(data), &offset)
	return offset, err
}

type OffsetSync interface {
	Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error)
	Release(instanceID string, tunnelID string) error

	GetOffset(tunnelID string) (Offset, error)
	PutOffset(offset Offset) error
}
