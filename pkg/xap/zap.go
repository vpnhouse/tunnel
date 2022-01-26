package xap

import (
	"fmt"

	"go.uber.org/zap"
)

// ZapType returns zap.String with a type name of v
func ZapType(v interface{}) zap.Field {
	return zap.String("type", fmt.Sprintf("%T", v))
}
