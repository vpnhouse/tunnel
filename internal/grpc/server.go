// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"sync/atomic"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/federation_keys"
	"github.com/vpnhouse/tunnel/pkg/tlsutils"
	"github.com/vpnhouse/tunnel/pkg/xnet"
	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	// Addr to listen for gRPC connections
	Addr        string             `json:"addr"`
	Tls         *TlsConfig         `json:"tls,omitempty"`
	TlsSelfSign *TlsSelfSignConfig `json:"tls_self_sign,omitempty"`
	CA          string             `json:"ca,omitempty"`
}

type TlsConfig struct {
	// Cert is a server certificate path
	Cert string `json:"cert"`
	// Key is a server certificate key
	Key string `json:"key"`
	// ClientCA certificate path
	ClientCA string `json:"client_ca"`
}

type TlsSelfSignConfig struct {
	TunnelKey string `json:"tunnel_key"`
}

// grpcServer wraps grpc.Server into the control.ServiceController interface
type grpcServer struct {
	running atomic.Value

	server   *grpc.Server
	ca       string
	keystore federation_keys.Keystore
}

func (g *grpcServer) CA() string {
	return g.ca
}

func (g *grpcServer) Running() bool {
	return g.running.Load().(bool)
}

func (g *grpcServer) Shutdown() error {
	g.running.Store(false)

	g.server.Stop()
	return nil
}

// New creates and starts gRPC services.
func New(config Config, eventLog eventlog.EventManager, keystore federation_keys.Keystore) (*grpcServer, error) {
	var withTls grpc.ServerOption
	var ca string
	var err error
	switch {
	case config.Tls != nil:
		withTls, ca, err = tlsCredentialsAndCA(config.Tls)
	case config.TlsSelfSign != nil:
		withTls, ca, err = tlsSelfSignCredentialsAndCA()
	default:
		return nil, errors.New("tls config is not defined")
	}
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer(withTls,
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	eventSrv := newEventServer(eventLog)
	proto.RegisterEventLogServiceServer(srv, eventSrv)

	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, err
	}

	wrapper := &grpcServer{
		server:   srv,
		ca:       ca,
		keystore: keystore,
	}
	wrapper.running.Store(true)

	go func() {
		zap.L().Debug("starting gRPC server", zap.String("addr", lis.Addr().String()))
		if err := srv.Serve(lis); err != nil {
			zap.L().Warn("gRPC listener stopped", zap.Error(err))
		}
		wrapper.running.Store(false)
	}()

	return wrapper, nil
}

func tlsCredentialsAndCA(tlsConfig *TlsConfig) (grpc.ServerOption, string, error) {
	// load server cert
	serverCert, err := tls.LoadX509KeyPair(tlsConfig.Cert, tlsConfig.Key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load server's cert, key pair: %w", err)
	}
	// load ca cert
	clientCABytes, err := os.ReadFile(tlsConfig.ClientCA)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load CA cert from %s: %w", tlsConfig.ClientCA, err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(clientCABytes); !ok {
		return nil, "", fmt.Errorf("failed to add ClientCA from %s: check the certificate content", tlsConfig.ClientCA)
	}

	tlsCreds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
	})

	return grpc.Creds(tlsCreds), "", nil
}

func tlsSelfSignCredentialsAndCA() (grpc.ServerOption, string, error) {
	externalIp, err := xnet.GetExternalIPv4Addr()
	if err != nil {
		return nil, "", err
	}

	optsCA := []tlsutils.SignGenOption{
		tlsutils.WithCA(),
		tlsutils.WithRsaSigner(4096),
	}

	signCA, err := tlsutils.GenerateSign(optsCA...)
	if err != nil {
		return nil, "", err
	}

	opts := []tlsutils.SignGenOption{
		tlsutils.WithParentSign(signCA),
		tlsutils.WithRsaSigner(4096),
		tlsutils.WithIPAddresses(externalIp),
		tlsutils.WithLocalIPAddresses(),
	}

	sign, err := tlsutils.GenerateSign(opts...)
	if err != nil {
		return nil, "", err
	}

	creds, err := sign.GrpcServerCredentials()
	if err != nil {
		return nil, "", err
	}

	return grpc.Creds(creds), string(signCA.GetCertPem()), nil
}
