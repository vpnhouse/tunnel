package xerror

import (
	"net"
	"time"

	openapiTypes "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
)

func StorageToValue(v interface{}) interface{} {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return t
	case int64:
		return t
	case bool:
		return t
	case uuid.UUID:
		return t.String()
	case *string:
		if t == nil {
			return nil
		}
		return *t
	case *int:
		if t == nil {
			return nil
		}
		return *t
	case *int64:
		if t == nil {
			return nil
		}
		return *t
	case *openapiTypes.Date:
		if t == nil {
			return nil
		}
		return t.Time
	case *net.IP:
		if t == nil {
			return nil
		}
		return t.String()
	case *time.Time:
		if t == nil {
			return nil
		}
		return t.Unix()
	case *bool:
		if t == nil {
			return nil
		}
		return *t
	case *uuid.UUID:
		if t == nil {
			return nil
		}
		return t.String()
	default:
		return nil
	}
}
