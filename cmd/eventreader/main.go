package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/vpnhouse/tunnel/pkg/tlsutils"
	"github.com/vpnhouse/tunnel/pkg/xap"
	"github.com/vpnhouse/tunnel/proto"

	"google.golang.org/grpc"
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

	// Get ca certificate
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/grpc/ca", serverHost), nil)
	if err != nil {
		zap.L().Fatal("get ca certificate failed", zap.Error(err))
		return
	}
	req.Header.Add("X-VPNHOUSE-FEDERATION-KEY", authSecret)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		zap.L().Fatal("get ca certificate failed", zap.Error(err))
		return
	}

	if resp.StatusCode != 200 {
		zap.L().Fatal("get ca certificate failed", zap.String("response_status", resp.Status))
		return
	}
	defer resp.Body.Close()

	var cas caList
	err = json.NewDecoder(resp.Body).Decode(&cas)
	if err != nil {
		zap.L().Fatal("get ca certificate failed", zap.Error(err))
		return
	}

	if len(cas.CA) == 0 {
		zap.L().Fatal("get ca certificate failed, ca cert list is empty")
		return
	}

	// TODO: Add resp header check (compare to expected one)
	tunnelKey := resp.Header.Get("X-VPNHOUSE-TUNNEL-KEY")
	if tunnelKey == "" {
		zap.L().Fatal("get ca certificate failed, unauthorised response")
		return
	}

	clientSign := tlsutils.Sign{CertPem: []byte(cas.CA[0])}
	creds, err := clientSign.GrpcClientCredentials()

	if err != nil {
		zap.L().Fatal("failed to create grps clent credentials")
		return
	}

	cc, err := grpc.Dial(
		net.JoinHostPort(serverHost, serverPort),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		zap.L().Fatal("failed to init grps client ", zap.Error(err))
		return
	}

	client := proto.NewEventLogServiceClient(cc)

	ctx, cancel := context.WithCancel(context.Background())
	fetchEventsClient, err := client.FetchEvents(ctx, &proto.FetchEventsRequest{})

	if err != nil {
		zap.L().Fatal("failed to setup grps fetch events client ", zap.Error(err))
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		zap.L().Info("start listening events")
		for {
			select {
			case <-ctx.Done():
				zap.L().Info("listen events interrupted")
				return
			default:
				event, err := fetchEventsClient.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						err := fetchEventsClient.CloseSend()
						if err != nil {
							zap.L().Error("failed to send close event to server", zap.Error(err))
						}
						return
					}
					zap.L().Error("failed to recieve events from server", zap.Error(err))
					continue
				}
				zap.L().Info("received", zap.Stringer("event", event))
			}
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
	cancel()
	<-done
	zap.L().Info("exited")
}
