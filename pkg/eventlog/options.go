package eventlog

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"time"
)

type options struct {
	// self signed client's tls setup
	// client negotiate against tunnel API to gather CA (self signed)
	SelfSigned bool
	// CA (pem) in case mutual tls setup only
	CA []byte
	// Certificate related to the client in case mutual tls setup only
	Cert tls.Certificate
	// tunnel port (GRPC)
	TunnelPort string
	// authSecret to sign requests to the server both grpc and http
	AuthSecret string
	// tunnel key to extra check
	// extra check is ommited if empty
	TunnelKey string
	// indicate to use http request to tunnel api (used for debug purposes)
	NoSSL bool
	// Stop idle timeout
	// > 0 the client stops listen events and close connection
	// in case silence (idle) from the server during timeout
	StopIdleTimeout time.Duration
	// Optional tunnel id (used in tracking offset)
	TunnelID string
	// Enforce to start reading from active log disregarding to stored offset
	ForceStartFromActiveLog bool
}

type Option func(opts *options) error

func WithSelfSignedTLS() Option {
	return func(opts *options) error {
		opts.SelfSigned = true
		return nil
	}
}

func WithTLSByFiles(caPath string, certFile string, keyFile string) Option {
	return func(opts *options) error {
		ca, err := ioutil.ReadFile(caPath)
		if err != nil {
			return fmt.Errorf("failed to load ca file: %w", err)
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load certificate: %w", err)
		}

		opts.CA = ca
		opts.Cert = cert
		return nil
	}
}

func WithTLS(ca []byte, cert tls.Certificate) Option {
	return func(opts *options) error {
		opts.CA = ca
		opts.Cert = cert
		return nil
	}
}

func WithTunnelPort(port string) Option {
	return func(opts *options) error {
		opts.TunnelPort = port
		return nil
	}
}

func WithNoSSL() Option {
	return func(opts *options) error {
		opts.NoSSL = true
		return nil
	}
}

func WithAuthSecret(authSecret string) Option {
	return func(opts *options) error {
		opts.AuthSecret = authSecret
		return nil
	}
}

func WithStopIdleTimeout(stopIdleTimeout time.Duration) Option {
	return func(opts *options) error {
		opts.StopIdleTimeout = stopIdleTimeout
		return nil
	}
}

func WithStartFromActiveLog() Option {
	return func(opts *options) error {
		opts.ForceStartFromActiveLog = true
		return nil
	}
}
