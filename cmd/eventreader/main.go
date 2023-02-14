package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/vpnhouse/tunnel/pkg/eventlog"
	"github.com/vpnhouse/tunnel/pkg/xap"
)

const defaultServerPort = "8089"

type caList struct {
	CA []string `json:"ca"`
}

func main() {
	zap.ReplaceGlobals(xap.Development())
	authSecret := os.Getenv("AUTH_SECRET")
	if authSecret == "" {
		zap.L().Fatal("auth secret is not provided", zap.String("env", "AUTH_SECRET"))
		return
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		zap.L().Fatal("server host is not provided", zap.String("env", "SERVER_HOST"))
		return
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		zap.L().Info("server port is not provided, use default one", zap.String("env", "SERVER_PORT"), zap.String("port", defaultServerPort))
		serverPort = defaultServerPort
	}

	run_client(serverHost, serverPort, authSecret)
	time.Sleep(time.Second)
}

func run_client(serverHost string, serverPort string, authSecret string) {
	client, err := eventlog.NewClient(
		eventlog.WithSelfSignedTLS(),
		eventlog.WithNoSSL(),
		eventlog.WithHost(serverHost, serverPort),
		eventlog.WithAuthSecret(authSecret),
	)
	if err != nil {
		zap.L().Fatal("failed to create eventlog client", zap.Error(err))
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer client.Close()

	for {
		select {
		case evt, ok := <-client.Events():
			if !ok {
				zap.L().Info("no events, exiting...")
				return
			}
			if evt.Err != nil {
				zap.L().Error("read event error", zap.Error(evt.Err))
				return
			}
			zap.L().Info("event", zap.Any("event", *evt))
		case <-ch:
			zap.L().Info("interrupted")
			return
		}
	}
}
