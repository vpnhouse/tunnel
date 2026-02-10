package runtime

import (
	"github.com/vpnhouse/common-lib-go/xhttp"
	"github.com/vpnhouse/tunnel/internal/settings"
	"go.uber.org/zap"
)

func MakeReverseHandlers(cfg []*settings.ReverseConfig) []*xhttp.HandleStruct {
	handlers := make([]*xhttp.HandleStruct, 0, len(cfg))
	for _, c := range cfg {
		handler, err := xhttp.ReverseProxyHandler(c.Patterns, c.URL)
		if err != nil {
			zap.L().Error("Skipping reverse proxy handler", zap.Any("proxy", c))
		}
		handlers = append(handlers, handler)
	}

	return handlers
}
