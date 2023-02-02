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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func (s *Client) connect() error {
	if s.opts.selfSigned {
		return s.connectSelfSignedTLS()
	} else if s.client == nil {
		return s.connectTLS()
	}
	return nil
}

func (s *Client) connectSelfSignedTLS() error {
	// Get ca certificate
	urlPrefix := "https"
	if s.opts.noSSL {
		urlPrefix = "http"
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/grpc/ca", urlPrefix, s.opts.tunnelHost), nil)
	if err != nil {
		return fmt.Errorf("get tunnel CA failed: %w", err)
	}
	req.Header.Add(federationAuthHeader, s.opts.authSecret)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("request tunnel CA failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("request tunnel CA failed: %s", resp.Status)
	}

	tunnelKey := resp.Header.Get(tunnelAuthHeader)
	if s.opts.tunnelKey != "" && tunnelKey == "" {
		return errors.New("request tunnel CA failed: unauthorized response, tunnel key is empty")
	}

	if s.opts.tunnelKey != "" && s.opts.tunnelKey != tunnelKey {
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
	if !certPool.AppendCertsFromPEM(s.opts.ca) {
		return errors.New("failed to add CA's certificate")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{s.opts.cert},
		RootCAs:      certPool,
	}

	creds := credentials.NewTLS(config)

	return s.connectWithCreds(creds)
}

func (s *Client) connectWithCreds(creds credentials.TransportCredentials) error {
	cc, err := grpc.Dial(
		net.JoinHostPort(s.opts.tunnelHost, s.opts.tunnelPort),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		zap.L().Fatal("failed to init grps client", zap.Error(err))
		return fmt.Errorf("failed to init grps client: %w", err)
	}

	s.client = proto.NewEventLogServiceClient(cc)
	return nil
}

func (s *Client) fetchEventsClient() (proto.EventLogService_FetchEventsClient, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())

	md := metadata.New(map[string]string{federationAuthHeader: s.opts.authSecret})
	ctx = metadata.NewOutgoingContext(ctx, md)
	var header metadata.MD
	fetchEventsClient, err := s.client.FetchEvents(ctx, &proto.FetchEventsRequest{}, grpc.Header(&header))
	if err != nil {
		cancel()
		return nil, nil, err
	}

	tunnelKey := header.Get(tunnelAuthHeader)
	if s.opts.tunnelKey != "" && (len(tunnelKey) == 0 || tunnelKey[0] == "") {
		cancel()
		return nil, nil, errors.New("request tunnel CA failed: unauthorized response, tunnel key is empty")
	}

	if s.opts.tunnelKey != "" && s.opts.tunnelKey != tunnelKey[0] {
		cancel()
		return nil, nil, errors.New("request tunnel CA failed: invalid response tunnel key")
	}

	return fetchEventsClient, cancel, nil
}
