package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/Codename-Uranium/tunnel/eventlog"
	"github.com/Codename-Uranium/tunnel/proto"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	// Addr to listen for gRPC connections
	Addr string `json:"addr"`
	// Cert is a server certificate path
	Cert string `json:"cert"`
	// Key is a server certificate key
	Key string `json:"key"`
	// ClientCA certificate path
	ClientCA string `json:"client_ca"`
}

// grpcServer wraps grpc.Server into the control.ServiceController interface
type grpcServer struct {
	mu      sync.Mutex
	running bool

	server *grpc.Server
}

func (g *grpcServer) Running() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.running
}

func (g *grpcServer) Shutdown() error {
	g.mu.Lock()
	g.running = false
	g.mu.Unlock()

	g.server.Stop()
	return nil
}

// New creates and starts gRPC services.
func New(config Config, eventLog eventlog.EventManager) (*grpcServer, error) {
	withTLS, err := tlsCredentialsFromConfig(config)
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer(withTLS,
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
		running: true,
		server:  srv,
	}

	go func() {
		zap.L().Debug("starting gRPC server", zap.String("addr", lis.Addr().String()))
		if err := srv.Serve(lis); err != nil {
			zap.L().Warn("gRPC listener stopped", zap.Error(err))
		}

		wrapper.mu.Lock()
		defer wrapper.mu.Unlock()

		wrapper.running = false
	}()

	return wrapper, nil
}

func tlsCredentialsFromConfig(config Config) (grpc.ServerOption, error) {
	// load server cert
	serverCert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to load server's cert, key pair: %v", err)
	}
	// load client ca cert
	caCertBytes, err := os.ReadFile(config.ClientCA)
	if err != nil {
		return nil, fmt.Errorf("failed to load client CA from %s: %v", config.ClientCA, err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCertBytes); !ok {
		return nil, fmt.Errorf("failed to add ClientCA from %s: check the certificate content", config.ClientCA)
	}

	tlsCreds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
	})

	return grpc.Creds(tlsCreds), nil
}
