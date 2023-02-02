package eventlog

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
)

type options struct {
	// self signed client's tls setup
	// client negotiate against tunnel API to gather CA (self signed)
	selfSigned bool
	// CA (pem) in case mutual tls setup only
	ca []byte
	// Certificate related to the client in case mutual tls setup only
	cert tls.Certificate
	// tunnel domain name or ip
	tunnelHost string
	// tunnel port (GRPC)
	tunnelPort string
	// authSecret to sign requests to the server both grpc and http
	authSecret string
	// tunnel key to extra check
	// extra check is ommited if empty
	tunnelKey string
	// indicate to use http request to tunnel api (used for debug purposes)
	noSSL bool
}

type Option func(opts *options) error

func WithSelfSignedTLS() Option {
	return func(opts *options) error {
		opts.selfSigned = true
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

		opts.ca = ca
		opts.cert = cert
		return nil
	}
}

func WithTLS(ca []byte, cert tls.Certificate) Option {
	return func(opts *options) error {
		opts.ca = ca
		opts.cert = cert
		return nil
	}
}

func WithTunnel(host string, port string) Option {
	return func(opts *options) error {
		opts.tunnelHost = host
		opts.tunnelPort = port
		return nil
	}
}

func WithNoSSL() Option {
	return func(opts *options) error {
		opts.noSSL = true
		return nil
	}
}

func WithHost(host string, port string) Option {
	return func(opts *options) error {
		opts.tunnelHost = host
		opts.tunnelPort = port
		return nil
	}
}

func WithAuthSecret(authSecret string) Option {
	return func(opts *options) error {
		opts.authSecret = authSecret
		return nil
	}
}
