package eventlog

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Client) connect() error {
	if s.opts.SelfSigned {
		return s.connectSelfSignedTLS()
	} else if s.client == nil {
		return s.connectTLS()
	}
	return nil
}

func (s *Client) connectSelfSignedTLS() error {
	// Get ca certificate
	urlPrefix := "https"
	if s.opts.NoSSL {
		urlPrefix = "http"
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/grpc/ca", urlPrefix, s.tunnelHost), nil)
	if err != nil {
		return fmt.Errorf("get tunnel CA failed: %w", err)
	}
	req.Header.Add(federationAuthHeader, s.opts.AuthSecret)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("request tunnel CA failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("request tunnel CA failed: %s", resp.Status)
	}

	tunnelKey := resp.Header.Get(tunnelAuthHeader)
	if s.opts.TunnelKey != "" && tunnelKey == "" {
		return errors.New("request tunnel CA failed: unauthorized response, tunnel key is empty")
	}

	if s.opts.TunnelKey != "" && s.opts.TunnelKey != tunnelKey {
		return errors.New("request tunnel CA failed: invalid response tunnel key")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("request tunnel CA failed: error read response body: %w", err)
	}

	var cas = struct {
		CA []string `json:"ca"`
	}{}
	err = json.Unmarshal(body, &cas)
	if err != nil {
		return fmt.Errorf("parse request tunnel CA failed: \n%s\n: %w", string(body), err)
	}

	if len(cas.CA) == 0 {
		return errors.New("request tunnel CA failed: empty tunnel CA list")
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(cas.CA[0])) {
		return errors.New("failed to add server CA's certificate")
	}

	creds := credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
	})

	return s.connectWithCreds(creds)
}

func (s *Client) connectTLS() error {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(s.opts.CA) {
		return errors.New("failed to add CA's certificate")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{s.opts.Cert},
		RootCAs:      certPool,
	}

	creds := credentials.NewTLS(config)

	return s.connectWithCreds(creds)
}

func (s *Client) connectWithCreds(creds credentials.TransportCredentials) error {
	cc, err := grpc.Dial(
		net.JoinHostPort(s.tunnelHost, s.opts.TunnelPort),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		zap.L().Fatal("failed to init grps client", zap.Error(err))
		return fmt.Errorf("failed to init grps client: %w", err)
	}

	s.client = proto.NewEventLogServiceClient(cc)
	return nil
}

func (s *Client) fetchEventsClient(ctx context.Context) (proto.EventLogService_FetchEventsClient, error) {
	md := metadata.New(map[string]string{federationAuthHeader: s.opts.AuthSecret})
	ctx = metadata.NewOutgoingContext(ctx, md)
	var header metadata.MD

	req := &proto.FetchEventsRequest{}

	offset, err := s.offsetSync.GetOffset(s.tunnelHost)
	if err == nil {
		req.StartPosition = &proto.EventLogPosition{
			LogId:  offset.LogID,
			Offset: offset.Offset,
		}
	} else {
		zap.L().Error("failed to get offset position. Start reading from the beginning of the active log")
	}

	fetchEventsClient, err := s.client.FetchEvents(ctx, req, grpc.Header(&header))
	if err != nil {
		if req.StartPosition == nil {
			return nil, fmt.Errorf("failed to connect to the active log stream: %w", err)
		}
		if status, ok := status.FromError(err); !ok || status.Code() != codes.NotFound {
			zap.L().Info("event log not found, use active event log", zap.String("log_id", req.StartPosition.LogId))
			req.StartPosition = nil // nil position means read from the active log
			fetchEventsClient, err = s.client.FetchEvents(ctx, req, grpc.Header(&header))
			if err != nil {
				return nil, fmt.Errorf("failed to connect to the active log stream: %w", err)
			}
		}
	}

	tunnelKey := header.Get(tunnelAuthHeader)
	if s.opts.TunnelKey != "" && (len(tunnelKey) == 0 || tunnelKey[0] == "") {
		return nil, errors.New("connect tunnel fetch events failed: unauthorized response, tunnel key is empty")
	}

	if s.opts.TunnelKey != "" && s.opts.TunnelKey != tunnelKey[0] {
		return nil, errors.New("connect tunnel fetch events failed: invalid response tunnel key")
	}

	return fetchEventsClient, nil
}

func (s *Client) eventFetchedClient(ctx context.Context) (proto.EventLogService_EventFetchedClient, error) {
	md := metadata.New(map[string]string{federationAuthHeader: s.opts.AuthSecret})
	ctx = metadata.NewOutgoingContext(ctx, md)
	var header metadata.MD
	eventFetchedClient, err := s.client.EventFetched(ctx, grpc.Header(&header))
	if err != nil {
		return nil, err
	}

	tunnelKey := header.Get(tunnelAuthHeader)
	if s.opts.TunnelKey != "" && (len(tunnelKey) == 0 || tunnelKey[0] == "") {
		return nil, errors.New("connect tunnel event fetched notification stream failed: unauthorized response, tunnel key is empty")
	}

	if s.opts.TunnelKey != "" && s.opts.TunnelKey != tunnelKey[0] {
		return nil, errors.New("connect tunnel event fetched notification stream failed: invalid response tunnel key")
	}

	return eventFetchedClient, nil
}
