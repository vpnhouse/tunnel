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
	tunnelHost := net.JoinHostPort(s.tunnelHost, s.opts.TunnelPort)

	conn, err := net.Dial("tcp", tunnelHost)
	if err != nil {
		return fmt.Errorf("failed to dial %s: %s", tunnelHost, err)
	}
	defer conn.Close()

	config := &tls.Config{ServerName: s.tunnelHost, InsecureSkipVerify: false}
	tlsConn := tls.Client(conn, config)
	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("failed TLS handshake: %s", err)
	}
	defer tlsConn.Close()

	tlsCert := tls.Certificate{}
	for _, cert := range tlsConn.ConnectionState().PeerCertificates {
		tlsCert.Certificate = append(tlsCert.Certificate, cert.Raw)
	}

	tlsConfig := &tls.Config{
		ServerName:   s.tunnelHost,
		Certificates: []tls.Certificate{tlsCert},
	}

	creds := credentials.NewTLS(tlsConfig)

	zap.L().Info("handshake tls succeed", zap.String("tunnel", tunnelHost), zap.Int("certificates", len(tlsConn.ConnectionState().PeerCertificates)))

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

	offset, err := s.eventlogSync.GetPosition(s.tunnelHost)
	if err == nil {
		req.Position = &proto.EventLogPosition{
			LogId:  offset.LogID,
			Offset: offset.Offset,
		}
		req.SkipEventAtPosition = offset.LogID != ""
	} else if errors.Is(err, ErrPositionNotFound) {
		zap.L().Info("failed to get offset position. Start reading from the beginning of the active log")
	} else {
		return nil, err
	}

	fetchEventsClient, err := s.client.FetchEvents(ctx, req, grpc.Header(&header))
	if err != nil {
		return nil, err
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
